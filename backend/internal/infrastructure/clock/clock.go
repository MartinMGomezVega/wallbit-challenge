// Package clock implementa ports.Clock: reloj real y reloj fijo (APP_NOW).
package clock

import "time"

// ARLocation devuelve America/Argentina/Buenos_Aires o, si no hay tzdata, UTC-3 fija.
func ARLocation() *time.Location { return loadLocation("America/Argentina/Buenos_Aires") }

func loadLocation(name string) *time.Location {
	if loc, err := time.LoadLocation(name); err == nil {
		return loc
	}
	return time.FixedZone("ART", -3*3600)
}

// Real usa el reloj del sistema.
type Real struct{}

// Now devuelve la hora actual en zona AR.
func (Real) Now() time.Time { return time.Now().In(ARLocation()) }

// Location devuelve la zona AR.
func (Real) Location() *time.Location { return ARLocation() }

// Fixed devuelve siempre el mismo instante.
type Fixed struct{ t time.Time }

// NewFixed crea un reloj fijo en t (expresado en zona AR).
func NewFixed(t time.Time) Fixed { return Fixed{t: t.In(ARLocation())} }

// Now devuelve el instante fijo.
func (f Fixed) Now() time.Time { return f.t }

// Location devuelve la zona AR.
func (Fixed) Location() *time.Location { return ARLocation() }
