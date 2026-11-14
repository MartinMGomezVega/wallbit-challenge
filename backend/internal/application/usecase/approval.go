package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/closing"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/notification"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/signal"
)

// ApproveClosingRequest contiene los parámetros para aprobar individualmente un cierre.
type ApproveClosingRequest struct {
	ClosingID  string `json:"closingId"`
	Version    int    `json:"version"`
	ClientNote string `json:"clientNote"`
	Batch      bool   `json:"batch"`
}

// ApproveClosingResult es la respuesta tras aprobar un cierre individual.
type ApproveClosingResult struct {
	Closing closing.Closing `json:"closing"`
	Warning string          `json:"warning,omitempty"`
}

// DispatchOutbox procesa mensajes salientes pendientes en el outbox con reintentos.
type DispatchOutbox struct {
	Outbox ports.OutboxRepository
	Sender ports.NotificationSender
	Clock  ports.Clock
}

// Execute despacha hasta limit mensajes salientes del outbox.
func (uc *DispatchOutbox) Execute(ctx context.Context, limit int) (int, error) {
	if limit <= 0 {
		limit = 50
	}
	pending, err := uc.Outbox.Pending(ctx, limit)
	if err != nil {
		return 0, err
	}

	sentCount := 0
	for _, m := range pending {
		if m.Attempts >= 3 {
			continue
		}
		m.Attempts++
		sendErr := uc.Sender.Send(ctx, m)
		if sendErr != nil {
			m.LastError = sendErr.Error()
			_ = uc.Outbox.Update(ctx, m)
		} else {
			now := uc.Clock.Now()
			m.SentAt = &now
			m.LastError = ""
			_ = uc.Outbox.Update(ctx, m)
			sentCount++
		}
	}
	return sentCount, nil
}

// ApproveClosing gestiona la aprobación individual de un cierre y el envío de notificaciones.
type ApproveClosing struct {
	Closings       ports.ClosingRepository
	Clients        ports.ClientRepository
	Notifications  ports.NotificationRepository
	Outbox         ports.OutboxRepository
	Sender         ports.NotificationSender
	Clock          ports.Clock
	IDGen          ports.IDGenerator
	DispatchOutbox *DispatchOutbox
	Tx             ports.TxManager
}

// Execute ejecuta la aprobación individual de un cierre según los criterios 1 y 2 de la spec 010.
func (uc *ApproveClosing) Execute(ctx context.Context, req ApproveClosingRequest) (ApproveClosingResult, error) {
	if len(req.ClientNote) > 500 {
		return ApproveClosingResult{}, domain.ValidationError{Field: "clientNote", Msg: "máximo 500 caracteres"}
	}

	var c closing.Closing
	var warning string
	err := withinTx(ctx, uc.Tx, func(ctx context.Context) error {
		var err error
		c, err = uc.Closings.Get(ctx, req.ClosingID)
		if err != nil {
			return err
		}

		// 2. Si ya está aprobado o notificado en la misma versión -> ALREADY_APPROVED
		if (c.Status == closing.StatusApproved || c.Status == closing.StatusNotified) && c.Version == req.Version {
			return domain.ErrAlreadyApproved
		}

		// 1. Si la versión no coincide -> VERSION_CONFLICT (concurrencia optimista)
		if c.Version != req.Version {
			return domain.ErrVersionConflict
		}

		// Estados permitidos: READY_FOR_REVIEW y PENDING_CLIENT_INPUT
		if c.Status != closing.StatusReadyForReview && c.Status != closing.StatusPendingClientInput {
			return fmt.Errorf("%w: estado actual %s no permite aprobación", domain.ErrInvalidTransition, c.Status)
		}

		if c.Status == closing.StatusPendingClientInput {
			warning = "CLIENT_INPUT_NOT_CONFIRMED"
		}

		now := uc.Clock.Now()
		approval := closing.Approval{
			ID:         uc.IDGen.New("appr"),
			Version:    c.Version,
			ResultHash: c.ResultHash,
			ApprovedBy: "accountant",
			ApprovedAt: now,
			Note:       req.ClientNote,
			Batch:      req.Batch,
		}

		if err := c.Approve(approval); err != nil {
			return err
		}

		cli, err := uc.Clients.Get(ctx, c.ClientID)
		if err != nil {
			return err
		}

		// Determinar si es primer aviso o aviso de actualización
		kind := notification.KindClosingReady
		if c.Published != nil || len(c.Approvals) > 1 {
			kind = notification.KindClosingUpdated
		}

		msg := notification.BuildClosingMessage(c, cli, req.ClientNote)

		// Guardar notificación in-app
		inApp := notification.Notification{
			ID:        uc.IDGen.New("notif"),
			ClientID:  c.ClientID,
			ClosingID: c.ID,
			Kind:      kind,
			Title:     msg.Title,
			Body:      msg.Body,
			Key:       string(kind),
			CreatedAt: now,
		}
		if err := uc.Notifications.Save(ctx, inApp); err != nil {
			return err
		}

		// Encolar mensajes salientes (email + push)
		monthLabel := notification.FormatPeriodLabel(c.Period)
		outboundMsgs := notification.OutboundFor(c.ClientID, c.ID, kind, monthLabel, now)
		for _, om := range outboundMsgs {
			om.ID = uc.IDGen.New("outbox")
			om.CreatedAt = now
			if err := uc.Outbox.Enqueue(ctx, om); err != nil {
				return err
			}
		}

		// Congelar Published y marcar cierre como NOTIFIED
		pub := closing.Published{
			Version:     c.Version,
			Title:       msg.Title,
			Body:        msg.Body,
			Figures:     buildPublishedFigures(c),
			PublishedAt: now,
		}
		if err := c.Publish(pub); err != nil {
			return err
		}

		return uc.Closings.SaveIfVersion(ctx, c, req.Version)
	})
	if err != nil {
		return ApproveClosingResult{}, err
	}

	// Despachar outbox de forma síncrona
	if uc.DispatchOutbox != nil {
		_, _ = uc.DispatchOutbox.Execute(ctx, 10)
	}

	return ApproveClosingResult{
		Closing: c,
		Warning: warning,
	}, nil
}

// buildPublishedFigures extrae los importes oficiales que el cliente visualiza.
func buildPublishedFigures(c closing.Closing) closing.PublishedFigures {
	var figures closing.PublishedFigures
	if c.Mono != nil {
		figures.BilledUSD = c.Mono.BilledInPeriodUSD
		figures.BilledARS = c.Mono.BilledInPeriodARS
		figures.Total = c.Mono.Fee
		figures.Items = []closing.PublishedFigureItem{
			{
				Concept: "Cuota Monotributo",
				Amount:  c.Mono.Fee,
				DueDate: c.Mono.FeeDueDate,
			},
		}
	} else if c.RI != nil {
		figures.BilledUSD = c.RI.BilledInPeriodUSD
		figures.BilledARS = c.RI.BilledInPeriodARS
		figures.Total = c.RI.TotalToPay
		figures.Items = []closing.PublishedFigureItem{
			{
				Concept: "IVA",
				Amount:  c.RI.VatToPay,
				DueDate: c.RI.VatDueDate,
			},
			{
				Concept: "Autónomos",
				Amount:  c.RI.SelfEmployedContribution,
				DueDate: c.RI.SelfEmployedDueDate,
			},
		}
		if c.RI.IncomeTaxAdvance != nil && c.RI.IncomeTaxAdvance.IsPositive() {
			due := c.RI.SelfEmployedDueDate
			if c.RI.IncomeTaxAdvanceDueDate != nil {
				due = *c.RI.IncomeTaxAdvanceDueDate
			}
			figures.Items = append(figures.Items, closing.PublishedFigureItem{
				Concept: "Anticipo Ganancias",
				Amount:  *c.RI.IncomeTaxAdvance,
				DueDate: due,
			})
		}
	}
	return figures
}

// ApproveBatchItemRequest representa un ítem individual en la solicitud de aprobación por lote.
type ApproveBatchItemRequest struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
}

// ApproveBatchRequest contiene la lista de ítems a aprobar en lote (máximo 500).
type ApproveBatchRequest struct {
	Items []ApproveBatchItemRequest `json:"items"`
}

// ApproveBatchItemResult describe el resultado de aprobar un ítem individual del lote.
type ApproveBatchItemResult struct {
	ID    string `json:"id"`
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// ApproveBatchResult consolida los resultados del lote.
type ApproveBatchResult struct {
	Approved int                      `json:"approved"`
	Failed   int                      `json:"failed"`
	Results  []ApproveBatchItemResult `json:"results"`
}

// ApproveBatch procesa la aprobación en lote en paralelo mediante un worker pool (spec 010 criterio 3).
type ApproveBatch struct {
	ApproveClosing *ApproveClosing
	Closings       ports.ClosingRepository
	NumWorkers     int
	Tx             ports.TxManager
}

// Execute procesa la aprobación en lote concurrentemente y devuelve el resumen.
func (uc *ApproveBatch) Execute(ctx context.Context, p period.Period, req ApproveBatchRequest) (ApproveBatchResult, error) {
	if len(req.Items) > 500 {
		return ApproveBatchResult{}, domain.ValidationError{Field: "items", Msg: "el lote no puede superar los 500 elementos"}
	}

	results := make([]ApproveBatchItemResult, len(req.Items))
	if len(req.Items) == 0 {
		return ApproveBatchResult{Approved: 0, Failed: 0, Results: results}, nil
	}

	workers := uc.NumWorkers
	if workers <= 0 {
		workers = 8
	}
	if workers > len(req.Items) {
		workers = len(req.Items)
	}

	type job struct {
		index int
		item  ApproveBatchItemRequest
	}

	jobs := make(chan job, len(req.Items))
	for i, it := range req.Items {
		jobs <- job{index: i, item: it}
	}
	close(jobs)

	var wg sync.WaitGroup
	wg.Add(workers)

	for w := 0; w < workers; w++ {
		go func() {
			defer wg.Done()
			for j := range jobs {
				res := uc.processItem(ctx, j.item)
				results[j.index] = res
			}
		}()
	}

	wg.Wait()

	approved := 0
	failed := 0
	for _, r := range results {
		if r.OK {
			approved++
		} else {
			failed++
		}
	}

	return ApproveBatchResult{
		Approved: approved,
		Failed:   failed,
		Results:  results,
	}, nil
}

func (uc *ApproveBatch) processItem(ctx context.Context, item ApproveBatchItemRequest) ApproveBatchItemResult {
	var res ApproveBatchItemResult
	_ = withinTx(ctx, uc.Tx, func(ctx context.Context) error {
		c, err := uc.Closings.Get(ctx, item.ID)
		if err != nil {
			res = ApproveBatchItemResult{ID: item.ID, OK: false, Error: "NOT_FOUND"}
			return err
		}

		if c.Severity != signal.SeverityGreen {
			res = ApproveBatchItemResult{ID: item.ID, OK: false, Error: "NOT_GREEN"}
			return domain.ErrValidation
		}

		if c.Status != closing.StatusReadyForReview {
			res = ApproveBatchItemResult{ID: item.ID, OK: false, Error: "NOT_READY"}
			return domain.ErrInvalidTransition
		}

		if c.Version != item.Version {
			res = ApproveBatchItemResult{ID: item.ID, OK: false, Error: "VERSION_CONFLICT"}
			return domain.ErrVersionConflict
		}

		_, err = uc.ApproveClosing.Execute(ctx, ApproveClosingRequest{
			ClosingID: item.ID,
			Version:   item.Version,
			Batch:     true,
		})
		if err != nil {
			if errors.Is(err, domain.ErrVersionConflict) {
				res = ApproveBatchItemResult{ID: item.ID, OK: false, Error: "VERSION_CONFLICT"}
			} else {
				res = ApproveBatchItemResult{ID: item.ID, OK: false, Error: "NOT_READY"}
			}
			return err
		}

		res = ApproveBatchItemResult{ID: item.ID, OK: true}
		return nil
	})
	return res
}

// CreateActionRequiredNotices genera avisos para clientes con cierres en PENDING_CLIENT_INPUT (spec 010 criterio 9).
type CreateActionRequiredNotices struct {
	Closings       ports.ClosingRepository
	Notifications  ports.NotificationRepository
	Outbox         ports.OutboxRepository
	IDGen          ports.IDGenerator
	Clock          ports.Clock
	DispatchOutbox *DispatchOutbox
}

// Execute genera los avisos requeridos para el período indicado.
func (uc *CreateActionRequiredNotices) Execute(ctx context.Context, p period.Period) (int, error) {
	closings, err := uc.Closings.ListByPeriod(ctx, p)
	if err != nil {
		return 0, err
	}

	createdCount := 0
	monthLabel := notification.FormatPeriodLabel(p)
	now := uc.Clock.Now()

	for _, c := range closings {
		if c.Status != closing.StatusPendingClientInput {
			continue
		}

		exists, err := uc.Notifications.ExistsFor(ctx, c.ClientID, c.ID, notification.KindActionRequired, "ACTION_REQUIRED")
		if err != nil || exists {
			continue
		}

		title := fmt.Sprintf("Acción requerida: cierre de %s", monthLabel)
		body := fmt.Sprintf("Subí tus facturas de %s y contanos si cobraste algo fuera de Wallbit.", monthLabel)

		inApp := notification.Notification{
			ID:        uc.IDGen.New("notif"),
			ClientID:  c.ClientID,
			ClosingID: c.ID,
			Kind:      notification.KindActionRequired,
			Title:     title,
			Body:      body,
			Key:       "ACTION_REQUIRED",
			CreatedAt: now,
		}
		if err := uc.Notifications.Save(ctx, inApp); err != nil {
			return createdCount, err
		}

		outbounds := notification.OutboundFor(c.ClientID, c.ID, notification.KindActionRequired, monthLabel, now)
		for _, om := range outbounds {
			om.ID = uc.IDGen.New("outbox")
			om.CreatedAt = now
			if err := uc.Outbox.Enqueue(ctx, om); err != nil {
				return createdCount, err
			}
		}
		createdCount++
	}

	if uc.DispatchOutbox != nil {
		_, _ = uc.DispatchOutbox.Execute(ctx, 50)
	}

	return createdCount, nil
}

// ConceptDueDate asocia un concepto tributario con su fecha límite de pago.
type ConceptDueDate struct {
	Concept string
	DueDate time.Time
}

// SendPaymentReminders procesa y despacha recordatorios de vencimiento impositivo (spec 010 criterio 11).
type SendPaymentReminders struct {
	Closings       ports.ClosingRepository
	Notifications  ports.NotificationRepository
	Outbox         ports.OutboxRepository
	Parameters     ports.ParametersRepository
	Clock          ports.Clock
	IDGen          ports.IDGenerator
	DispatchOutbox *DispatchOutbox
}

// Execute evalúa cierres NOTIFIED sin paidAt y envía recordatorios para vencimientos en [now, now + reminder_days_before_due].
func (uc *SendPaymentReminders) Execute(ctx context.Context, now time.Time) (int, error) {
	reminderDays := 3
	if uc.Parameters != nil {
		if params, err := uc.Parameters.Get(ctx); err == nil && params.ReminderDaysBeforeDue > 0 {
			reminderDays = params.ReminderDaysBeforeDue
		}
	}

	windowEnd := now.AddDate(0, 0, reminderDays)

	pNow := period.FromTime(now)
	periodsToCheck := []period.Period{pNow, pNow.Prev()}

	remindedCount := 0

	for _, p := range periodsToCheck {
		closings, err := uc.Closings.ListByPeriod(ctx, p)
		if err != nil {
			continue
		}

		for _, c := range closings {
			if c.Status != closing.StatusNotified || c.PaidAt != nil {
				continue
			}

			dueDates := extractDueDates(c)
			for _, due := range dueDates {
				// Ventana: [now, now + reminder_days_before_due]
				if due.DueDate.Before(now) || due.DueDate.After(windowEnd) {
					continue
				}

				key := fmt.Sprintf("%s:%s", due.Concept, due.DueDate.In(period.Location()).Format("2006-01-02"))
				exists, err := uc.Notifications.ExistsFor(ctx, c.ClientID, c.ID, notification.KindPaymentReminder, key)
				if err != nil || exists {
					continue
				}

				monthLabel := notification.FormatPeriodLabel(c.Period)
				title := "Recordatorio de vencimiento impositivo"
				body := fmt.Sprintf("Recordá que se acerca el vencimiento de tus impuestos de %s. Entrá a Wallbit para ver el detalle.", monthLabel)

				inApp := notification.Notification{
					ID:        uc.IDGen.New("notif"),
					ClientID:  c.ClientID,
					ClosingID: c.ID,
					Kind:      notification.KindPaymentReminder,
					Title:     title,
					Body:      body,
					Key:       key,
					CreatedAt: now,
				}
				if err := uc.Notifications.Save(ctx, inApp); err != nil {
					return remindedCount, err
				}

				outbound := notification.OutboundFor(c.ClientID, c.ID, notification.KindPaymentReminder, monthLabel, now)
				for _, om := range outbound {
					om.ID = uc.IDGen.New("outbox")
					om.CreatedAt = now
					if err := uc.Outbox.Enqueue(ctx, om); err != nil {
						return remindedCount, err
					}
				}
				remindedCount++
			}
		}
	}

	if uc.DispatchOutbox != nil {
		_, _ = uc.DispatchOutbox.Execute(ctx, 50)
	}

	return remindedCount, nil
}

func extractDueDates(c closing.Closing) []ConceptDueDate {
	var out []ConceptDueDate
	if c.Mono != nil && !c.Mono.FeeDueDate.IsZero() && c.Mono.Fee.IsPositive() {
		out = append(out, ConceptDueDate{
			Concept: "MONO_FEE",
			DueDate: c.Mono.FeeDueDate,
		})
	}
	if c.RI != nil {
		if !c.RI.VatDueDate.IsZero() && c.RI.VatToPay.IsPositive() {
			out = append(out, ConceptDueDate{
				Concept: "VAT",
				DueDate: c.RI.VatDueDate,
			})
		}
		if !c.RI.SelfEmployedDueDate.IsZero() && c.RI.SelfEmployedContribution.IsPositive() {
			out = append(out, ConceptDueDate{
				Concept: "SELF_EMPLOYED",
				DueDate: c.RI.SelfEmployedDueDate,
			})
		}
		if c.RI.IncomeTaxAdvance != nil && c.RI.IncomeTaxAdvance.IsPositive() && c.RI.IncomeTaxAdvanceDueDate != nil {
			out = append(out, ConceptDueDate{
				Concept: "INCOME_TAX_ADVANCE",
				DueDate: *c.RI.IncomeTaxAdvanceDueDate,
			})
		}
	}
	return out
}
