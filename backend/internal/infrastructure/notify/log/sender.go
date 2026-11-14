package log

import (
	"context"
	"io"
	"log/slog"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/notification"
)

var _ ports.NotificationSender = (*Sender)(nil)

// Sender es un adaptador de NotificationSender que registra los envíos por log
// garantizando que nunca se incluya el cuerpo del mensaje en el registro (criterio 8).
type Sender struct {
	logger *slog.Logger
}

// NewSender crea un nuevo Sender con el logger provisto o slog.Default() si es nil.
func NewSender(logger *slog.Logger) *Sender {
	if logger == nil {
		logger = slog.Default()
	}
	return &Sender{logger: logger}
}

// NewSenderWithWriter crea un Sender que escribe registros JSON en el io.Writer provisto.
func NewSenderWithWriter(w io.Writer) *Sender {
	handler := slog.NewJSONHandler(w, nil)
	return &Sender{logger: slog.New(handler)}
}

// Send registra en log los metadatos del mensaje saliente sin incluir el cuerpo del mensaje.
func (s *Sender) Send(ctx context.Context, m notification.OutboundMessage) error {
	s.logger.InfoContext(ctx, "outbound message sent",
		"channel", string(m.Channel),
		"clientId", m.ClientID,
		"subject", m.Subject,
	)
	return nil
}
