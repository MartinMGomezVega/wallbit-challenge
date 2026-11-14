package usecase

import (
	"context"
	"strings"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/clientinput"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
)

// ConfirmInputOnBehalfInput agrupa los parámetros para confirmar el período en nombre del cliente.
type ConfirmInputOnBehalfInput struct {
	ClientID       string
	Period         period.Period
	None           bool
	ExternalIncome []clientinput.ExternalIncomeItem
	Note           string
}

// ConfirmInputOnBehalf sella la declaración de ingresos externos y confirma el período en nombre del cliente (spec 016).
type ConfirmInputOnBehalf struct {
	Inputs   ports.ClientInputRepository
	Closings ports.ClosingRepository
	Clock    ports.Clock
	Notifier ports.ChangeNotifier
	Tx       ports.TxManager
}

// Execute valida la nota, persiste el input confirmado por el contador y emite el evento de recálculo.
func (uc *ConfirmInputOnBehalf) Execute(ctx context.Context, in ConfirmInputOnBehalfInput) (clientinput.ClientPeriodInput, error) {
	trimmedNote := strings.TrimSpace(in.Note)
	if len(trimmedNote) < 1 || len(trimmedNote) > 300 {
		return clientinput.ClientPeriodInput{}, domain.NewError(
			domain.ErrValidation,
			"NOTE_REQUIRED",
			"La nota es obligatoria cuando el contador confirma el período (1 a 300 caracteres)",
		)
	}

	var res clientinput.ClientPeriodInput
	err := withinTx(ctx, uc.Tx, func(ctx context.Context) error {
		existing, err := uc.Inputs.Get(ctx, in.ClientID, in.Period)
		if err != nil {
			if errorsIsNotFound(err) {
				existing = clientinput.ClientPeriodInput{
					ClientID: in.ClientID,
					Period:   in.Period,
				}
			} else {
				return err
			}
		}

		if existing.IsConfirmed() {
			return domain.NewError(domain.ErrConflict, "ALREADY_CONFIRMED", "El período ya fue confirmado")
		}

		if in.None {
			existing.DeclaredNoExternalIncome = true
			existing.ExternalIncome = []clientinput.ExternalIncomeItem{}
		} else {
			if len(in.ExternalIncome) == 0 {
				return domain.NewError(
					domain.ErrValidation,
					"EXTERNAL_INCOME_NOT_ANSWERED",
					"Se requiere responder sobre ingresos externos antes de confirmar el período",
				)
			}
			existing.DeclaredNoExternalIncome = false
			existing.ExternalIncome = in.ExternalIncome
		}

		now := uc.Clock.Now()
		if err := existing.ConfirmBy(clientinput.ConfirmedByAccountant, trimmedNote, now); err != nil {
			return err
		}

		if err := uc.Inputs.Save(ctx, existing); err != nil {
			return err
		}

		res = existing
		return nil
	})
	if err != nil {
		return clientinput.ClientPeriodInput{}, err
	}

	if uc.Notifier != nil {
		_ = uc.Notifier.NotifyClosingInputsChanged(ctx, ports.ClosingInputsChanged{
			ClientID: in.ClientID,
			Period:   in.Period,
		})
	}

	return res, nil
}
