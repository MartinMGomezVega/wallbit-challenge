// Package usecase implementa los casos de uso de la aplicación (architecture.md §2.2).
package usecase

import (
	"context"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/parameters"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
)

// GetParameters obtiene los parámetros vigentes del contador.
type GetParameters struct {
	Repo ports.ParametersRepository
}

// Execute ejecuta el caso de uso.
func (uc *GetParameters) Execute(ctx context.Context) (parameters.AccountantParameters, error) {
	return uc.Repo.Get(ctx)
}

// UpdateParametersResult contiene los parámetros actualizados y la cantidad de cierres recalculados.
type UpdateParametersResult struct {
	Parameters   parameters.AccountantParameters
	Recalculated int
}

// UpdateParameters valida y actualiza los parámetros globales del contador.
type UpdateParameters struct {
	Repo         ports.ParametersRepository
	Clock        ports.Clock
	RecalcPeriod *RecalcPeriod
}

// Execute valida los parámetros, sella fecha y usuario, persiste y recalcula cierres no aprobados.
func (uc *UpdateParameters) Execute(ctx context.Context, p parameters.AccountantParameters) (UpdateParametersResult, error) {
	if err := p.Validate(); err != nil {
		return UpdateParametersResult{}, err
	}
	p.UpdatedAt = uc.Clock.Now()
	p.UpdatedBy = "accountant"

	if err := uc.Repo.Save(ctx, p); err != nil {
		return UpdateParametersResult{}, err
	}

	recalculated := 0
	if uc.RecalcPeriod != nil {
		// Último período cerrado respecto a clock.Now()
		lastClosed := period.FromTime(uc.Clock.Now()).Prev()
		n, err := uc.RecalcPeriod.Execute(ctx, lastClosed)
		if err == nil {
			recalculated = n
		}
	}

	return UpdateParametersResult{
		Parameters:   p,
		Recalculated: recalculated,
	}, nil
}
