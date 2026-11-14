package usecase

import (
	"context"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/movement"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
)

// OverrideClassificationInput agrupa los datos necesarios para sobrescribir la clasificación.
type OverrideClassificationInput struct {
	MovementID string
	Kind       movement.ClassificationKind
	Note       string
	Operator   string // por defecto "accountant"
}

// OverrideClassification permite al contador corregir manualmente la clasificación de un movimiento.
type OverrideClassification struct {
	Movements       ports.MovementRepository
	Classifications ports.ClassificationRepository
	Clock           ports.Clock
	Notifier        ports.ChangeNotifier
}

// Execute valida los parámetros, registra la corrección y notifica el cambio para recálculo de cierres.
func (uc *OverrideClassification) Execute(ctx context.Context, input OverrideClassificationInput) (movement.Classification, error) {
	if input.Kind != movement.KindIncome && input.Kind != movement.KindNotIncome {
		return movement.Classification{}, domain.ValidationErrors{
			{Field: "kind", Msg: "debe ser INCOME o NOT_INCOME"},
		}
	}

	m, err := uc.Movements.Get(ctx, input.MovementID)
	if err != nil {
		return movement.Classification{}, err
	}

	now := uc.Clock.Now()
	operator := input.Operator
	if operator == "" {
		operator = "accountant"
	}

	c := movement.Classification{
		MovementID:   m.ID,
		Kind:         input.Kind,
		Confidence:   movement.ConfidenceHigh,
		Rule:         movement.RuleManualOverride,
		OverriddenBy: operator,
		OverriddenAt: &now,
		Note:         input.Note,
	}

	if err := uc.Classifications.Save(ctx, c); err != nil {
		return movement.Classification{}, err
	}

	p := period.FromTime(m.OccurredAt)
	if uc.Notifier != nil {
		_ = uc.Notifier.NotifyClosingInputsChanged(ctx, ports.ClosingInputsChanged{
			ClientID: m.ClientID,
			Period:   p,
		})
	}

	return c, nil
}
