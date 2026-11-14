package seed

import (
	"context"
	"fmt"
	"math/rand/v2"
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

// SeedConfig define los parámetros para la generación de datos semilla.
type SeedConfig struct {
	Seed    uint64
	Clients int
	AppNow  time.Time
}

// Dataset agrupa todos los datos generados por el seed.
type Dataset struct {
	Clients          []client.Client
	Movements        []movement.Movement
	Invoices         []invoice.Invoice
	PurchaseInvoices []invoice.PurchaseInvoice
	ClientInputs     []clientinput.ClientPeriodInput
}

var firstNames = []string{
	"Santiago", "Mateo", "Joaquín", "Nicolás", "Tomás", "Lucas", "Benjamín", "Felipe",
	"Martín", "Agustín", "Facundo", "Ignacio", "Manuel", "Franco", "Gonzalo", "Bautista",
	"Valentina", "Sofía", "Camila", "Martina", "Lucía", "Catalina", "Julieta", "Paula",
	"Agustina", "Carla", "Florencia", "Mariana", "Victoria", "Milagros", "Carolina", "Sol",
}

var lastNames = []string{
	"González", "Rodríguez", "López", "García", "Martínez", "Pérez", "Álvarez", "Romero",
	"Sánchez", "Díaz", "Torres", "Fernández", "Benítez", "Ruiz", "Herrera", "Medina",
	"Castro", "Peralta", "Molina", "Acosta", "Rojas", "Silva", "Morales", "Suárez",
	"Vázquez", "Giménez", "Ríos", "Mendoza", "Navarro", "Guzmán", "Bustos", "Domínguez",
}

type payerTemplate struct {
	name    string
	country string
	kind    movement.CounterpartyKind
}

var foreignPayers = []payerTemplate{
	{"Northwind Labs LLC", "US", movement.CounterpartyCompany},
	{"Brightloop GmbH", "DE", movement.CounterpartyCompany},
	{"Kiwi Studio Ltd", "GB", movement.CounterpartyCompany},
	{"Acme Digital Inc", "US", movement.CounterpartyCompany},
	{"Orbital SaaS Corp", "US", movement.CounterpartyCompany},
	{"PixelCraft Studio LLC", "US", movement.CounterpartyCompany},
	{"Apex Cloud Systems Inc", "US", movement.CounterpartyCompany},
	{"Vanguard Software BV", "NL", movement.CounterpartyCompany},
	{"Zenith AI Labs Inc", "US", movement.CounterpartyCompany},
	{"BlueOcean Tech Ltd", "GB", movement.CounterpartyCompany},
	{"Nordic Code Hub AB", "SE", movement.CounterpartyCompany},
	{"Alpen Consulting GmbH", "CH", movement.CounterpartyCompany},
	{"Pacific Media Group LLC", "US", movement.CounterpartyCompany},
	{"Beacon Analytics Inc", "US", movement.CounterpartyCompany},
	{"Crestview Ventures LLC", "US", movement.CounterpartyCompany},
	{"Summit Cloud Services Inc", "US", movement.CounterpartyCompany},
	{"Solaria Interactive SL", "ES", movement.CounterpartyCompany},
	{"Horizon Data Systems Inc", "US", movement.CounterpartyCompany},
	{"Hyperion Labs LLC", "US", movement.CounterpartyCompany},
	{"NextGen Commerce Inc", "US", movement.CounterpartyCompany},
	{"Kinetix Global Ltd", "GB", movement.CounterpartyCompany},
	{"Streamline Systems LLC", "US", movement.CounterpartyCompany},
	{"AeroTech Dynamics Inc", "US", movement.CounterpartyCompany},
	{"Silverline Solutions GmbH", "DE", movement.CounterpartyCompany},
	{"Atlas Digital Corp", "US", movement.CounterpartyCompany},
	{"Deel", "US", movement.CounterpartyPlatform},
	{"Upwork", "US", movement.CounterpartyPlatform},
	{"Fiverr", "IL", movement.CounterpartyPlatform},
	{"Toptal", "US", movement.CounterpartyPlatform},
	{"Ontop", "US", movement.CounterpartyPlatform},
}

var cardMerchants = []struct {
	name    string
	country string
}{
	{"Amazon", "US"},
	{"Spotify", "SE"},
	{"Netflix", "US"},
	{"GitHub", "US"},
	{"Google Cloud", "US"},
	{"AWS", "US"},
	{"DigitalOcean", "US"},
	{"OpenAI", "US"},
	{"Figma", "US"},
	{"Uber", "US"},
	{"Apple", "US"},
	{"Starbucks", "US"},
}

// Generate genera de forma determinística los clientes y movimientos requeridos.
func Generate(ctx context.Context, cfg SeedConfig, fx ports.ExchangeRateProvider, idGen ports.IDGenerator) (Dataset, error) {
	if cfg.Clients <= 0 {
		cfg.Clients = 300
	}
	if cfg.AppNow.IsZero() {
		cfg.AppNow = time.Date(2026, 11, 13, 12, 0, 0, 0, period.Location())
	}

	// 1. Cargar los 15 clientes demo
	demoData, err := BuildDemoClients(ctx, fx)
	if err != nil {
		return Dataset{}, fmt.Errorf("error construyendo demo clients: %w", err)
	}

	clients := make([]client.Client, 0, cfg.Clients)
	clients = append(clients, demoData.Clients...)

	movements := make([]movement.Movement, 0, len(demoData.Movements)+cfg.Clients*50)
	movements = append(movements, demoData.Movements...)

	invoices := make([]invoice.Invoice, 0, len(demoData.Invoices)+cfg.Clients*15)
	invoices = append(invoices, demoData.Invoices...)

	purchaseInvoices := make([]invoice.PurchaseInvoice, 0, len(demoData.PurchaseInvoices)+cfg.Clients*5)
	purchaseInvoices = append(purchaseInvoices, demoData.PurchaseInvoices...)

	clientInputs := make([]clientinput.ClientPeriodInput, 0, len(demoData.ClientInputs)+cfg.Clients)
	clientInputs = append(clientInputs, demoData.ClientInputs...)

	// RNG determinístico con PCG
	pcg := rand.NewPCG(cfg.Seed, cfg.Seed)
	rng := rand.New(pcg)

	remainingClients := cfg.Clients - len(demoData.Clients)
	if remainingClients < 0 {
		remainingClients = 0
	}

	// Para lograr exactamente 85% Monotributo y 15% RI del total (cfg.Clients)
	targetTotalRI := (cfg.Clients * 15) / 100
	demoRICount := 0
	for _, c := range demoData.Clients {
		if c.Regime == client.RegimeRI {
			demoRICount++
		}
	}
	remainingRI := targetTotalRI - demoRICount
	if remainingRI < 0 {
		remainingRI = 0
	}

	startDate := time.Date(2025, 10, 1, 0, 0, 0, 0, period.Location())

	for i := 0; i < remainingClients; i++ {
		cid := idGen.New("cli")
		fn := firstNames[rng.IntN(len(firstNames))]
		ln := lastNames[rng.IntN(len(lastNames))]
		fullName := fn + " " + ln
		email := fmt.Sprintf("%s.%s_%d@wallbit-test.io", fn, ln, i+1)

		isRI := i < remainingRI
		regime := client.RegimeMonotributo
		var monoCategory tax.Category
		var riProfile *client.RIProfile

		var baseIncomeUSD int64

		if isRI {
			regime = client.RegimeRI
			riProfile = &client.RIProfile{
				SelfEmployedMonthlyContribution: money.New(int64(45000+rng.IntN(50000))*100, money.ARS),
				PreviousVatCredit:               money.Zero(money.ARS),
			}
			baseIncomeUSD = int64(1800 + rng.IntN(2500))
		} else {
			monoIdx := i - remainingRI
			switch {
			case monoIdx < 16:
				// 16 clientes RED: superan tope K (pasan a RI)
				baseIncomeUSD = int64(7500 + rng.IntN(500))
				monoCategory = tax.CatK
			case monoIdx < 16+46:
				// 46 clientes YELLOW: cerca del tope de categoría
				switch monoIdx % 4 {
				case 0:
					baseIncomeUSD = int64(1600 + rng.IntN(50))
					monoCategory = tax.CatD
				case 1:
					baseIncomeUSD = int64(1900 + rng.IntN(50))
					monoCategory = tax.CatE
				case 2:
					baseIncomeUSD = int64(2800 + rng.IntN(50))
					monoCategory = tax.CatG
				default:
					baseIncomeUSD = int64(6700 + rng.IntN(100))
					monoCategory = tax.CatK
				}
			default:
				// 180 clientes GREEN: cómodos dentro de su categoría (< 80% del tope)
				if monoIdx%5 < 2 {
					baseIncomeUSD = int64(350 + rng.IntN(100))
					monoCategory = tax.CatA
				} else {
					baseIncomeUSD = int64(3300 + rng.IntN(50))
					monoCategory = tax.CatH
				}
			}
		}

		// CUIT con terminación rotativa 0..9 para distribución uniforme
		termination := i % 10
		prefix := 20
		if rng.IntN(2) == 0 {
			prefix = 27
		}
		cuit := MakeValidCUIT(prefix, 2000000+i*13, termination)

		newClient := client.Client{
			ID:                  cid,
			Name:                fullName,
			Email:               email,
			CUIT:                cuit,
			Regime:              regime,
			MonotributoCategory: monoCategory,
			RIProfile:           riProfile,
			CreatedAt:           startDate,
		}
		clients = append(clients, newClient)

		// 1 a 3 pagadores extranjeros asignados a este cliente
		numPayers := 1 + rng.IntN(3)
		clientPayers := make([]payerTemplate, numPayers)
		for pi := 0; pi < numPayers; pi++ {
			clientPayers[pi] = foreignPayers[rng.IntN(len(foreignPayers))]
		}

		payDay := 5 + rng.IntN(20) // día de cobro habitual

		// 13 meses de movimientos
		for m := 0; m < 13; m++ {
			year := 2025
			month := 10 + m
			if month > 12 {
				year = 2026
				month -= 12
			}

			// Cobro principal con variación de ±15%
			variation := 0.85 + rng.Float64()*0.30
			monthUSD := int64(float64(baseIncomeUSD) * variation)
			if monthUSD < 100 {
				monthUSD = 100
			}

			payer := clientPayers[m%len(clientPayers)]
			payDate := time.Date(year, time.Month(month), payDay, 10+rng.IntN(8), rng.IntN(60), 0, 0, period.Location())

			q, err := fx.USDQuote(ctx, payDate)
			if err == nil {
				movements = append(movements, movement.Movement{
					ID:         idGen.New("mov"),
					ClientID:   cid,
					Type:       movement.TypeInboundTransfer,
					Amount:     money.New(monthUSD*100, money.USD),
					FXRate:     q.Buy,
					OccurredAt: payDate,
					Counterparty: movement.Counterparty{
						Name:    payer.name,
						Country: payer.country,
						Kind:    payer.kind,
					},
					Description: "Professional services payment",
				})

				// Generar factura E fechada 0 a 5 días antes del cobro (criterio 18)
				invDate := payDate.AddDate(0, 0, -rng.IntN(5))
				prevDay := period.PrevBusinessDay(invDate.AddDate(0, 0, -1))
				var invFXRate *money.Rate
				if qPrev, err := fx.USDQuote(ctx, prevDay); err == nil {
					invFXRate = &qPrev.Buy
				}
				invoices = append(invoices, invoice.Invoice{
					ID:           idGen.New("inv"),
					ClientID:     cid,
					Type:         invoice.TypeE,
					PointOfSale:  1,
					Number:       m + 1,
					IssueDate:    invDate,
					Currency:     money.USD,
					NetAmount:    money.New(monthUSD*100, money.USD),
					VatAmount:    money.Zero(money.USD),
					TotalAmount:  money.New(monthUSD*100, money.USD),
					FXRate:       invFXRate,
					ReceiverName: payer.name,
					ReceiverDoc:  fmt.Sprintf("US%08d", rng.IntN(100000000)),
					Source:       invoice.SourceSeed,
					CreatedAt:    invDate,
				})
			}

			// Para clientes RI: generar factura de compra mensual para crédito fiscal
			if isRI && rng.IntN(2) == 0 {
				piDate := time.Date(year, time.Month(month), 10+rng.IntN(15), 10, 0, 0, 0, period.Location())
				netARS := int64(100000+rng.IntN(400000)) * 100 // ARS 100,000 to 500,000
				vatARS := netARS * 21 / 100
				purchaseInvoices = append(purchaseInvoices, invoice.PurchaseInvoice{
					ID:           idGen.New("pi"),
					ClientID:     cid,
					IssueDate:    piDate,
					SupplierName: "Proveedor Local de Insumos SA",
					SupplierCUIT: string(MakeValidCUIT(30, 7100000+rng.IntN(100000), rng.IntN(10))),
					NetAmount:    money.New(netARS, money.ARS),
					VatAmount:    money.New(vatARS, money.ARS),
					CreatedAt:    piDate,
				})
			}

			// Gastos con tarjeta: 5 a 20 consumos por mes (USD 5 a 300)
			numSpends := 5 + rng.IntN(16)
			for s := 0; s < numSpends; s++ {
				spDay := 1 + rng.IntN(28)
				spDate := time.Date(year, time.Month(month), spDay, rng.IntN(24), rng.IntN(60), 0, 0, period.Location())
				qSpend, err := fx.USDQuote(ctx, spDate)
				if err != nil {
					continue
				}
				merchant := cardMerchants[rng.IntN(len(cardMerchants))]
				spendUSD := int64(5 + rng.IntN(295))

				movements = append(movements, movement.Movement{
					ID:         idGen.New("mov"),
					ClientID:   cid,
					Type:       movement.TypeCardSpend,
					Amount:     money.New(spendUSD*100, money.USD),
					FXRate:     qSpend.Buy,
					OccurredAt: spDate,
					Counterparty: movement.Counterparty{
						Name:    merchant.name,
						Country: merchant.country,
						Kind:    movement.CounterpartyCompany,
					},
					Description: fmt.Sprintf("%s purchase", merchant.name),
				})
			}

			// Ocasional transferencia saliente (0 a 1)
			if rng.IntN(3) == 0 {
				outDay := 10 + rng.IntN(15)
				outDate := time.Date(year, time.Month(month), outDay, 12, 0, 0, 0, period.Location())
				qOut, err := fx.USDQuote(ctx, outDate)
				if err == nil {
					movements = append(movements, movement.Movement{
						ID:         idGen.New("mov"),
						ClientID:   cid,
						Type:       movement.TypeOutboundTransfer,
						Amount:     money.New(int64(100+rng.IntN(500))*100, money.USD),
						FXRate:     qOut.Buy,
						OccurredAt: outDate,
						Counterparty: movement.Counterparty{
							Name:    "External Bank",
							Country: "US",
							Kind:    movement.CounterpartyCompany,
						},
						Description: "Outbound transfer",
					})
				}
			}

			// Ocasional fee (0 a 1)
			if rng.IntN(4) == 0 {
				feeDate := time.Date(year, time.Month(month), 28, 23, 0, 0, 0, period.Location())
				qFee, err := fx.USDQuote(ctx, feeDate)
				if err == nil {
					movements = append(movements, movement.Movement{
						ID:         idGen.New("mov"),
						ClientID:   cid,
						Type:       movement.TypeFee,
						Amount:     money.New(500, money.USD), // USD 5.00
						FXRate:     qFee.Buy,
						OccurredAt: feeDate,
						Counterparty: movement.Counterparty{
							Name:    "Wallbit",
							Country: "US",
							Kind:    movement.CounterpartyCompany,
						},
						Description: "Monthly account fee",
					})
				}
			}
		}

		// Movimientos adicionales en noviembre 2026 hasta AppNow
		if cfg.AppNow.After(time.Date(2026, 11, 1, 0, 0, 0, 0, period.Location())) {
			novSpends := 1 + rng.IntN(3)
			for ns := 0; ns < novSpends; ns++ {
				day := 1 + rng.IntN(12)
				t := time.Date(2026, 11, day, 10+ns, 0, 0, 0, period.Location())
				if t.After(cfg.AppNow) {
					continue
				}
				qNov, err := fx.USDQuote(ctx, t)
				if err == nil {
					mMerchant := cardMerchants[rng.IntN(len(cardMerchants))]
					movements = append(movements, movement.Movement{
						ID:         idGen.New("mov"),
						ClientID:   cid,
						Type:       movement.TypeCardSpend,
						Amount:     money.New(int64(10+rng.IntN(40))*100, money.USD),
						FXRate:     qNov.Buy,
						OccurredAt: t,
						Counterparty: movement.Counterparty{
							Name:    mMerchant.name,
							Country: mMerchant.country,
							Kind:    movement.CounterpartyCompany,
						},
						Description: fmt.Sprintf("%s purchase", mMerchant.name),
					})
				}
			}
		}

		// Input del período octubre 2026 (criterio 18: 90% confirmados, 10% sin confirmar)
		octPeriod := period.Period{Year: 2026, Month: 10}
		if rng.IntN(10) != 0 {
			// 90% confirmados
			confDate := time.Date(2026, 11, 1+rng.IntN(5), 9+rng.IntN(10), rng.IntN(60), 0, 0, period.Location())
			clientInputs = append(clientInputs, clientinput.ClientPeriodInput{
				ClientID:                 cid,
				Period:                   octPeriod,
				DeclaredNoExternalIncome: true,
				ConfirmedAt:              &confDate,
				UpdatedAt:                confDate,
			})
		} else {
			// 10% sin confirmar
			clientInputs = append(clientInputs, clientinput.ClientPeriodInput{
				ClientID:                 cid,
				Period:                   octPeriod,
				DeclaredNoExternalIncome: false,
				ConfirmedAt:              nil,
				UpdatedAt:                time.Date(2026, 10, 31, 0, 0, 0, 0, period.Location()),
			})
		}
	}

	return Dataset{
		Clients:          clients,
		Movements:        movements,
		Invoices:         invoices,
		PurchaseInvoices: purchaseInvoices,
		ClientInputs:     clientInputs,
	}, nil
}
