package period

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestPeriodBasic(t *testing.T) {
	p, err := Parse("2026-10")
	if err != nil {
		t.Fatalf("Parse falló: %v", err)
	}
	if p.Year != 2026 || p.Month != time.October {
		t.Fatalf("Period = %+v", p)
	}
	if p.String() != "2026-10" {
		t.Fatalf("String = %s", p.String())
	}

	// Start / End
	start := p.Start()
	if start.Year() != 2026 || start.Month() != time.October || start.Day() != 1 || start.Hour() != 0 {
		t.Fatalf("Start = %v", start)
	}

	end := p.End()
	if end.Year() != 2026 || end.Month() != time.October || end.Day() != 31 || end.Hour() != 23 || end.Minute() != 59 {
		t.Fatalf("End = %v", end)
	}

	// Prev / Next con cambio de año
	prev := p.Prev()
	if prev.String() != "2026-09" {
		t.Fatalf("Prev = %s", prev)
	}
	next := p.Next()
	if next.String() != "2026-11" {
		t.Fatalf("Next = %s", next)
	}

	pJan := New(2026, time.January)
	if pJan.Prev().String() != "2025-12" {
		t.Fatalf("Jan.Prev = %s", pJan.Prev())
	}
	pDec := New(2025, time.December)
	if pDec.Next().String() != "2026-01" {
		t.Fatalf("Dec.Next = %s", pDec.Next())
	}

	// FromTime
	fromT := FromTime(time.Date(2026, 2, 15, 12, 0, 0, 0, Location()))
	if fromT.String() != "2026-02" {
		t.Fatalf("FromTime = %s", fromT)
	}

	// Parse errors
	if _, err := Parse("invalid"); !errors.Is(err, ErrInvalidPeriod) {
		t.Fatalf("esperaba ErrInvalidPeriod, got %v", err)
	}
	if _, err := Parse("2026-13"); !errors.Is(err, ErrInvalidPeriod) {
		t.Fatalf("esperaba ErrInvalidPeriod, got %v", err)
	}
}

func TestLast12MonthsWindow(t *testing.T) {
	p, _ := Parse("2026-10")
	w := p.Last12MonthsWindow()

	// Ventana: (2025-10-31 23:59:59.999..., 2026-10-31 23:59:59.999...]
	// 2025-10-31 23:59:59.999 NO está
	if w.Contains(w.FromExclusive) {
		t.Fatal("FromExclusive no debe estar incluido")
	}
	// 2025-11-01 00:00:00 SÍ está
	firstDay := time.Date(2025, 11, 1, 0, 0, 0, 0, Location())
	if !w.Contains(firstDay) {
		t.Fatal("2025-11-01 debe estar incluido")
	}
	// 2026-10-31 23:59:59 SÍ está
	if !w.Contains(w.ToInclusive) {
		t.Fatal("ToInclusive debe estar incluido")
	}
	// 2026-11-01 00:00:00 NO está
	nextDay := time.Date(2026, 11, 1, 0, 0, 0, 0, Location())
	if w.Contains(nextDay) {
		t.Fatal("2026-11-01 no debe estar incluido")
	}
}

func TestBusinessDays(t *testing.T) {
	// 2026-10-16 es Viernes
	fri := time.Date(2026, 10, 16, 10, 0, 0, 0, Location())
	if NextBusinessDay(fri) != fri {
		t.Fatal("Viernes no debe cambiar")
	}
	if PrevBusinessDay(fri) != fri {
		t.Fatal("Viernes no debe cambiar")
	}

	// 2026-10-17 es Sábado -> Lunes 19
	sat := time.Date(2026, 10, 17, 10, 0, 0, 0, Location())
	nextMon := NextBusinessDay(sat)
	if nextMon.Weekday() != time.Monday || nextMon.Day() != 19 {
		t.Fatalf("Sábado NextBusinessDay = %v", nextMon)
	}
	prevFri := PrevBusinessDay(sat)
	if prevFri.Weekday() != time.Friday || prevFri.Day() != 16 {
		t.Fatalf("Sábado PrevBusinessDay = %v", prevFri)
	}

	// 2026-10-18 es Domingo -> Lunes 19 y Viernes 16
	sun := time.Date(2026, 10, 18, 10, 0, 0, 0, Location())
	if got := NextBusinessDay(sun); got.Weekday() != time.Monday || got.Day() != 19 {
		t.Fatalf("Domingo NextBusinessDay = %v", got)
	}
	if got := PrevBusinessDay(sun); got.Weekday() != time.Friday || got.Day() != 16 {
		t.Fatalf("Domingo PrevBusinessDay = %v", got)
	}
}

func TestFebrero(t *testing.T) {
	// 2026 no es bisiesto (28 días)
	p2026, _ := Parse("2026-02")
	if p2026.End().Day() != 28 {
		t.Fatalf("Feb 2026 días = %d", p2026.End().Day())
	}

	// 2024 fue bisiesto (29 días)
	p2024, _ := Parse("2024-02")
	if p2024.End().Day() != 29 {
		t.Fatalf("Feb 2024 días = %d", p2024.End().Day())
	}
}

func TestPeriod_JSON_And_SetLocation(t *testing.T) {
	p := MustParse("2026-10")
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if string(b) != `"2026-10"` {
		t.Fatalf("Marshal = %s, want \"2026-10\"", b)
	}
	var got Period
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if got != p {
		t.Fatalf("Unmarshal = %v, want %v", got, p)
	}
	if err := json.Unmarshal([]byte(`123`), &got); err == nil {
		t.Fatal("expected error for non-string JSON")
	}
	if err := json.Unmarshal([]byte(`"invalid"`), &got); err == nil {
		t.Fatal("expected error for invalid period string")
	}

	orig := Location()
	defer SetLocation(orig)
	SetLocation(time.UTC)
	if Location() != time.UTC {
		t.Fatalf("Location = %v, want UTC", Location())
	}
	SetLocation(nil)
	if Location() != time.UTC {
		t.Fatalf("SetLocation(nil) should not overwrite")
	}
}
