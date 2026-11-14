package memory_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/closing"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/signal"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/tax"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/memory"
)

func TestClosingRepositoryCRUD(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewClosingRepository()
	p10 := period.MustParse("2026-10")
	p11 := period.MustParse("2026-11")
	now := time.Date(2026, 11, 1, 0, 0, 0, 0, period.Location())

	// 1. Get inexistente -> ErrNotFound
	_, err := repo.Get(ctx, "nonexistent")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("esperaba ErrNotFound, obtuvo: %v", err)
	}

	// 2. ByClientPeriod inexistente -> found = false
	_, found, err := repo.ByClientPeriod(ctx, "cli_1", p10)
	if err != nil || found {
		t.Fatalf("esperaba found=false, err=nil; got found=%v, err=%v", found, err)
	}

	// 3. Save
	c1 := closing.NewClosing(
		"clo_1", "cli_1", p10, true, "hash1",
		signal.SeverityGreen, nil, "desc", nil, nil, closing.Reconciliation{}, now,
	)
	if err := repo.Save(ctx, c1); err != nil {
		t.Fatalf("error al guardar: %v", err)
	}

	// 4. Get existente
	got, err := repo.Get(ctx, "clo_1")
	if err != nil {
		t.Fatalf("error en Get: %v", err)
	}
	if got.ID != "clo_1" || got.ClientID != "cli_1" {
		t.Fatalf("datos incorrectos en Get: %+v", got)
	}

	// 5. ByClientPeriod existente
	got2, found, err := repo.ByClientPeriod(ctx, "cli_1", p10)
	if err != nil || !found {
		t.Fatalf("esperaba found=true en ByClientPeriod: got found=%v, err=%v", found, err)
	}
	if got2.ID != "clo_1" {
		t.Fatalf("ID incorrecto: %s", got2.ID)
	}

	// 6. Upsert con mismo ID pero datos modificados
	c1.Status = closing.StatusApproved
	if err := repo.Save(ctx, c1); err != nil {
		t.Fatalf("error en upsert: %v", err)
	}
	got3, _ := repo.Get(ctx, "clo_1")
	if got3.Status != closing.StatusApproved {
		t.Fatalf("estado no actualizado: %s", got3.Status)
	}

	// 7. Upsert con nuevo ID para el mismo (clientID, period) reemplaza el viejo ID
	c1b := closing.NewClosing(
		"clo_1b", "cli_1", p10, true, "hash2",
		signal.SeverityYellow, nil, "desc", nil, nil, closing.Reconciliation{}, now,
	)
	if err := repo.Save(ctx, c1b); err != nil {
		t.Fatalf("error en upsert con nuevo ID: %v", err)
	}
	// El viejo ID ya no debe existir
	_, err = repo.Get(ctx, "clo_1")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("el viejo ID debió ser eliminado")
	}
	got1b, err := repo.Get(ctx, "clo_1b")
	if err != nil || got1b.ID != "clo_1b" {
		t.Fatalf("el nuevo ID debe existir")
	}

	// 8. ListByPeriod
	c2 := closing.NewClosing(
		"clo_2", "cli_2", p10, false, "hash2",
		signal.SeverityYellow, nil, "desc", nil, nil, closing.Reconciliation{}, now,
	)
	c3 := closing.NewClosing(
		"clo_3", "cli_1", p11, true, "hash3",
		signal.SeverityGreen, nil, "desc", nil, nil, closing.Reconciliation{}, now,
	)
	_ = repo.Save(ctx, c2)
	_ = repo.Save(ctx, c3)

	list10, err := repo.ListByPeriod(ctx, p10)
	if err != nil {
		t.Fatalf("error en ListByPeriod: %v", err)
	}
	if len(list10) != 2 {
		t.Fatalf("esperaba 2 cierres en p10, obtuvo %d", len(list10))
	}
	// Deben estar ordenados por ClientID: cli_1, cli_2
	if list10[0].ClientID != "cli_1" || list10[1].ClientID != "cli_2" {
		t.Fatalf("orden por ClientID incorrecto: %s, %s", list10[0].ClientID, list10[1].ClientID)
	}
}

func TestLatestExclusionDate(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewClosingRepository()
	p09 := period.MustParse("2026-09")
	p10 := period.MustParse("2026-10")
	now := time.Date(2026, 11, 1, 0, 0, 0, 0, period.Location())

	// Sin exclusiones
	d, err := repo.LatestExclusionDate(ctx, "cli_1")
	if err != nil || d != nil {
		t.Fatalf("esperaba nil exclusion date, obtuvo %v, err=%v", d, err)
	}

	// Con exclusión en p10
	exclDate := time.Date(2026, 10, 14, 0, 0, 0, 0, period.Location())
	c := closing.NewClosing(
		"clo_1", "cli_1", p10, true, "h",
		signal.SeverityRed, nil, "desc",
		&tax.MonotributoAssessment{
			Excluded:      true,
			ExclusionDate: &exclDate,
		},
		nil, closing.Reconciliation{}, now,
	)
	_ = repo.Save(ctx, c)

	got, err := repo.LatestExclusionDate(ctx, "cli_1")
	if err != nil || got == nil {
		t.Fatalf("esperaba exclusión, obtuvo %v", got)
	}
	if !got.Equal(exclDate) {
		t.Fatalf("fecha de exclusión incorrecta: got %v, want %v", got, exclDate)
	}

	// Para otro cliente -> nil
	gotOther, err := repo.LatestExclusionDate(ctx, "cli_other")
	if err != nil || gotOther != nil {
		t.Fatalf("esperaba nil para otro cliente")
	}

	// Cierre anterior no excluido no sobreescribe
	cOld := closing.NewClosing(
		"clo_0", "cli_1", p09, true, "h0",
		signal.SeverityGreen, nil, "desc",
		&tax.MonotributoAssessment{
			Excluded: false,
		},
		nil, closing.Reconciliation{}, now,
	)
	_ = repo.Save(ctx, cOld)
	gotStill, _ := repo.LatestExclusionDate(ctx, "cli_1")
	if gotStill == nil || !gotStill.Equal(exclDate) {
		t.Fatalf("fecha de exclusión debió permanecer %v", exclDate)
	}
}

func TestClosingRepositoryConcurrency(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewClosingRepository()
	p := period.MustParse("2026-10")
	now := time.Date(2026, 11, 1, 0, 0, 0, 0, period.Location())

	const numGoroutines = 50
	var wg sync.WaitGroup
	wg.Add(numGoroutines * 3)

	// Goroutines que escriben
	for i := 0; i < numGoroutines; i++ {
		clientID := fmt.Sprintf("cli_%03d", i%10)
		closingID := fmt.Sprintf("clo_%03d", i)
		go func(cid, clid string) {
			defer wg.Done()
			c := closing.NewClosing(
				cid, clid, p, true, "hash",
				signal.SeverityGreen, nil, "desc", nil, nil, closing.Reconciliation{}, now,
			)
			_ = repo.Save(ctx, c)
		}(closingID, clientID)
	}

	// Goroutines que leen por clientPeriod
	for i := 0; i < numGoroutines; i++ {
		clientID := fmt.Sprintf("cli_%03d", i%10)
		go func(clid string) {
			defer wg.Done()
			_, _, _ = repo.ByClientPeriod(ctx, clid, p)
		}(clientID)
	}

	// Goroutines que listan
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			_, _ = repo.ListByPeriod(ctx, p)
			_, _ = repo.LatestExclusionDate(ctx, "cli_001")
		}()
	}

	wg.Wait()
}
