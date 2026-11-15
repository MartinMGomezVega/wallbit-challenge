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
	UploadedBy   string   `json:"uploadedBy,omitempty"`
	CreatedAt    string   `json:"createdAt"`
}

func toInvoiceDTO(inv invoice.Invoice) InvoiceDTO {
	var fxRateStr string
	if inv.FXRate != nil {
		fxRateStr = inv.FXRate.DecimalString()
	}
	uploadedBy := string(inv.UploadedBy)
	if uploadedBy == "" {
		uploadedBy = "CLIENT"
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
		UploadedBy:   uploadedBy,
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
	ConfirmedBy              string                  `json:"confirmedBy,omitempty"`
	ConfirmationNote         string                  `json:"confirmationNote,omitempty"`
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
		ConfirmedBy:              string(in.ConfirmedBy),
		ConfirmationNote:         in.ConfirmationNote,
		UpdatedAt:                in.UpdatedAt.Format(time.RFC3339),
	}
}

type invoicesHandler struct {
	importUC               *usecase.ImportInvoicesCSV
	addInvoiceUC           *usecase.AddInvoice
	listInvoicesUC         *usecase.ListInvoices
	addPurchaseUC          *usecase.AddPurchaseInvoice
	setExternalIncomeUC    *usecase.SetExternalIncome
	confirmPeriodUC        *usecase.ConfirmPeriodInput
	confirmInputOnBehalfUC *usecase.ConfirmInputOnBehalf
}

// importCSV maneja POST /v1/clients/{clientId}/invoice-imports (multipart, max 2MB)
func (h *invoicesHandler) importCSV(w http.ResponseWriter, r *http.Request) {
	clientID := r.PathValue("clientId")
	ident, _ := FromContext(r.Context())

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

	uploader := invoice.UploadedByClient
	if ident.Role == RoleAccountant {
		uploader = invoice.UploadedByAccountant
	}

	res, err := h.importUC.Execute(r.Context(), clientID, file, uploader)
	if err != nil {
		writeError(w, err)
		return
	}

	w.Header().Set("Location", "/v1/clients/"+clientID+"/invoices")
	writeJSON(w, http.StatusCreated, res)
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

// createInvoice maneja POST /v1/clients/{clientId}/invoices
func (h *invoicesHandler) createInvoice(w http.ResponseWriter, r *http.Request) {
	clientID := r.PathValue("clientId")

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
		ClientID:     clientID,
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

	w.Header().Set("Location", "/v1/clients/"+clientID+"/invoices/"+created.ID)
	writeJSON(w, http.StatusCreated, toInvoiceDTO(created))
}

// listClientInvoices maneja GET /v1/clients/{clientId}/invoices?period=2026-10
func (h *invoicesHandler) listClientInvoices(w http.ResponseWriter, r *http.Request) {
	clientID := r.PathValue("clientId")
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

// createPurchaseInvoice maneja POST /v1/clients/{clientId}/purchase-invoices
func (h *invoicesHandler) createPurchaseInvoice(w http.ResponseWriter, r *http.Request) {
	clientID := r.PathValue("clientId")

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
		ClientID:     clientID,
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

	w.Header().Set("Location", "/v1/clients/"+clientID+"/purchase-invoices/"+pi.ID)
	writeJSON(w, http.StatusCreated, toPurchaseInvoiceDTO(pi))
}

type setExternalIncomeRequest struct {
	Items []struct {
		Amount      MoneyDTO `json:"amount"`
		Description string   `json:"description"`
	} `json:"items"`
	None bool `json:"none"`
}

// setExternalIncome maneja PUT /v1/clients/{clientId}/periods/{period}/external-income
func (h *invoicesHandler) setExternalIncome(w http.ResponseWriter, r *http.Request) {
	clientID := r.PathValue("clientId")

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
		ClientID: clientID,
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

type confirmOnBehalfRequest struct {
	ExternalIncome struct {
		None  bool `json:"none"`
		Items []struct {
			Amount struct {
				Amount   any    `json:"amount"`
				Currency string `json:"currency"`
			} `json:"amount"`
			Description string `json:"description"`
		} `json:"items"`
	} `json:"externalIncome"`
	Note string `json:"note"`
}

// confirmPeriodInput maneja POST /v1/clients/{clientId}/periods/{period}/confirmation (unificado)
func (h *invoicesHandler) confirmPeriodInput(w http.ResponseWriter, r *http.Request) {
	clientID := r.PathValue("clientId")
	ident, _ := FromContext(r.Context())

	pStr := r.PathValue("period")
	p, err := period.Parse(pStr)
	if err != nil {
		writeErrorCode(w, http.StatusBadRequest, "BAD_REQUEST", "Período inválido (YYYY-MM)", nil)
		return
	}

	if ident.Role == RoleAccountant {
		var req confirmOnBehalfRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErrorCode(w, http.StatusBadRequest, "BAD_REQUEST", "JSON inválido", nil)
			return
		}
		if req.Note == "" {
			writeErrorCode(w, http.StatusBadRequest, "BAD_REQUEST", "La nota es obligatoria cuando confirma el contador", nil)
			return
		}

		items := make([]clientinput.ExternalIncomeItem, len(req.ExternalIncome.Items))
		for i, it := range req.ExternalIncome.Items {
			curr := money.USD
			if it.Amount.Currency != "" {
				curr = money.Currency(it.Amount.Currency)
			}
			var amt money.Money
			switch v := it.Amount.Amount.(type) {
			case float64:
				amt = money.New(int64(v), curr)
			case string:
				parsed, err := money.ParseDecimal(v, curr)
				if err == nil {
					amt = parsed
				}
			}
			items[i] = clientinput.ExternalIncomeItem{
				Amount:      amt,
				Description: it.Description,
			}
		}

		res, err := h.confirmInputOnBehalfUC.Execute(r.Context(), usecase.ConfirmInputOnBehalfInput{
			ClientID:       clientID,
			Period:         p,
			None:           req.ExternalIncome.None,
			ExternalIncome: items,
			Note:           req.Note,
		})
		if err != nil {
			writeError(w, err)
			return
		}

		w.Header().Set("Location", "/v1/clients/"+clientID+"/closings")
		writeJSON(w, http.StatusCreated, toClientPeriodInputDTO(res))
		return
	}

	// Cliente confirmando su propio período
	confirmed, err := h.confirmPeriodUC.Execute(r.Context(), clientID, p)
	if err != nil {
		writeError(w, err)
		return
	}

	w.Header().Set("Location", "/v1/clients/"+clientID+"/closings")
	writeJSON(w, http.StatusCreated, toClientPeriodInputDTO(confirmed))
}
