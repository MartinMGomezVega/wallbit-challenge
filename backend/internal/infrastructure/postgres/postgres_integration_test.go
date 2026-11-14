//go:build integration

package postgres_test

import (
	"context"
	"log/slog"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports/repotest"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/usecase"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/movement"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/arca/mock"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/clock"
	fxmock "github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/fx/mock"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/idgen"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/memory"
	notifylog "github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/notify/log"
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

	fxClient, err := fxmock.New()
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

func TestMemoryVsPostgresClosingQueryFullSeed(t *testing.T) {
	pool := openTestPool(t)
	truncateAll(t, pool)
	ctx := context.Background()

	appNow := time.Date(2026, 11, 5, 9, 0, 0, 0, period.Location())
	clk := clock.NewFixed(appNow)
	fxClient, err := fxmock.New()
	if err != nil {
		t.Fatalf("fxmock.New: %v", err)
	}
	arcaClient, err := mock.New()
	if err != nil {
		t.Fatalf("arca mock.New: %v", err)
	}

	ds, err := seed.Generate(ctx, seed.SeedConfig{
		Seed:    42,
		Clients: 300,
		AppNow:  appNow,
	}, fxClient, idgen.NewSequential())
	if err != nil {
		t.Fatalf("seed.Generate: %v", err)
	}

	// Clasificación inicial determinística
	classifications := make([]movement.Classification, 0, len(ds.Movements))
	for _, m := range ds.Movements {
		classifications = append(classifications, movement.Classification{
			MovementID: m.ID,
			Kind:       movement.KindIncome,
			Confidence: movement.ConfidenceHigh,
			Rule:       movement.RuleForeignCompany,
		})
	}

	// 1. Cargar y ejecutar en Memoria
	memClients := memory.NewClientRepository()
	memMovs := memory.NewMovementRepository()
	memClass := memory.NewClassificationRepository()
	memInvs := memory.NewInvoiceRepository()
	memPurch := memory.NewPurchaseInvoiceRepository()
	memInputs := memory.NewClientInputRepository()
	memParams := memory.NewParametersRepository()
	memClosings := memory.NewClosingRepository()

	_ = memClients.SaveAll(ctx, ds.Clients)
	_ = memMovs.SaveAll(ctx, ds.Movements)
	_ = memInvs.SaveAll(ctx, ds.Invoices)
	_ = memPurch.SaveAll(ctx, ds.PurchaseInvoices)
	_ = memInputs.SaveAll(ctx, ds.ClientInputs)
	_ = memClass.SaveAll(ctx, classifications)

	p := period.MustParse("2026-10")
	memCompute := &usecase.ComputeClosing{
		Clients: memClients, Movements: memMovs, Classifications: memClass,
		Invoices: memInvs, PurchaseInvoices: memPurch, ClientInputs: memInputs,
		Parameters: memParams, Closings: memClosings,
		TaxScales: arcaClient, DueDates: arcaClient, ExchangeRates: fxClient,
		Clock: clk, IDGen: idgen.NewSequential(), Tx: memory.NewTxManager(),
	}
	memRun := &usecase.RunMonthlyClosing{
		Clients: memClients, Compute: memCompute, Workers: 1,
	}
	if _, err := memRun.Execute(ctx, p); err != nil {
		t.Fatalf("mem RunMonthlyClosing: %v", err)
	}

	// 2. Cargar y ejecutar en Postgres
	if _, err := postgres.LoadSeed(ctx, pool, &ds, classifications, true); err != nil {
		t.Fatalf("postgres.LoadSeed: %v", err)
	}
	pgClients := postgres.NewClientRepository(pool)
	pgMovs := postgres.NewMovementRepository(pool)
	pgClass := postgres.NewClassificationRepository(pool)
	pgInvs := postgres.NewInvoiceRepository(pool)
	pgPurch := postgres.NewPurchaseInvoiceRepository(pool)
	pgInputs := postgres.NewClientInputRepository(pool)
	pgParams := postgres.NewParametersRepository(pool)
	pgClosings := postgres.NewClosingRepository(pool)
	pgTx := postgres.NewTxManager(pool)

	pgCompute := &usecase.ComputeClosing{
		Clients: pgClients, Movements: pgMovs, Classifications: pgClass,
		Invoices: pgInvs, PurchaseInvoices: pgPurch, ClientInputs: pgInputs,
		Parameters: pgParams, Closings: pgClosings,
		TaxScales: arcaClient, DueDates: arcaClient, ExchangeRates: fxClient,
		Clock: clk, IDGen: idgen.NewSequential(), Tx: pgTx,
	}
	pgRun := &usecase.RunMonthlyClosing{
		Clients: pgClients, Compute: pgCompute, Workers: 1,
	}
	if _, err := pgRun.Execute(ctx, p); err != nil {
		t.Fatalf("pg RunMonthlyClosing: %v", err)
	}

	memCQ := memory.NewClosingQuery(memClosings, memClients)
	pgCQ := postgres.NewClosingQuery(pool)
	repotest.CompareClosingQueries(t, memCQ, pgCQ, p)
}

func TestIdempotentStartup_WithoutReset(t *testing.T) {
	pool := openTestPool(t)
	truncateAll(t, pool)
	ctx := context.Background()

	fxClient, err := fxmock.New()
	if err != nil {
		t.Fatalf("fxmock.New: %v", err)
	}
	arcaClient, err := mock.New()
	if err != nil {
		t.Fatalf("mock.New: %v", err)
	}
	clk := clock.NewFixed(time.Date(2026, 11, 5, 9, 0, 0, 0, period.Location()))

	pSep := period.MustParse("2026-09")
	pOct := period.MustParse("2026-10")

	// Helper to simulate main.go startup
	runStartup := func() {
		idGen := idgen.NewSequential()
		if err := postgres.InitIDGen(ctx, pool, idGen); err != nil {
			t.Fatalf("InitIDGen: %v", err)
		}

		ds, err := seed.Generate(ctx, seed.SeedConfig{
			Seed:    42,
			Clients: 20,
			AppNow:  clk.Now(),
		}, fxClient, idGen)
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

		if _, err := postgres.LoadSeed(ctx, pool, &ds, classifications, false); err != nil {
			t.Fatalf("LoadSeed: %v", err)
		}

		clients := postgres.NewClientRepository(pool)
		movs := postgres.NewMovementRepository(pool)
		classRepo := postgres.NewClassificationRepository(pool)
		invs := postgres.NewInvoiceRepository(pool)
		purch := postgres.NewPurchaseInvoiceRepository(pool)
		inputs := postgres.NewClientInputRepository(pool)
		closings := postgres.NewClosingRepository(pool)
		notifs := postgres.NewNotificationRepository(pool)
		outbox := postgres.NewOutboxRepository(pool)
		params := postgres.NewParametersRepository(pool)
		tx := postgres.NewTxManager(pool)

		computeClosing := &usecase.ComputeClosing{
			Clients: clients, Movements: movs, Classifications: classRepo,
			Invoices: invs, PurchaseInvoices: purch, ClientInputs: inputs,
			Parameters: params, Closings: closings,
			TaxScales: arcaClient, DueDates: arcaClient, ExchangeRates: fxClient,
			Clock: clk, IDGen: idGen, Tx: tx,
		}

		runClosingUC := &usecase.RunMonthlyClosing{
			Clients: clients, Compute: computeClosing, Workers: 2,
		}

		if _, err := runClosingUC.Execute(ctx, pSep); err != nil {
			t.Fatalf("RunMonthlyClosing Sep: %v", err)
		}
		if err := usecase.SeedSeptemberClosings(ctx, closings, clients, idGen, 42, "C.P. Mariana Quiroga", "Matrícula de ejemplo"); err != nil {
			t.Fatalf("SeedSeptemberClosings: %v", err)
		}
		if _, err := runClosingUC.Execute(ctx, pOct); err != nil {
			t.Fatalf("RunMonthlyClosing Oct: %v", err)
		}

		octSeedClk := clock.NewFixed(time.Date(2026, 11, 4, 10, 0, 0, 0, period.Location()))
		octSeedApproveUC := &usecase.ApproveClosing{
			Closings:          closings,
			Clients:           clients,
			Notifications:     notifs,
			Outbox:            outbox,
			Sender:            notifylog.NewSender(slog.Default()),
			Clock:             octSeedClk,
			IDGen:             idGen,
			Tx:                tx,
			AccountantName:    "C.P. Mariana Quiroga",
			AccountantLicense: "Matrícula de ejemplo",
		}
		if err := usecase.SeedOctoberApprovedClosings(ctx, octSeedApproveUC, closings); err != nil {
			t.Fatalf("SeedOctoberApprovedClosings: %v", err)
		}
	}

	// 1er arranque
	runStartup()

	// Capturar métricas después del primer arranque
	statsQ := postgres.NewStatsQuery(pool)
	stats1, err := statsQ.PeriodStats(ctx, pOct)
	if err != nil {
		t.Fatalf("statsQ.PeriodStats 1: %v", err)
	}

	var aprCnt1, notifCnt1, cloCnt1 int
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM approvals`).Scan(&aprCnt1)
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM notifications`).Scan(&notifCnt1)
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM closings`).Scan(&cloCnt1)

	// 2do arranque sin SEED_RESET (idéntica base)
	runStartup()

	// Capturar métricas después del segundo arranque
	stats2, err := statsQ.PeriodStats(ctx, pOct)
	if err != nil {
		t.Fatalf("statsQ.PeriodStats 2: %v", err)
	}

	var aprCnt2, notifCnt2, cloCnt2 int
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM approvals`).Scan(&aprCnt2)
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM notifications`).Scan(&notifCnt2)
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM closings`).Scan(&cloCnt2)

	if aprCnt1 != aprCnt2 {
		t.Fatalf("approvals count changed: %d vs %d (duplicados)", aprCnt1, aprCnt2)
	}
	if notifCnt1 != notifCnt2 {
		t.Fatalf("notifications count changed: %d vs %d (duplicados)", notifCnt1, notifCnt2)
	}
	if cloCnt1 != cloCnt2 {
		t.Fatalf("closings count changed: %d vs %d", cloCnt1, cloCnt2)
	}
	if !reflect.DeepEqual(stats1, stats2) {
		t.Fatalf("stats mismatch between first and second startup:\n%+v\n%+v", stats1, stats2)
	}
}
