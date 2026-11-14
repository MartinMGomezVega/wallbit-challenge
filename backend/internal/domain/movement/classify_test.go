package movement

import (
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
)

func TestClassify(t *testing.T) {
	now := time.Date(2026, 10, 15, 10, 0, 0, 0, time.UTC)
	usdAmount := money.New(100000, money.USD)
	rate := money.Rate(15000000)

	baseMovement := Movement{
		ID:         "mov_test",
		ClientID:   "cli_01",
		Amount:     usdAmount,
		FXRate:     rate,
		OccurredAt: now,
	}

	knownPayers := map[string]bool{
		"local client sa": true,
		"global corp":     true,
	}

	t.Run("R1: OWN_ACCOUNT_TRANSFER", func(t *testing.T) {
		m := baseMovement
		m.Type = TypeInboundTransfer
		m.Counterparty = Counterparty{
			Name:         "Martin Gomez",
			Country:      "US",
			Kind:         CounterpartyPerson,
			IsOwnAccount: true,
		}
		c := Classify(m, knownPayers)
		if c.Kind != KindNotIncome || c.Confidence != ConfidenceHigh || c.Rule != RuleOwnAccountTransfer {
			t.Fatalf("R1 got %+v", c)
		}
	})

	t.Run("Precedence: R1 > R6 (own account from foreign company)", func(t *testing.T) {
		m := baseMovement
		m.Type = TypeInboundTransfer
		m.Counterparty = Counterparty{
			Name:         "Northwind Labs LLC",
			Country:      "US",
			Kind:         CounterpartyCompany,
			IsOwnAccount: true, // R1 matches first
		}
		c := Classify(m, knownPayers)
		if c.Kind != KindNotIncome || c.Confidence != ConfidenceHigh || c.Rule != RuleOwnAccountTransfer {
			t.Fatalf("R1 > R6 failed, got %+v", c)
		}
	})

	t.Run("R2: STABLECOIN_CONVERSION buy", func(t *testing.T) {
		m := baseMovement
		m.Type = TypeStablecoinBuy
		c := Classify(m, knownPayers)
		if c.Kind != KindNotIncome || c.Confidence != ConfidenceHigh || c.Rule != RuleStablecoinConversion {
			t.Fatalf("R2 buy got %+v", c)
		}
	})

	t.Run("R2: STABLECOIN_CONVERSION sell", func(t *testing.T) {
		m := baseMovement
		m.Type = TypeStablecoinSell
		c := Classify(m, knownPayers)
		if c.Kind != KindNotIncome || c.Confidence != ConfidenceHigh || c.Rule != RuleStablecoinConversion {
			t.Fatalf("R2 sell got %+v", c)
		}
	})

	t.Run("R3: CARD_EXPENSE spend", func(t *testing.T) {
		m := baseMovement
		m.Type = TypeCardSpend
		c := Classify(m, knownPayers)
		if c.Kind != KindNotIncome || c.Confidence != ConfidenceHigh || c.Rule != RuleCardExpense {
			t.Fatalf("R3 spend got %+v", c)
		}
	})

	t.Run("R3: CARD_EXPENSE refund", func(t *testing.T) {
		m := baseMovement
		m.Type = TypeCardRefund
		c := Classify(m, knownPayers)
		if c.Kind != KindNotIncome || c.Confidence != ConfidenceHigh || c.Rule != RuleCardExpense {
			t.Fatalf("R3 refund got %+v", c)
		}
	})

	t.Run("R4: OUTBOUND transfer", func(t *testing.T) {
		m := baseMovement
		m.Type = TypeOutboundTransfer
		c := Classify(m, knownPayers)
		if c.Kind != KindNotIncome || c.Confidence != ConfidenceHigh || c.Rule != RuleOutbound {
			t.Fatalf("R4 transfer got %+v", c)
		}
	})

	t.Run("R4: OUTBOUND fee", func(t *testing.T) {
		m := baseMovement
		m.Type = TypeFee
		c := Classify(m, knownPayers)
		if c.Kind != KindNotIncome || c.Confidence != ConfidenceHigh || c.Rule != RuleOutbound {
			t.Fatalf("R4 fee got %+v", c)
		}
	})

	t.Run("R5: KNOWN_PAYER (and R5 > R8 precedence when country is AR)", func(t *testing.T) {
		m := baseMovement
		m.Type = TypeInboundTransfer
		m.Counterparty = Counterparty{
			Name:    "Local Client S.A.", // Normalizes to "local client"
			Country: "AR",
			Kind:    CounterpartyCompany,
		}
		// Notice that knownPayers has "local client" (without suffix)
		known := map[string]bool{
			"local client": true,
		}
		c := Classify(m, known)
		if c.Kind != KindIncome || c.Confidence != ConfidenceHigh || c.Rule != RuleKnownPayer {
			t.Fatalf("R5 got %+v", c)
		}
	})

	t.Run("R6: FOREIGN_COMPANY (company)", func(t *testing.T) {
		m := baseMovement
		m.Type = TypeInboundTransfer
		m.Counterparty = Counterparty{
			Name:    "Acme Corp",
			Country: "US",
			Kind:    CounterpartyCompany,
		}
		c := Classify(m, nil)
		if c.Kind != KindIncome || c.Confidence != ConfidenceHigh || c.Rule != RuleForeignCompany {
			t.Fatalf("R6 company got %+v", c)
		}
	})

	t.Run("R6: FOREIGN_COMPANY (platform)", func(t *testing.T) {
		m := baseMovement
		m.Type = TypeInboundTransfer
		m.Counterparty = Counterparty{
			Name:    "Upwork",
			Country: "US",
			Kind:    CounterpartyPlatform,
		}
		c := Classify(m, nil)
		if c.Kind != KindIncome || c.Confidence != ConfidenceHigh || c.Rule != RuleForeignCompany {
			t.Fatalf("R6 platform got %+v", c)
		}
	})

	t.Run("R7: FOREIGN_PERSON", func(t *testing.T) {
		m := baseMovement
		m.Type = TypeInboundTransfer
		m.Counterparty = Counterparty{
			Name:    "John Doe",
			Country: "US",
			Kind:    CounterpartyPerson,
		}
		c := Classify(m, nil)
		if c.Kind != KindIncome || c.Confidence != ConfidenceLow || c.Rule != RuleForeignPerson {
			t.Fatalf("R7 got %+v", c)
		}
	})

	t.Run("R8: LOCAL_INBOUND", func(t *testing.T) {
		m := baseMovement
		m.Type = TypeInboundTransfer
		m.Counterparty = Counterparty{
			Name:    "Juan Pérez",
			Country: "AR",
			Kind:    CounterpartyPerson,
		}
		c := Classify(m, nil)
		if c.Kind != KindUnclassified || c.Confidence != ConfidenceLow || c.Rule != RuleLocalInbound {
			t.Fatalf("R8 got %+v", c)
		}
	})

	t.Run("R9: FALLBACK", func(t *testing.T) {
		m := baseMovement
		m.Type = "UNKNOWN_TYPE"
		c := Classify(m, nil)
		if c.Kind != KindUnclassified || c.Confidence != ConfidenceLow || c.Rule != RuleFallback {
			t.Fatalf("R9 got %+v", c)
		}
	})
}
