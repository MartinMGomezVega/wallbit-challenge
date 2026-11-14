package signal

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
)

// formatDate formatea una fecha como dd/MM/yyyy en la zona horaria argentina.
func formatDate(t time.Time) string {
	return t.In(period.Location()).Format("02/01/2006")
}

// formatNumberES formatea un monto en centavos con separador de miles '.' y coma decimal ','.
func formatNumberES(cents int64) string {
	negative := cents < 0
	if negative {
		cents = -cents
	}

	integerPart := cents / 100
	decimalPart := cents % 100

	intStr := strconv.FormatInt(integerPart, 10)
	// Insertar puntos como separadores de miles desde la derecha
	var b strings.Builder
	l := len(intStr)
	for i, ch := range intStr {
		b.WriteRune(ch)
		remaining := l - 1 - i
		if remaining > 0 && remaining%3 == 0 {
			b.WriteRune('.')
		}
	}

	res := fmt.Sprintf("%s,%02d", b.String(), decimalPart)
	if negative {
		return "-" + res
	}
	return res
}

// formatCents formatea centavos a representación decimal estándar "123.45" sin usar float64.
func formatCents(cents int64) string {
	neg := ""
	if cents < 0 {
		neg = "-"
		cents = -cents
	}
	return fmt.Sprintf("%s%d.%02d", neg, cents/100, cents%100)
}

// formatARS formatea un importe en pesos argentinos según domain-context.md §8 (ej. "$ 1.234.567,89").
func formatARS(m money.Money) string {
	if m.Amount < 0 {
		return "-$ " + formatNumberES(-m.Amount)
	}
	return "$ " + formatNumberES(m.Amount)
}

// formatUSD formatea un importe en dólares según domain-context.md §8 (ej. "US$ 1.234,56").
func formatUSD(m money.Money) string {
	if m.Amount < 0 {
		return "-US$ " + formatNumberES(-m.Amount)
	}
	return "US$ " + formatNumberES(m.Amount)
}

// formatPct formatea un porcentaje expresado en basis points (1% = 100 bp) redondeando hacia abajo (ej. "87%").
func formatPct(bp int64) string {
	pct := bp / 100
	return fmt.Sprintf("%d%%", pct)
}

// formatRatio formatea una relación (amount / median) con un decimal y símbolo '×' (ej. "3,1×").
func formatRatio(amount, median int64) string {
	if median <= 0 {
		return "0,0×"
	}

	// Redondeo al decimal más cercano (half-up)
	scaled := (amount*100 + median/2) / median // centésimos
	dec := (scaled + 5) / 10                   // décimos redondeados
	integer := dec / 10
	fraction := dec % 10

	return fmt.Sprintf("%d,%d×", integer, fraction)
}
