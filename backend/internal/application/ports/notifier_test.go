package ports

import (
	"context"
	"testing"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
)

func TestNoOpChangeNotifier(t *testing.T) {
	n := &NoOpChangeNotifier{}
	p, _ := period.Parse("2026-10")
	err := n.NotifyClosingInputsChanged(context.Background(), ClosingInputsChanged{
		ClientID: "cli_1",
		Period:   p,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

