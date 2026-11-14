package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports/fakes"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/usecase"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/clientinput"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/closing"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/invoice"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/movement"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/parameters"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/signal"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/tax"
)

type testHarness struct {
	clients          *fakes.ClientRepository
	movements        *fakes.MovementRepository
	classifications  *fakes.ClassificationRepository
	invoices         *fakes.InvoiceRepository
	purchaseInvoices *fakes.PurchaseInvoiceRepository
	clientInputs     *fakes.ClientInputRepository
	params           *fakes.ParametersRepository
	closings         *fakes.ClosingRepository
	taxScales        *fakes.TaxScaleProvider
	dueDates         *fakes.DueDateProvider
	fx               *fakes.ExchangeRateProvider
	clock            fakes.Clock
	idGen            *fakes.IDGenerator
	uc               *usecase.ComputeClosing
}

func newTestHarness(now time.Time) *testHarness {
	clients := fakes.NewClientRepository()
	movs := fakes.NewMovementRepository()
	classes := fakes.NewClassificationRepository()
	invs := fakes.NewInvoiceRepository()
	purchases := fakes.NewPurchaseInvoiceRepository()
	inputs := fakes.NewClientInputRepository()
	params := &fakes.ParametersRepository{Params: parameters.Defaults()}
	closings := fakes.NewClosingRepository()

	// Escala de Monotributo estándar de prueba
	scale := tax.Scale{
		ValidFrom: time.Date(2026, 8, 1, 0, 0, 0, 0, period.Location()),
		Rows: []tax.CategoryRow{
			{Category: tax.CatA, CapARS: money.New(1200941045, money.ARS), ServicesFeeARS: money.New(4952718, money.ARS)},
			{Category: tax.CatB, CapARS: money.New(1759518300, money.ARS), ServicesFeeARS: money.New(5637908, money.ARS)},
			{Category: tax.CatC, CapARS: money.New(2467049400, money.ARS), ServicesFeeARS: money.New(6602000, money.ARS)},
			{Category: tax.CatD, CapARS: money.New(3062865100, money.ARS), ServicesFeeARS: money.New(8461300, money.ARS)},
			{Category: tax.CatK, CapARS: money.New(12661083875, money.ARS), ServicesFeeARS: money.New(161444604, money.ARS)},
		},
	}
	scales := &fakes.TaxScaleProvider{Scale: scale}

	due := ports.DueDates{
		Monotributo:  time.Date(2026, 11, 20, 0, 0, 0, 0, period.Location()),
		VAT:          time.Date(2026, 11, 20, 0, 0, 0, 0, period.Location()),
		SelfEmployed: time.Date(2026, 11, 5, 0, 0, 0, 0, period.Location()),
	}
	dueDates := &fakes.DueDateProvider{Dates: due}

	fx := &fakes.ExchangeRateProvider{
		Default: ports.Quote{
			Day:  now,
			Buy:  150000, // 1500.00
			Sell: 155000,
		},
	}

	clock := fakes.Clock{Current: now}
	idGen := fakes.NewIDGenerator()

	uc := &usecase.ComputeClosing{
		Clients:          clients,
		Movements:        movs,
		Classifications:  classes,
		Invoices:         invs,
		PurchaseInvoices: purchases,
		ClientInputs:     inputs,
		Parameters:       params,
		Closings:         closings,
		TaxScales:        scales,
		DueDates:         dueDates,
		ExchangeRates:    fx,
		Clock:            clock,
		IDGen:            idGen,
	}

	return &testHarness{
		clients:          clients,
		movements:        movs,
		classifications:  classes,
		invoices:         invs,
		purchaseInvoices: purchases,
		clientInputs:     inputs,
		params:           params,
		closings:         closings,
		taxScales:        scales,
		dueDates:         dueDates,
		fx:               fx,
		clock:            clock,
		idGen:            idGen,
		uc:               uc,
	}
}

func TestComputeClosing_MonotributoVerde(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 11, 1, 9, 0, 0, 0, period.Location())
	h := newTestHarness(now)
	p := period.MustParse("2026-10")

	// Cliente Monotributo Cat A
	cuit, _ := client.ParseCUIT("20-11111111-2")
	cli := client.Client{
		ID:                  "cli_mono",
		Name:                "Martín Gómez",
		CUIT:                cuit,
		Regime:              client.RegimeMonotributo,
		MonotributoCategory: tax.CatA,
		CreatedAt:           time.Date(2025, 1, 1, 0, 0, 0, 0, period.Location()),
	}
	_ = h.clients.Save(ctx, cli)

	// Factura E en el período de USD 1.000
	rate := money.Rate(150000)
	invDate := time.Date(2026, 10, 15, 10, 0, 0, 0, period.Location())
	inv := invoice.Invoice{
		ID:           "inv_1",
		ClientID:     cli.ID,
		Type:         invoice.TypeE,
		PointOfSale:  1,
		Number:       1,
		IssueDate:    invDate,
		Currency:     money.USD,
		TotalAmount:  money.New(100000, money.USD), // US$ 1000.00
		ReceiverName: "Acme Corp",
		FXRate:       &rate,
	}
	_ = h.invoices.Save(ctx, inv)

	// Movimiento coincidente en el período (cobro de US$ 1000)
	mov := movement.Movement{
		ID:         "mov_1",
		ClientID:   cli.ID,
		Type:       movement.TypeInboundTransfer,
		Amount:     money.New(100000, money.USD),
		FXRate:     rate,
		OccurredAt: time.Date(2026, 10, 16, 15, 0, 0, 0, period.Location()),
		Counterparty: movement.Counterparty{
			Name:    "Acme Corp",
			Country: "US",
			Kind:    movement.CounterpartyCompany,
		},
	}
	_ = h.movements.Save(ctx, mov)

	// Input del cliente confirmado
	_ = h.clientInputs.Save(ctx, clientinput.ClientPeriodInput{
		ClientID:    cli.ID,
		Period:      p,
		ConfirmedAt: &now,
	})

	res, err := h.uc.Execute(ctx, cli.ID, p)
	if err != nil {
		t.Fatalf("error en Execute: %v", err)
	}

	if res.Status != usecase.ComputeStatusCreated {
		t.Fatalf("esperaba status CREATED, obtuvo %s", res.Status)
	}
	clo := res.Closing
	if clo.Status != closing.StatusReadyForReview {
		t.Fatalf("con input confirmado el estado debe ser READY_FOR_REVIEW, obtuvo %s", clo.Status)
	}
	if clo.Severity != signal.SeverityGreen {
		t.Fatalf("esperaba semáforo GREEN, obtuvo %s", clo.Severity)
	}
	if clo.Mono == nil {
		t.Fatal("liquidación Monotributo debe estar presente")
	}
	if clo.Mono.Excluded {
		t.Fatal("no debe estar excluido")
	}
	if len(clo.Reconciliation.MatchedPairs) != 1 {
		t.Fatalf("esperaba 1 par conciliado, obtuvo %d", len(clo.Reconciliation.MatchedPairs))
	}
}

func TestComputeClosing_MonotributoExclusion(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 11, 1, 9, 0, 0, 0, period.Location())
	h := newTestHarness(now)
	p := period.MustParse("2026-10")

	cuit, _ := client.ParseCUIT("20-22222222-3")
	cli := client.Client{
		ID:                  "cli_excl",
		Name:                "Excluido Test",
		CUIT:                cuit,
		Regime:              client.RegimeMonotributo,
		MonotributoCategory: tax.CatK,
		CreatedAt:           time.Date(2025, 1, 1, 0, 0, 0, 0, period.Location()),
	}
	_ = h.clients.Save(ctx, cli)

	// Factura gigante que supera los 126 millones del tope K
	rate := money.Rate(150000)
	inv := invoice.Invoice{
		ID:           "inv_big",
		ClientID:     cli.ID,
		Type:         invoice.TypeC,
		PointOfSale:  1,
		Number:       1,
		IssueDate:    time.Date(2026, 10, 10, 0, 0, 0, 0, period.Location()),
		Currency:     money.ARS,
		TotalAmount:  money.New(15000000000, money.ARS), // 150 millones ARS > 126M cap K
		ReceiverName: "Cliente Local",
		FXRate:       &rate,
	}
	_ = h.invoices.Save(ctx, inv)

	res, err := h.uc.Execute(ctx, cli.ID, p)
	if err != nil {
		t.Fatalf("error en Execute: %v", err)
	}

	clo := res.Closing
	if clo.Severity != signal.SeverityRed {
		t.Fatalf("esperaba semáforo RED por exclusión, obtuvo %s", clo.Severity)
	}
	if clo.Mono == nil || !clo.Mono.Excluded {
		t.Fatal("esperaba Mono.Excluded = true")
	}
	if clo.Mono.ExclusionDate == nil {
		t.Fatal("esperaba ExclusionDate != nil")
	}

	// Verificar LatestExclusionDate en el repo
	latestExcl, err := h.closings.LatestExclusionDate(ctx, cli.ID)
	if err != nil || latestExcl == nil {
		t.Fatalf("LatestExclusionDate debió registrar la fecha de exclusión: %v", err)
	}
}

func TestComputeClosing_RI(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 11, 1, 9, 0, 0, 0, period.Location())
	h := newTestHarness(now)
	p := period.MustParse("2026-10")

	cuit, _ := client.ParseCUIT("20-33333333-4")
	cli := client.Client{
		ID:        "cli_ri",
		Name:      "RI Contribuyente",
		CUIT:      cuit,
		Regime:    client.RegimeRI,
		CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, period.Location()),
		RIProfile: &client.RIProfile{
			SelfEmployedMonthlyContribution: money.New(10000000, money.ARS),
			PreviousVatCredit:               money.Zero(money.ARS),
		},
	}
	_ = h.clients.Save(ctx, cli)

	// Factura A de venta local: Neto 100.000 ARS + IVA 21.000 ARS = Total 121.000 ARS
	invA := invoice.Invoice{
		ID:           "inv_a",
		ClientID:     cli.ID,
		Type:         invoice.TypeA,
		PointOfSale:  1,
		Number:       1,
		IssueDate:    time.Date(2026, 10, 5, 0, 0, 0, 0, period.Location()),
		Currency:     money.ARS,
		NetAmount:    money.New(10000000, money.ARS),
		VatAmount:    money.New(2100000, money.ARS),
		TotalAmount:  money.New(12100000, money.ARS),
		ReceiverName: "Empresa SA",
	}
	_ = h.invoices.Save(ctx, invA)

	// Factura de compra: Crédito fiscal IVA 5.000 ARS
	pi := invoice.PurchaseInvoice{
		ID:           "pi_1",
		ClientID:     cli.ID,
		SupplierCUIT: "30-55555555-9",
		SupplierName: "Proveedor SA",
		IssueDate:    time.Date(2026, 10, 8, 0, 0, 0, 0, period.Location()),
		NetAmount:    money.New(2380952, money.ARS),
		VatAmount:    money.New(500000, money.ARS),
	}
	_ = h.purchaseInvoices.Save(ctx, pi)

	res, err := h.uc.Execute(ctx, cli.ID, p)
	if err != nil {
		t.Fatalf("error en Execute: %v", err)
	}

	clo := res.Closing
	if clo.RI == nil {
		t.Fatal("esperaba liquidación RI")
	}
	// OutputVat 21.000 - InputVat 5.000 = VatToPay 16.000 ARS
	if clo.RI.VatToPay.Amount != 1600000 {
		t.Fatalf("esperaba VatToPay 16.000 ARS (1600000 centavos), obtuvo %d", clo.RI.VatToPay.Amount)
	}
	// TotalToPay = VatToPay (16.000) + Autónomos (100.000) = 116.000 ARS
	if clo.RI.TotalToPay.Amount != 11600000 {
		t.Fatalf("esperaba TotalToPay 116.000 ARS, obtuvo %d", clo.RI.TotalToPay.Amount)
	}
}

func TestComputeClosing_RegimeChangeInPeriod(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 11, 1, 9, 0, 0, 0, period.Location())
	h := newTestHarness(now)
	p := period.MustParse("2026-10")

	// Cliente que pasó a RI el 15/10/2026
	cuit, _ := client.ParseCUIT("20-44444444-5")
	effDate := time.Date(2026, 10, 15, 0, 0, 0, 0, period.Location())
	toMono := effDate.AddDate(0, 0, -1)

	cli := client.Client{
		ID:        "cli_trans",
		Name:      "Transición Test",
		CUIT:      cuit,
		Regime:    client.RegimeRI,
		CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, period.Location()),
		RegimeHistory: []client.RegimeHistoryItem{
			{Regime: client.RegimeMonotributo, From: time.Date(2025, 1, 1, 0, 0, 0, 0, period.Location()), To: &toMono},
			{Regime: client.RegimeRI, From: effDate, To: nil},
		},
		RIProfile: &client.RIProfile{
			SelfEmployedMonthlyContribution: money.New(5000000, money.ARS),
		},
	}
	_ = h.clients.Save(ctx, cli)

	// Emitió una factura C el 20/10/2026 (posterior a su pase a RI)
	invC := invoice.Invoice{
		ID:           "inv_c_post",
		ClientID:     cli.ID,
		Type:         invoice.TypeC,
		PointOfSale:  1,
		Number:       1,
		IssueDate:    time.Date(2026, 10, 20, 0, 0, 0, 0, period.Location()),
		Currency:     money.ARS,
		TotalAmount:  money.New(5000000, money.ARS),
		ReceiverName: "Cliente",
	}
	_ = h.invoices.Save(ctx, invC)

	res, err := h.uc.Execute(ctx, cli.ID, p)
	if err != nil {
		t.Fatalf("error en Execute: %v", err)
	}

	clo := res.Closing
	// Al final del período es RI
	if clo.RI == nil {
		t.Fatal("esperaba liquidación RI al final del período")
	}
	// Debe detectar la señal POST_EXCLUSION_C_INVOICES (RED)
	foundSignal := false
	for _, s := range clo.Signals {
		if s.Code == signal.CodePostExclusionCInvoices {
			foundSignal = true
			break
		}
	}
	if !foundSignal {
		t.Fatal("esperaba señal POST_EXCLUSION_C_INVOICES por factura C emitida tras pasar a RI")
	}
	if clo.Severity != signal.SeverityRed {
		t.Fatalf("esperaba severidad RED, obtuvo %s", clo.Severity)
	}
}

func TestComputeClosing_IdempotencyAndReopen(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 11, 1, 9, 0, 0, 0, period.Location())
	h := newTestHarness(now)
	p := period.MustParse("2026-10")

	cuit, _ := client.ParseCUIT("20-55555555-6")
	cli := client.Client{
		ID:                  "cli_idemp",
		Name:                "Idempotency Test",
		CUIT:                cuit,
		Regime:              client.RegimeMonotributo,
		MonotributoCategory: tax.CatA,
		CreatedAt:           time.Date(2025, 1, 1, 0, 0, 0, 0, period.Location()),
	}
	_ = h.clients.Save(ctx, cli)

	// 1. Primera ejecución -> CREATED, versión 1
	res1, err := h.uc.Execute(ctx, cli.ID, p)
	if err != nil {
		t.Fatalf("error en res1: %v", err)
	}
	if res1.Status != usecase.ComputeStatusCreated || res1.Closing.Version != 1 {
		t.Fatalf("esperaba CREATED v1, obtuvo %s v%d", res1.Status, res1.Closing.Version)
	}

	// 2. Segunda ejecución idéntica -> UNCHANGED, versión 1
	res2, err := h.uc.Execute(ctx, cli.ID, p)
	if err != nil {
		t.Fatalf("error en res2: %v", err)
	}
	if res2.Status != usecase.ComputeStatusUnchanged || res2.Closing.Version != 1 {
		t.Fatalf("esperaba UNCHANGED v1, obtuvo %s v%d", res2.Status, res2.Closing.Version)
	}

	// 3. Aprobar el cierre en versión 1
	clo := res2.Closing
	_ = clo.Approve(closing.Approval{
		ID:         "app_1",
		ApprovedBy: "contador",
		ApprovedAt: now,
	})
	_ = h.closings.Save(ctx, clo)

	// 4. Agregar una nueva factura en el período que modifica el resultado
	inv := invoice.Invoice{
		ID:           "inv_new",
		ClientID:     cli.ID,
		Type:         invoice.TypeC,
		PointOfSale:  1,
		Number:       2,
		IssueDate:    time.Date(2026, 10, 25, 0, 0, 0, 0, period.Location()),
		Currency:     money.ARS,
		TotalAmount:  money.New(500000000, money.ARS), // Salto de categoría
		ReceiverName: "Cliente Nuevo",
	}
	_ = h.invoices.Save(ctx, inv)

	// 5. Ejecutar recálculo -> reabre a READY_FOR_REVIEW con versión 2 y preserva historial
	res3, err := h.uc.Execute(ctx, cli.ID, p)
	if err != nil {
		t.Fatalf("error en res3: %v", err)
	}
	if res3.Status != usecase.ComputeStatusUpdated {
		t.Fatalf("esperaba status UPDATED, obtuvo %s", res3.Status)
	}
	if res3.Closing.Version != 2 {
		t.Fatalf("esperaba versión 2, obtuvo %d", res3.Closing.Version)
	}
	if res3.Closing.Status != closing.StatusReadyForReview {
		t.Fatalf("cierre aprobado debió reabrirse a READY_FOR_REVIEW tras cambio de hash, obtuvo %s", res3.Closing.Status)
	}
	if len(res3.Closing.Approvals) != 1 {
		t.Fatalf("debió preservar la aprobación anterior: %+v", res3.Closing.Approvals)
	}
}
