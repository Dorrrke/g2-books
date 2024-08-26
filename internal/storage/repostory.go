package storage

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Dorrrke/g2-books/internal/domain/models"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const ctxTimeout = 2 * time.Second

type Repository struct {
	conn *pgxpool.Pool
}

func NewRepo(ctx context.Context, dbAddr string) (*Repository, error) {
	conn, err := pgxpool.New(ctx, dbAddr)
	if err != nil {
		return nil, err
	}
	return &Repository{
		conn: conn,
	}, nil
}

func (r *Repository) SaveUser(user models.User) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer cancel()
	UID := uuid.New().String()
	_, err := r.conn.Exec(ctx, "INSERT INTO users(uid, name, email, pass) VALUES ($1, $2, $3, $4)",
		UID, user.Name, user.Email, user.Pass)
	if err != nil {
		return "", err
	}
	return UID, nil
}

func (r *Repository) ValidateUser(user models.User) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer cancel()
	row := r.conn.QueryRow(ctx, "SELECT uid, pass FROM users WHERE email = $1", user.Email)
	var UID string
	var pass string
	if err := row.Scan(&UID, &pass); err != nil {
		return "", "", err
	}
	return UID, pass, nil
}

func (r *Repository) GetBooks() ([]models.Book, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer cancel()
	rows, err := r.conn.Query(ctx, "SELECT * FROM books")
	if err != nil {
		return nil, err
	}
	var books []models.Book
	for rows.Next() {
		var book models.Book
		if err := rows.Scan(&book.BID, &book.Lable, &book.Author, &book.UID); err != nil {
			return nil, err
		}
		books = append(books, book)
	}
	if len(books) == 0 {
		return nil, fmt.Errorf("no books in db")
	}
	return books, nil
}

func (r *Repository) GetBookById(bID string) (models.Book, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer cancel()
	row := r.conn.QueryRow(ctx, "SELECT lable, author FROM books WHERE bid = $1", bID)
	var book models.Book
	if err := row.Scan(&book.Lable, &book.Author); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Book{}, fmt.Errorf("book with id = %s, does not exist", bID)
		}
		return models.Book{}, err
	}
	return book, nil
}

func (r *Repository) SaveBook(book models.Book) error {
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer cancel()
	_, err := r.conn.Exec(ctx, "INSERT INTO books(bid, lable, author, uid) VALUES ($1, $2, $3, $4)",
		uuid.New().String(), book.Lable, book.Author, book.UID)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) DeleteBook(bID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer cancel()
	tx, err := r.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Prepare(ctx, "delete book", "DELETE FROM books WHERE bid = $1"); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, "delete book", bID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func Migrations(dbAddr, migrationPath string) error {
	migratePath := fmt.Sprintf("file://%s", migrationPath)
	m, err := migrate.New(migratePath, dbAddr)
	if err != nil {
		return err
	}
	if err = m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("No migrations apply")
			return nil
		}
		return err
	}
	log.Println("Migrations complete")
	return nil
}
