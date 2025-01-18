package repository

import (
	"database/sql"
	"errors"
	"fmt"

	errorsUtils "linkshrink/internal/utils/errors"
	"linkshrink/internal/utils/logger"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PingableRepository interface {
	URLRepository
	Ping() error
}

type Transaction interface {
	Commit() error
	Rollback() error
}

type TransactableRepository interface {
	URLRepository
	Begin() (*sql.Tx, error)
}

type PostgresRepository struct {
	db     *sql.DB
	logger logger.Logger
}

func NewPostgresRepository(dsn string, log logger.Logger) (*PostgresRepository, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Создание таблицы, если она не существует
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS urls (
		id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
		original_url VARCHAR(2048) NOT NULL
	);`)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	_, err = db.Exec(`CREATE UNIQUE INDEX idx_original_url ON urls (original_url);`)
	if err != nil {
		return nil, fmt.Errorf("failed to create index: %w", err)
	}

	return &PostgresRepository{db: db, logger: log}, nil
}

func (p *PostgresRepository) Save(originalURL string) (string, error) {
	var lastInsertID string
	err := p.db.QueryRow(`
	INSERT INTO urls (original_url)
	 VALUES ($1) RETURNING id
	  ON CONFLICT (original_url) DO NOTHING;`, originalURL).Scan(&lastInsertID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return lastInsertID, &errorsUtils.UniqueViolationError{Err: err}
			} else {
				return "", errors.New("Ошибка:" + pgErr.Message + ", Код:" + pgErr.Code)
			}
		} else {
			return "", errors.New("error inserting URL" + err.Error())
		}
	}

	return lastInsertID, nil
}

func (p *PostgresRepository) Find(id string) (string, error) {
	var originalURL string
	err := p.db.QueryRow("SELECT original_url FROM urls WHERE uuid = $1", id).Scan(&originalURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrURLNotFound
		}
		return "", errors.New("не удалось найти URL в бд файл: " + err.Error())
	}
	return originalURL, nil
}

func (p *PostgresRepository) Close() error {
	if err := p.db.Close(); err != nil {
		return errors.New("Ошибка при закрытии базы данных:" + err.Error())
	}
	return nil
}

func (p *PostgresRepository) Ping() error {
	if err := p.db.Ping(); err != nil {
		return errors.New("Ошибка при ping до базы данных:" + err.Error())
	}
	return nil
}

func (p *PostgresRepository) Begin() (*sql.Tx, error) {
	transaction, err := p.db.Begin()
	if err != nil {
		return nil, errors.New("Ошибка при начале транзакции в базе данных:" + err.Error())
	}
	return transaction, nil
}
