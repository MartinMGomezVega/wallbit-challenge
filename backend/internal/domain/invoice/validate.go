package invoice

import (
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
)

// ValidateManual valida una factura cargada manualmente según las reglas de negocio (criterio 10).
func ValidateManual(inv Invoice, regime string, now time.Time) error {
	var errs domain.ValidationErrors

	if inv.ClientID == "" {
		errs = append(errs, domain.ValidationError{Field: "clientId", Msg: "el ID de cliente es requerido"})
	}

	if inv.PointOfSale <= 0 {
		errs = append(errs, domain.ValidationError{Field: "pointOfSale", Msg: "el punto de venta debe ser mayor a 0"})
	}

	if inv.Number <= 0 {
		errs = append(errs, domain.ValidationError{Field: "number", Msg: "el número de comprobante debe ser mayor a 0"})
	}

	if inv.IssueDate.IsZero() {
		errs = append(errs, domain.ValidationError{Field: "issueDate", Msg: "la fecha de emisión es requerida"})
	} else if inv.IssueDate.After(now) {
		errs = append(errs, domain.ValidationError{Field: "issueDate", Msg: "la fecha de emisión no puede ser futura"})
	}

	if inv.Currency != money.ARS && inv.Currency != money.USD {
		errs = append(errs, domain.ValidationError{Field: "currency", Msg: "la moneda debe ser ARS o USD"})
	}

	if !inv.TotalAmount.IsPositive() {
		errs = append(errs, domain.ValidationError{Field: "totalAmount", Msg: "el importe total debe ser mayor a 0"})
	}

	if inv.ReceiverName == "" {
		errs = append(errs, domain.ValidationError{Field: "receiverName", Msg: "la denominación del receptor es requerida"})
	}

	// Validación de tipo de factura según el régimen vigente a la fecha de emisión
	switch regime {
	case "MONOTRIBUTO":
		if inv.Type != TypeC && inv.Type != TypeE {
			errs = append(errs, domain.ValidationError{
				Field: "type",
				Msg:   "comprobante no permitido para Monotributo (solo facturas C o E)",
			})
		}
	case "RI":
		if inv.Type != TypeA && inv.Type != TypeB && inv.Type != TypeE {
			errs = append(errs, domain.ValidationError{
				Field: "type",
				Msg:   "comprobante no permitido para Responsable Inscripto (solo facturas A, B o E)",
			})
		}
	}

	return errs.OrNil()
}
