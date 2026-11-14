package usecase

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/closing"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/notification"
)

// MeDTO representa la información básica del cliente autenticado (011 §2).
type MeDTO struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	CUITMasked string `json:"cuitMasked"`
	Regime     string `json:"regime"`
	Category   string `json:"category,omitempty"`
}

// GetMe obtiene el perfil público del cliente autenticado.
type GetMe struct {
	Clients ports.ClientRepository
}

// Execute ejecuta la obtención del perfil de cliente.
func (uc *GetMe) Execute(ctx context.Context, clientID string) (MeDTO, error) {
	c, err := uc.Clients.Get(ctx, clientID)
	if err != nil {
		return MeDTO{}, err
	}

	dto := MeDTO{
		ID:         c.ID,
		Name:       c.Name,
		CUITMasked: c.CUIT.Masked(),
		Regime:     string(c.Regime),
	}
	if c.Regime == client.RegimeMonotributo && c.MonotributoCategory != "" {
		dto.Category = string(c.MonotributoCategory)
	}
	return dto, nil
}

// DemoClientDTO representa un cliente para el switcher de la demo (011 §1).
type DemoClientDTO struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Regime string `json:"regime"`
}

// ListDemoClients lista los clientes para el RoleSwitcher, ordenando primero los 15 de demo.
type ListDemoClients struct {
	Clients ports.ClientRepository
}

// Execute obtiene y ordena los clientes de demo.
func (uc *ListDemoClients) Execute(ctx context.Context) ([]DemoClientDTO, error) {
	list, err := uc.Clients.List(ctx)
	if err != nil {
		return nil, err
	}

	sort.Slice(list, func(i, j int) bool {
		isDemoI := strings.HasPrefix(list[i].ID, "cli_demo_")
		isDemoJ := strings.HasPrefix(list[j].ID, "cli_demo_")
		if isDemoI && isDemoJ {
			return list[i].ID < list[j].ID
		}
		if isDemoI {
			return true
		}
		if isDemoJ {
			return false
		}
		return list[i].Name < list[j].Name
	})

	out := make([]DemoClientDTO, len(list))
	for i, c := range list {
		out[i] = DemoClientDTO{
			ID:     c.ID,
			Name:   c.Name,
			Regime: string(c.Regime),
		}
	}
	return out, nil
}

// ClientClosingInputDTO resume el estado de los inputs cargados por el cliente para el período.
type ClientClosingInputDTO struct {
	Answered      bool `json:"answered"`
	Confirmed     bool `json:"confirmed"`
	InvoicesCount int  `json:"invoicesCount"`
}

// ClientClosingDTO es el modelo que el cliente ve para cada período de cierre (011 §4).
type ClientClosingDTO struct {
	ClosingID string                `json:"closingId"`
	Period    string                `json:"period"`
	State     string                `json:"state"` // "ACTION_REQUIRED" | "IN_REVIEW" | "READY"
	Updating  bool                  `json:"updating,omitempty"`
	Published *closing.Published    `json:"published,omitempty"`
	PaidAt    *time.Time            `json:"paidAt,omitempty"`
	Input     ClientClosingInputDTO `json:"input"`
}

// ListMyClosings devuelve los cierres del cliente autenticado con su estado y snapshot publicado.
type ListMyClosings struct {
	Closings     ports.ClosingRepository
	ClientInputs ports.ClientInputRepository
	Invoices     ports.InvoiceRepository
}

// Execute obtiene los cierres del cliente.
func (uc *ListMyClosings) Execute(ctx context.Context, clientID string) ([]ClientClosingDTO, error) {
	closingsList, err := uc.Closings.ByClient(ctx, clientID)
	if err != nil {
		return nil, err
	}

	out := make([]ClientClosingDTO, 0, len(closingsList))
	for _, c := range closingsList {
		input, err := uc.ClientInputs.Get(ctx, clientID, c.Period)
		answered := false
		confirmed := false
		if err == nil {
			answered = input.Answered()
			confirmed = input.IsConfirmed()
		}

		invs, _ := uc.Invoices.ByClientBetween(ctx, clientID, c.Period.Start(), c.Period.End())
		invoicesCount := len(invs)

		var state string
		var updating bool

		if !confirmed {
			state = "ACTION_REQUIRED"
		} else if c.Status == closing.StatusNotified && c.Published != nil {
			state = "READY"
		} else {
			state = "IN_REVIEW"
			if c.Published != nil {
				updating = true
			}
		}

		out = append(out, ClientClosingDTO{
			ClosingID: c.ID,
			Period:    c.Period.String(),
			State:     state,
			Updating:  updating,
			Published: c.Published,
			PaidAt:    c.PaidAt,
			Input: ClientClosingInputDTO{
				Answered:      answered,
				Confirmed:     confirmed,
				InvoicesCount: invoicesCount,
			},
		})
	}
	return out, nil
}

// MarkClosingPaid marca un cierre como pagado por el cliente de forma idempotente (011 §5).
type MarkClosingPaid struct {
	Closings ports.ClosingRepository
	Clock    ports.Clock
}

// Execute registra la fecha de pago si el cierre está en estado READY.
func (uc *MarkClosingPaid) Execute(ctx context.Context, closingID, clientID string) (closing.Closing, error) {
	c, err := uc.Closings.Get(ctx, closingID)
	if err != nil {
		return closing.Closing{}, err
	}

	// 1. Verificación de propiedad (403)
	if c.ClientID != clientID {
		return closing.Closing{}, domain.ErrForbidden
	}

	// 2. Verificación de estado READY (409)
	if c.Status != closing.StatusNotified || c.Published == nil {
		return closing.Closing{}, domain.NewError(domain.ErrConflict, "CLOSING_NOT_READY", "Solo se puede marcar como pagado un cierre listo")
	}

	// 3. Idempotente: si ya está pagado, no modificar fecha
	if c.PaidAt == nil {
		now := uc.Clock.Now()
		_ = c.MarkPaid(now)
		if err := uc.Closings.Save(ctx, c); err != nil {
			return closing.Closing{}, err
		}
	}

	return c, nil
}

// ListMyNotifications devuelve las notificaciones del cliente autenticado ordenadas descendentemente.
type ListMyNotifications struct {
	Notifications ports.NotificationRepository
}

// Execute obtiene las notificaciones del cliente.
func (uc *ListMyNotifications) Execute(ctx context.Context, clientID string) ([]notification.Notification, error) {
	return uc.Notifications.ByClient(ctx, clientID)
}

// MarkNotificationRead marca una notificación como leída verificando propiedad (011 §6).
type MarkNotificationRead struct {
	Notifications ports.NotificationRepository
	Clock         ports.Clock
}

// Execute valida que la notificación pertenezca al cliente y la marca leída.
func (uc *MarkNotificationRead) Execute(ctx context.Context, notificationID, clientID string) error {
	n, err := uc.Notifications.Get(ctx, notificationID)
	if err != nil {
		return err
	}

	// 1. Verificación de pertenencia (403)
	if n.ClientID != clientID {
		return domain.ErrForbidden
	}

	now := uc.Clock.Now()
	return uc.Notifications.MarkRead(ctx, notificationID, now)
}
