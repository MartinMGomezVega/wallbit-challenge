// Package notification implementa los modelos, reglas de privacidad y generación de
// mensajes para avisos in-app y salientes según specs/010-approval-notifications/spec.md.
package notification

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/closing"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/signal"
)

// Kind representa el tipo de notificación.
type Kind string

const (
	KindActionRequired  Kind = "ACTION_REQUIRED"
	KindClosingReady    Kind = "CLOSING_READY"
	KindClosingUpdated  Kind = "CLOSING_UPDATED"
	KindPaymentReminder Kind = "PAYMENT_REMINDER"
)

// Channel representa el canal de un mensaje saliente.
type Channel string

const (
	ChannelEmail Channel = "EMAIL"
	ChannelPush  Channel = "PUSH"
)

// Notification representa una notificación in-app para el cliente.
type Notification struct {
	ID        string     `json:"id"`
	ClientID  string     `json:"clientId"`
	ClosingID string     `json:"closingId,omitempty"`
	Kind      Kind       `json:"kind"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	Key       string     `json:"key,omitempty"` // Para deduplicación (ej. recordatorios por vencimiento)
	CreatedAt time.Time  `json:"createdAt"`
	ReadAt    *time.Time `json:"readAt,omitempty"`
}

// OutboundMessage representa un mensaje saliente (email/push) gestionado mediante outbox.
type OutboundMessage struct {
	ID        string     `json:"id"`
	Channel   Channel    `json:"channel"`
	ClientID  string     `json:"clientId"`
	ClosingID string     `json:"closingId,omitempty"`
	Subject   string     `json:"subject"`
	Body      string     `json:"body"`
	Attempts  int        `json:"attempts"`
	SentAt    *time.Time `json:"sentAt,omitempty"`
	LastError string     `json:"lastError,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
}

// Message representa un título y cuerpo de aviso generado para el cliente.
type Message struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

var spanishMonths = [...]string{
	"", "enero", "febrero", "marzo", "abril", "mayo", "junio",
	"julio", "agosto", "septiembre", "octubre", "noviembre", "diciembre",
}

// MonthNameES devuelve el nombre del mes en español (1 -> "enero", 10 -> "octubre").
func MonthNameES(m time.Month) string {
	if int(m) >= 1 && int(m) <= 12 {
		return spanishMonths[m]
	}
	return m.String()
}

// FormatPeriodLabel formatea un período como "octubre 2026".
func FormatPeriodLabel(p period.Period) string {
	return fmt.Sprintf("%s %d", MonthNameES(p.Month), p.Year)
}

// formatNumberES formatea centavos a formato "1.234,56".
func formatNumberES(cents int64) string {
	negative := cents < 0
	if negative {
		cents = -cents
	}

	integerPart := cents / 100
	decimalPart := cents % 100

	intStr := strconv.FormatInt(integerPart, 10)
	var b strings.Builder
	l := len(intStr)
	for i, ch := range intStr {
		b.WriteRune(ch)
		remaining := l - 1 - i
		if remaining > 0 && remaining%3 == 0 {
			b.WriteRune('.')
		}
	}

	res := fmt.Sprintf("%s,%02d", b.String(), decimalPart)
	if negative {
		return "-" + res
	}
	return res
}

func formatARS(m money.Money) string {
	if m.Amount < 0 {
		return "-$ " + formatNumberES(-m.Amount)
	}
	return "$ " + formatNumberES(m.Amount)
}

func formatDate(t time.Time) string {
	return t.In(period.Location()).Format("02/01/2006")
}

// BuildClosingMessage genera el título y cuerpo del aviso para el cliente según el criterio 4 de la spec 010.
func BuildClosingMessage(c closing.Closing, cl client.Client, clientNote ...string) Message {
	monthLabel := FormatPeriodLabel(c.Period)
	title := fmt.Sprintf("Tu cierre de %s está listo", monthLabel)

	var note string
	if len(clientNote) > 0 && strings.TrimSpace(clientNote[0]) != "" {
		note = strings.TrimSpace(clientNote[0])
	} else if len(c.Approvals) > 0 {
		lastApproval := c.Approvals[len(c.Approvals)-1]
		if strings.TrimSpace(lastApproval.Note) != "" {
			note = strings.TrimSpace(lastApproval.Note)
		}
	}

	var sentences []string

	if cl.Regime == client.RegimeMonotributo && c.Mono != nil {
		usdStr := formatNumberES(c.Mono.BilledInPeriodUSD.Amount)
		arsStr := formatARS(c.Mono.BilledInPeriodARS)
		catStr := string(c.Mono.CurrentCategory)
		feeStr := formatARS(c.Mono.Fee)
		dueStr := formatDate(c.Mono.FeeDueDate)

		monoBase := fmt.Sprintf("Facturaste US$ %s (%s). Cuota de Monotributo categoría %s: %s, vence el %s.",
			usdStr, arsStr, catStr, feeStr, dueStr)
		sentences = append(sentences, monoBase)
	} else if cl.Regime == client.RegimeRI && c.RI != nil {
		usdStr := formatNumberES(c.RI.BilledInPeriodUSD.Amount)
		arsStr := formatARS(c.RI.BilledInPeriodARS)

		var vatPart string
		if c.RI.VatToPay.IsZero() && c.RI.NewVatCredit.IsPositive() {
			vatPart = fmt.Sprintf("No tenés IVA a pagar; te queda un saldo a favor de %s.",
				formatARS(c.RI.NewVatCredit))
		} else {
			vatPart = fmt.Sprintf("IVA a pagar: %s (vence %s).",
				formatARS(c.RI.VatToPay), formatDate(c.RI.VatDueDate))
		}

		autPart := fmt.Sprintf("Aportes de autónomos: %s (vence %s).",
			formatARS(c.RI.SelfEmployedContribution), formatDate(c.RI.SelfEmployedDueDate))

		var anticipoPart string
		if c.RI.IncomeTaxAdvance != nil && c.RI.IncomeTaxAdvance.IsPositive() {
			due := c.RI.SelfEmployedDueDate
			if c.RI.IncomeTaxAdvanceDueDate != nil {
				due = *c.RI.IncomeTaxAdvanceDueDate
			}
			anticipoPart = fmt.Sprintf("Anticipo Ganancias: %s (vence %s).",
				formatARS(*c.RI.IncomeTaxAdvance), formatDate(due))
		}

		totalPart := fmt.Sprintf("Total: %s.", formatARS(c.RI.TotalToPay))

		riParts := []string{fmt.Sprintf("Facturaste US$ %s (%s).", usdStr, arsStr), vatPart, autPart}
		if anticipoPart != "" {
			riParts = append(riParts, anticipoPart)
		}
		riParts = append(riParts, totalPart)
		sentences = append(sentences, strings.Join(riParts, " "))
	}

	// Agregados por señal (solo códigos específicos)
	for _, sig := range c.Signals {
		switch sig.Code {
		case signal.CodeMonoExclusionKCapExceeded:
			sentences = append(sentences, "Superaste el tope del Monotributo. Tu contador te va a contactar para el pase a Responsable Inscripto.")

		case signal.CodeMonoCategoryChangeRequired:
			req := sig.Data["required"]
			if req == "" && c.Mono != nil && c.Mono.RequiredCategory != nil {
				req = string(*c.Mono.RequiredCategory)
			}
			if req != "" {
				sentences = append(sentences, fmt.Sprintf("En la próxima recategorización tenés que pasar a la categoría %s.", req))
			}

		case signal.CodeMonoCategoryDowngradePossible:
			req := sig.Data["required"]
			if req == "" && c.Mono != nil && c.Mono.RequiredCategory != nil {
				req = string(*c.Mono.RequiredCategory)
			}
			if req != "" {
				sentences = append(sentences, fmt.Sprintf("Buenas noticias: en la próxima recategorización podrías bajar a la categoría %s.", req))
			}

		case signal.CodeMissingInvoices:
			count := sig.Data["count"]
			if count == "" {
				count = fmt.Sprintf("%d", len(c.Reconciliation.UnmatchedIncomeIDs))
			}
			sentences = append(sentences, fmt.Sprintf("Tenés %s cobros sin factura. Emitilas y subilas para que quede todo en orden.", count))
		}
	}

	// Nota del contador si existe
	if note != "" {
		sentences = append(sentences, fmt.Sprintf("Nota de tu contador: %s", note))
	}

	return Message{
		Title: title,
		Body:  strings.Join(sentences, " "),
	}
}

// amountRegex detecta signos monetarios o cifras con formato de moneda.
// Se usa para verificar la garantía de privacidad en mensajes salientes (criterio 7).
var amountRegex = regexp.MustCompile(`(\$|US\$|\d{1,3}(\.\d{3})+(,\d{2})?)`)

// ContainsAmounts verifica si un texto contiene montos o formatos monetarios.
func ContainsAmounts(s string) bool {
	return amountRegex.MatchString(s)
}

// OutboundFor genera los mensajes salientes (EMAIL y PUSH) con textos fijos sin montos (criterio 7).
func OutboundFor(clientID string, closingID string, kind Kind, monthLabel string, createdAt time.Time) []OutboundMessage {
	var subject, body string

	switch kind {
	case KindActionRequired:
		subject = fmt.Sprintf("Acción requerida: cierre de %s", monthLabel)
		body = fmt.Sprintf("Subí tus facturas de %s y contanos si cobraste algo fuera de Wallbit.", monthLabel)

	case KindClosingReady:
		subject = "Tenés novedades sobre tus impuestos en Wallbit"
		body = fmt.Sprintf("Tu cierre de %s está listo. Entrá a Wallbit para verlo.", monthLabel)

	case KindClosingUpdated:
		subject = fmt.Sprintf("Actualizamos tu cierre de %s", monthLabel)
		body = fmt.Sprintf("Actualizamos tu cierre de %s. Entrá a Wallbit para verlo.", monthLabel)

	case KindPaymentReminder:
		subject = "Recordatorio de vencimiento impositivo"
		body = fmt.Sprintf("Recordá que se acerca el vencimiento de tus impuestos de %s. Entrá a Wallbit para ver el detalle.", monthLabel)

	default:
		subject = "Tenés novedades sobre tus impuestos en Wallbit"
		body = "Tenés novedades sobre tus impuestos en Wallbit. Entrá a la aplicación para revisarlas."
	}

	return []OutboundMessage{
		{
			Channel:   ChannelEmail,
			ClientID:  clientID,
			ClosingID: closingID,
			Subject:   subject,
			Body:      body,
			CreatedAt: createdAt,
		},
		{
			Channel:   ChannelPush,
			ClientID:  clientID,
			ClosingID: closingID,
			Subject:   subject,
			Body:      body,
			CreatedAt: createdAt,
		},
	}
}
