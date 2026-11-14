package http

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/api"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports/fakes"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/clock"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/memory"
)

func fullTestDeps() Deps {
	clk := clock.NewFixed(testNow)
	clientRepo := memory.NewClientRepository()
	movRepo := memory.NewMovementRepository()
	classRepo := memory.NewClassificationRepository()
	invRepo := memory.NewInvoiceRepository()
	purchRepo := memory.NewPurchaseInvoiceRepository()
	inputRepo := memory.NewClientInputRepository()
	paramRepo := memory.NewParametersRepository()
	closingRepo := memory.NewClosingRepository()
	notifRepo := memory.NewNotificationRepository()
	outboxRepo := memory.NewOutboxRepository()

	return Deps{
		Clock:                     clk,
		ParametersRepo:            paramRepo,
		ClientRepo:                clientRepo,
		MovementRepo:              movRepo,
		ClassificationRepo:        classRepo,
		InvoiceRepo:               invRepo,
		PurchaseInvoiceRepo:       purchRepo,
		ClientInputRepo:           inputRepo,
		LatestExclusionDateReader: &fakes.LatestExclusionDateReader{},
		ClosingRepo:               closingRepo,
		ClosingQuery:              memory.NewClosingQuery(closingRepo, clientRepo),
		StatsQuery:                memory.NewStatsQuery(closingRepo),
		TxManager:                 memory.NewTxManager(),
		TaxScales:                 &fakes.TaxScaleProvider{},
		DueDates:                  &fakes.DueDateProvider{},
		ExchangeRates:             &fakes.ExchangeRateProvider{},
		IDGen:                     &fakes.IDGenerator{},
		ChangeNotifier:            &ports.NoOpChangeNotifier{},
		NotificationRepo:          notifRepo,
		OutboxRepo:                outboxRepo,
		NotificationSender:        &fakes.NotificationSender{},
	}
}

func TestOpenAPIAndDocsEndpoints(t *testing.T) {
	h := NewRouter(testDeps())

	rec := do(t, h, "GET", "/api/openapi.yaml", "", nil)
	if rec.Code != 200 {
		t.Fatalf("GET /api/openapi.yaml status %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "application/yaml") {
		t.Fatalf("unexpected Content-Type for openapi.yaml: %s", ct)
	}
	if !strings.Contains(rec.Body.String(), "openapi: 3.0.3") {
		t.Fatalf("openapi.yaml body missing openapi header")
	}

	recDocs := do(t, h, "GET", "/docs", "", nil)
	if recDocs.Code != 200 {
		t.Fatalf("GET /docs status %d", recDocs.Code)
	}
	if ct := recDocs.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Fatalf("unexpected Content-Type for /docs: %s", ct)
	}
	if !strings.Contains(recDocs.Body.String(), "swagger-ui") {
		t.Fatalf("/docs body missing swagger-ui")
	}
	if !strings.Contains(recDocs.Body.String(), "/api/openapi.yaml") {
		t.Fatalf("/docs body missing /api/openapi.yaml")
	}

	recFavicon := do(t, h, "GET", "/favicon.svg", "", nil)
	if recFavicon.Code != 200 {
		t.Fatalf("GET /favicon.svg status %d", recFavicon.Code)
	}
	if ct := recFavicon.Header().Get("Content-Type"); !strings.Contains(ct, "image/svg+xml") {
		t.Fatalf("unexpected Content-Type for /favicon.svg: %s", ct)
	}

	rootDocs, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "docs", "index.html"))
	if err == nil && string(rootDocs) != string(api.DocsHTML) {
		t.Fatalf("backend/api/docs.html está desincronizado con docs/index.html")
	}
}

func TestRouterMatchesOpenAPI(t *testing.T) {
	pathRe := regexp.MustCompile(`^  (/api/v1/[^:]+):$`)
	methodRe := regexp.MustCompile(`^    (get|post|put|patch|delete):$`)

	specRoutes := map[string]bool{}
	var currentPath string
	for _, line := range strings.Split(string(api.Spec), "\n") {
		if m := pathRe.FindStringSubmatch(line); m != nil {
			currentPath = m[1]
			continue
		}
		if currentPath != "" {
			if m := methodRe.FindStringSubmatch(line); m != nil {
				specRoutes[strings.ToUpper(m[1])+" "+currentPath] = true
			} else if len(line) > 0 && line[0] != ' ' {
				currentPath = ""
			}
		}
	}

	routerRoutes := map[string]bool{}
	for _, r := range Routes(fullTestDeps()) {
		routerRoutes[r] = true
	}

	for r := range routerRoutes {
		if !specRoutes[r] {
			t.Errorf("ruta registrada en el router pero faltante en openapi.yaml: %s", r)
		}
	}
	for r := range specRoutes {
		if !routerRoutes[r] {
			t.Errorf("ruta documentada en openapi.yaml pero faltante en el router: %s", r)
		}
	}
}
