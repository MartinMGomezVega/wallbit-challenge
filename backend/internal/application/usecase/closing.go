// Package usecase implementa los casos de uso de la aplicación según Clean Architecture
// (architecture.md §2.2).
package usecase

import (
	"context"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/clientinput"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/closing"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/invoice"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/movement"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/reconciliation"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/signal"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/tax"
)

// ComputeStatus indica el resultado del cálculo de un cierre respecto a su estado previo.
type ComputeStatus string

const (
	ComputeStatusCreated   ComputeStatus = "CREATED"
	ComputeStatusUpdated   ComputeStatus = "UPDATED"
	ComputeStatusUnchanged ComputeStatus = "UNCHANGED"
)

// ComputeResult agrupa el cierre resultante y su estado de actualización.
type ComputeResult struct {
	Closing closing.Closing
	Status  ComputeStatus
}

// ComputeClosing orquesta el cálculo integral del cierre mensual para un cliente (spec 007 criterio 1).
type ComputeClosing struct {
	Clients          ports.ClientRepository
	Movements        ports.MovementRepository
	Classifications  ports.ClassificationRepository
	Invoices         ports.InvoiceRepository
	PurchaseInvoices ports.PurchaseInvoiceRepository
	ClientInputs     ports.ClientInputRepository
	Parameters       ports.ParametersRepository
	Closings         ports.ClosingRepository
	TaxScales        ports.TaxScaleProvider
	DueDates         ports.DueDateProvider
	ExchangeRates    ports.ExchangeRateProvider
	Clock            ports.Clock
	IDGen            ports.IDGenerator
	Tx               ports.TxManager
}

// Execute liquida el mes para el cliente indicado según el flujo de la spec 007.
func (uc *ComputeClosing) Execute(ctx context.Context, clientID string, p period.Period) (ComputeResult, error) {
	var res ComputeResult
	err := withinTx(ctx, uc.Tx, func(ctx context.Context) error {
		var err error
		res, err = uc.execute(ctx, clientID, p)
		return err
	})
	return res, err
}

func (uc *ComputeClosing) execute(ctx context.Context, clientID string, p period.Period) (ComputeResult, error) {
	c, err := uc.Clients.Get(ctx, clientID)
	if err != nil {
		return ComputeResult{}, err
	}

	now := uc.Clock.Now()
	periodStart := p.Start()
	periodEnd := p.End()

	params, err := uc.Parameters.Get(ctx)
	if err != nil {
		return ComputeResult{}, err
	}

	input, err := uc.ClientInputs.Get(ctx, clientID, p)
	if err != nil {
		input = clientinput.ClientPeriodInput{
			ClientID: clientID,
			Period:   p,
		}
	}

	// 1. Cargar facturas del cliente
	allInvoices, err := uc.Invoices.ByClient(ctx, clientID)
	if err != nil {
		return ComputeResult{}, err
	}

	// knownPayers = receptores de facturas anteriores al período
	knownPayers := make(map[string]bool)
	for _, inv := range allInvoices {
		if inv.IssueDate.Before(periodStart) {
			norm := movement.NormalizeName(inv.ReceiverName)
			if norm != "" {
				knownPayers[norm] = true
			}
		}
	}

	// Pesificar facturas hasta periodEnd
	quoteLookup := func(day time.Time) (invoice.Quote, error) {
		q, err := uc.ExchangeRates.USDQuote(ctx, day)
		if err != nil {
			return invoice.Quote{}, err
		}
		return invoice.Quote{Buy: q.Buy, Sell: q.Sell}, nil
	}

	var pricedInvoices []tax.PricedInvoice
	var periodInvoices []invoice.Invoice

	for _, inv := range allInvoices {
		if inv.IssueDate.After(periodEnd) {
			continue
		}
		ars, err := invoice.AmountARS(inv, params.FXRule, quoteLookup)
		if err != nil {
			return ComputeResult{}, err
		}
		var usd *money.Money
		if inv.Currency == money.USD {
			usd = &inv.TotalAmount
		}
		pi := tax.PricedInvoice{
			Invoice:   inv,
			AmountARS: ars,
			AmountUSD: usd,
		}
		pricedInvoices = append(pricedInvoices, pi)

		if period.FromTime(inv.IssueDate) == p {
			periodInvoices = append(periodInvoices, inv)
		}
	}

	// 2. Cargar y clasificar movimientos
	allMovements, err := uc.Movements.ByClient(ctx, clientID)
	if err != nil {
		return ComputeResult{}, err
	}

	var allMovIDs []string
	for _, m := range allMovements {
		if !m.OccurredAt.After(periodEnd) {
			allMovIDs = append(allMovIDs, m.ID)
		}
	}

	classMap, err := uc.Classifications.ByMovementIDs(ctx, allMovIDs)
	if err != nil {
		classMap = make(map[string]movement.Classification)
	}

	classifyMov := func(m movement.Movement) movement.Classification {
		if cl, ok := classMap[m.ID]; ok && cl.Rule == movement.RuleManualOverride {
			return cl
		}
		return movement.Classify(m, knownPayers)
	}

	var periodMovements []movement.Movement
	var classifiedPeriodMovements []signal.ClassifiedMovement
	var periodIncomeMovements []movement.Movement
	var inboundMovementsARS []tax.DatedAmount
	var cardSpendARS []tax.DatedAmount
	var historicIncomeUSD []money.Money

	histWindowStart := periodStart.AddDate(0, -params.AtypicalWindowMonths, 0)

	for _, m := range allMovements {
		if m.OccurredAt.After(periodEnd) {
			continue
		}

		class := classifyMov(m)
		inPeriod := !m.OccurredAt.Before(periodStart) && !m.OccurredAt.After(periodEnd)

		if inPeriod {
			periodMovements = append(periodMovements, m)
			classifiedPeriodMovements = append(classifiedPeriodMovements, signal.ClassifiedMovement{
				Movement:       m,
				Classification: class,
			})
			if class.Kind == movement.KindIncome {
				periodIncomeMovements = append(periodIncomeMovements, m)
			}
		}

		// Movimientos históricos para cobros atípicos
		if m.OccurredAt.After(histWindowStart) && m.OccurredAt.Before(periodStart) {
			if class.Kind == movement.KindIncome {
				historicIncomeUSD = append(historicIncomeUSD, m.Amount)
			}
		}

		// Depósitos y gastos con tarjeta para Monotributo (en ventana de 12 meses móviles)
		if m.Type == movement.TypeInboundTransfer && !m.Counterparty.IsOwnAccount {
			ars, err := money.ConvertUSDToARS(m.Amount, m.FXRate)
			if err == nil {
				inboundMovementsARS = append(inboundMovementsARS, tax.DatedAmount{
					Day:    m.OccurredAt,
					Amount: ars,
				})
			}
		}
		if m.Type == movement.TypeCardSpend {
			ars, err := money.ConvertUSDToARS(m.Amount, m.FXRate)
			if err == nil {
				cardSpendARS = append(cardSpendARS, tax.DatedAmount{
					Day:    m.OccurredAt,
					Amount: ars,
				})
			}
		}
		if m.Type == movement.TypeCardRefund {
			ars, err := money.ConvertUSDToARS(m.Amount, m.FXRate)
			if err == nil {
				cardSpendARS = append(cardSpendARS, tax.DatedAmount{
					Day:    m.OccurredAt,
					Amount: money.New(-ars.Amount, money.ARS),
				})
			}
		}
	}

	// 3. Conciliación
	recResult := reconciliation.Reconcile(
		periodIncomeMovements,
		periodInvoices,
		reconciliation.Params{
			ToleranceBp:      params.ReconciliationToleranceBp,
			WindowBeforeDays: params.ReconciliationWindowBeforeDays,
			WindowAfterDays:  params.ReconciliationWindowAfterDays,
		},
	)

	matchedPairs := make([]closing.MatchedPair, 0, len(recResult.Matched))
	for _, pair := range recResult.Matched {
		matchedPairs = append(matchedPairs, closing.MatchedPair{
			MovementID: pair.Movement.ID,
			InvoiceID:  pair.Invoice.ID,
			DiffAmount: pair.DiffAmount,
		})
	}
	unmatchedIncomeIDs := make([]string, 0, len(recResult.UnmatchedIncome))
	for _, m := range recResult.UnmatchedIncome {
		unmatchedIncomeIDs = append(unmatchedIncomeIDs, m.ID)
	}
	unmatchedInvoiceIDs := make([]string, 0, len(recResult.UnmatchedInvoices))
	for _, inv := range recResult.UnmatchedInvoices {
		unmatchedInvoiceIDs = append(unmatchedInvoiceIDs, inv.ID)
	}
	recData := closing.Reconciliation{
		MatchedPairs:        matchedPairs,
		UnmatchedIncomeIDs:  unmatchedIncomeIDs,
		UnmatchedInvoiceIDs: unmatchedInvoiceIDs,
	}

	// 4. Liquidación según régimen al último día del período
	dueDates, err := uc.DueDates.DueDates(ctx, p, c.CUIT)
	if err != nil {
		return ComputeResult{}, err
	}

	regime := c.RegimeAt(periodEnd)
	var monoAss *tax.MonotributoAssessment
	var riAss *tax.RIAssessment

	if regime == client.RegimeMonotributo {
		cat := c.MonotributoCategory
		if cat == "" {
			cat = tax.CatA
		}
		ma, err := tax.AssessMonotributo(tax.MonotributoInput{
			Period:          p,
			CurrentCategory: cat,
			Invoices:        pricedInvoices,
			ScaleAt: func(t time.Time) tax.Scale {
				sc, _ := uc.TaxScales.MonotributoScale(ctx, t)
				return sc
			},
			FeeDueDate:          dueDates.Monotributo,
			Params:              params,
			InboundMovementsARS: inboundMovementsARS,
			CardSpendARS:        cardSpendARS,
		})
		if err != nil {
			return ComputeResult{}, err
		}
		monoAss = &ma
	} else {
		purchases, err := uc.PurchaseInvoices.ByClientBetween(ctx, clientID, periodStart, periodEnd)
		if err != nil {
			purchases = nil
		}
		profile := tax.RIProfile{}
		if c.RIProfile != nil {
			profile = *c.RIProfile
		}
		ra, err := tax.AssessRI(tax.RIInput{
			Period:           p,
			Invoices:         pricedInvoices,
			PurchaseInvoices: purchases,
			Profile:          profile,
			DueDates: tax.RIDueDates{
				VatDueDate:          dueDates.VAT,
				SelfEmployedDueDate: dueDates.SelfEmployed,
			},
			Params: params,
		})
		if err != nil {
			return ComputeResult{}, err
		}
		riAss = &ra
	}

	// 5. Señales de riesgo
	var cInvoicesAfterRI []signal.CInvoiceAfterRI
	if regime == client.RegimeRI {
		riFrom, ok := c.RIFrom()
		if ok {
			for _, pi := range pricedInvoices {
				if period.FromTime(pi.Invoice.IssueDate) == p && pi.Invoice.Type == invoice.TypeC {
					if !pi.Invoice.IssueDate.Before(riFrom) {
						cInvoicesAfterRI = append(cInvoicesAfterRI, signal.CInvoiceAfterRI{
							Invoice:   pi.Invoice,
							AmountARS: pi.AmountARS,
						})
					}
				}
			}
		}
	}

	evalRes := signal.Evaluate(signal.Input{
		Regime:            regime,
		Mono:              monoAss,
		RI:                riAss,
		Reconciliation:    recResult,
		PeriodMovements:   classifiedPeriodMovements,
		HistoricIncomeUSD: historicIncomeUSD,
		ClientInput:       input,
		CInvoicesAfterRI:  cInvoicesAfterRI,
		Params:            params,
	})

	// 6. Hash canónico y persistencia
	hash := closing.ComputeHash(monoAss, riAss, evalRes.Signals)

	existing, found, err := uc.Closings.ByClientPeriod(ctx, clientID, p)
	if err != nil {
		return ComputeResult{}, err
	}

	if !found {
		cloID := uc.IDGen.New("clo")
		newClo := closing.NewClosing(
			cloID,
			clientID,
			p,
			input.IsConfirmed(),
			hash,
			evalRes.Severity,
			evalRes.Signals,
			evalRes.Description,
			monoAss,
			riAss,
			recData,
			now,
		)
		if err := uc.Closings.Save(ctx, newClo); err != nil {
			return ComputeResult{}, err
		}
		return ComputeResult{Closing: newClo, Status: ComputeStatusCreated}, nil
	}

	// Si existía, evaluar si cambió el resultado o si el cliente confirmó el input
	statusChanged := false
	if input.IsConfirmed() && existing.Status == closing.StatusPendingClientInput {
		_ = existing.MarkReady()
		statusChanged = true
	}

	hashChanged := existing.ApplyResult(
		hash,
		evalRes.Severity,
		evalRes.Signals,
		evalRes.Description,
		monoAss,
		riAss,
		recData,
		now,
	)

	if hashChanged || statusChanged {
		if err := uc.Closings.Save(ctx, existing); err != nil {
			return ComputeResult{}, err
		}
		return ComputeResult{Closing: existing, Status: ComputeStatusUpdated}, nil
	}

	// Sin cambios
	_ = uc.Closings.Save(ctx, existing)
	return ComputeResult{Closing: existing, Status: ComputeStatusUnchanged}, nil
}
