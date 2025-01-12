package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"linkshrink/internal/utils/logger"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type IPingableRepository interface {
	IURLRepository
	Ping() error
}

type ITransaction interface {
	Commit() error
	Rollback() error
}

type ITransactableRepository interface {
	IURLRepository
	Begin() (*sql.Tx, error)
	Commit() error
	Rollback() error
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
		id SERIAL PRIMARY KEY,
		uuid TEXT NOT NULL UNIQUE,
		original_url TEXT NOT NULL
	);`)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &PostgresRepository{db: db, logger: log}, nil
}

func (p *PostgresRepository) Save(id string, originalURL string) error {
	_, err := p.db.Exec("INSERT INTO urls (uuid, original_url) VALUES ($1, $2)", id, originalURL)
	if err != nil {
		return errors.New("не удалось сохранить в бд файл: " + err.Error())
	}
	return nil
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
