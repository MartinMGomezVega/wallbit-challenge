package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/notification"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/memory"
)

func TestNotificationRepository(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewNotificationRepository()

	n1 := notification.Notification{
		ID:        "notif_01",
		ClientID:  "cli_01",
		ClosingID: "clo_01",
		Kind:      notification.KindClosingReady,
		Title:     "Cierre listo",
		Body:      "Cuerpo de prueba",
		Key:       "READY",
		CreatedAt: time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC),
	}
	n2 := notification.Notification{
		ID:        "notif_02",
		ClientID:  "cli_01",
		ClosingID: "clo_01",
		Kind:      notification.KindPaymentReminder,
		Title:     "Recordatorio",
		Body:      "Vence pronto",
		Key:       "DUE:2026-11-20",
		CreatedAt: time.Date(2026, 11, 2, 10, 0, 0, 0, time.UTC),
	}

	if err := repo.Save(ctx, n1); err != nil {
		t.Fatalf("Save(n1) failed: %v", err)
	}
	if err := repo.Save(ctx, n2); err != nil {
		t.Fatalf("Save(n2) failed: %v", err)
	}

	// ByClient returns sorted desc by CreatedAt
	list, err := repo.ByClient(ctx, "cli_01")
	if err != nil {
		t.Fatalf("ByClient failed: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 notifications, got %d", len(list))
	}
	if list[0].ID != "notif_02" || list[1].ID != "notif_01" {
		t.Errorf("expected list sorted desc, got %s then %s", list[0].ID, list[1].ID)
	}

	// MarkRead
	readAt := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)
	if err := repo.MarkRead(ctx, "notif_01", readAt); err != nil {
		t.Fatalf("MarkRead failed: %v", err)
	}
	listAfter, _ := repo.ByClient(ctx, "cli_01")
	if listAfter[1].ReadAt == nil || !listAfter[1].ReadAt.Equal(readAt) {
		t.Errorf("expected ReadAt set, got %v", listAfter[1].ReadAt)
	}

	// MarkRead not found
	if err := repo.MarkRead(ctx, "not_found", readAt); err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}

	// ExistsFor
	exists, err := repo.ExistsFor(ctx, "cli_01", "clo_01", notification.KindPaymentReminder, "DUE:2026-11-20")
	if err != nil || !exists {
		t.Errorf("ExistsFor expected true, got %v (err: %v)", exists, err)
	}
	notExists, err := repo.ExistsFor(ctx, "cli_01", "clo_01", notification.KindPaymentReminder, "DUE:2026-11-25")
	if err != nil || notExists {
		t.Errorf("ExistsFor expected false, got %v (err: %v)", notExists, err)
	}
}

func TestOutboxRepository(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewOutboxRepository()

	m1 := notification.OutboundMessage{
		ID:        "msg_01",
		Channel:   notification.ChannelEmail,
		ClientID:  "cli_01",
		ClosingID: "clo_01",
		Subject:   "Asunto 1",
		Body:      "Cuerpo 1",
		Attempts:  0,
		CreatedAt: time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC),
	}
	m2 := notification.OutboundMessage{
		ID:        "msg_02",
		Channel:   notification.ChannelPush,
		ClientID:  "cli_01",
		ClosingID: "clo_01",
		Subject:   "Asunto 2",
		Body:      "Cuerpo 2",
		Attempts:  0,
		CreatedAt: time.Date(2026, 11, 1, 9, 0, 0, 0, time.UTC),
	}

	if err := repo.Enqueue(ctx, m1); err != nil {
		t.Fatalf("Enqueue(m1) failed: %v", err)
	}
	if err := repo.Enqueue(ctx, m2); err != nil {
		t.Fatalf("Enqueue(m2) failed: %v", err)
	}

	// Pending sorted asc by CreatedAt
	pending, err := repo.Pending(ctx, 10)
	if err != nil {
		t.Fatalf("Pending failed: %v", err)
	}
	if len(pending) != 2 {
		t.Fatalf("expected 2 pending, got %d", len(pending))
	}
	if pending[0].ID != "msg_02" || pending[1].ID != "msg_01" {
		t.Errorf("expected pending sorted asc, got %s then %s", pending[0].ID, pending[1].ID)
	}

	// Pending with limit
	pendingLimited, _ := repo.Pending(ctx, 1)
	if len(pendingLimited) != 1 {
		t.Fatalf("expected 1 pending with limit, got %d", len(pendingLimited))
	}

	// Update to sent
	sentAt := time.Date(2026, 11, 1, 11, 0, 0, 0, time.UTC)
	m2.SentAt = &sentAt
	m2.Attempts = 1
	if err := repo.Update(ctx, m2); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	pendingAfter, _ := repo.Pending(ctx, 10)
	if len(pendingAfter) != 1 || pendingAfter[0].ID != "msg_01" {
		t.Errorf("expected only msg_01 pending, got %v", pendingAfter)
	}

	// Update not found
	if err := repo.Update(ctx, notification.OutboundMessage{ID: "non_existent"}); err != domain.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

