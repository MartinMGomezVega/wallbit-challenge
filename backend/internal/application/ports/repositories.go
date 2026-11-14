package ports

import (
	"context"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/clientinput"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/closing"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/invoice"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/movement"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/notification"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
)

// ClientRepository define el acceso y persistencia de clientes.
type ClientRepository interface {
	List(ctx context.Context) ([]client.Client, error)
	Get(ctx context.Context, id string) (client.Client, error)
	Save(ctx context.Context, c client.Client) error
	SaveAll(ctx context.Context, cs []client.Client) error
}

// MovementRepository define el acceso y persistencia de movimientos.
type MovementRepository interface {
	ByClientBetween(ctx context.Context, clientID string, from, to time.Time) ([]movement.Movement, error)
	ByClient(ctx context.Context, clientID string) ([]movement.Movement, error)
	Get(ctx context.Context, id string) (movement.Movement, error)
	Save(ctx context.Context, m movement.Movement) error
	SaveAll(ctx context.Context, ms []movement.Movement) error
	ListAll(ctx context.Context) ([]movement.Movement, error)
}

// ClassificationRepository define el acceso y persistencia de clasificaciones de movimientos.
type ClassificationRepository interface {
	Get(ctx context.Context, movementID string) (movement.Classification, bool, error)
	Save(ctx context.Context, c movement.Classification) error
	SaveAll(ctx context.Context, cs []movement.Classification) error
	ByMovementIDs(ctx context.Context, ids []string) (map[string]movement.Classification, error)
}

// InvoiceRepository define el acceso y persistencia de facturas emitidas.
type InvoiceRepository interface {
	ByClientBetween(ctx context.Context, clientID string, from, to time.Time) ([]invoice.Invoice, error)
	ByClient(ctx context.Context, clientID string) ([]invoice.Invoice, error)
	Exists(ctx context.Context, clientID string, key invoice.Key) (bool, error)
	Get(ctx context.Context, id string) (invoice.Invoice, error)
	Save(ctx context.Context, inv invoice.Invoice) error
	SaveAll(ctx context.Context, invs []invoice.Invoice) error
}

// PurchaseInvoiceRepository define el acceso y persistencia de facturas de compra.
type PurchaseInvoiceRepository interface {
	ByClientBetween(ctx context.Context, clientID string, from, to time.Time) ([]invoice.PurchaseInvoice, error)
	ByClient(ctx context.Context, clientID string) ([]invoice.PurchaseInvoice, error)
	Save(ctx context.Context, pi invoice.PurchaseInvoice) error
	SaveAll(ctx context.Context, pis []invoice.PurchaseInvoice) error
}

// ClientInputRepository define el acceso y persistencia de inputs del cliente por período.
type ClientInputRepository interface {
	Get(ctx context.Context, clientID string, p period.Period) (clientinput.ClientPeriodInput, error)
	Save(ctx context.Context, input clientinput.ClientPeriodInput) error
	SaveAll(ctx context.Context, inputs []clientinput.ClientPeriodInput) error
}

// LatestExclusionDateReader lee la fecha de exclusión detectada en el último cierre del cliente (spec 005).
type LatestExclusionDateReader interface {
	LatestExclusionDate(ctx context.Context, clientID string) (*time.Time, error)
}

// ClosingRepository define el acceso y persistencia de los cierres contables mensuales.
type ClosingRepository interface {
	Get(ctx context.Context, id string) (closing.Closing, error)
	ByClientPeriod(ctx context.Context, clientID string, p period.Period) (closing.Closing, bool, error)
	ListByPeriod(ctx context.Context, p period.Period) ([]closing.Closing, error)
	ByClient(ctx context.Context, clientID string) ([]closing.Closing, error)
	Save(ctx context.Context, c closing.Closing) error
	SaveIfVersion(ctx context.Context, c closing.Closing, expectedVersion int) error
	LatestExclusionDate(ctx context.Context, clientID string) (*time.Time, error)
}

// NotificationRepository define el acceso y persistencia de notificaciones in-app.
type NotificationRepository interface {
	Get(ctx context.Context, id string) (notification.Notification, error)
	Save(ctx context.Context, n notification.Notification) error
	ByClient(ctx context.Context, clientID string) ([]notification.Notification, error)
	MarkRead(ctx context.Context, id string, at time.Time) error
	ExistsFor(ctx context.Context, clientID, closingID string, kind notification.Kind, key string) (bool, error)
}

// OutboxRepository define el acceso y persistencia de mensajes salientes pendientes de envío.
type OutboxRepository interface {
	Enqueue(ctx context.Context, m notification.OutboundMessage) error
	Pending(ctx context.Context, limit int) ([]notification.OutboundMessage, error)
	Update(ctx context.Context, m notification.OutboundMessage) error
}

// NotificationSender envía un mensaje saliente (email/push).
type NotificationSender interface {
	Send(ctx context.Context, m notification.OutboundMessage) error
}
