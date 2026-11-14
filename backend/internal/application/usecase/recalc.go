package usecase

import (
	"context"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/closing"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
)

var _ ports.ChangeNotifier = (*RecalcOnChange)(nil)

// RecalcOnChange implementa ports.ChangeNotifier recalculando sincrónicamente el cierre afectado
// y el del período siguiente si ya existe en el repositorio (spec 007 criterio 11).
type RecalcOnChange struct {
	Compute  *ComputeClosing
	Closings ports.ClosingRepository
}

// NotifyClosingInputsChanged recalcula el cierre del período del evento y el siguiente si existe.
func (r *RecalcOnChange) NotifyClosingInputsChanged(ctx context.Context, event ports.ClosingInputsChanged) error {
	if r.Compute == nil {
		return nil
	}

	// 1. Recalcular el período afectado
	if _, err := r.Compute.Execute(ctx, event.ClientID, event.Period); err != nil {
		return err
	}

	// 2. Si existe un cierre del período siguiente para ese cliente, recalcularlo también
	nextPeriod := event.Period.Next()
	if r.Closings != nil {
		_, found, err := r.Closings.ByClientPeriod(ctx, event.ClientID, nextPeriod)
		if err == nil && found {
			_, _ = r.Compute.Execute(ctx, event.ClientID, nextPeriod)
		}
	}

	return nil
}

// RecalcPeriod recalcula todos los cierres no aprobados de un período (spec 007 criterio 12).
type RecalcPeriod struct {
	Closings ports.ClosingRepository
	Compute  *ComputeClosing
}

// Execute recorre los cierres del período y recalcula los que no estén aprobados ni notificados.
func (uc *RecalcPeriod) Execute(ctx context.Context, p period.Period) (int, error) {
	if uc.Closings == nil || uc.Compute == nil {
		return 0, nil
	}

	list, err := uc.Closings.ListByPeriod(ctx, p)
	if err != nil {
		return 0, err
	}

	recalculated := 0
	for _, c := range list {
		if c.Status == closing.StatusApproved || c.Status == closing.StatusNotified {
			continue
		}
		if _, err := uc.Compute.Execute(ctx, c.ClientID, p); err == nil {
			recalculated++
		}
	}

	return recalculated, nil
}

