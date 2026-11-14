package usecase_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/usecase"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/tax"
)

func TestRunMonthlyClosing(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 11, 1, 9, 0, 0, 0, period.Location())
	h := newTestHarness(now)
	p := period.MustParse("2026-10")

	// Crear 10 clientes
	for i := 1; i <= 10; i++ {
		cuit, _ := client.ParseCUIT(fmt.Sprintf("20-%08d-1", i))
		cli := client.Client{
			ID:                  fmt.Sprintf("cli_%02d", i),
			Name:                fmt.Sprintf("Cliente %d", i),
			CUIT:                cuit,
			Regime:              client.RegimeMonotributo,
			MonotributoCategory: tax.CatA,
			CreatedAt:           time.Date(2025, 1, 1, 0, 0, 0, 0, period.Location()),
		}
		_ = h.clients.Save(ctx, cli)
	}

	runner := &usecase.RunMonthlyClosing{
		Clients: h.clients,
		Compute: h.uc,
		Workers: 4,
	}

	// 1. Primera corrida -> todos procesados y creados
	sum1, err := runner.Execute(ctx, p)
	if err != nil {
		t.Fatalf("error en run 1: %v", err)
	}
	if sum1.Processed != 10 || sum1.Created != 10 || sum1.Updated != 0 || sum1.Unchanged != 0 || len(sum1.Failed) != 0 {
		t.Fatalf("resultado de run 1 inesperado: %+v", sum1)
	}

	// 2. Segunda corrida -> idempotente: created=0, updated=0, unchanged=10
	sum2, err := runner.Execute(ctx, p)
	if err != nil {
		t.Fatalf("error en run 2: %v", err)
	}
	if sum2.Processed != 10 || sum2.Created != 0 || sum2.Updated != 0 || sum2.Unchanged != 10 || len(sum2.Failed) != 0 {
		t.Fatalf("resultado de run 2 (idempotente) inesperado: %+v", sum2)
	}
}

func TestRunMonthlyClosing_ErrorInOneClientDoesNotHalt(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 11, 1, 9, 0, 0, 0, period.Location())
	h := newTestHarness(now)
	p := period.MustParse("2026-10")

	// Crear 5 clientes válidos y 1 cliente con CUIT inválida que fallará en DueDates
	for i := 1; i <= 5; i++ {
		cuit, _ := client.ParseCUIT(fmt.Sprintf("20-%08d-1", i))
		cli := client.Client{
			ID:                  fmt.Sprintf("cli_ok_%d", i),
			Name:                fmt.Sprintf("Cliente OK %d", i),
			CUIT:                cuit,
			Regime:              client.RegimeMonotributo,
			MonotributoCategory: tax.CatA,
			CreatedAt:           time.Date(2025, 1, 1, 0, 0, 0, 0, period.Location()),
		}
		_ = h.clients.Save(ctx, cli)
	}

	// Cliente 6: CUIT sin configurar / inválida
	cliBad := client.Client{
		ID:                  "cli_bad",
		Name:                "Cliente Fallido",
		CUIT:                client.CUIT("invalid-cuit"),
		Regime:              client.RegimeMonotributo,
		MonotributoCategory: tax.CatA,
		CreatedAt:           time.Date(2025, 1, 1, 0, 0, 0, 0, period.Location()),
	}
	_ = h.clients.Save(ctx, cliBad)

	runner := &usecase.RunMonthlyClosing{
		Clients: h.clients,
		Compute: h.uc,
		Workers: 3,
	}

	sum, err := runner.Execute(ctx, p)
	if err != nil {
		t.Fatalf("Execute no debió fallar a nivel global: %v", err)
	}

	if sum.Processed != 6 {
		t.Fatalf("esperaba 6 procesados, obtuvo %d", sum.Processed)
	}
	// Los 5 clientes válidos deben haber sido procesados con éxito
	if sum.Created != 5 && sum.Unchanged != 5 {
		// En la primera corrida deben crearse 5
		if sum.Created != 5 {
			t.Fatalf("esperaba 5 creados, obtuvo %d", sum.Created)
		}
	}
	if len(sum.Failed) != 1 || sum.Failed[0].ClientID != "cli_bad" {
		t.Fatalf("esperaba 1 fallo en cli_bad, obtuvo: %+v", sum.Failed)
	}
}
