// Package period implementa el manejo de períodos mensuales (YYYY-MM),
// cálculo de ventanas temporales (12 meses móviles) y días hábiles en zona horaria
// argentina (domain-context.md §1, §2, architecture.md §2.2).
package period

import (
	"errors"
	"fmt"
	"time"
)

var arLocation = time.FixedZone("ART", -3*3600)

func init() {
	if loc, err := time.LoadLocation("America/Argentina/Buenos_Aires"); err == nil {
		arLocation = loc
	}
}

// Location devuelve la zona horaria argentina.
func Location() *time.Location {
	return arLocation
}

// SetLocation permite sobreescribir la zona (ej. tests).
func SetLocation(loc *time.Location) {
	if loc != nil {
		arLocation = loc
	}
}

// ErrInvalidPeriod indica un formato de período inválido.
var ErrInvalidPeriod = errors.New("período inválido (formato esperado YYYY-MM)")

// Period representa un mes calendario (ej. 2026-10).
type Period struct {
	Year  int
	Month time.Month
}

// New crea un Period a partir de año y mes.
func New(year int, month time.Month) Period {
	return Period{Year: year, Month: month}
}

// FromTime crea un Period a partir de un instante en zona AR.
func FromTime(t time.Time) Period {
	tAR := t.In(arLocation)
	return Period{Year: tAR.Year(), Month: tAR.Month()}
}

// Parse parsea "YYYY-MM".
func Parse(s string) (Period, error) {
	if len(s) != 7 || s[4] != '-' {
		return Period{}, ErrInvalidPeriod
	}
	t, err := time.Parse("2006-01", s)
	if err != nil {
		return Period{}, fmt.Errorf("%w: %v", ErrInvalidPeriod, err)
	}
	return Period{Year: t.Year(), Month: t.Month()}, nil
}

// MustParse parsea "YYYY-MM" o produce panic si el formato es inválido.
func MustParse(s string) Period {
	p, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return p
}

// String formatea como "YYYY-MM".
func (p Period) String() string {
	return fmt.Sprintf("%04d-%02d", p.Year, p.Month)
}

// Start devuelve el primer instante del período (00:00:00 en zona AR).
func (p Period) Start() time.Time {
	return time.Date(p.Year, p.Month, 1, 0, 0, 0, 0, arLocation)
}

// End devuelve el último instante del período (23:59:59.999999999 en zona AR).
func (p Period) End() time.Time {
	// Primer instante del mes siguiente menos 1 nanosegundo
	nextMonth := time.Date(p.Year, p.Month+1, 1, 0, 0, 0, 0, arLocation)
	return nextMonth.Add(-time.Nanosecond)
}

// Prev devuelve el período anterior.
func (p Period) Prev() Period {
	t := time.Date(p.Year, p.Month-1, 1, 0, 0, 0, 0, time.UTC)
	return Period{Year: t.Year(), Month: t.Month()}
}

// Next devuelve el período siguiente.
func (p Period) Next() Period {
	t := time.Date(p.Year, p.Month+1, 1, 0, 0, 0, 0, time.UTC)
	return Period{Year: t.Year(), Month: t.Month()}
}

// Window12Months representa el rango (from, to] de los últimos 12 meses móviles.
type Window12Months struct {
	FromExclusive time.Time
	ToInclusive   time.Time
}

// Contains reporta si t está en (FromExclusive, ToInclusive].
func (w Window12Months) Contains(t time.Time) bool {
	return t.After(w.FromExclusive) && !t.After(w.ToInclusive)
}

// Last12MonthsWindow devuelve el rango móvil de 12 meses que termina en periodEnd:
// (periodEnd − 12 meses, periodEnd].
func (p Period) Last12MonthsWindow() Window12Months {
	end := p.End()
	// Hace 12 meses exactos al mismo nanosegundo
	from := time.Date(p.Year-1, p.Month+1, 1, 0, 0, 0, 0, arLocation).Add(-time.Nanosecond)
	return Window12Months{FromExclusive: from, ToInclusive: end}
}

// NextBusinessDay devuelve date si es día hábil (lunes a viernes); si es sábado o domingo,
// devuelve el lunes siguiente manteniendo hora, minuto, segundo y zona.
func NextBusinessDay(d time.Time) time.Time {
	switch d.Weekday() {
	case time.Saturday:
		return d.AddDate(0, 0, 2)
	case time.Sunday:
		return d.AddDate(0, 0, 1)
	default:
		return d
	}
}

// PrevBusinessDay devuelve date si es día hábil; si es sábado devuelve el viernes anterior (-1 día);
// si es domingo devuelve el viernes anterior (-2 días).
func PrevBusinessDay(d time.Time) time.Time {
	switch d.Weekday() {
	case time.Saturday:
		return d.AddDate(0, 0, -1)
	case time.Sunday:
		return d.AddDate(0, 0, -2)
	default:
		return d
	}
}
