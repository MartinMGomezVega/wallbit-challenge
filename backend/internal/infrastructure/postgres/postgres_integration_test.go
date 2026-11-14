//go:build integration

package postgres_test

import (
	"context"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports/repotest"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/movement"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/fx/mock"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/idgen"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/postgres"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/seed"
	"github.com/jackc/pgx/v5/pgxpool"
)

func openTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		dsn = "postgres://wallbit:wallbit@localhost:5432/accountant_test?sslmode=disable"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	ensureTestDB(ctx, dsn)

	pool, err := postgres.OpenPool(ctx, dsn)
	if err != nil {
		t.Fatalf("OpenPool failed: %v", err)
	}
	t.Cleanup(pool.Close)

	if err := postgres.Migrate(ctx, pool); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}
	return pool
}

func ensureTestDB(ctx context.Context, dsn string) {
	u, err := url.Parse(dsn)
	if err != nil {
		return
	}
	dbName := strings.TrimPrefix(u.Path, "/")
	if dbName == "" || dbName == "postgres" || dbName == "accountant" {
		return
	}
	adminURL := *u
	adminURL.Path = "/accountant"
	adminPool, err := postgres.OpenPool(ctx, adminURL.String())
	if err != nil {
		adminURL.Path = "/postgres"
		adminPool, err = postgres.OpenPool(ctx, adminURL.String())
		if err != nil {
			return
		}
	}
	defer adminPool.Close()
	_, _ = adminPool.Exec(ctx, "CREATE DATABASE "+dbName)
}

func truncateAll(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	_, err := pool.Exec(ctx, `
		TRUNCATE TABLE
			approvals, closing_signals, closings,
			external_income_items, client_period_inputs,
			purchase_invoices, invoices,
			classifications, movements,
			income_tax_advances, regime_history, clients,
			notifications, outbox, parameters
		CASCADE
	`)
	if err != nil {
		t.Fatalf("truncate failed: %v", err)
	}
}

func TestMigrateIdempotent(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()
	if err := postgres.Migrate(ctx, pool); err != nil {
		t.Fatalf("second Migrate failed: %v", err)
	}
	var cnt int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&cnt); err != nil {
		t.Fatalf("count schema_migrations: %v", err)
	}
	if cnt < 2 {
		t.Fatalf("expected >= 2 migrations recorded, got %d", cnt)
	}
}

func TestPostgresContract(t *testing.T) {
	pool := openTestPool(t)
	newBundle := func(t *testing.T) repotest.Bundle {
		t.Helper()
		truncateAll(t, pool)
		return repotest.Bundle{
			TxManager:        postgres.NewTxManager(pool),
			Clients:          postgres.NewClientRepository(pool),
			Movements:        postgres.NewMovementRepository(pool),
			Classifications:  postgres.NewClassificationRepository(pool),
			Invoices:         postgres.NewInvoiceRepository(pool),
			PurchaseInvoices: postgres.NewPurchaseInvoiceRepository(pool),
			ClientInputs:     postgres.NewClientInputRepository(pool),
			Closings:         postgres.NewClosingRepository(pool),
			ClosingQuery:     postgres.NewClosingQuery(pool),
			StatsQuery:       postgres.NewStatsQuery(pool),
			Notifications:    postgres.NewNotificationRepository(pool),
			Outbox:           postgres.NewOutboxRepository(pool),
			Parameters:       postgres.NewParametersRepository(pool),
		}
	}
	repotest.RunAll(t, newBundle, true)
}

func TestLoadSeedIdempotentAndReset(t *testing.T) {
	pool := openTestPool(t)
	truncateAll(t, pool)
	ctx := context.Background()

	fxClient, err := mock.New()
	if err != nil {
		t.Fatalf("fxmock.New: %v", err)
	}
	ds, err := seed.Generate(ctx, seed.SeedConfig{
		Seed:    42,
		Clients: 10,
		AppNow:  time.Date(2026, 11, 5, 9, 0, 0, 0, period.Location()),
	}, fxClient, idgen.NewSequential())
	if err != nil {
		t.Fatalf("seed.Generate: %v", err)
	}
	var classifications []movement.Classification
	for _, m := range ds.Movements {
		classifications = append(classifications, movement.Classification{
			MovementID: m.ID,
			Kind:       movement.KindIncome,
			Confidence: movement.ConfidenceHigh,
			Rule:       movement.RuleForeignCompany,
		})
	}

	inserted1, err := postgres.LoadSeed(ctx, pool, &ds, classifications, false)
	if err != nil || !inserted1 {
		t.Fatalf("first LoadSeed = (%v, %v), want (true, nil)", inserted1, err)
	}

	inserted2, err := postgres.LoadSeed(ctx, pool, &ds, classifications, false)
	if err != nil || inserted2 {
		t.Fatalf("second LoadSeed = (%v, %v), want (false, nil)", inserted2, err)
	}

	var cnt int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM clients`).Scan(&cnt); err != nil {
		t.Fatalf("count clients: %v", err)
	}
	if cnt != len(ds.Clients) {
		t.Fatalf("clients count = %d, want %d", cnt, len(ds.Clients))
	}

	inserted3, err := postgres.LoadSeed(ctx, pool, &ds, classifications, true)
	if err != nil || !inserted3 {
		t.Fatalf("reset LoadSeed = (%v, %v), want (true, nil)", inserted3, err)
	}
}
