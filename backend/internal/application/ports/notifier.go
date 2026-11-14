package ports

import (
	"context"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
)

// ClosingInputsChanged representa el evento cuando cambian datos de entrada que afectan un cierre mensual.
type ClosingInputsChanged struct {
	ClientID string
	Period   period.Period
}

// ChangeNotifier desacopla la notificación de cambios del recálculo de cierres.
type ChangeNotifier interface {
	NotifyClosingInputsChanged(ctx context.Context, event ClosingInputsChanged) error
}

// NoOpChangeNotifier es la implementación por defecto / placeholder cuando el recálculo no está activo.
type NoOpChangeNotifier struct{}

// NotifyClosingInputsChanged no realiza ninguna acción.
func (n *NoOpChangeNotifier) NotifyClosingInputsChanged(_ context.Context, _ ClosingInputsChanged) error {
	return nil
}
