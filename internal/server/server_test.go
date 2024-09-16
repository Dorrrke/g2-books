package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Dorrrke/g2-books/internal/domain/models"
	"github.com/Dorrrke/g2-books/internal/logger"
	"github.com/Dorrrke/g2-books/internal/storage"
	mocks "github.com/Dorrrke/g2-books/moks"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestRegisterHandler(t *testing.T) {
	var srv Server
	r := gin.Default()
	r.POST("/register", srv.RegisterHandler)
	httpSrv := httptest.NewServer(r)

	type want struct {
		errFlag    bool
		mockFlag   bool
		statusCode int
	}
	type test struct {
		name    string
		method  string
		request string
		user    string
		uid     string
		err     error
		want    want
	}

	tests := []test{
		{
			name:    "Test RegisterHandler; Case 1:",
			method:  http.MethodPost,
			request: "/register",
			user:    `{"uid":"uid","name":"Sergei","email":"testemail@ya.ru","pass":"qwerty1234"}`,
			uid:     "testUid",
			err:     nil,
			want: want{
				statusCode: http.StatusOK,
				errFlag:    false,
				mockFlag:   true,
			},
		},
		{
			name:    "Test RegisterHandler; Case 2:",
			method:  http.MethodPost,
			request: "/register",
			user:    `{uid""uid","email":"testemail@ya.ru","pass":"qwerty1234"`,
			want: want{
				statusCode: http.StatusBadRequest,
				errFlag:    true,
				mockFlag:   false,
			},
		},
		{
			name:    "Test RegisterHandler; Case 3:",
			method:  http.MethodPost,
			request: "/register",
			user:    `{"uid":"uid","name":"Sergei","email":"testemail@ya.ru","pass":"qwerty1234"}`,
			uid:     "",
			err:     fmt.Errorf("test error"),
			want: want{
				statusCode: http.StatusInternalServerError,
				errFlag:    true,
				mockFlag:   true,
			},
		},
	}

	// log := logger.Get(true)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			m := mocks.NewMockStorage(ctrl)
			defer ctrl.Finish()
			if tc.want.mockFlag {
				m.EXPECT().SaveUser(gomock.Any()).Return(tc.uid, tc.err)
				srv.storage = m
			}
			req := resty.New().R()
			req.Method = tc.method
			req.URL = httpSrv.URL + tc.request
			req.Body = tc.user
			resp, err := req.Send()
			if !tc.want.errFlag {
				assert.NoError(t, err)
				assert.NotEmpty(t, resp.Header().Get("Authorization"))
			}
			assert.Equal(t, resp.StatusCode(), tc.want.statusCode)
		})
	}
}

func TestAuthHandler(t *testing.T) {
	var srv Server
	r := gin.Default()
	r.POST("/auth", srv.AuthHandler)
	httpSrv := httptest.NewServer(r)

	type want struct {
		errFlag    bool
		mockFlag   bool
		statusCode int
	}
	type test struct {
		name    string
		method  string
		request string
		user    string
		pass    string
		uid     string
		err     error
		want    want
	}

	tests := []test{
		{
			name:    "Test AuthHandler; Case 1:",
			method:  http.MethodPost,
			request: "/auth",
			user:    `{"uid":"uid","name":"Sergei","email":"testemail@ya.ru","pass":"qwerty1234"}`,
			pass:    "qwerty1234",
			uid:     "testUid",
			err:     nil,
			want: want{
				statusCode: http.StatusOK,
				errFlag:    false,
				mockFlag:   true,
			},
		},
		{
			name:    "Test AuthHandler; Case 2:",
			method:  http.MethodPost,
			request: "/auth",
			user:    `{uid""uid","email":"testemail@ya.ru","pass":"qwerty1234"`,
			pass:    "qwerty1234",
			want: want{
				statusCode: http.StatusBadRequest,
				errFlag:    true,
				mockFlag:   false,
			},
		},
		{
			name:    "Test AuthHandler; Case 3:",
			method:  http.MethodPost,
			request: "/auth",
			user:    `{"uid":"uid","name":"Sergei","email":"testemail@ya.ru","pass":"qwerty1234"}`,
			pass:    "qwerty1234",
			uid:     "",
			err:     fmt.Errorf("test error"),
			want: want{
				statusCode: http.StatusInternalServerError,
				errFlag:    true,
				mockFlag:   true,
			},
		},
		{
			name:    "Test AuthHandler; Case 4:",
			method:  http.MethodPost,
			request: "/auth",
			user:    `{"uid":"testuid","name":"Sergei","email":"testemail@ya.ru","pass":"1qaz!QAZ"}`,
			pass:    "qwerty1234",
			uid:     "testuid",
			err:     nil,
			want: want{
				statusCode: http.StatusUnauthorized,
				errFlag:    true,
				mockFlag:   true,
			},
		},
		{
			name:    "Test AuthHandler; Case 5:",
			method:  http.MethodPost,
			request: "/auth",
			user:    `{"uid":"testuid","name":"Sergei","email":"testemail@ya.ru","pass":"1qaz!QAZ"}`,
			pass:    "qwerty1234",
			uid:     "testuid",
			err:     storage.ErrInvalidAuthData,
			want: want{
				statusCode: http.StatusUnauthorized,
				errFlag:    true,
				mockFlag:   true,
			},
		},
	}

	log := logger.Get(true)
	for _, tc := range tests {
		passHash, err := bcrypt.GenerateFromPassword([]byte(tc.pass), bcrypt.DefaultCost)
		assert.NoError(t, err)
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			m := mocks.NewMockStorage(ctrl)
			defer ctrl.Finish()
			if tc.want.mockFlag {
				m.EXPECT().ValidateUser(gomock.Any()).Return(tc.uid, string(passHash), tc.err)
				srv.storage = m
			}
			req := resty.New().R()
			req.Method = tc.method
			req.URL = httpSrv.URL + tc.request
			req.Body = tc.user
			resp, err := req.Send()
			log.Debug().Err(err).Str("body", string(resp.Body())).Send()
			if !tc.want.errFlag {
				assert.NoError(t, err)
				assert.NotEmpty(t, resp.Header().Get("Authorization"))
			}
			assert.Equal(t, resp.StatusCode(), tc.want.statusCode)
		})
	}
}

func TestAllBookHandler(t *testing.T) {
	var srv Server
	r := gin.Default()
	r.GET("/all", srv.AllBookHandler)
	httpSrv := httptest.NewServer(r)

	type want struct {
		errFlag    bool
		statusCode int
		books      string
	}
	type test struct {
		name    string
		method  string
		request string
		books   []models.Book
		err     error
		want    want
	}

	tests := []test{
		{
			name:    "Test AllBookHandler; Case 1:",
			method:  http.MethodGet,
			request: "/all",
			err:     nil,
			books: []models.Book{
				{
					BID:    "test1",
					Lable:  "b_lable",
					Author: "b_author",
					Delete: false,
					UID:    "test",
				},
			},
			want: want{
				statusCode: http.StatusOK,
				books:      `[{"b_id":"test1","lable":"b_lable","author":"b_author","delete":false,"uid":"test"}]`,
				errFlag:    false,
			},
		},
		{
			name:    "Test AllBookHandler; Case 2:",
			method:  http.MethodGet,
			request: "/all",
			err:     storage.ErrBooksListEmpty,
			books:   nil,
			want: want{
				statusCode: http.StatusNoContent,
				books:      ``,
				errFlag:    true,
			},
		},
		{
			name:    "Test AllBookHandler; Case 3:",
			method:  http.MethodGet,
			request: "/all",
			err:     fmt.Errorf("test error"),
			books:   nil,
			want: want{
				statusCode: http.StatusInternalServerError,
				books:      `{"error":"test error"}`,
				errFlag:    true,
			},
		},
	}

	log := logger.Get(true)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			m := mocks.NewMockStorage(ctrl)
			defer ctrl.Finish()
			m.EXPECT().GetBooks().Return(tc.books, tc.err)
			srv.storage = m
			req := resty.New().R()
			req.Method = tc.method
			req.URL = httpSrv.URL + tc.request
			resp, err := req.Send()
			log.Debug().Err(err).Str("body", string(resp.Body())).Any("str", resp.String()).Send()
			if !tc.want.errFlag {
				assert.NoError(t, err)
			}
			assert.Equal(t, resp.StatusCode(), tc.want.statusCode)
			assert.Equal(t, tc.want.books, string(resp.Body()))
		})
	}
}

func TestDeleter(t *testing.T) {
	type want struct {
		err error
	}
	type test struct {
		name string
		want want
	}
	tests := []test{
		{
			name: "Test deleter func; Case 1:",
			want: want{
				err: nil,
			},
		},
		{
			name: "Test deleter func; Case 2:",
			want: want{
				err: fmt.Errorf("test err"),
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
			defer cancel()
			ctrl := gomock.NewController(t)
			m := mocks.NewMockStorage(ctrl)
			defer ctrl.Finish()
			m.EXPECT().DeleteBooks().Return(tc.want.err)
			srv := New("0.0.0.0:8080", m)
			for i := 0; i < 5; i++ {
				srv.deleteChan <- i
			}
			go srv.deleter(ctx)
			for {
				select {
				case err := <-srv.ErrChan:
					assert.Equal(t, tc.want.err, err)
					return
				case <-time.After(time.Second):
					if tc.want.err != nil {
						t.Fatalf("Exp err = %s; actual = nil", tc.want.err)
						return
					}
					return
				}
			}
		})
	}
}
