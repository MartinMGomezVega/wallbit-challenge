package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/tax"
)

// ChangeRegimeInput define los parámetros para cambiar el régimen de un cliente.
type ChangeRegimeInput struct {
	ClientID      string
	To            client.Regime
	EffectiveDate time.Time
	Note          string
	Force         bool
}

// ChangeRegime ejecuta la transición de régimen (Monotributo -> RI) de un cliente (spec 005 criterio 5).
type ChangeRegime struct {
	Clients    ports.ClientRepository
	Exclusions ports.LatestExclusionDateReader
	Clock      ports.Clock
	Notifier   ports.ChangeNotifier
}

func (uc *ChangeRegime) Execute(ctx context.Context, in ChangeRegimeInput) (client.Client, error) {
	c, err := uc.Clients.Get(ctx, in.ClientID)
	if err != nil {
		return client.Client{}, err
	}

	loc := period.Location()
	effAR := in.EffectiveDate.In(loc)
	effDayStr := effAR.Format("2006-01-02")

	// Si hay lector de fecha de exclusión y no está forzado, verificar coincidencia
	if uc.Exclusions != nil {
		latestExcl, err := uc.Exclusions.LatestExclusionDate(ctx, in.ClientID)
		if err != nil {
			return client.Client{}, err
		}
		if latestExcl != nil {
			exclAR := latestExcl.In(loc)
			exclDayStr := exclAR.Format("2006-01-02")
			if exclDayStr != effDayStr {
				if !in.Force {
					return client.Client{}, domain.NewError(
						domain.ErrValidation,
						"EFFECTIVE_DATE_MISMATCH",
						fmt.Sprintf("la fecha efectiva (%s) no coincide con la fecha de exclusión (%s)", effDayStr, exclDayStr),
					)
				}
				// Si está forzado, registrar en la nota
				forcedTag := fmt.Sprintf("[Forzado por contador; exclusión sugerida: %s]", exclDayStr)
				if in.Note != "" {
					in.Note = fmt.Sprintf("%s %s", in.Note, forcedTag)
				} else {
					in.Note = forcedTag
				}
			}
		}
	}

	now := uc.Clock.Now()

	if err := c.ChangeRegime(in.To, in.EffectiveDate, "EXCLUSION_K_CAP", in.Note, now); err != nil {
		return client.Client{}, err
	}

	if err := uc.Clients.Save(ctx, c); err != nil {
		return client.Client{}, err
	}

	if uc.Notifier != nil {
		_ = uc.Notifier.NotifyClosingInputsChanged(ctx, ports.ClosingInputsChanged{
			ClientID: c.ID,
			Period:   period.FromTime(in.EffectiveDate),
		})
	}

	return c, nil
}

// UpdateRIProfileInput define los parámetros para actualizar la configuración de un RI.
type UpdateRIProfileInput struct {
	ClientID                        string
	SelfEmployedMonthlyContribution money.Money
	PreviousVatCredit               money.Money
	IncomeTaxAdvances               []tax.IncomeTaxAdvance
}

// UpdateRIProfile actualiza la configuración impositiva de un Responsable Inscripto (spec 005 criterio 6).
type UpdateRIProfile struct {
	Clients  ports.ClientRepository
	Clock    ports.Clock
	Notifier ports.ChangeNotifier
}

func (uc *UpdateRIProfile) Execute(ctx context.Context, in UpdateRIProfileInput) (tax.RIProfile, error) {
	c, err := uc.Clients.Get(ctx, in.ClientID)
	if err != nil {
		return tax.RIProfile{}, err
	}

	var ves domain.ValidationErrors

	if in.SelfEmployedMonthlyContribution.Amount < 0 {
		ves = append(ves, domain.ValidationError{
			Field: "selfEmployedMonthlyContribution",
			Msg:   "el aporte mensual a autónomos no puede ser negativo",
		})
	}
	if in.SelfEmployedMonthlyContribution.Currency != "" && in.SelfEmployedMonthlyContribution.Currency != money.ARS {
		ves = append(ves, domain.ValidationError{
			Field: "selfEmployedMonthlyContribution",
			Msg:   "el aporte a autónomos debe estar expresado en ARS",
		})
	}

	if in.PreviousVatCredit.Amount < 0 {
		ves = append(ves, domain.ValidationError{
			Field: "previousVatCredit",
			Msg:   "el saldo a favor de IVA previo no puede ser negativo",
		})
	}
	if in.PreviousVatCredit.Currency != "" && in.PreviousVatCredit.Currency != money.ARS {
		ves = append(ves, domain.ValidationError{
			Field: "previousVatCredit",
			Msg:   "el saldo a favor previo debe estar expresado en ARS",
		})
	}

	for i, adv := range in.IncomeTaxAdvances {
		if adv.Amount.Amount < 0 {
			ves = append(ves, domain.ValidationError{
				Field: fmt.Sprintf("incomeTaxAdvances[%d].amount", i),
				Msg:   "el importe del anticipo no puede ser negativo",
			})
		}
		if adv.Amount.Currency != "" && adv.Amount.Currency != money.ARS {
			ves = append(ves, domain.ValidationError{
				Field: fmt.Sprintf("incomeTaxAdvances[%d].amount", i),
				Msg:   "el importe del anticipo debe ser en ARS",
			})
		}
		if adv.Period.Year <= 0 || adv.Period.Month <= 0 {
			ves = append(ves, domain.ValidationError{
				Field: fmt.Sprintf("incomeTaxAdvances[%d].period", i),
				Msg:   "período de anticipo inválido",
			})
		}
	}

	if len(ves) > 0 {
		return tax.RIProfile{}, ves
	}

	// Normalizar monedas
	if in.SelfEmployedMonthlyContribution.Currency == "" {
		in.SelfEmployedMonthlyContribution.Currency = money.ARS
	}
	if in.PreviousVatCredit.Currency == "" {
		in.PreviousVatCredit.Currency = money.ARS
	}
	for i := range in.IncomeTaxAdvances {
		if in.IncomeTaxAdvances[i].Amount.Currency == "" {
			in.IncomeTaxAdvances[i].Amount.Currency = money.ARS
		}
	}

	profile := tax.RIProfile{
		SelfEmployedMonthlyContribution: in.SelfEmployedMonthlyContribution,
		PreviousVatCredit:               in.PreviousVatCredit,
		IncomeTaxAdvances:               in.IncomeTaxAdvances,
	}

	c.RIProfile = &profile
	if err := uc.Clients.Save(ctx, c); err != nil {
		return tax.RIProfile{}, err
	}

	if uc.Notifier != nil {
		_ = uc.Notifier.NotifyClosingInputsChanged(ctx, ports.ClosingInputsChanged{
			ClientID: c.ID,
		})
	}

	return profile, nil
}
