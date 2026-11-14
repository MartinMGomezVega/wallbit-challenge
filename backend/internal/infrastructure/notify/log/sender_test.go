package log_test

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/notification"
	notifyLog "github.com/MartinMGomezVega/wallbit-challenge/backend/internal/infrastructure/notify/log"
)

func TestSender_LogsMetadataWithoutBody(t *testing.T) {
	buf := &bytes.Buffer{}
	sender := notifyLog.NewSenderWithWriter(buf)

	msg := notification.OutboundMessage{
		ID:        "msg_123",
		Channel:   notification.ChannelEmail,
		ClientID:  "cli_01",
		ClosingID: "clo_01",
		Subject:   "Tu cierre de octubre 2026 está listo",
		Body:      "SECRET_BODY_WITH_CONFIDENTIAL_CONTENT",
		Attempts:  0,
		CreatedAt: time.Now(),
	}

	ctx := context.Background()
	if err := sender.Send(ctx, msg); err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	logOutput := buf.String()
	if strings.Contains(logOutput, "SECRET_BODY_WITH_CONFIDENTIAL_CONTENT") {
		t.Errorf("CRITICAL PRIVACY VIOLATION: log output contains message body: %s", logOutput)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("failed to parse log JSON: %v", err)
	}

	if parsed["channel"] != "EMAIL" {
		t.Errorf("channel = %v, want EMAIL", parsed["channel"])
	}
	if parsed["clientId"] != "cli_01" {
		t.Errorf("clientId = %v, want cli_01", parsed["clientId"])
	}
	if parsed["subject"] != "Tu cierre de octubre 2026 está listo" {
		t.Errorf("subject = %v, want subject", parsed["subject"])
	}

	// Also test NewSender(nil)
	defaultSender := notifyLog.NewSender(nil)
	if err := defaultSender.Send(ctx, msg); err != nil {
		t.Errorf("defaultSender.Send failed: %v", err)
	}
}
