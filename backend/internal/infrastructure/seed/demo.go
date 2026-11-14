package seed

import (
	"context"
	"fmt"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/clientinput"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/invoice"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/movement"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/tax"
)

// DemoClientsData contiene los 15 clientes de demo, sus movimientos, facturas emitidas,
// facturas de compra e inputs de período (spec 002 y 003).
type DemoClientsData struct {
	Clients          []client.Client
	Movements        []movement.Movement
	Invoices         []invoice.Invoice
	PurchaseInvoices []invoice.PurchaseInvoice
	ClientInputs     []clientinput.ClientPeriodInput
}

// BuildDemoClients construye los 15 clientes de demo y su dataset completo de 13 meses
// según los escenarios detallados en specs/002-clients-and-movements/plan.md.
func BuildDemoClients(ctx context.Context, fx ports.ExchangeRateProvider) (DemoClientsData, error) {
	startDate := time.Date(2025, 10, 1, 0, 0, 0, 0, period.Location())

	// Definición de los 15 clientes
	c01 := client.Client{
		ID:                  "cli_demo_01",
		Name:                "Lucía Fernández",
		Email:               "lucia.fernandez@demo.wallbit.io",
		CUIT:                MakeValidCUIT(27, 3211111, 0),
		Regime:              client.RegimeMonotributo,
		MonotributoCategory: tax.CatA,
		CreatedAt:           startDate,
	}

	c02 := client.Client{
		ID:                  "cli_demo_02",
		Name:                "Ignacio Sosa",
		Email:               "ignacio.sosa@demo.wallbit.io",
		CUIT:                MakeValidCUIT(20, 2822222, 1),
		Regime:              client.RegimeMonotributo,
		MonotributoCategory: tax.CatK,
		CreatedAt:           startDate,
	}

	c03 := client.Client{
		ID:                  "cli_demo_03",
		Name:                "Carla Benítez",
		Email:               "carla.benitez@demo.wallbit.io",
		CUIT:                MakeValidCUIT(27, 3433333, 2),
		Regime:              client.RegimeMonotributo,
		MonotributoCategory: tax.CatE,
		CreatedAt:           startDate,
	}

	c04 := client.Client{
		ID:                  "cli_demo_04",
		Name:                "Diego Romero",
		Email:               "diego.romero@demo.wallbit.io",
		CUIT:                MakeValidCUIT(20, 3044444, 3),
		Regime:              client.RegimeMonotributo,
		MonotributoCategory: tax.CatC,
		CreatedAt:           startDate,
	}

	c05 := client.Client{
		ID:                  "cli_demo_05",
		Name:                "Valentina Ruiz",
		Email:               "valentina.ruiz@demo.wallbit.io",
		CUIT:                MakeValidCUIT(27, 3655555, 4),
		Regime:              client.RegimeMonotributo,
		MonotributoCategory: tax.CatF,
		CreatedAt:           startDate,
	}

	c06 := client.Client{
		ID:                  "cli_demo_06",
		Name:                "Tomás Álvarez",
		Email:               "tomas.alvarez@demo.wallbit.io",
		CUIT:                MakeValidCUIT(20, 3266666, 5),
		Regime:              client.RegimeMonotributo,
		MonotributoCategory: tax.CatD,
		CreatedAt:           startDate,
	}

	c07 := client.Client{
		ID:                  "cli_demo_07",
		Name:                "Sofía Medina",
		Email:               "sofia.medina@demo.wallbit.io",
		CUIT:                MakeValidCUIT(27, 3877777, 6),
		Regime:              client.RegimeMonotributo,
		MonotributoCategory: tax.CatC,
		CreatedAt:           startDate,
	}

	c08 := client.Client{
		ID:                  "cli_demo_08",
		Name:                "Joaquín Herrera",
		Email:               "joaquin.herrera@demo.wallbit.io",
		CUIT:                MakeValidCUIT(20, 3488888, 7),
		Regime:              client.RegimeMonotributo,
		MonotributoCategory: tax.CatK,
		CreatedAt:           startDate,
	}

	c09 := client.Client{
		ID:        "cli_demo_09",
		Name:      "Camila Torres",
		Email:     "camila.torres@demo.wallbit.io",
		CUIT:      MakeValidCUIT(27, 3999999, 8),
		Regime:    client.RegimeRI,
		CreatedAt: startDate,
		RIProfile: &client.RIProfile{
			SelfEmployedMonthlyContribution: money.New(5500000, money.ARS),
			PreviousVatCredit:               money.New(20000000, money.ARS), // 200,000 ARS a favor
		},
	}

	c10 := client.Client{
		ID:        "cli_demo_10",
		Name:      "Federico Díaz",
		Email:     "federico.diaz@demo.wallbit.io",
		CUIT:      MakeValidCUIT(20, 3100000, 9),
		Regime:    client.RegimeRI,
		CreatedAt: startDate,
		RIProfile: &client.RIProfile{
			SelfEmployedMonthlyContribution: money.New(7500000, money.ARS),
			PreviousVatCredit:               money.Zero(money.ARS),
		},
	}

	riFromAgustina := time.Date(2026, 8, 12, 0, 0, 0, 0, period.Location())
	toMonoAgustina := riFromAgustina.Add(-time.Second)
	c11 := client.Client{
		ID:        "cli_demo_11",
		Name:      "Agustina Castro",
		Email:     "agustina.castro@demo.wallbit.io",
		CUIT:      MakeValidCUIT(27, 3500000, 0),
		Regime:    client.RegimeRI,
		CreatedAt: startDate,
		RegimeHistory: []client.RegimeHistoryItem{
			{Regime: client.RegimeMonotributo, From: startDate, To: &toMonoAgustina, Reason: "INICIO"},
			{Regime: client.RegimeRI, From: riFromAgustina, To: nil, Reason: "EXCLUSION_K_CAP", Note: "Superó tope K"},
		},
		RIProfile: &client.RIProfile{
			SelfEmployedMonthlyContribution: money.New(8500000, money.ARS),
			PreviousVatCredit:               money.Zero(money.ARS),
		},
	}

	c12 := client.Client{
		ID:                  "cli_demo_12",
		Name:                "Nicolás Peralta",
		Email:               "nicolas.peralta@demo.wallbit.io",
		CUIT:                MakeValidCUIT(20, 3311111, 1),
		Regime:              client.RegimeMonotributo,
		MonotributoCategory: tax.CatG,
		CreatedAt:           startDate,
	}

	c13 := client.Client{
		ID:                  "cli_demo_13",
		Name:                "Julieta Molina",
		Email:               "julieta.molina@demo.wallbit.io",
		CUIT:                MakeValidCUIT(27, 3722222, 2),
		Regime:              client.RegimeMonotributo,
		MonotributoCategory: tax.CatD,
		CreatedAt:           startDate,
	}

	c14 := client.Client{
		ID:                  "cli_demo_14",
		Name:                "Bruno Acosta",
		Email:               "bruno.acosta@demo.wallbit.io",
		CUIT:                MakeValidCUIT(20, 3533333, 3),
		Regime:              client.RegimeMonotributo,
		MonotributoCategory: tax.CatH,
		CreatedAt:           startDate,
	}

	c15 := client.Client{
		ID:                  "cli_demo_15",
		Name:                "Paula Rojas",
		Email:               "paula.rojas@demo.wallbit.io",
		CUIT:                MakeValidCUIT(27, 3344444, 4),
		Regime:              client.RegimeMonotributo,
		MonotributoCategory: tax.CatB,
		CreatedAt:           startDate,
	}

	clients := []client.Client{c01, c02, c03, c04, c05, c06, c07, c08, c09, c10, c11, c12, c13, c14, c15}

	var movements []movement.Movement
	var invoices []invoice.Invoice
	var purchaseInvoices []invoice.PurchaseInvoice
	var clientInputs []clientinput.ClientPeriodInput

	movIdx := 1
	invIdx := 1
	piIdx := 1

	addMovement := func(clientID string, mType movement.MovementType, usdCents int64, t time.Time, cp movement.Counterparty, desc string) error {
		q, err := fx.USDQuote(ctx, t)
		if err != nil {
			return fmt.Errorf("quote error para %s: %w", t.Format("2006-01-02"), err)
		}
		movements = append(movements, movement.Movement{
			ID:           fmt.Sprintf("mov_demo_%05d", movIdx),
			ClientID:     clientID,
			Type:         mType,
			Amount:       money.New(usdCents, money.USD),
			FXRate:       q.Buy,
			OccurredAt:   t,
			Counterparty: cp,
			Description:  desc,
		})
		movIdx++
		return nil
	}

	addInvoice := func(clientID string, invType invoice.InvoiceType, pos, num int, d time.Time, c money.Currency, totalCents int64, receiver string) error {
		var fxRate *money.Rate
		if c == money.USD {
			prevDay := period.PrevBusinessDay(d.AddDate(0, 0, -1))
			q, err := fx.USDQuote(ctx, prevDay)
			if err == nil {
				fxRate = &q.Buy
			}
		}
		invoices = append(invoices, invoice.Invoice{
			ID:           fmt.Sprintf("inv_demo_%05d", invIdx),
			ClientID:     clientID,
			Type:         invType,
			PointOfSale:  pos,
			Number:       num,
			IssueDate:    d,
			Currency:     c,
			NetAmount:    money.New(totalCents, c),
			VatAmount:    money.Zero(c),
			TotalAmount:  money.New(totalCents, c),
			FXRate:       fxRate,
			ReceiverName: receiver,
			Source:       invoice.SourceSeed,
			CreatedAt:    d,
		})
		invIdx++
		return nil
	}

	oct2026Period, _ := period.Parse("2026-10")

	// 13 meses: 2025-10 a 2026-10
	for m := 0; m < 13; m++ {
		year := 2025
		month := 10 + m
		if month > 12 {
			year = 2026
			month -= 12
		}

		isOct2026 := (year == 2026 && month == 10)

		// cli_demo_01: Lucía Fernández (MONO A -> GREEN)
		_ = addMovement("cli_demo_01", movement.TypeInboundTransfer, 50000,
			time.Date(year, time.Month(month), 10, 14, 30, 0, 0, period.Location()),
			movement.Counterparty{Name: "Brightloop GmbH", Country: "DE", Kind: movement.CounterpartyCompany}, "Honorarios desarrollo backend")
		_ = addInvoice("cli_demo_01", invoice.TypeE, 1, m+1, time.Date(year, time.Month(month), 6, 10, 0, 0, 0, period.Location()), money.USD, 50000, "Brightloop GmbH")

		_ = addMovement("cli_demo_01", movement.TypeCardSpend, 2500, time.Date(year, time.Month(month), 12, 10, 0, 0, 0, period.Location()),
			movement.Counterparty{Name: "Spotify", Country: "SE", Kind: movement.CounterpartyCompany}, "Suscripción Spotify")
		_ = addMovement("cli_demo_01", movement.TypeCardSpend, 6500, time.Date(year, time.Month(month), 20, 18, 0, 0, 0, period.Location()),
			movement.Counterparty{Name: "Amazon", Country: "US", Kind: movement.CounterpartyCompany}, "Compra insumos de trabajo")

		// cli_demo_02: Ignacio Sosa (MONO K -> MONO_EXCLUSION_K_CAP_EXCEEDED el 14/10/2026)
		if isOct2026 {
			_ = addMovement("cli_demo_02", movement.TypeInboundTransfer, 400000,
				time.Date(2026, 10, 5, 11, 0, 0, 0, period.Location()),
				movement.Counterparty{Name: "Northwind Labs LLC", Country: "US", Kind: movement.CounterpartyCompany}, "Pago servicios consultoría cloud")
			_ = addInvoice("cli_demo_02", invoice.TypeE, 1, 20, time.Date(2026, 10, 2, 10, 0, 0, 0, period.Location()), money.USD, 400000, "Northwind Labs LLC")

			_ = addMovement("cli_demo_02", movement.TypeInboundTransfer, 350000,
				time.Date(2026, 10, 14, 15, 0, 0, 0, period.Location()),
				movement.Counterparty{Name: "Orbital SaaS Corp", Country: "US", Kind: movement.CounterpartyCompany}, "Hito final migración arquitectura")
			_ = addInvoice("cli_demo_02", invoice.TypeE, 1, 21, time.Date(2026, 10, 14, 10, 0, 0, 0, period.Location()), money.USD, 350000, "Orbital SaaS Corp")
		} else {
			_ = addMovement("cli_demo_02", movement.TypeInboundTransfer, 710000,
				time.Date(year, time.Month(month), 8, 11, 0, 0, 0, period.Location()),
				movement.Counterparty{Name: "Northwind Labs LLC", Country: "US", Kind: movement.CounterpartyCompany}, "Servicios consultoría cloud")
			_ = addInvoice("cli_demo_02", invoice.TypeE, 1, m+1, time.Date(year, time.Month(month), 5, 10, 0, 0, 0, period.Location()), money.USD, 710000, "Northwind Labs LLC")
		}

		// cli_demo_03: Carla Benítez (MONO E -> MISSING_INVOICES: 2 cobros en oct 2026 sin factura)
		if isOct2026 {
			_ = addMovement("cli_demo_03", movement.TypeInboundTransfer, 180000,
				time.Date(2026, 10, 8, 12, 0, 0, 0, period.Location()),
				movement.Counterparty{Name: "Orbital SaaS Corp", Country: "US", Kind: movement.CounterpartyCompany}, "Desarrollo frontend")
			_ = addMovement("cli_demo_03", movement.TypeInboundTransfer, 120000,
				time.Date(2026, 10, 22, 16, 0, 0, 0, period.Location()),
				movement.Counterparty{Name: "Kiwi Studio Ltd", Country: "GB", Kind: movement.CounterpartyCompany}, "Servicios UI/UX")
			// No se emiten facturas en octubre -> MISSING_INVOICES!
		} else {
			_ = addMovement("cli_demo_03", movement.TypeInboundTransfer, 200000,
				time.Date(year, time.Month(month), 10, 12, 0, 0, 0, period.Location()),
				movement.Counterparty{Name: "Orbital SaaS Corp", Country: "US", Kind: movement.CounterpartyCompany}, "Desarrollo frontend")
			_ = addInvoice("cli_demo_03", invoice.TypeE, 1, m+1, time.Date(year, time.Month(month), 7, 10, 0, 0, 0, period.Location()), money.USD, 200000, "Orbital SaaS Corp")
		}

		// cli_demo_04: Diego Romero (MONO C -> MONO_CATEGORY_CHANGE_REQUIRED necesita E)
		_ = addMovement("cli_demo_04", movement.TypeInboundTransfer, 190000,
			time.Date(year, time.Month(month), 15, 10, 0, 0, 0, period.Location()),
			movement.Counterparty{Name: "Kiwi Studio Ltd", Country: "GB", Kind: movement.CounterpartyCompany}, "Desarrollo mobile")
		_ = addInvoice("cli_demo_04", invoice.TypeE, 1, m+1, time.Date(year, time.Month(month), 12, 10, 0, 0, 0, period.Location()), money.USD, 190000, "Kiwi Studio Ltd")

		// cli_demo_05: Valentina Ruiz (MONO F -> MONO_NEAR_CATEGORY_CAP ~90% de F)
		_ = addMovement("cli_demo_05", movement.TypeInboundTransfer, 228000,
			time.Date(year, time.Month(month), 11, 10, 0, 0, 0, period.Location()),
			movement.Counterparty{Name: "Northwind Labs LLC", Country: "US", Kind: movement.CounterpartyCompany}, "Data engineering")
		_ = addInvoice("cli_demo_05", invoice.TypeE, 1, m+1, time.Date(year, time.Month(month), 8, 10, 0, 0, 0, period.Location()), money.USD, 228000, "Northwind Labs LLC")

		// cli_demo_06: Tomás Álvarez (MONO D -> ATYPICAL_INCOME en oct 2026 con cobro 3x mediana)
		if isOct2026 {
			_ = addMovement("cli_demo_06", movement.TypeInboundTransfer, 350000,
				time.Date(2026, 10, 16, 14, 0, 0, 0, period.Location()),
				movement.Counterparty{Name: "Acme Digital Inc", Country: "US", Kind: movement.CounterpartyCompany}, "Bonus por entrega anticipada")
			_ = addInvoice("cli_demo_06", invoice.TypeE, 1, 20, time.Date(2026, 10, 14, 10, 0, 0, 0, period.Location()), money.USD, 350000, "Acme Digital Inc")
		} else {
			_ = addMovement("cli_demo_06", movement.TypeInboundTransfer, 95000,
				time.Date(year, time.Month(month), 14, 10, 0, 0, 0, period.Location()),
				movement.Counterparty{Name: "Acme Digital Inc", Country: "US", Kind: movement.CounterpartyCompany}, "Soporte técnico y mantenimiento")
			_ = addInvoice("cli_demo_06", invoice.TypeE, 1, m+1, time.Date(year, time.Month(month), 10, 10, 0, 0, 0, period.Location()), money.USD, 95000, "Acme Digital Inc")
		}

		// cli_demo_07: Sofía Medina (MONO C -> UNCLASSIFIED_MOVEMENTS en oct 2026)
		_ = addMovement("cli_demo_07", movement.TypeInboundTransfer, 110000,
			time.Date(year, time.Month(month), 12, 10, 0, 0, 0, period.Location()),
			movement.Counterparty{Name: "Brightloop GmbH", Country: "DE", Kind: movement.CounterpartyCompany}, "Consultoría QA")
		_ = addInvoice("cli_demo_07", invoice.TypeE, 1, m+1, time.Date(year, time.Month(month), 9, 10, 0, 0, 0, period.Location()), money.USD, 110000, "Brightloop GmbH")
		if isOct2026 {
			_ = addMovement("cli_demo_07", movement.TypeInboundTransfer, 60000,
				time.Date(2026, 10, 19, 16, 0, 0, 0, period.Location()),
				movement.Counterparty{Name: "Esteban Morales", Country: "AR", Kind: movement.CounterpartyPerson}, "Transferencia recibida AR")
		}

		// cli_demo_08: Joaquín Herrera (MONO K -> MONO_NEAR_K_CAP ~88% de K)
		_ = addMovement("cli_demo_08", movement.TypeInboundTransfer, 625000,
			time.Date(year, time.Month(month), 9, 10, 0, 0, 0, period.Location()),
			movement.Counterparty{Name: "Deel", Country: "US", Kind: movement.CounterpartyPlatform}, "Contrato de software Deel")
		_ = addInvoice("cli_demo_08", invoice.TypeE, 1, m+1, time.Date(year, time.Month(month), 6, 10, 0, 0, 0, period.Location()), money.USD, 625000, "Deel")

		// cli_demo_09: Camila Torres (RI -> GREEN, Factura E exportación + saldo a favor IVA)
		_ = addMovement("cli_demo_09", movement.TypeInboundTransfer, 350000,
			time.Date(year, time.Month(month), 15, 10, 0, 0, 0, period.Location()),
			movement.Counterparty{Name: "Upwork", Country: "US", Kind: movement.CounterpartyPlatform}, "Honorarios Upwork exportación")
		_ = addInvoice("cli_demo_09", invoice.TypeE, 1, m+1, time.Date(year, time.Month(month), 12, 10, 0, 0, 0, period.Location()), money.USD, 350000, "Upwork")

		if isOct2026 {
			// Factura de compra para crédito fiscal
			purchaseInvoices = append(purchaseInvoices, invoice.PurchaseInvoice{
				ID:           fmt.Sprintf("pi_demo_%05d", piIdx),
				ClientID:     "cli_demo_09",
				IssueDate:    time.Date(2026, 10, 18, 0, 0, 0, 0, period.Location()),
				SupplierName: "Equipos y Servidores SA",
				SupplierCUIT: string(MakeValidCUIT(30, 7122222, 1)),
				NetAmount:    money.New(400000000, money.ARS), // 4,000,000 ARS
				VatAmount:    money.New(84000000, money.ARS),  // 840,000 ARS crédito fiscal
				CreatedAt:    time.Date(2026, 10, 18, 0, 0, 0, 0, period.Location()),
			})
			piIdx++
		}

		// cli_demo_10: Federico Díaz (RI -> INVOICE_WITHOUT_PAYMENT + IVA a pagar por Factura A local)
		if isOct2026 {
			// Factura A local sin cobro registrado en Wallbit
			invoices = append(invoices, invoice.Invoice{
				ID:           fmt.Sprintf("inv_demo_%05d", invIdx),
				ClientID:     "cli_demo_10",
				Type:         invoice.TypeA,
				PointOfSale:  1,
				Number:       15,
				IssueDate:    time.Date(2026, 10, 10, 10, 0, 0, 0, period.Location()),
				Currency:     money.ARS,
				NetAmount:    money.New(500000000, money.ARS), // 5,000,000 ARS
				VatAmount:    money.New(105000000, money.ARS), // 1,050,000 ARS IVA débito fiscal
				TotalAmount:  money.New(605000000, money.ARS), // 6,050,000 ARS
				ReceiverName: "Servicios Tecnológicos Argentinos SA",
				ReceiverDoc:  "30-65432109-8",
				Source:       invoice.SourceSeed,
				CreatedAt:    time.Date(2026, 10, 10, 10, 0, 0, 0, period.Location()),
			})
			invIdx++
		} else {
			_ = addMovement("cli_demo_10", movement.TypeInboundTransfer, 200000,
				time.Date(year, time.Month(month), 18, 10, 0, 0, 0, period.Location()),
				movement.Counterparty{Name: "Acme Digital Inc", Country: "US", Kind: movement.CounterpartyCompany}, "Servicios informáticos")
			_ = addInvoice("cli_demo_10", invoice.TypeE, 1, m+1, time.Date(year, time.Month(month), 15, 10, 0, 0, 0, period.Location()), money.USD, 200000, "Acme Digital Inc")
		}

		// cli_demo_11: Agustina Castro (RI desde 2026-08-12 -> POST_EXCLUSION_C_INVOICES)
		_ = addMovement("cli_demo_11", movement.TypeInboundTransfer, 420000,
			time.Date(year, time.Month(month), 7, 10, 0, 0, 0, period.Location()),
			movement.Counterparty{Name: "Northwind Labs LLC", Country: "US", Kind: movement.CounterpartyCompany}, "Servicios devops")
		if isOct2026 {
			// En octubre emite erróneamente 2 facturas C siendo ya RI -> POST_EXCLUSION_C_INVOICES
			_ = addInvoice("cli_demo_11", invoice.TypeC, 1, 40, time.Date(2026, 10, 5, 10, 0, 0, 0, period.Location()), money.ARS, 80000000, "Cliente Local 1")
			_ = addInvoice("cli_demo_11", invoice.TypeC, 1, 41, time.Date(2026, 10, 18, 10, 0, 0, 0, period.Location()), money.ARS, 120000000, "Cliente Local 2")
		} else if m < 10 {
			// Período Monotributo
			_ = addInvoice("cli_demo_11", invoice.TypeE, 1, m+1, time.Date(year, time.Month(month), 4, 10, 0, 0, 0, period.Location()), money.USD, 420000, "Northwind Labs LLC")
		} else {
			// Período RI
			_ = addInvoice("cli_demo_11", invoice.TypeE, 1, m+1, time.Date(year, time.Month(month), 4, 10, 0, 0, 0, period.Location()), money.USD, 420000, "Northwind Labs LLC")
		}

		// cli_demo_12: Nicolás Peralta (MONO G -> DEPOSITS_EXCEED_K_CAP y HIGH_CARD_SPENDING)
		if isOct2026 {
			_ = addMovement("cli_demo_12", movement.TypeInboundTransfer, 250000,
				time.Date(2026, 10, 6, 10, 0, 0, 0, period.Location()),
				movement.Counterparty{Name: "Brightloop GmbH", Country: "DE", Kind: movement.CounterpartyCompany}, "Pagos acumulados")
			_ = addInvoice("cli_demo_12", invoice.TypeE, 1, m+1, time.Date(2026, 10, 3, 10, 0, 0, 0, period.Location()), money.USD, 250000, "Brightloop GmbH")
		} else {
			_ = addMovement("cli_demo_12", movement.TypeInboundTransfer, 1000000,
				time.Date(year, time.Month(month), 6, 10, 0, 0, 0, period.Location()),
				movement.Counterparty{Name: "Brightloop GmbH", Country: "DE", Kind: movement.CounterpartyCompany}, "Pagos acumulados")
			_ = addInvoice("cli_demo_12", invoice.TypeE, 1, m+1, time.Date(year, time.Month(month), 3, 10, 0, 0, 0, period.Location()), money.USD, 250000, "Brightloop GmbH")
		}
		_ = addMovement("cli_demo_12", movement.TypeCardSpend, 350000,
			time.Date(year, time.Month(month), 14, 15, 0, 0, 0, period.Location()),
			movement.Counterparty{Name: "Apple Store", Country: "US", Kind: movement.CounterpartyCompany}, "Equipos de hardware")

		// cli_demo_13: Julieta Molina (MONO D -> EXTERNAL_INCOME_DECLARED)
		_ = addMovement("cli_demo_13", movement.TypeInboundTransfer, 140000,
			time.Date(year, time.Month(month), 10, 10, 0, 0, 0, period.Location()),
			movement.Counterparty{Name: "Kiwi Studio Ltd", Country: "GB", Kind: movement.CounterpartyCompany}, "Diseño digital")
		_ = addInvoice("cli_demo_13", invoice.TypeE, 1, m+1, time.Date(year, time.Month(month), 7, 10, 0, 0, 0, period.Location()), money.USD, 140000, "Kiwi Studio Ltd")

		// cli_demo_14: Bruno Acosta (MONO H -> MONO_CATEGORY_DOWNGRADE_POSSIBLE)
		_ = addMovement("cli_demo_14", movement.TypeInboundTransfer, 180000,
			time.Date(year, time.Month(month), 12, 10, 0, 0, 0, period.Location()),
			movement.Counterparty{Name: "Acme Digital Inc", Country: "US", Kind: movement.CounterpartyCompany}, "Mantenimiento servidores")
		_ = addInvoice("cli_demo_14", invoice.TypeE, 1, m+1, time.Date(year, time.Month(month), 9, 10, 0, 0, 0, period.Location()), money.USD, 180000, "Acme Digital Inc")

		// cli_demo_15: Paula Rojas (MONO B -> PENDING_CLIENT_INPUT y MISSING_INVOICES)
		// Recibe cobro de USD 850 en octubre 2026 pero NO tiene facturas subidas todavía
		_ = addMovement("cli_demo_15", movement.TypeInboundTransfer, 85000,
			time.Date(year, time.Month(month), 15, 10, 0, 0, 0, period.Location()),
			movement.Counterparty{Name: "Brightloop GmbH", Country: "DE", Kind: movement.CounterpartyCompany}, "Traducción técnica")
		if !isOct2026 {
			_ = addInvoice("cli_demo_15", invoice.TypeE, 1, m+1, time.Date(year, time.Month(month), 12, 10, 0, 0, 0, period.Location()), money.USD, 85000, "Brightloop GmbH")
		}
	}

	// Inputs de período para los 15 clientes demo
	confirmedDate := time.Date(2026, 11, 2, 11, 0, 0, 0, period.Location())
	for _, c := range clients {
		if c.ID == "cli_demo_15" {
			// Paula Rojas: NO confirma aún (PENDING_CLIENT_INPUT)
			clientInputs = append(clientInputs, clientinput.ClientPeriodInput{
				ClientID:                 c.ID,
				Period:                   oct2026Period,
				DeclaredNoExternalIncome: false,
				ConfirmedAt:              nil,
				UpdatedAt:                time.Date(2026, 10, 31, 0, 0, 0, 0, period.Location()),
			})
			continue
		}

		if c.ID == "cli_demo_13" {
			// Julieta Molina: declara ingresos externos
			clientInputs = append(clientInputs, clientinput.ClientPeriodInput{
				ClientID: c.ID,
				Period:   oct2026Period,
				ExternalIncome: []clientinput.ExternalIncomeItem{
					{Amount: money.New(120000, money.USD), Description: "Consultoría directa empresa Chile"},
				},
				DeclaredNoExternalIncome: false,
				ConfirmedAt:              &confirmedDate,
				UpdatedAt:                confirmedDate,
			})
			continue
		}

		// Todos los demás clientes confirman sin ingresos externos
		clientInputs = append(clientInputs, clientinput.ClientPeriodInput{
			ClientID:                 c.ID,
			Period:                   oct2026Period,
			DeclaredNoExternalIncome: true,
			ConfirmedAt:              &confirmedDate,
			UpdatedAt:                confirmedDate,
		})
	}

	// Noviembre 2026 días 1 al 12
	for d := 1; d <= 12; d++ {
		t := time.Date(2026, 11, d, 11, 0, 0, 0, period.Location())
		if d == 3 {
			_ = addMovement("cli_demo_01", movement.TypeCardSpend, 3500, t, movement.Counterparty{Name: "Uber", Country: "US", Kind: movement.CounterpartyCompany}, "Viaje Uber")
		}
		if d == 5 {
			_ = addMovement("cli_demo_02", movement.TypeInboundTransfer, 400000, t, movement.Counterparty{Name: "Northwind Labs LLC", Country: "US", Kind: movement.CounterpartyCompany}, "Adelanto noviembre")
		}
		if d == 9 {
			_ = addMovement("cli_demo_09", movement.TypeCardSpend, 1500, t, movement.Counterparty{Name: "GitHub", Country: "US", Kind: movement.CounterpartyCompany}, "Suscripción GitHub Pro")
		}
	}

	return DemoClientsData{
		Clients:          clients,
		Movements:        movements,
		Invoices:         invoices,
		PurchaseInvoices: purchaseInvoices,
		ClientInputs:     clientInputs,
	}, nil
}
