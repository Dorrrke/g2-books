package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/Dorrrke/g2-books/internal/domain/models"
	authservicev1 "github.com/Dorrrke/g2-books/internal/go"
	"github.com/Dorrrke/g2-books/internal/logger"
	"github.com/Dorrrke/g2-books/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const SecretKey = "VerySecretKey2000"

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

type Storage interface {
	SaveUser(models.User) (string, error)
	ValidateUser(models.User) (string, string, error)
	GetBooks() ([]models.Book, error)
	GetBookById(string) (models.Book, error)
	GetBookByUID(string) ([]models.Book, error)
	SaveBook(models.Book) error
	DeleteBook(string) error
	DeleteBooks() error
}

type Server struct {
	serve      *http.Server
	storage    Storage
	deleteChan chan int
	authClient authservicev1.AuthServiceClient
	ErrChan    chan error
}

func New(host string, storage Storage, authClien authservicev1.AuthServiceClient) *Server {
	serve := http.Server{
		Addr: host,
	}
	dChan := make(chan int, 5)
	errChan := make(chan error)
	return &Server{
		serve:      &serve,
		storage:    storage,
		deleteChan: dChan,
		ErrChan:    errChan,
		authClient: authClien,
	}
}

func (s *Server) Run(ctx context.Context) error {
	go s.deleter(ctx)
	r := gin.Default()
	userGroup := r.Group("/user")
	{
		userGroup.POST("/register", s.RegisterHandler)
		userGroup.POST("/auth", s.AuthHandler)
	}
	bookGroup := r.Group("/books")
	{
		bookGroup.GET("/my-books", s.BooksByUser)
		bookGroup.GET("/all-books", s.AllBookHandler)
		bookGroup.GET("/:id", s.GetBookByIdHandler)
		bookGroup.POST("/add-book", s.SaveBookHandler)
		bookGroup.DELETE("/delete/:id", s.DeleteBookHandler)
	}
	s.serve.Handler = r
	if err := s.serve.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

func (s *Server) RegisterHandler(ctx *gin.Context) {
	log := logger.Get()
	var user models.User
	if err := ctx.ShouldBindBodyWithJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req, err := s.authClient.Register(context.TODO(), &authservicev1.User{
		Name:  user.Name,
		Email: user.Email,
		Pass:  user.Pass,
	})
	if err != nil {
		if status.Code(err) == codes.Aborted {
			log.Error().Err(err).Msg("register request failed; user alredy exist")
			ctx.JSON(http.StatusConflict, gin.H{"error": "user with this e-mail address is already registered"})
			return
		}
		log.Error().Err(err).Msg("user register request failed")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Debug().Str("token", req.Token).Str("msg", req.Message).Msg("grpc register request")
	ctx.Header("Authorization", req.Token)
	ctx.String(http.StatusOK, req.Message)
}

func (s *Server) AuthHandler(ctx *gin.Context) {
	log := logger.Get()
	var user models.User
	if err := ctx.ShouldBindBodyWithJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req, err := s.authClient.Login(context.TODO(), &authservicev1.UserCreds{
		Email: user.Email,
		Pass:  user.Pass,
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			log.Error().Err(err).Msg("request failed; user not found")
			ctx.String(http.StatusUnauthorized, "user not found")
		} else if status.Code(err) == codes.Unauthenticated {
			log.Error().Err(err).Msg("request failed; incorrect password")
			ctx.String(http.StatusUnauthorized, "incorrect password")
		}
		log.Error().Err(err).Msg("user login request failed")
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
	log.Debug().Str("token", req.Token).Str("msg", req.Message).Msg("grpc login request")
	ctx.Header("Authorization", req.Token)
	ctx.String(http.StatusOK, req.Message)
}

func (s *Server) AllBookHandler(ctx *gin.Context) {
	books, err := s.storage.GetBooks()
	if err != nil {
		if errors.Is(err, storage.ErrBooksListEmpty) {
			ctx.String(http.StatusNoContent, err.Error())
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, books)
}

func (s *Server) GetBookByIdHandler(ctx *gin.Context) {
	bid := ctx.Param("id")
	log.Println(bid)
	book, err := s.storage.GetBookById(bid)
	if err != nil {
		if errors.Is(err, storage.ErrBookNotFound) {
			ctx.String(http.StatusNoContent, err.Error())
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, book)
}

func (s *Server) BooksByUser(ctx *gin.Context) {
	log := logger.Get()
	token := ctx.GetHeader("Authorization")
	uid, err := getUID(token)
	if err != nil {
		log.Error().Err(err).Msg("get UID failed")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}
	books, err := s.storage.GetBookByUID(uid)
	if err != nil {
		if errors.Is(err, storage.ErrBooksListEmpty) {
			ctx.String(http.StatusNoContent, err.Error())
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, books)
}

func (s *Server) SaveBookHandler(ctx *gin.Context) {
	var book models.Book
	if err := ctx.ShouldBindBodyWithJSON(&book); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token := ctx.GetHeader("Authorization")
	uid, err := getUID(token)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}
	book.UID = uid
	if err := s.storage.SaveBook(book); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.String(http.StatusCreated, "book was saved")
}

func (s *Server) DeleteBookHandler(ctx *gin.Context) {
	bid := ctx.Param("id")
	if err := s.storage.DeleteBook(bid); err != nil {
		if errors.Is(err, storage.ErrBookNotFound) {
			ctx.String(http.StatusNoContent, err.Error())
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	s.deleteChan <- 1
	ctx.String(http.StatusOK, "book was deleted")
}

func (s *Server) ShutdownServer(ctx context.Context) error {
	log := logger.Get()
	defer log.Debug().Msg("server shutdowner - end")
	close(s.ErrChan)
	if err := s.serve.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("server shutdown failed")
		return err
	}
	return nil
}

func (s *Server) deleter(ctx context.Context) {
	log := logger.Get()
	defer log.Debug().Msg("deleter end")
	for {
		select {
		case <-ctx.Done():
			log.Debug().Msg("deleter: ctx done")
			return
		default:
			if len(s.deleteChan) == 5 {
				log.Debug().Int("delete count", len(s.deleteChan)).Msg("start deleting")
				for i := 0; i < 5; i++ {
					<-s.deleteChan
				}
				if err := s.storage.DeleteBooks(); err != nil {
					log.Error().Err(err).Msg("deleting books failed")
					s.ErrChan <- err
					return
				}
			}
		}
	}
}

func getUID(tokenStr string) (string, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})
	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", fmt.Errorf("invalid token")
	}
	return claims.UserID, nil
}
