// Package scheduler implementa la tarea programada que dispara el cierre mensual
// el día 1 de cada mes (spec 007 criterio 8).
package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/usecase"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
)

// Runner define la interfaz para ejecutar el cierre mensual masivo.
type Runner interface {
	Execute(ctx context.Context, p period.Period) (usecase.RunSummary, error)
}

// ReminderRunner define la interfaz para ejecutar los recordatorios de vencimiento.
type ReminderRunner interface {
	Execute(ctx context.Context, now time.Time) (int, error)
}

// Scheduler supervisa el reloj y dispara automáticamente el cierre mensual y recordatorios diarios.
type Scheduler struct {
	Runner         Runner
	ReminderRunner ReminderRunner
	Clock          ports.Clock
	Interval       time.Duration

	mu              sync.Mutex
	lastRan         period.Period
	lastReminderDay string
}

// NewScheduler crea un nuevo programador de tareas.
func NewScheduler(runner Runner, clock ports.Clock) *Scheduler {
	return &Scheduler{
		Runner:   runner,
		Clock:    clock,
		Interval: time.Minute,
	}
}

// CheckAndRun verifica si hoy es día 1 y el mes anterior aún no fue corrido, y corre recordatorios diarios.
// Devuelve true si disparó la corrida de cierre mensual.
func (s *Scheduler) CheckAndRun(ctx context.Context) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.Clock.Now().In(period.Location())

	// Recordatorios diarios
	todayStr := now.Format("2006-01-02")
	if s.ReminderRunner != nil && s.lastReminderDay != todayStr {
		s.lastReminderDay = todayStr
		_, _ = s.ReminderRunner.Execute(ctx, now)
	}

	if now.Day() != 1 {
		return false, nil
	}

	targetPeriod := period.FromTime(now).Prev()
	if s.lastRan == targetPeriod {
		return false, nil
	}

	s.lastRan = targetPeriod
	if s.Runner != nil {
		_, err := s.Runner.Execute(ctx, targetPeriod)
		return true, err
	}
	return true, nil
}

// Start inicia el ciclo periódico del scheduler en segundo plano hasta que se cancele ctx.
func (s *Scheduler) Start(ctx context.Context) {
	interval := s.Interval
	if interval <= 0 {
		interval = time.Minute
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	_, _ = s.CheckAndRun(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_, _ = s.CheckAndRun(ctx)
		}
	}
}
