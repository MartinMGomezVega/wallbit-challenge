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
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/notification"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/parameters"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/reconciliation"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/signal"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/tax"
)

// ListClosingsQuery encapsula la petición de listado de cierres.
type ListClosingsQuery struct {
	Period   period.Period
	Filter   closing.Filter
	Sort     closing.Sort
	Page     int
	PageSize int
}

// ListClosingsResult es la respuesta paginada con filas de cierres.
type ListClosingsResult struct {
	Items      []closing.Row `json:"items"`
	Total      int           `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"pageSize"`
	TotalPages int           `json:"totalPages"`
}

// ListClosingIDsQuery encapsula la petición de listado de IDs de cierres (spec 009 criterio 1).
type ListClosingIDsQuery struct {
	Period period.Period
	Filter closing.Filter
	Sort   closing.Sort
}

// ListClosingIDsResult contiene los IDs ordenados y el total de cierres que cumplen el filtro.
type ListClosingIDsResult struct {
	IDs   []string `json:"ids"`
	Total int      `json:"total"`
}

func buildClosingRows(ctx context.Context, closingsRepo ports.ClosingRepository, clientsRepo ports.ClientRepository, p period.Period) ([]closing.Row, error) {
	closingsList, err := closingsRepo.ListByPeriod(ctx, p)
	if err != nil {
		return nil, err
	}

	clientsList, err := clientsRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	clientMap := make(map[string]client.Client, len(clientsList))
	for _, c := range clientsList {
		clientMap[c.ID] = c
	}

	rows := make([]closing.Row, 0, len(closingsList))
	for _, c := range closingsList {
		cli := clientMap[c.ClientID]
		rows = append(rows, buildRow(c, cli))
	}
	return rows, nil
}

// ListClosings construye la vista tabular de los cierres mensuales para el contador (spec 007 criterio 14).
type ListClosings struct {
	Closings ports.ClosingRepository
	Clients  ports.ClientRepository
}

// Execute obtiene los cierres, los enriquece con datos del cliente y aplica filtros, ordenamiento y paginación.
func (uc *ListClosings) Execute(ctx context.Context, q ListClosingsQuery) (ListClosingsResult, error) {
	rows, err := buildClosingRows(ctx, uc.Closings, uc.Clients, q.Period)
	if err != nil {
		return ListClosingsResult{}, err
	}

	filtered := closing.Apply(rows, q.Filter, q.Sort)

	total := len(filtered)
	pageSize := q.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 500 {
		pageSize = 500
	}

	page := q.Page
	if page <= 0 {
		page = 1
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	items := filtered[start:end]
	if items == nil {
		items = []closing.Row{}
	}

	return ListClosingsResult{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// ListClosingIDs devuelve únicamente los IDs de los cierres filtrados y ordenados, sin paginación (spec 009 criterio 1).
type ListClosingIDs struct {
	Closings ports.ClosingRepository
	Clients  ports.ClientRepository
}

// Execute obtiene los cierres, los filtra y ordena según los criterios indicados, y devuelve la lista completa de IDs.
func (uc *ListClosingIDs) Execute(ctx context.Context, q ListClosingIDsQuery) (ListClosingIDsResult, error) {
	rows, err := buildClosingRows(ctx, uc.Closings, uc.Clients, q.Period)
	if err != nil {
		return ListClosingIDsResult{}, err
	}

	filtered := closing.Apply(rows, q.Filter, q.Sort)
	ids := make([]string, len(filtered))
	for i, r := range filtered {
		ids[i] = r.ClosingID
	}
	if ids == nil {
		ids = []string{}
	}

	return ListClosingIDsResult{
		IDs:   ids,
		Total: len(ids),
	}, nil
}

// ClosingsSummary calcula los totales y agrupadores para los chips de la UI (spec 007 criterio 15).
type ClosingsSummary struct {
	Closings ports.ClosingRepository
	Clients  ports.ClientRepository
}

// Execute aplica los filtros indicados excepto severidad y calcula el resumen.
func (uc *ClosingsSummary) Execute(ctx context.Context, p period.Period, filter closing.Filter) (closing.Summary, error) {
	closingsList, err := uc.Closings.ListByPeriod(ctx, p)
	if err != nil {
		return closing.Summary{}, err
	}

	clientsList, err := uc.Clients.List(ctx)
	if err != nil {
		return closing.Summary{}, err
	}

	clientMap := make(map[string]client.Client, len(clientsList))
	for _, c := range clientsList {
		clientMap[c.ID] = c
	}

	rows := make([]closing.Row, 0, len(closingsList))
	for _, c := range closingsList {
		cli := clientMap[c.ClientID]
		rows = append(rows, buildRow(c, cli))
	}

	// Filtros sin severidad para no sesgar los contadores de los chips de semáforo
	noSevFilter := filter
	noSevFilter.Severities = nil

	filtered := closing.Apply(rows, noSevFilter, closing.Sort{})
	return closing.ComputeSummary(filtered), nil
}

func buildRow(c closing.Closing, cli client.Client) closing.Row {
	var cat *tax.Category
	if c.Mono != nil {
		catCopy := c.Mono.CurrentCategory
		cat = &catCopy
	} else if cli.Regime == client.RegimeMonotributo && cli.MonotributoCategory != "" {
		catCopy := cli.MonotributoCategory
		cat = &catCopy
	}

	var primary *closing.SignalSummary
	var others []closing.SignalSummary
	if len(c.Signals) > 0 {
		first := c.Signals[0]
		label := string(first.Code)
		if meta, ok := signal.MetaFor(first.Code); ok {
			label = meta.Label
		}
		primary = &closing.SignalSummary{
			Code:     first.Code,
			Label:    label,
			Severity: first.Severity.String(),
		}
		for _, s := range c.Signals[1:] {
			sLabel := string(s.Code)
			if meta, ok := signal.MetaFor(s.Code); ok {
				sLabel = meta.Label
			}
			others = append(others, closing.SignalSummary{
				Code:     s.Code,
				Label:    sLabel,
				Severity: s.Severity.String(),
			})
		}
	}
	if others == nil {
		others = []closing.SignalSummary{}
	}

	var billed closing.BilledInPeriod
	var accum closing.Accumulated12m
	var toPay closing.ToPay

	if c.Mono != nil {
		billed = closing.BilledInPeriod{
			USD: c.Mono.BilledInPeriodUSD,
			ARS: c.Mono.BilledInPeriodARS,
		}
		accum = closing.Accumulated12m{
			ARS:         c.Mono.Billed12mARS,
			PctOfKCapBp: c.Mono.PctOfKCapBp,
		}
		toPay = closing.ToPay{
			Amount:  c.Mono.Fee,
			DueDate: c.Mono.FeeDueDate,
		}
	} else if c.RI != nil {
		billed = closing.BilledInPeriod{
			USD: c.RI.BilledInPeriodUSD,
			ARS: c.RI.BilledInPeriodARS,
		}
		accum = closing.Accumulated12m{
			ARS:         money.Zero(money.ARS),
			PctOfKCapBp: 0,
		}
		minDue := c.RI.VatDueDate
		if !c.RI.SelfEmployedDueDate.IsZero() && (minDue.IsZero() || c.RI.SelfEmployedDueDate.Before(minDue)) {
			minDue = c.RI.SelfEmployedDueDate
		}
		toPay = closing.ToPay{
			Amount:  c.RI.TotalToPay,
			DueDate: minDue,
		}
	}

	return closing.Row{
		ClosingID:      c.ID,
		ClientID:       c.ClientID,
		ClientName:     cli.Name,
		CUIT:           cli.CUIT.String(),
		Regime:         cli.RegimeAt(c.Period.End()),
		Category:       cat,
		Severity:       c.Severity,
		PrimarySignal:  primary,
		Description:    c.Description,
		OtherSignals:   others,
		BilledInPeriod: billed,
		Accumulated12m: accum,
		ToPay:          toPay,
		Status:         c.Status,
		Version:        c.Version,
		PaidAt:         c.PaidAt,
	}
}

// PeriodMovementDTO modela un movimiento del período con su clasificación.
type PeriodMovementDTO struct {
	Movement       movement.Movement       `json:"movement"`
	Classification movement.Classification `json:"classification"`
}

// PeriodInvoiceDTO modela una factura del período con su importe en ARS y el TC utilizado.
type PeriodInvoiceDTO struct {
	Invoice    invoice.Invoice `json:"invoice"`
	AmountARS  money.Money     `json:"amountArs"`
	FXRateUsed *money.Rate     `json:"fxRateUsed,omitempty"`
}

// ReconciliationDetailDTO detalla los pares conciliados y los ítems pendientes.
type ReconciliationDetailDTO struct {
	MatchedPairs      []reconciliation.Pair `json:"matchedPairs"`
	UnmatchedIncome   []movement.Movement   `json:"unmatchedIncome"`
	UnmatchedInvoices []invoice.Invoice     `json:"unmatchedInvoices"`
}

// ClosingDetailDTO agrupa toda la información detallada para el modal de revisión (spec 007 criterio 16).
type ClosingDetailDTO struct {
	Closing         closing.Closing               `json:"closing"`
	Client          client.Client                 `json:"client"`
	PeriodMovements []PeriodMovementDTO           `json:"periodMovements"`
	PeriodInvoices  []PeriodInvoiceDTO            `json:"periodInvoices"`
	Reconciliation  ReconciliationDetailDTO       `json:"reconciliation"`
	ClientInput     clientinput.ClientPeriodInput `json:"clientInput"`
	DraftMessage    notification.Message          `json:"draftMessage"`
}

// GetClosingDetail recupera toda la información atómica que compone un cierre mensual.
type GetClosingDetail struct {
	Closings        ports.ClosingRepository
	Clients         ports.ClientRepository
	Movements       ports.MovementRepository
	Classifications ports.ClassificationRepository
	Invoices        ports.InvoiceRepository
	ClientInputs    ports.ClientInputRepository
	Parameters      ports.ParametersRepository
	ExchangeRates   ports.ExchangeRateProvider
}

// Execute obtiene el detalle completo del cierre dado por ID.
func (uc *GetClosingDetail) Execute(ctx context.Context, id string) (ClosingDetailDTO, error) {
	c, err := uc.Closings.Get(ctx, id)
	if err != nil {
		return ClosingDetailDTO{}, err
	}

	cli, err := uc.Clients.Get(ctx, c.ClientID)
	if err != nil {
		return ClosingDetailDTO{}, err
	}

	p := c.Period
	periodStart := p.Start()
	periodEnd := p.End()

	input, err := uc.ClientInputs.Get(ctx, c.ClientID, p)
	if err != nil {
		input = clientinput.ClientPeriodInput{
			ClientID: c.ClientID,
			Period:   p,
		}
	}

	params, err := uc.Parameters.Get(ctx)
	if err != nil {
		params = parameters.Defaults()
	}

	allInvoices, err := uc.Invoices.ByClient(ctx, c.ClientID)
	if err != nil {
		allInvoices = nil
	}

	knownPayers := make(map[string]bool)
	for _, inv := range allInvoices {
		if inv.IssueDate.Before(periodStart) {
			norm := movement.NormalizeName(inv.ReceiverName)
			if norm != "" {
				knownPayers[norm] = true
			}
		}
	}

	quoteLookup := func(day time.Time) (invoice.Quote, error) {
		q, err := uc.ExchangeRates.USDQuote(ctx, day)
		if err != nil {
			return invoice.Quote{}, err
		}
		return invoice.Quote{Buy: q.Buy, Sell: q.Sell}, nil
	}

	var periodInvoicesDTO []PeriodInvoiceDTO
	invMap := make(map[string]invoice.Invoice)

	for _, inv := range allInvoices {
		invMap[inv.ID] = inv
		if period.FromTime(inv.IssueDate) == p {
			ars, err := invoice.AmountARS(inv, params.FXRule, quoteLookup)
			if err != nil {
				ars = inv.TotalAmount
			}
			rateUsed := inv.FXRate
			if rateUsed == nil && inv.Currency == money.USD {
				if q, err := quoteLookup(inv.IssueDate); err == nil {
					r := q.Buy
					rateUsed = &r
				}
			}
			periodInvoicesDTO = append(periodInvoicesDTO, PeriodInvoiceDTO{
				Invoice:    inv,
				AmountARS:  ars,
				FXRateUsed: rateUsed,
			})
		}
	}

	allMovements, err := uc.Movements.ByClient(ctx, c.ClientID)
	if err != nil {
		allMovements = nil
	}

	var periodMovIDs []string
	movMap := make(map[string]movement.Movement)
	for _, m := range allMovements {
		movMap[m.ID] = m
		if !m.OccurredAt.Before(periodStart) && !m.OccurredAt.After(periodEnd) {
			periodMovIDs = append(periodMovIDs, m.ID)
		}
	}

	classMap, _ := uc.Classifications.ByMovementIDs(ctx, periodMovIDs)
	if classMap == nil {
		classMap = make(map[string]movement.Classification)
	}

	var periodMovementsDTO []PeriodMovementDTO
	for _, m := range allMovements {
		if !m.OccurredAt.Before(periodStart) && !m.OccurredAt.After(periodEnd) {
			class, ok := classMap[m.ID]
			if !ok || class.Rule != movement.RuleManualOverride {
				class = movement.Classify(m, knownPayers)
			}
			periodMovementsDTO = append(periodMovementsDTO, PeriodMovementDTO{
				Movement:       m,
				Classification: class,
			})
		}
	}

	// Reconstruir detalle de conciliación a partir de los IDs registrados en Closing
	matchedPairs := make([]reconciliation.Pair, 0, len(c.Reconciliation.MatchedPairs))
	for _, pair := range c.Reconciliation.MatchedPairs {
		m, okM := movMap[pair.MovementID]
		inv, okI := invMap[pair.InvoiceID]
		if okM && okI {
			matchedPairs = append(matchedPairs, reconciliation.Pair{
				Movement:   m,
				Invoice:    inv,
				DiffAmount: pair.DiffAmount,
			})
		}
	}

	unmatchedIncome := make([]movement.Movement, 0, len(c.Reconciliation.UnmatchedIncomeIDs))
	for _, id := range c.Reconciliation.UnmatchedIncomeIDs {
		if m, ok := movMap[id]; ok {
			unmatchedIncome = append(unmatchedIncome, m)
		}
	}

	unmatchedInvoices := make([]invoice.Invoice, 0, len(c.Reconciliation.UnmatchedInvoiceIDs))
	for _, id := range c.Reconciliation.UnmatchedInvoiceIDs {
		if inv, ok := invMap[id]; ok {
			unmatchedInvoices = append(unmatchedInvoices, inv)
		}
	}

	recDetail := ReconciliationDetailDTO{
		MatchedPairs:      matchedPairs,
		UnmatchedIncome:   unmatchedIncome,
		UnmatchedInvoices: unmatchedInvoices,
	}

	draft := notification.BuildClosingMessage(c, cli)

	return ClosingDetailDTO{
		Closing:         c,
		Client:          cli,
		PeriodMovements: periodMovementsDTO,
		PeriodInvoices:  periodInvoicesDTO,
		Reconciliation:  recDetail,
		ClientInput:     input,
		DraftMessage:    draft,
	}, nil
}
