package postgres

import (
	"context"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ ports.TxManager = (*TxManager)(nil)

// TxManager implementa ports.TxManager usando transacciones de Postgres.
type TxManager struct {
	pool *pgxpool.Pool
}

// NewTxManager crea un TxManager respaldado por pgxpool.
func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

// WithinTx ejecuta fn dentro de una transacción SQL. Si ya hay una transacción activa en ctx, la reutiliza.
func (m *TxManager) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if existing, ok := ctx.Value(txCtxKey{}).(pgx.Tx); ok && existing != nil {
		return fn(ctx)
	}

	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return err
	}
	txCtx := context.WithValue(ctx, txCtxKey{}, tx)
	if err := fn(txCtx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}
