package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/usecase"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/clientinput"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/invoice"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
)

// InvoiceDTO representa una factura emitida para la API JSON.
type InvoiceDTO struct {
	ID           string   `json:"id"`
	ClientID     string   `json:"clientId"`
	Type         string   `json:"type"`
	PointOfSale  int      `json:"pointOfSale"`
	Number       int      `json:"number"`
	IssueDate    string   `json:"issueDate"`
	Currency     string   `json:"currency"`
	NetAmount    MoneyDTO `json:"netAmount"`
	VatAmount    MoneyDTO `json:"vatAmount"`
	TotalAmount  MoneyDTO `json:"totalAmount"`
	FXRate       string   `json:"fxRate,omitempty"`
	ReceiverName string   `json:"receiverName"`
	ReceiverDoc  string   `json:"receiverDoc,omitempty"`
	Source       string   `json:"source"`
	CreatedAt    string   `json:"createdAt"`
}

func toInvoiceDTO(inv invoice.Invoice) InvoiceDTO {
	var fxRateStr string
	if inv.FXRate != nil {
		fxRateStr = inv.FXRate.DecimalString()
	}
	return InvoiceDTO{
		ID:           inv.ID,
		ClientID:     inv.ClientID,
		Type:         string(inv.Type),
		PointOfSale:  inv.PointOfSale,
		Number:       inv.Number,
		IssueDate:    inv.IssueDate.Format("2006-01-02"),
		Currency:     string(inv.Currency),
		NetAmount:    MoneyDTO{Amount: inv.NetAmount.DecimalString(), Currency: string(inv.NetAmount.Currency)},
		VatAmount:    MoneyDTO{Amount: inv.VatAmount.DecimalString(), Currency: string(inv.VatAmount.Currency)},
		TotalAmount:  MoneyDTO{Amount: inv.TotalAmount.DecimalString(), Currency: string(inv.TotalAmount.Currency)},
		FXRate:       fxRateStr,
		ReceiverName: inv.ReceiverName,
		ReceiverDoc:  inv.ReceiverDoc,
		Source:       string(inv.Source),
		CreatedAt:    inv.CreatedAt.Format(time.RFC3339),
	}
}

// PurchaseInvoiceDTO representa una factura de compra en formato JSON.
type PurchaseInvoiceDTO struct {
	ID           string   `json:"id"`
	ClientID     string   `json:"clientId"`
	IssueDate    string   `json:"issueDate"`
	SupplierName string   `json:"supplierName"`
	SupplierCUIT string   `json:"supplierCuit"`
	NetAmount    MoneyDTO `json:"netAmount"`
	VatAmount    MoneyDTO `json:"vatAmount"`
	CreatedAt    string   `json:"createdAt"`
}

func toPurchaseInvoiceDTO(pi invoice.PurchaseInvoice) PurchaseInvoiceDTO {
	return PurchaseInvoiceDTO{
		ID:           pi.ID,
		ClientID:     pi.ClientID,
		IssueDate:    pi.IssueDate.Format("2006-01-02"),
		SupplierName: pi.SupplierName,
		SupplierCUIT: string(pi.SupplierCUIT),
		NetAmount:    MoneyDTO{Amount: pi.NetAmount.DecimalString(), Currency: string(pi.NetAmount.Currency)},
		VatAmount:    MoneyDTO{Amount: pi.VatAmount.DecimalString(), Currency: string(pi.VatAmount.Currency)},
		CreatedAt:    pi.CreatedAt.Format(time.RFC3339),
	}
}

// ClientPeriodInputDTO representa la declaración de inputs del período.
type ClientPeriodInputDTO struct {
	ClientID                 string                  `json:"clientId"`
	Period                   string                  `json:"period"`
	ExternalIncome           []ExternalIncomeItemDTO `json:"externalIncome"`
	DeclaredNoExternalIncome bool                    `json:"declaredNoExternalIncome"`
	ConfirmedAt              string                  `json:"confirmedAt,omitempty"`
	UpdatedAt                string                  `json:"updatedAt"`
}

// ExternalIncomeItemDTO representa un ítem de ingreso externo en formato JSON.
type ExternalIncomeItemDTO struct {
	Amount      MoneyDTO `json:"amount"`
	Description string   `json:"description"`
}

func toClientPeriodInputDTO(in clientinput.ClientPeriodInput) ClientPeriodInputDTO {
	var confirmedAt string
	if in.ConfirmedAt != nil && !in.ConfirmedAt.IsZero() {
		confirmedAt = in.ConfirmedAt.Format(time.RFC3339)
	}
	items := make([]ExternalIncomeItemDTO, len(in.ExternalIncome))
	for i, it := range in.ExternalIncome {
		items[i] = ExternalIncomeItemDTO{
			Amount:      MoneyDTO{Amount: it.Amount.DecimalString(), Currency: string(it.Amount.Currency)},
			Description: it.Description,
		}
	}
	return ClientPeriodInputDTO{
		ClientID:                 in.ClientID,
		Period:                   in.Period.String(),
		ExternalIncome:           items,
		DeclaredNoExternalIncome: in.DeclaredNoExternalIncome,
		ConfirmedAt:              confirmedAt,
		UpdatedAt:                in.UpdatedAt.Format(time.RFC3339),
	}
}

type invoicesHandler struct {
	importUC            *usecase.ImportInvoicesCSV
	addInvoiceUC        *usecase.AddInvoice
	listInvoicesUC      *usecase.ListInvoices
	addPurchaseUC       *usecase.AddPurchaseInvoice
	setExternalIncomeUC *usecase.SetExternalIncome
	confirmPeriodUC     *usecase.ConfirmPeriodInput
}

// importCSV maneja POST /api/v1/me/invoices/import (multipart, max 2MB)
func (h *invoicesHandler) importCSV(w http.ResponseWriter, r *http.Request) {
	id, _ := FromContext(r.Context())
	if id.ClientID == "" {
		writeErrorCode(w, http.StatusBadRequest, "BAD_REQUEST", "ClientID no identificado", nil)
		return
	}

	// Limitar a 2 MB
	r.Body = http.MaxBytesReader(w, r.Body, 2<<20)
	if err := r.ParseMultipartForm(2 << 20); err != nil {
		writeErrorCode(w, http.StatusBadRequest, "BAD_REQUEST", "Archivo supera el límite de 2 MB o formato multipart inválido", nil)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		writeErrorCode(w, http.StatusBadRequest, "BAD_REQUEST", "Falta el campo 'file' en el formulario multipart", nil)
		return
	}
	defer func() { _ = file.Close() }()

	res, err := h.importUC.Execute(r.Context(), id.ClientID, file)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, res)
}

type createInvoiceRequest struct {
	Type         string   `json:"type"`
	PointOfSale  int      `json:"pointOfSale"`
	Number       int      `json:"number"`
	IssueDate    string   `json:"issueDate"`
	Currency     string   `json:"currency"`
	NetAmount    MoneyDTO `json:"netAmount"`
	VatAmount    MoneyDTO `json:"vatAmount"`
	TotalAmount  MoneyDTO `json:"totalAmount"`
	FXRate       string   `json:"fxRate,omitempty"`
	ReceiverName string   `json:"receiverName"`
	ReceiverDoc  string   `json:"receiverDoc,omitempty"`
}

// createInvoice maneja POST /api/v1/me/invoices
func (h *invoicesHandler) createInvoice(w http.ResponseWriter, r *http.Request) {
	id, _ := FromContext(r.Context())
	if id.ClientID == "" {
		writeErrorCode(w, http.StatusBadRequest, "BAD_REQUEST", "ClientID no identificado", nil)
		return
	}

	var req createInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorCode(w, http.StatusBadRequest, "BAD_REQUEST", "JSON inválido", nil)
		return
	}

	issueDate, err := time.Parse("2006-01-02", req.IssueDate)
	if err != nil {
		issueDate, _ = time.Parse(time.RFC3339, req.IssueDate)
	}

	curr := money.Currency(req.Currency)
	tot, _ := money.ParseDecimal(req.TotalAmount.Amount, curr)
	net, _ := money.ParseDecimal(req.NetAmount.Amount, curr)
	vat, _ := money.ParseDecimal(req.VatAmount.Amount, curr)

	var fxRate *money.Rate
	if req.FXRate != "" {
		r, err := money.RateFromDecimal(req.FXRate)
		if err == nil {
			fxRate = &r
		}
	}

	inv := invoice.Invoice{
		ClientID:     id.ClientID,
		Type:         invoice.InvoiceType(req.Type),
		PointOfSale:  req.PointOfSale,
		Number:       req.Number,
		IssueDate:    issueDate,
		Currency:     curr,
		NetAmount:    net,
		VatAmount:    vat,
		TotalAmount:  tot,
		FXRate:       fxRate,
		ReceiverName: req.ReceiverName,
		ReceiverDoc:  req.ReceiverDoc,
	}

	created, err := h.addInvoiceUC.Execute(r.Context(), inv)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toInvoiceDTO(created))
}

// listMeInvoices maneja GET /api/v1/me/invoices?period=2026-10
func (h *invoicesHandler) listMeInvoices(w http.ResponseWriter, r *http.Request) {
	id, _ := FromContext(r.Context())
	h.listInvoicesForClient(w, r, id.ClientID)
}

// listClientInvoicesAccountant maneja GET /api/v1/accountant/clients/{id}/invoices?period=2026-10
func (h *invoicesHandler) listClientInvoicesAccountant(w http.ResponseWriter, r *http.Request) {
	clientID := r.PathValue("id")
	if clientID == "" {
		writeErrorCode(w, http.StatusBadRequest, "BAD_REQUEST", "ID de cliente requerido", nil)
		return
	}
	h.listInvoicesForClient(w, r, clientID)
}

func (h *invoicesHandler) listInvoicesForClient(w http.ResponseWriter, r *http.Request, clientID string) {
	periodQuery := r.URL.Query().Get("period")
	var p *period.Period
	if periodQuery != "" {
		parsed, err := period.Parse(periodQuery)
		if err != nil {
			writeErrorCode(w, http.StatusBadRequest, "BAD_REQUEST", "Formato de período inválido (esperado YYYY-MM)", nil)
			return
		}
		p = &parsed
	}

	list, err := h.listInvoicesUC.Execute(r.Context(), clientID, p)
	if err != nil {
		writeError(w, err)
		return
	}

	dtos := make([]InvoiceDTO, len(list))
	for i, inv := range list {
		dtos[i] = toInvoiceDTO(inv)
	}

	writeJSON(w, http.StatusOK, dtos)
}

type createPurchaseInvoiceRequest struct {
	IssueDate    string   `json:"issueDate"`
	SupplierName string   `json:"supplierName"`
	SupplierCUIT string   `json:"supplierCuit"`
	NetAmount    MoneyDTO `json:"netAmount"`
	VatAmount    MoneyDTO `json:"vatAmount"`
}

// createPurchaseInvoice maneja POST /api/v1/me/purchase-invoices
func (h *invoicesHandler) createPurchaseInvoice(w http.ResponseWriter, r *http.Request) {
	id, _ := FromContext(r.Context())
	if id.ClientID == "" {
		writeErrorCode(w, http.StatusBadRequest, "BAD_REQUEST", "ClientID no identificado", nil)
		return
	}

	var req createPurchaseInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorCode(w, http.StatusBadRequest, "BAD_REQUEST", "JSON inválido", nil)
		return
	}

	issueDate, err := time.Parse("2006-01-02", req.IssueDate)
	if err != nil {
		issueDate, _ = time.Parse(time.RFC3339, req.IssueDate)
	}

	net, _ := money.ParseDecimal(req.NetAmount.Amount, money.ARS)
	vat, _ := money.ParseDecimal(req.VatAmount.Amount, money.ARS)

	pi, err := h.addPurchaseUC.Execute(r.Context(), usecase.AddPurchaseInvoiceInput{
		ClientID:     id.ClientID,
		IssueDate:    issueDate,
		SupplierName: req.SupplierName,
		SupplierCUIT: req.SupplierCUIT,
		NetAmount:    net,
		VatAmount:    vat,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toPurchaseInvoiceDTO(pi))
}

type setExternalIncomeRequest struct {
	Items []struct {
		Amount      MoneyDTO `json:"amount"`
		Description string   `json:"description"`
	} `json:"items"`
	None bool `json:"none"`
}

// setExternalIncome maneja PUT /api/v1/me/periods/{period}/external-income
func (h *invoicesHandler) setExternalIncome(w http.ResponseWriter, r *http.Request) {
	id, _ := FromContext(r.Context())
	pStr := r.PathValue("period")
	p, err := period.Parse(pStr)
	if err != nil {
		writeErrorCode(w, http.StatusBadRequest, "BAD_REQUEST", "Período inválido (YYYY-MM)", nil)
		return
	}

	var req setExternalIncomeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorCode(w, http.StatusBadRequest, "BAD_REQUEST", "JSON inválido", nil)
		return
	}

	items := make([]clientinput.ExternalIncomeItem, len(req.Items))
	for i, it := range req.Items {
		amt, _ := money.ParseDecimal(it.Amount.Amount, money.USD)
		items[i] = clientinput.ExternalIncomeItem{
			Amount:      amt,
			Description: it.Description,
		}
	}

	updated, err := h.setExternalIncomeUC.Execute(r.Context(), usecase.SetExternalIncomeInput{
		ClientID: id.ClientID,
		Period:   p,
		Items:    items,
		None:     req.None,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toClientPeriodInputDTO(updated))
}

// confirmPeriod maneja POST /api/v1/me/periods/{period}/confirm
func (h *invoicesHandler) confirmPeriod(w http.ResponseWriter, r *http.Request) {
	id, _ := FromContext(r.Context())
	pStr := r.PathValue("period")
	p, err := period.Parse(pStr)
	if err != nil {
		writeErrorCode(w, http.StatusBadRequest, "BAD_REQUEST", "Período inválido (YYYY-MM)", nil)
		return
	}

	confirmed, err := h.confirmPeriodUC.Execute(r.Context(), id.ClientID, p)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toClientPeriodInputDTO(confirmed))
}
