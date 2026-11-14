package invoice

import (
	"fmt"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
)

// Tipos de comprobante válidos para emisión.
const (
	TypeA InvoiceType = "A"
	TypeB InvoiceType = "B"
	TypeC InvoiceType = "C"
	TypeE InvoiceType = "E"
)

// InvoiceType define el tipo de factura (A, B, C, E).
type InvoiceType string

// Orígenes posibles de una factura en el sistema.
const (
	SourceCSV    InvoiceSource = "CSV"
	SourceManual InvoiceSource = "MANUAL"
	SourceSeed   InvoiceSource = "SEED"
)

// InvoiceSource identifica el origen del registro.
type InvoiceSource string

// Key representa la clave natural y única de una factura emitida por cliente.
type Key struct {
	ClientID    string      `json:"clientId"`
	Type        InvoiceType `json:"type"`
	PointOfSale int         `json:"pointOfSale"`
	Number      int         `json:"number"`
}

func (k Key) String() string {
	return fmt.Sprintf("%s-%s-%04d-%08d", k.ClientID, k.Type, k.PointOfSale, k.Number)
}

// Invoice modela una factura emitida por un cliente (spec 003 criterio 1).
type Invoice struct {
	ID           string         `json:"id"`
	ClientID     string         `json:"clientId"`
	Type         InvoiceType    `json:"type"`
	PointOfSale  int            `json:"pointOfSale"`
	Number       int            `json:"number"`
	IssueDate    time.Time      `json:"issueDate"`
	Currency     money.Currency `json:"currency"`
	NetAmount    money.Money    `json:"netAmount"`
	VatAmount    money.Money    `json:"vatAmount"`
	TotalAmount  money.Money    `json:"totalAmount"`
	FXRate       *money.Rate    `json:"fxRate,omitempty"`
	ReceiverName string         `json:"receiverName"`
	ReceiverDoc  string         `json:"receiverDoc"`
	Source       InvoiceSource  `json:"source"`
	CreatedAt    time.Time      `json:"createdAt"`
}

// Key devuelve la clave única de la factura.
func (inv Invoice) Key() Key {
	return Key{
		ClientID:    inv.ClientID,
		Type:        inv.Type,
		PointOfSale: inv.PointOfSale,
		Number:      inv.Number,
	}
}

// PurchaseInvoice representa una factura de compra para clientes RI (spec 003 criterio 2).
type PurchaseInvoice struct {
	ID           string      `json:"id"`
	ClientID     string      `json:"clientId"`
	IssueDate    time.Time   `json:"issueDate"`
	SupplierName string      `json:"supplierName"`
	SupplierCUIT string      `json:"supplierCuit"`
	NetAmount    money.Money `json:"netAmount"` // ARS
	VatAmount    money.Money `json:"vatAmount"` // ARS
	CreatedAt    time.Time   `json:"createdAt"`
}
