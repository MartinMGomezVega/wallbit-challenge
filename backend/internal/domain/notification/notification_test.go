package notification_test

import (
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/closing"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/notification"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/signal"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/tax"
)

func TestBuildClosingMessage_Monotributo(t *testing.T) {
	p := period.New(2026, time.October)
	cl := client.Client{
		ID:     "cli_01",
		Name:   "Ignacio Sosa",
		CUIT:   "20-28222221-1",
		Regime: client.RegimeMonotributo,
	}

	feeDue := time.Date(2026, time.November, 20, 0, 0, 0, 0, period.Location())
	c := closing.Closing{
		ID:       "clo_01",
		ClientID: "cli_01",
		Period:   p,
		Status:   closing.StatusReadyForReview,
		Mono: &tax.MonotributoAssessment{
			BilledInPeriodUSD: money.New(150000, money.USD),    // 1.500,00
			BilledInPeriodARS: money.New(225000000, money.ARS), // 2.250.000,00
			CurrentCategory:   tax.CatK,
			Fee:               money.New(16144460, money.ARS), // 161.444,60
			FeeDueDate:        feeDue,
		},
		Signals: []signal.Signal{
			{
				Code: signal.CodeMonoExclusionKCapExceeded,
			},
		},
	}

	msg := notification.BuildClosingMessage(c, cl, "Revisar los comprobantes adjuntos")

	expectedTitle := "Tu cierre de octubre 2026 está listo"
	expectedBody := "Facturaste US$ 1.500,00 ($ 2.250.000,00). Cuota de Monotributo categoría K: $ 161.444,60, vence el 20/11/2026. Superaste el tope del Monotributo. Tu contador te va a contactar para el pase a Responsable Inscripto. Nota de tu contador: Revisar los comprobantes adjuntos"

	if msg.Title != expectedTitle {
		t.Errorf("Title = %q, want %q", msg.Title, expectedTitle)
	}
	if msg.Body != expectedBody {
		t.Errorf("Body = %q, want %q", msg.Body, expectedBody)
	}
}

func TestBuildClosingMessage_RI_WithVatToPayAndAnticipo(t *testing.T) {
	p := period.New(2026, time.October)
	cl := client.Client{
		ID:     "cli_02",
		Name:   "Lucía Méndez",
		CUIT:   "27-33333333-3",
		Regime: client.RegimeRI,
	}

	vatDue := time.Date(2026, time.November, 18, 0, 0, 0, 0, period.Location())
	autDue := time.Date(2026, time.November, 10, 0, 0, 0, 0, period.Location())
	anticipoDue := time.Date(2026, time.November, 25, 0, 0, 0, 0, period.Location())

	anticipo := money.New(5000000, money.ARS) // 50.000,00

	c := closing.Closing{
		ID:       "clo_02",
		ClientID: "cli_02",
		Period:   p,
		Status:   closing.StatusReadyForReview,
		RI: &tax.RIAssessment{
			BilledInPeriodUSD:        money.New(200000, money.USD),    // 2.000,00
			BilledInPeriodARS:        money.New(300000000, money.ARS), // 3.000.000,00
			VatToPay:                 money.New(12000000, money.ARS),  // 120.000,00
			VatDueDate:               vatDue,
			SelfEmployedContribution: money.New(4500000, money.ARS), // 45.000,00
			SelfEmployedDueDate:      autDue,
			IncomeTaxAdvance:         &anticipo,
			IncomeTaxAdvanceDueDate:  &anticipoDue,
			TotalToPay:               money.New(21500000, money.ARS), // 215.000,00
		},
	}

	msg := notification.BuildClosingMessage(c, cl)

	expectedTitle := "Tu cierre de octubre 2026 está listo"
	expectedBody := "Facturaste US$ 2.000,00 ($ 3.000.000,00). IVA a pagar: $ 120.000,00 (vence 18/11/2026). Aportes de autónomos: $ 45.000,00 (vence 10/11/2026). Anticipo Ganancias: $ 50.000,00 (vence 25/11/2026). Total: $ 215.000,00."

	if msg.Title != expectedTitle {
		t.Errorf("Title = %q, want %q", msg.Title, expectedTitle)
	}
	if msg.Body != expectedBody {
		t.Errorf("Body = %q, want %q", msg.Body, expectedBody)
	}
}

func TestBuildClosingMessage_RI_ZeroVatWithCredit(t *testing.T) {
	p := period.New(2026, time.October)
	cl := client.Client{
		ID:     "cli_02",
		Name:   "Lucía Méndez",
		CUIT:   "27-33333333-3",
		Regime: client.RegimeRI,
	}

	vatDue := time.Date(2026, time.November, 18, 0, 0, 0, 0, period.Location())
	autDue := time.Date(2026, time.November, 10, 0, 0, 0, 0, period.Location())

	c := closing.Closing{
		ID:       "clo_02",
		ClientID: "cli_02",
		Period:   p,
		Status:   closing.StatusReadyForReview,
		RI: &tax.RIAssessment{
			BilledInPeriodUSD:        money.New(100000, money.USD),    // 1.000,00
			BilledInPeriodARS:        money.New(150000000, money.ARS), // 1.500.000,00
			VatToPay:                 money.Zero(money.ARS),
			NewVatCredit:             money.New(3500000, money.ARS), // 35.000,00
			VatDueDate:               vatDue,
			SelfEmployedContribution: money.New(4500000, money.ARS), // 45.000,00
			SelfEmployedDueDate:      autDue,
			IncomeTaxAdvance:         nil,
			TotalToPay:               money.New(4500000, money.ARS), // 45.000,00
		},
	}

	msg := notification.BuildClosingMessage(c, cl)

	expectedBody := "Facturaste US$ 1.000,00 ($ 1.500.000,00). No tenés IVA a pagar; te queda un saldo a favor de $ 35.000,00. Aportes de autónomos: $ 45.000,00 (vence 10/11/2026). Total: $ 45.000,00."
	if msg.Body != expectedBody {
		t.Errorf("Body = %q, want %q", msg.Body, expectedBody)
	}
}

func TestBuildClosingMessage_Signals(t *testing.T) {
	p := period.New(2026, time.October)
	cl := client.Client{
		ID:     "cli_01",
		Name:   "Paula Rojas",
		Regime: client.RegimeMonotributo,
	}

	feeDue := time.Date(2026, time.November, 20, 0, 0, 0, 0, period.Location())
	c := closing.Closing{
		ID:       "clo_03",
		ClientID: "cli_01",
		Period:   p,
		Mono: &tax.MonotributoAssessment{
			BilledInPeriodUSD: money.Zero(money.USD),
			BilledInPeriodARS: money.Zero(money.ARS),
			CurrentCategory:   tax.CatE,
			Fee:               money.New(1200000, money.ARS),
			FeeDueDate:        feeDue,
		},
		Signals: []signal.Signal{
			{
				Code: signal.CodeMonoCategoryChangeRequired,
				Data: map[string]string{"required": "F"},
			},
			{
				Code: signal.CodeMonoCategoryDowngradePossible,
				Data: map[string]string{"required": "D"},
			},
			{
				Code: signal.CodeMissingInvoices,
				Data: map[string]string{"count": "3"},
			},
		},
	}

	msg := notification.BuildClosingMessage(c, cl)

	expectedSnippets := []string{
		"En la próxima recategorización tenés que pasar a la categoría F.",
		"Buenas noticias: en la próxima recategorización podrías bajar a la categoría D.",
		"Tenés 3 cobros sin factura. Emitilas y subilas para que quede todo en orden.",
	}

	for _, snippet := range expectedSnippets {
		if !contains(msg.Body, snippet) {
			t.Errorf("Body expected to contain %q, but got %q", snippet, msg.Body)
		}
	}
}

func TestContainsAmounts(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"$ 1.500,00", true},
		{"$ 100", true},
		{"US$ 50", true},
		{"Facturaste 1.500 pesos", true},
		{"Total a pagar: 1.234.567,89", true},
		{"Tu cierre de octubre 2026 está listo. Entrá a Wallbit para verlo.", false},
		{"Tenés novedades sobre tus impuestos en Wallbit", false},
		{"Subí tus facturas de octubre 2026 y contanos si cobraste algo fuera de Wallbit.", false},
		{"Recordá que se acerca el vencimiento de tus impuestos de octubre 2026. Entrá a Wallbit para ver el detalle.", false},
	}

	for _, tc := range tests {
		got := notification.ContainsAmounts(tc.input)
		if got != tc.want {
			t.Errorf("ContainsAmounts(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestOutboundFor_NeverContainsAmounts(t *testing.T) {
	kinds := []notification.Kind{
		notification.KindActionRequired,
		notification.KindClosingReady,
		notification.KindClosingUpdated,
		notification.KindPaymentReminder,
	}

	now := time.Date(2026, 10, 1, 10, 0, 0, 0, time.UTC)
	for _, k := range kinds {
		msgs := notification.OutboundFor("cli_01", "clo_01", k, "octubre 2026", now)
		if len(msgs) != 2 {
			t.Fatalf("expected 2 messages (email and push), got %d", len(msgs))
		}

		for _, m := range msgs {
			if notification.ContainsAmounts(m.Subject) {
				t.Errorf("Subject contains amounts: %q", m.Subject)
			}
			if notification.ContainsAmounts(m.Body) {
				t.Errorf("Body contains amounts: %q", m.Body)
			}
			if !m.CreatedAt.Equal(now) {
				t.Errorf("expected CreatedAt %v, got %v", now, m.CreatedAt)
			}
		}
	}
}

func TestBuildClosingMessage_NoteFromApproval(t *testing.T) {
	p := period.New(2026, time.October)
	cl := client.Client{
		ID:     "cli_01",
		Name:   "Paula Rojas",
		Regime: client.RegimeMonotributo,
	}

	feeDue := time.Date(2026, time.November, 20, 0, 0, 0, 0, period.Location())
	c := closing.Closing{
		ID:       "clo_03",
		ClientID: "cli_01",
		Period:   p,
		Mono: &tax.MonotributoAssessment{
			BilledInPeriodUSD: money.New(-5000, money.USD), // negative
			BilledInPeriodARS: money.New(-7500000, money.ARS),
			CurrentCategory:   tax.CatE,
			Fee:               money.New(1200000, money.ARS),
			FeeDueDate:        feeDue,
		},
		Approvals: []closing.Approval{
			{
				Note: "Nota aprobada previamente",
			},
		},
	}

	msg := notification.BuildClosingMessage(c, cl)
	if !contains(msg.Body, "Nota de tu contador: Nota aprobada previamente") {
		t.Errorf("expected body to contain approval note, got: %s", msg.Body)
	}

	// MonthNameES out of range
	invalidMonth := notification.MonthNameES(time.Month(15))
	if invalidMonth == "" {
		t.Errorf("expected fallback month name")
	}

	// Outbound default kind
	now := time.Date(2026, 10, 1, 10, 0, 0, 0, time.UTC)
	defMsgs := notification.OutboundFor("cli_01", "clo_01", notification.Kind("OTHER"), "octubre 2026", now)
	if len(defMsgs) != 2 {
		t.Fatalf("expected 2 default messages")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && len(substr) > 0 && searchSubstring(s, substr))
}

func searchSubstring(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
