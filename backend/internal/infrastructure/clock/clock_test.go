package clock

import (
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
)

var (
	_ ports.Clock = Real{}
	_ ports.Clock = Fixed{}
)

func TestFixed(t *testing.T) {
	ts, _ := time.Parse(time.RFC3339, "2026-11-05T12:00:00Z")
	c := NewFixed(ts)
	if !c.Now().Equal(ts) {
		t.Fatal("Now distinto")
	}
	if c.Now().Hour() != 9 {
		t.Fatalf("esperaba 09h AR, got %d", c.Now().Hour())
	}
	if c.Location() == nil {
		t.Fatal("sin location")
	}
}

func TestReal(t *testing.T) {
	c := Real{}
	if d := time.Since(c.Now()); d < 0 || d > time.Minute {
		t.Fatalf("reloj real desfasado: %v", d)
	}
	_, off := c.Now().Zone()
	if off != -3*3600 {
		t.Fatalf("offset %d", off)
	}
	if c.Location() == nil {
		t.Fatal("sin location")
	}
}

func TestLoadLocationFallback(t *testing.T) {
	loc := loadLocation("No/Existe")
	_, off := time.Date(2026, 1, 1, 0, 0, 0, 0, loc).Zone()
	if off != -3*3600 {
		t.Fatalf("fallback offset %d", off)
	}
}
