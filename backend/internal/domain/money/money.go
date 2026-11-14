// Package money implementa el modelo de dinero, tipo de cambio y porcentajes
// con enteros (centavos y basis points) y redondeo half-even (domain-context.md §3).
// Prohibido float64.
package money

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

// Monedas soportadas.
const (
	ARS Currency = "ARS"
	USD Currency = "USD"
)

// Currency representa una moneda ISO-4217.
type Currency string

// Errores de dinero.
var (
	ErrCurrencyMismatch = errors.New("currency mismatch")
	ErrInvalidAmount    = errors.New("invalid amount")
	ErrDivisionByZero   = errors.New("division by zero")
)

// Money representa una cantidad monetaria en unidades menores (centavos).
type Money struct {
	Amount   int64    `json:"amount"` // centavos
	Currency Currency `json:"currency"`
}

// New crea un valor Money.
func New(amount int64, c Currency) Money {
	return Money{Amount: amount, Currency: c}
}

// Zero crea un valor cero en la moneda indicada.
func Zero(c Currency) Money {
	return Money{Amount: 0, Currency: c}
}

// IsZero reporta si el monto es cero.
func (m Money) IsZero() bool {
	return m.Amount == 0
}

// IsPositive reporta si el monto es mayor a cero.
func (m Money) IsPositive() bool {
	return m.Amount > 0
}

// Add suma dos importes de la misma moneda.
func (m Money) Add(o Money) (Money, error) {
	if m.Currency != o.Currency {
		return Money{}, fmt.Errorf("%w: %s vs %s", ErrCurrencyMismatch, m.Currency, o.Currency)
	}
	return Money{Amount: m.Amount + o.Amount, Currency: m.Currency}, nil
}

// Sub resta o de m si tienen la misma moneda.
func (m Money) Sub(o Money) (Money, error) {
	if m.Currency != o.Currency {
		return Money{}, fmt.Errorf("%w: %s vs %s", ErrCurrencyMismatch, m.Currency, o.Currency)
	}
	return Money{Amount: m.Amount - o.Amount, Currency: m.Currency}, nil
}

// Rate representa un tipo de cambio escalado x10.000 (4 decimales).
// Ej: 1539.87 -> 15398700.
type Rate int64

// RateFromDecimal parsea un string decimal a Rate escalado x10.000.
// Acepta tanto "." como "," como separador decimal.
func RateFromDecimal(s string) (Rate, error) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ",", ".")
	parts := strings.Split(s, ".")
	if len(parts) == 1 {
		v, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("rate inválido: %w", err)
		}
		return Rate(v * 10000), nil
	}
	if len(parts) == 2 {
		intPart, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("rate inválido: %w", err)
		}
		decStr := parts[1]
		if len(decStr) > 4 {
			decStr = decStr[:4]
		}
		for len(decStr) < 4 {
			decStr += "0"
		}
		decPart, err := strconv.ParseInt(decStr, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("rate inválido: %w", err)
		}
		sign := int64(1)
		if strings.HasPrefix(s, "-") {
			sign = -1
		}
		return Rate(intPart*10000 + sign*decPart), nil
	}
	return 0, errors.New("rate inválido: demasiados puntos")
}

// DecimalString devuelve el rate formateado con 4 decimales ("1539.8700").
func (r Rate) DecimalString() string {
	val := int64(r)
	sign := ""
	if val < 0 {
		sign = "-"
		val = -val
	}
	intPart := val / 10000
	decPart := val % 10000
	return fmt.Sprintf("%s%d.%04d", sign, intPart, decPart)
}

// DecimalString devuelve el monto en formato estándar "1234.56".
func (m Money) DecimalString() string {
	val := m.Amount
	sign := ""
	if val < 0 {
		sign = "-"
		val = -val
	}
	intPart := val / 100
	decPart := val % 100
	return fmt.Sprintf("%s%d.%02d", sign, intPart, decPart)
}

// ParseDecimal parsea "1234.56" a Money en centavos.
func ParseDecimal(s string, c Currency) (Money, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Money{}, ErrInvalidAmount
	}
	parts := strings.Split(s, ".")
	if len(parts) == 1 {
		intPart, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return Money{}, fmt.Errorf("%w: %v", ErrInvalidAmount, err)
		}
		return Money{Amount: intPart * 100, Currency: c}, nil
	}
	if len(parts) == 2 {
		intPart, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return Money{}, fmt.Errorf("%w: %v", ErrInvalidAmount, err)
		}
		decStr := parts[1]
		if len(decStr) > 2 {
			decStr = decStr[:2]
		}
		for len(decStr) < 2 {
			decStr += "0"
		}
		decPart, err := strconv.ParseInt(decStr, 10, 64)
		if err != nil {
			return Money{}, fmt.Errorf("%w: %v", ErrInvalidAmount, err)
		}
		sign := int64(1)
		if strings.HasPrefix(s, "-") {
			sign = -1
		}
		return Money{Amount: intPart*100 + sign*decPart, Currency: c}, nil
	}
	return Money{}, ErrInvalidAmount
}

// ParseLocalized acepta formato es-AR ("1.234,56") o estándar ("1234.56").
func ParseLocalized(s string, c Currency) (Money, error) {
	s = strings.TrimSpace(s)
	// Quitar puntos de miles y reemplazar coma por punto
	if strings.Contains(s, ",") {
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", ".")
	}
	return ParseDecimal(s, c)
}

// RoundHalfEven divide n por d redondeando half-even (a par más cercano).
// Usa math/big.Int para evitar cualquier riesgo de overflow.
func RoundHalfEven(num, den *big.Int) int64 {
	zero := big.NewInt(0)
	neg := (num.Sign() < 0) != (den.Sign() < 0)
	absNum := new(big.Int).Abs(num)
	absDen := new(big.Int).Abs(den)

	quot := new(big.Int)
	rem := new(big.Int)
	quot.QuoRem(absNum, absDen, rem)

	rem2 := new(big.Int).Mul(rem, big.NewInt(2))
	cmp := rem2.Cmp(absDen)
	if cmp > 0 {
		quot.Add(quot, big.NewInt(1))
	} else if cmp == 0 {
		// Exactamente la mitad: si quot es impar, redondea al par siguiente
		if quot.Bit(0) == 1 {
			quot.Add(quot, big.NewInt(1))
		}
	}

	if neg && quot.Cmp(zero) != 0 {
		quot.Neg(quot)
	}
	return quot.Int64()
}

// ConvertUSDToARS convierte USD a ARS usando rate escalado x10.000 con redondeo half-even.
// ars_cents = round_half_even(usd_cents * rate_scaled / 10_000)
func ConvertUSDToARS(usd Money, rate Rate) (Money, error) {
	if usd.Currency != USD {
		return Money{}, fmt.Errorf("esperaba USD, got %s", usd.Currency)
	}
	num := new(big.Int).Mul(big.NewInt(usd.Amount), big.NewInt(int64(rate)))
	den := big.NewInt(10000)
	arsCents := RoundHalfEven(num, den)
	return Money{Amount: arsCents, Currency: ARS}, nil
}

// BasisPoints representa un porcentaje expresado en puntos básicos (100 bp = 1%, 10.000 bp = 100%).
type BasisPoints int64

// ApplyBp calcula m * bp / 10.000 redondeando half-even.
func ApplyBp(m Money, bp BasisPoints) Money {
	num := new(big.Int).Mul(big.NewInt(m.Amount), big.NewInt(int64(bp)))
	den := big.NewInt(10000)
	res := RoundHalfEven(num, den)
	return Money{Amount: res, Currency: m.Currency}
}

// RatioBp calcula a / b en puntos básicos redondeando hacia abajo (según spec 004 criterio 5).
func RatioBp(a, b Money) (BasisPoints, error) {
	if a.Currency != b.Currency {
		return 0, ErrCurrencyMismatch
	}
	if b.Amount == 0 {
		return 0, ErrDivisionByZero
	}
	num := new(big.Int).Mul(big.NewInt(a.Amount), big.NewInt(10000))
	res := new(big.Int).Quo(num, big.NewInt(b.Amount))
	return BasisPoints(res.Int64()), nil
}
