package storage

import (
	"log"

	"github.com/google/uuid"

	"github.com/Dorrrke/g2-books/internal/domain/models"
)

type MemStorage struct {
	usersMap map[string]models.User
	booksMap map[string]models.Book
}

func New() *MemStorage {
	uMap := make(map[string]models.User)
	bMap := make(map[string]models.Book)
	return &MemStorage{
		usersMap: uMap,
		booksMap: bMap,
	}
}

func (ms *MemStorage) SaveUser(user models.User) error {
	uid := uuid.New().String()
	ms.usersMap[uid] = user
	return nil
}

func (ms *MemStorage) ValidateUser(user models.User) (string, error) {
	for uid, value := range ms.usersMap {
		if value.Email == user.Email {
			if value.Pass != user.Pass {
				return "", ErrInvalidAuthData
			}
			return uid, nil
		}
	}
	return "", ErrUserNotFound
}

func (ms *MemStorage) GetBooks() ([]models.Book, error) {
	books := []models.Book{}
	for bid, value := range ms.booksMap {
		book := value
		book.BID = bid
		books = append(books, book)
	}
	if len(books) == 0 {
		return nil, ErrBooksListEmpty
	}
	return books, nil
}

func (ms *MemStorage) GetBookByID(bID string) (models.Book, error) {
	log.Printf("BID: %s\n", bID)
	for _, val := range ms.booksMap {
		log.Println(val.Lable, val.BID)
	}
	book, ok := ms.booksMap[bID]
	if !ok {
		return models.Book{}, ErrBookNotFound
	}
	return book, nil
}

func (ms *MemStorage) SaveBook(book models.Book) error {
	bID := uuid.New().String()
	ms.booksMap[bID] = book
	return nil
}

func (ms *MemStorage) DeleteBook(bID string) error {
	_, ok := ms.booksMap[bID]
	if !ok {
		return ErrBookNotFound
	}
	delete(ms.booksMap, bID)
	return nil
}
