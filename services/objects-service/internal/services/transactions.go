package services

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/v-egorov/service-boilerplate/services/objects-service/internal/repository"
)

// Transaction wraps multiple repositories under a single database transaction
type Transaction interface {
	ObjectTypeRepository() repository.ObjectTypeRepository
	ObjectRepository() repository.ObjectRepository
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// txDB implements Transaction using pgx.Tx
type txDB struct {
	tx             pgx.Tx
	objectTypeRepo repository.ObjectTypeRepository
	objectRepo     repository.ObjectRepository
}

// NewTransaction creates a new transaction wrapper
func NewTransaction(tx pgx.Tx, objectTypeRepo repository.ObjectTypeRepository, objectRepo repository.ObjectRepository) Transaction {
	return &txDB{
		tx:             tx,
		objectTypeRepo: objectTypeRepo,
		objectRepo:     objectRepo,
	}
}

// ObjectTypeRepository returns the wrapped object type repository
func (t *txDB) ObjectTypeRepository() repository.ObjectTypeRepository {
	return t.objectTypeRepo
}

// ObjectRepository returns the wrapped object repository
func (t *txDB) ObjectRepository() repository.ObjectRepository {
	return t.objectRepo
}

// Commit commits the transaction
func (t *txDB) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

// Rollback rolls back the transaction
func (t *txDB) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}

// TransactionalDB wraps a pgxpool.Pool and provides transaction support
type TransactionalDB struct {
	pool *pgxpool.Pool
}

// NewTransactionalDB creates a new TransactionalDB
func NewTransactionalDB(pool *pgxpool.Pool) *TransactionalDB {
	return &TransactionalDB{pool: pool}
}

// Begin starts a new transaction with default options
func (db *TransactionalDB) Begin(ctx context.Context) (Transaction, error) {
	return db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:       pgx.Serializable,
		AccessMode:     pgx.ReadWrite,
		DeferrableMode: pgx.NotDeferrable,
	})
}

// BeginTx starts a new transaction with custom options
func (db *TransactionalDB) BeginTx(ctx context.Context, opts pgx.TxOptions) (Transaction, error) {
	tx, err := db.pool.BeginTx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	wrappedTx := &txWrapper{tx: tx}

	objectTypeRepo := repository.NewObjectTypeRepository(wrappedTx, nil)
	objectRepo := repository.NewObjectRepository(wrappedTx, nil)

	return &txDB{
		tx:             tx,
		objectTypeRepo: objectTypeRepo,
		objectRepo:     objectRepo,
	}, nil
}

// txWrapper wraps pgx.Tx to implement repository.DBInterface
type txWrapper struct {
	tx pgx.Tx
}

func (w *txWrapper) Query(ctx context.Context, sql string, args ...any) (repository.Rows, error) {
	return w.tx.Query(ctx, sql, args...)
}

func (w *txWrapper) QueryRow(ctx context.Context, sql string, args ...any) repository.Row {
	return w.tx.QueryRow(ctx, sql, args...)
}

func (w *txWrapper) Exec(ctx context.Context, sql string, args ...any) (repository.CommandTag, error) {
	return w.tx.Exec(ctx, sql, args...)
}

// WithinTx executes a function within a transaction
// The function receives the transaction and should return an error
// The transaction is automatically committed on success or rolled back on error
func WithinTx(ctx context.Context, db *pgxpool.Pool, fn func(tx Transaction) error) error {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:       pgx.Serializable,
		AccessMode:     pgx.ReadWrite,
		DeferrableMode: pgx.NotDeferrable,
	})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	txWrapper := NewTransaction(tx, nil, nil)

	if err := fn(txWrapper); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("error: %v, rollback error: %w", err, rbErr)
		}
		return err
	}

	return tx.Commit(ctx)
}
