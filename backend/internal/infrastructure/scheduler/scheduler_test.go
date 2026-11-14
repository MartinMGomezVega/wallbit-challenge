package scheduler_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports/fakes"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/usecase"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/scheduler"
)

type mockRunner struct {
	mu    sync.Mutex
	calls []period.Period
}

func (m *mockRunner) Execute(_ context.Context, p period.Period) (usecase.RunSummary, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, p)
	return usecase.RunSummary{Processed: 1}, nil
}

func (m *mockRunner) CallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.calls)
}

func TestScheduler_Day1_RunsOnce(t *testing.T) {
	ctx := context.Background()
	// Día 1 de Noviembre 2026 -> debe disparar para Octubre 2026
	day1 := time.Date(2026, 11, 1, 9, 0, 0, 0, period.Location())
	clk := fakes.Clock{Current: day1}
	runner := &mockRunner{}

	sched := scheduler.NewScheduler(runner, clk)

	// Primera verificación -> debe correr
	ran, err := sched.CheckAndRun(ctx)
	if err != nil {
		t.Fatalf("error en CheckAndRun: %v", err)
	}
	if !ran {
		t.Fatal("esperaba que disparara en día 1")
	}
	if runner.CallCount() != 1 {
		t.Fatalf("esperaba 1 llamada, obtuvo %d", runner.CallCount())
	}
	if runner.calls[0] != period.MustParse("2026-10") {
		t.Fatalf("esperaba período objetivo 2026-10, obtuvo %s", runner.calls[0])
	}

	// Segunda verificación con mismo reloj fijo -> no debe correr de nuevo
	ranAgain, err := sched.CheckAndRun(ctx)
	if err != nil {
		t.Fatalf("error en segunda CheckAndRun: %v", err)
	}
	if ranAgain {
		t.Fatal("no debió correr nuevamente para el mismo período")
	}
	if runner.CallCount() != 1 {
		t.Fatalf("el conteo debe seguir siendo 1, obtuvo %d", runner.CallCount())
	}
}

func TestScheduler_NotDay1_DoesNotRun(t *testing.T) {
	ctx := context.Background()
	// Día 15 de Noviembre 2026
	day15 := time.Date(2026, 11, 15, 9, 0, 0, 0, period.Location())
	clk := fakes.Clock{Current: day15}
	runner := &mockRunner{}

	sched := scheduler.NewScheduler(runner, clk)

	ran, err := sched.CheckAndRun(ctx)
	if err != nil {
		t.Fatalf("error en CheckAndRun: %v", err)
	}
	if ran {
		t.Fatal("no debe correr en día 15")
	}
	if runner.CallCount() != 0 {
		t.Fatalf("no debió haber llamadas al runner, hubo %d", runner.CallCount())
	}
}

func TestScheduler_BackgroundStart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	day1 := time.Date(2026, 11, 1, 9, 0, 0, 0, period.Location())
	clk := fakes.Clock{Current: day1}
	runner := &mockRunner{}

	sched := scheduler.NewScheduler(runner, clk)
	sched.Interval = 10 * time.Millisecond

	go sched.Start(ctx)

	// Esperar brevemente para verificar que ejecutó el chequeo inicial
	time.Sleep(50 * time.Millisecond)
	cancel()

	if runner.CallCount() != 1 {
		t.Fatalf("esperaba exactamente 1 ejecución en segundo plano, obtuvo %d", runner.CallCount())
	}
}

type mockReminderRunner struct {
	mu    sync.Mutex
	calls []time.Time
}

func (m *mockReminderRunner) Execute(_ context.Context, now time.Time) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, now)
	return len(m.calls), nil
}

func TestScheduler_ReminderRunner_Daily(t *testing.T) {
	ctx := context.Background()
	day1 := time.Date(2026, 11, 10, 9, 0, 0, 0, period.Location())
	clk := &fakes.Clock{Current: day1}
	runner := &mockRunner{}
	reminders := &mockReminderRunner{}

	sched := scheduler.NewScheduler(runner, clk)
	sched.ReminderRunner = reminders

	// Primera corrida del día 10
	_, err := sched.CheckAndRun(ctx)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(reminders.calls) != 1 {
		t.Fatalf("esperaba 1 llamada a reminders, obtuvo %d", len(reminders.calls))
	}

	// Misma fecha -> no vuelve a correr
	_, _ = sched.CheckAndRun(ctx)
	if len(reminders.calls) != 1 {
		t.Fatalf("no debió correr dos veces en el mismo día, obtuvo %d", len(reminders.calls))
	}

	// Avanzar al día 11 -> debe correr nuevamente
	clk.Current = clk.Current.AddDate(0, 0, 1)
	_, _ = sched.CheckAndRun(ctx)
	if len(reminders.calls) != 2 {
		t.Fatalf("debía correr en el nuevo día, obtuvo %d", len(reminders.calls))
	}
}
