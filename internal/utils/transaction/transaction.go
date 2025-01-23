package transaction

import (
	"context"
	"database/sql"
	"errors"
	"linkshrink/internal/utils"
)

type Transactor interface {
	BeginTx(ctx context.Context, txOptions *sql.TxOptions) (*sql.Tx, error)
}

// Handler - функция, которая выполняется в транзакции.
type Handler func(ctx context.Context) error

// TxManager менеджер транзакций, который выполняет указанный пользователем обработчик в транзакции.
type TxManager interface {
	ReadCommitted(ctx context.Context, f Handler) error
}

type manager struct {
	db Transactor
}

// NewTransactionManager создает новый менеджер транзакций, который удовлетворяет интерфейсу TxManager.
func NewTransactionManager(db Transactor) TxManager {
	return &manager{
		db: db,
	}
}

// transaction основная функция, которая выполняет указанный пользователем обработчик в транзакции.
func (m *manager) transaction(ctx context.Context, opts *sql.TxOptions, fn Handler) (err error) {
	// Если это вложенная транзакция, пропускаем инициацию новой транзакции и выполняем обработчик.
	tx, ok := ctx.Value(utils.TxKey).(*sql.Tx)
	if ok {
		return fn(ctx)
	}

	// Стартуем новую транзакцию.
	tx, err = m.db.BeginTx(ctx, opts)
	if err != nil {
		return errors.New("не удалось начать транзакцию: " + err.Error())
	}

	// Кладем транзакцию в контекст.
	ctx = context.WithValue(ctx, utils.TxKey, tx)

	// Настраиваем функцию отсрочки для отката или коммита транзакции.
	defer func() {
		// восстанавливаемся после паники
		if r := recover(); r != nil {
			err = errors.New("panic recovered")
		}

		// откатываем транзакцию, если произошла ошибка
		if err != nil {
			if errRollback := tx.Rollback(); errRollback != nil {
				err = errors.New("errRollback: " + errRollback.Error())
			}

			return
		}

		// если ошибок не было, коммитим транзакцию
		if err == nil {
			err = tx.Commit()
			if err != nil {
				err = errors.New("tx commit failed: " + err.Error())
			}
		}
	}()

	// Выполните код внутри транзакции.
	// Если функция терпит неудачу, возвращаем ошибку, и функция отсрочки выполняет откат
	// или в противном случае транзакция коммитится.
	if err = fn(ctx); err != nil {
		err = errors.New("failed executing code inside transaction: " + err.Error())
	}

	return err
}

func (m *manager) ReadCommitted(ctx context.Context, f Handler) error {
	txOpts := sql.TxOptions{Isolation: sql.LevelReadCommitted}
	return m.transaction(ctx, &txOpts, f)
}
