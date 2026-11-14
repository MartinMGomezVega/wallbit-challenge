// Package postgres implementa los puertos de persistencia sobre PostgreSQL usando pgx/v5.
package postgres

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/idgen"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBTX abstrae *pgxpool.Pool y pgx.Tx para ejecutar queries dentro o fuera de una transacción.
type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type txCtxKey struct{}

// OpenPool abre un pool pgxpool con reintentos exponenciales breves hasta que Postgres responda al Ping.
func OpenPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt < 15; attempt++ {
		pool, err := pgxpool.NewWithConfig(ctx, cfg)
		if err == nil {
			if pingErr := pool.Ping(ctx); pingErr == nil {
				return pool, nil
			} else {
				lastErr = pingErr
				pool.Close()
			}
		} else {
			lastErr = err
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
	return nil, fmt.Errorf("connect postgres after retries: %w", lastErr)
}

func conn(ctx context.Context, pool *pgxpool.Pool) DBTX {
	if tx, ok := ctx.Value(txCtxKey{}).(pgx.Tx); ok && tx != nil {
		return tx
	}
	return pool
}

// InitIDGen inicializa los contadores de Sequential con los IDs máximos existentes en Postgres.
func InitIDGen(ctx context.Context, pool *pgxpool.Pool, gen *idgen.Sequential) error {
	queries := []struct {
		table  string
		prefix string
	}{
		{"notifications", "notif"},
		{"outbox", "outbox"},
		{"closings", "clo"},
		{"clients", "cli"},
		{"movements", "mov"},
		{"invoices", "inv"},
		{"purchase_invoices", "pinv"},
		{"approvals", "apr"},
	}

	for _, q := range queries {
		var maxID string
		err := pool.QueryRow(ctx, fmt.Sprintf("SELECT id FROM %s ORDER BY id DESC LIMIT 1", q.table)).Scan(&maxID)
		if err != nil {
			continue
		}
		prefix := q.prefix + "_"
		if strings.HasPrefix(maxID, prefix) {
			if num, err := strconv.Atoi(strings.TrimPrefix(maxID, prefix)); err == nil {
				gen.SetMin(q.prefix, num)
			}
		}
	}
	return nil
}
