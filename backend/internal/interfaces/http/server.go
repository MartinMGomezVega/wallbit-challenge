// Package http expone la API REST (architecture.md §2.6).
package http

import (
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/api"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/usecase"
)

// Deps son las dependencias del router.
type Deps struct {
	Clock                     ports.Clock
	Logger                    *slog.Logger
	CORSOrigin                string
	ParametersRepo            ports.ParametersRepository
	ClientRepo                ports.ClientRepository
	MovementRepo              ports.MovementRepository
	ClassificationRepo        ports.ClassificationRepository
	InvoiceRepo               ports.InvoiceRepository
	PurchaseInvoiceRepo       ports.PurchaseInvoiceRepository
	ClientInputRepo           ports.ClientInputRepository
	LatestExclusionDateReader ports.LatestExclusionDateReader
	ClosingRepo               ports.ClosingRepository
	ClosingQuery              ports.ClosingQuery
	StatsQuery                ports.StatsQuery
	TxManager                 ports.TxManager
	TaxScales                 ports.TaxScaleProvider
	DueDates                  ports.DueDateProvider
	ExchangeRates             ports.ExchangeRateProvider
	Workers                   int
	IDGen                     ports.IDGenerator
	ChangeNotifier            ports.ChangeNotifier
	NotificationRepo          ports.NotificationRepository
	OutboxRepo                ports.OutboxRepository
	NotificationSender        ports.NotificationSender
	// RealNow es el reloj real para medir duraciones en el log. Default: Clock.Now.
	RealNow func() time.Time
}

type router struct {
	mux    *http.ServeMux
	deps   Deps
	routes []string
}

func (rt *router) public(pattern string, h http.HandlerFunc) {
	rt.routes = append(rt.routes, pattern)
	rt.mux.Handle(pattern, h)
}

func (rt *router) accountant(pattern string, h http.HandlerFunc) {
	rt.routes = append(rt.routes, pattern)
	rt.mux.Handle(pattern, RequireAccountant(h))
}

func (rt *router) client(pattern string, h http.HandlerFunc) {
	rt.routes = append(rt.routes, pattern)
	rt.mux.Handle(pattern, RequireClient(h))
}

// Routes devuelve las rutas registradas en el router de producción.
func Routes(deps Deps) []string {
	rt := newRouter(deps)
	out := make([]string, len(rt.routes))
	copy(out, rt.routes)
	return out
}

// NewRouter arma el handler HTTP con todas las rutas y middlewares.
func NewRouter(deps Deps) http.Handler {
	rt := newRouter(deps)
	return rt.handler()
}

func newRouter(deps Deps) *router {
	if deps.Logger == nil {
		deps.Logger = slog.New(slog.NewJSONHandler(io.Discard, nil))
	}
	if deps.RealNow == nil {
		deps.RealNow = deps.Clock.Now
	}
	rt := &router{mux: http.NewServeMux(), deps: deps}

	rt.public("GET /api/v1/health", rt.health)
	rt.mux.HandleFunc("GET /api/openapi.yaml", serveOpenAPI)
	rt.mux.HandleFunc("GET /docs", serveDocs)
	rt.mux.HandleFunc("GET /favicon.svg", serveFavicon)
	rt.accountant("GET /api/v1/accountant/signals/catalog", signalsCatalog)

	var computeClosing *usecase.ComputeClosing
	if deps.ClientRepo != nil && deps.MovementRepo != nil && deps.ClassificationRepo != nil &&
		deps.InvoiceRepo != nil && deps.PurchaseInvoiceRepo != nil && deps.ClientInputRepo != nil &&
		deps.ParametersRepo != nil && deps.ClosingRepo != nil && deps.TaxScales != nil &&
		deps.DueDates != nil && deps.ExchangeRates != nil && deps.IDGen != nil {
		computeClosing = &usecase.ComputeClosing{
			Clients:          deps.ClientRepo,
			Movements:        deps.MovementRepo,
			Classifications:  deps.ClassificationRepo,
			Invoices:         deps.InvoiceRepo,
			PurchaseInvoices: deps.PurchaseInvoiceRepo,
			ClientInputs:     deps.ClientInputRepo,
			Parameters:       deps.ParametersRepo,
			Closings:         deps.ClosingRepo,
			TaxScales:        deps.TaxScales,
			DueDates:         deps.DueDates,
			ExchangeRates:    deps.ExchangeRates,
			Clock:            deps.Clock,
			IDGen:            deps.IDGen,
			Tx:               deps.TxManager,
		}
	}

	if deps.ParametersRepo != nil {
		var recalcPeriod *usecase.RecalcPeriod
		if deps.ClosingRepo != nil && computeClosing != nil {
			recalcPeriod = &usecase.RecalcPeriod{
				Closings: deps.ClosingRepo,
				Compute:  computeClosing,
			}
		}
		paramH := &parametersHandler{
			getUsecase: &usecase.GetParameters{Repo: deps.ParametersRepo},
			updateUsecase: &usecase.UpdateParameters{
				Repo:         deps.ParametersRepo,
				Clock:        deps.Clock,
				RecalcPeriod: recalcPeriod,
			},
		}
		rt.accountant("GET /api/v1/accountant/parameters", paramH.get)
		rt.accountant("PUT /api/v1/accountant/parameters", paramH.update)
	}

	if deps.ClosingRepo != nil && deps.ClientRepo != nil {
		var dispatchOutbox *usecase.DispatchOutbox
		if deps.OutboxRepo != nil && deps.NotificationSender != nil && deps.Clock != nil {
			dispatchOutbox = &usecase.DispatchOutbox{
				Outbox: deps.OutboxRepo,
				Sender: deps.NotificationSender,
				Clock:  deps.Clock,
			}
		}

		var actionNotices *usecase.CreateActionRequiredNotices
		if deps.NotificationRepo != nil && deps.OutboxRepo != nil && deps.IDGen != nil && deps.Clock != nil {
			actionNotices = &usecase.CreateActionRequiredNotices{
				Closings:       deps.ClosingRepo,
				Notifications:  deps.NotificationRepo,
				Outbox:         deps.OutboxRepo,
				Clock:          deps.Clock,
				IDGen:          deps.IDGen,
				DispatchOutbox: dispatchOutbox,
			}
		}

		var approveUC *usecase.ApproveClosing
		var approveBatchUC *usecase.ApproveBatch
		var sendRemindersUC *usecase.SendPaymentReminders

		if deps.NotificationRepo != nil && deps.OutboxRepo != nil && deps.IDGen != nil && deps.Clock != nil {
			approveUC = &usecase.ApproveClosing{
				Closings:       deps.ClosingRepo,
				Clients:        deps.ClientRepo,
				Notifications:  deps.NotificationRepo,
				Outbox:         deps.OutboxRepo,
				Sender:         deps.NotificationSender,
				Clock:          deps.Clock,
				IDGen:          deps.IDGen,
				DispatchOutbox: dispatchOutbox,
				Tx:             deps.TxManager,
			}

			approveBatchUC = &usecase.ApproveBatch{
				ApproveClosing: approveUC,
				Closings:       deps.ClosingRepo,
				NumWorkers:     deps.Workers,
				Tx:             deps.TxManager,
			}

			sendRemindersUC = &usecase.SendPaymentReminders{
				Closings:       deps.ClosingRepo,
				Notifications:  deps.NotificationRepo,
				Outbox:         deps.OutboxRepo,
				Parameters:     deps.ParametersRepo,
				Clock:          deps.Clock,
				IDGen:          deps.IDGen,
				DispatchOutbox: dispatchOutbox,
			}
		}

		closH := &closingsHandler{
			closingQuery: deps.ClosingQuery,
			runClosingUC: &usecase.RunMonthlyClosing{
				Clients:       deps.ClientRepo,
				Compute:       computeClosing,
				Workers:       deps.Workers,
				ActionNotices: actionNotices,
			},
			listClosingsUC: &usecase.ListClosings{
				Query: deps.ClosingQuery,
			},
			listClosingIDsUC: &usecase.ListClosingIDs{
				Query: deps.ClosingQuery,
			},
			closingsSummaryUC: &usecase.ClosingsSummary{
				Query: deps.ClosingQuery,
			},
			getStatsUC: &usecase.GetPeriodStats{
				Query: deps.StatsQuery,
			},
			getDetailUC: &usecase.GetClosingDetail{
				Closings:        deps.ClosingRepo,
				Clients:         deps.ClientRepo,
				Movements:       deps.MovementRepo,
				Classifications: deps.ClassificationRepo,
				Invoices:        deps.InvoiceRepo,
				ClientInputs:    deps.ClientInputRepo,
				Parameters:      deps.ParametersRepo,
				ExchangeRates:   deps.ExchangeRates,
			},
			approveUC:       approveUC,
			approveBatchUC:  approveBatchUC,
			sendRemindersUC: sendRemindersUC,
			clock:           deps.Clock,
		}
		rt.accountant("GET /api/v1/accountant/periods", closH.periods)
		rt.accountant("POST /api/v1/accountant/periods/{period}/run", closH.run)
		rt.accountant("GET /api/v1/accountant/periods/{period}/closings/ids", closH.ids)
		rt.accountant("GET /api/v1/accountant/periods/{period}/closings", closH.list)
		rt.accountant("GET /api/v1/accountant/periods/{period}/closings/summary", closH.summary)
		if deps.StatsQuery != nil {
			rt.accountant("GET /api/v1/accountant/periods/{period}/stats", closH.stats)
		}
		rt.accountant("GET /api/v1/accountant/closings/{id}", closH.detail)
		if approveUC != nil {
			rt.accountant("POST /api/v1/accountant/closings/{id}/approve", closH.approve)
		}
		if approveBatchUC != nil {
			rt.accountant("POST /api/v1/accountant/periods/{period}/closings/approve-batch", closH.approveBatch)
		}
		if sendRemindersUC != nil {
			rt.accountant("POST /api/v1/accountant/reminders/run", closH.runReminders)
		}
	}

	if deps.MovementRepo != nil && deps.ClassificationRepo != nil {
		movH := &movementsHandler{
			overrideUsecase: &usecase.OverrideClassification{
				Movements:       deps.MovementRepo,
				Classifications: deps.ClassificationRepo,
				Clock:           deps.Clock,
				Notifier:        deps.ChangeNotifier,
			},
		}
		rt.accountant("PATCH /api/v1/accountant/movements/{id}/classification", movH.overrideClassification)
	}

	if deps.InvoiceRepo != nil && deps.ClientRepo != nil {
		invH := &invoicesHandler{
			importUC: &usecase.ImportInvoicesCSV{
				Clients:  deps.ClientRepo,
				Invoices: deps.InvoiceRepo,
				IDGen:    deps.IDGen,
				Clock:    deps.Clock,
				Notifier: deps.ChangeNotifier,
				Tx:       deps.TxManager,
			},
			addInvoiceUC: &usecase.AddInvoice{
				Clients:  deps.ClientRepo,
				Invoices: deps.InvoiceRepo,
				IDGen:    deps.IDGen,
				Clock:    deps.Clock,
				Notifier: deps.ChangeNotifier,
			},
			listInvoicesUC: &usecase.ListInvoices{Invoices: deps.InvoiceRepo},
			addPurchaseUC: &usecase.AddPurchaseInvoice{
				Clients:          deps.ClientRepo,
				PurchaseInvoices: deps.PurchaseInvoiceRepo,
				IDGen:            deps.IDGen,
				Clock:            deps.Clock,
				Notifier:         deps.ChangeNotifier,
			},
			setExternalIncomeUC: &usecase.SetExternalIncome{
				Inputs:   deps.ClientInputRepo,
				Clock:    deps.Clock,
				Notifier: deps.ChangeNotifier,
			},
			confirmPeriodUC: &usecase.ConfirmPeriodInput{
				Inputs:   deps.ClientInputRepo,
				Clock:    deps.Clock,
				Notifier: deps.ChangeNotifier,
			},
		}

		rt.client("POST /api/v1/me/invoices/import", invH.importCSV)
		rt.client("POST /api/v1/me/invoices", invH.createInvoice)
		rt.client("GET /api/v1/me/invoices", invH.listMeInvoices)
		rt.accountant("GET /api/v1/accountant/clients/{id}/invoices", invH.listClientInvoicesAccountant)
		rt.client("POST /api/v1/me/purchase-invoices", invH.createPurchaseInvoice)
		rt.client("PUT /api/v1/me/periods/{period}/external-income", invH.setExternalIncome)
		rt.client("POST /api/v1/me/periods/{period}/confirm", invH.confirmPeriod)
	}

	if deps.ClientRepo != nil {
		regimeH := &regimeHandler{
			changeRegimeUC: &usecase.ChangeRegime{
				Clients:    deps.ClientRepo,
				Exclusions: deps.LatestExclusionDateReader,
				Clock:      deps.Clock,
				Notifier:   deps.ChangeNotifier,
				Tx:         deps.TxManager,
			},
			updateRIProfileUC: &usecase.UpdateRIProfile{
				Clients:  deps.ClientRepo,
				Clock:    deps.Clock,
				Notifier: deps.ChangeNotifier,
			},
		}
		rt.accountant("POST /api/v1/accountant/clients/{id}/regime-change", regimeH.changeRegime)
		rt.accountant("PUT /api/v1/accountant/clients/{id}/ri-profile", regimeH.updateRIProfile)

		cliH := &clientHandler{
			getMeUC: &usecase.GetMe{Clients: deps.ClientRepo},
		}
		if deps.ClosingRepo != nil && deps.ClientInputRepo != nil && deps.InvoiceRepo != nil {
			cliH.listMyClosingsUC = &usecase.ListMyClosings{
				Closings:     deps.ClosingRepo,
				ClientInputs: deps.ClientInputRepo,
				Invoices:     deps.InvoiceRepo,
			}
		}
		if deps.ClosingRepo != nil && deps.Clock != nil {
			cliH.markClosingPaidUC = &usecase.MarkClosingPaid{
				Closings: deps.ClosingRepo,
				Clock:    deps.Clock,
			}
		}
		if deps.NotificationRepo != nil {
			cliH.listMyNotificationsUC = &usecase.ListMyNotifications{Notifications: deps.NotificationRepo}
		}
		if deps.NotificationRepo != nil && deps.Clock != nil {
			cliH.markNotificationReadUC = &usecase.MarkNotificationRead{
				Notifications: deps.NotificationRepo,
				Clock:         deps.Clock,
			}
		}

		rt.client("GET /api/v1/me", cliH.getMe)
		if cliH.listMyClosingsUC != nil {
			rt.client("GET /api/v1/me/closings", cliH.listMyClosings)
		}
		if cliH.markClosingPaidUC != nil {
			rt.client("POST /api/v1/me/closings/{id}/paid", cliH.markPaid)
		}
		if cliH.listMyNotificationsUC != nil {
			rt.client("GET /api/v1/me/notifications", cliH.listNotifications)
		}
		if cliH.markNotificationReadUC != nil {
			rt.client("POST /api/v1/me/notifications/{id}/read", cliH.markNotificationRead)
		}
	}

	rt.mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		writeErrorCode(w, http.StatusNotFound, "NOT_FOUND", "Ruta inexistente", nil)
	})

	return rt
}

func (rt *router) handler() http.Handler {
	return Chain(rt.mux,
		Recover(rt.deps.Logger),
		RequestLog(rt.deps.Logger, rt.deps.RealNow),
		CORS(rt.deps.CORSOrigin),
		DemoIdentity,
	)
}

func (rt *router) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"now":    rt.deps.Clock.Now().Format(time.RFC3339),
	})
}

func serveOpenAPI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(api.Spec)
}

func serveDocs(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(api.DocsHTML)
}

func serveFavicon(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(api.FaviconSVG)
}
