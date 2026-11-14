// Package ports define las interfaces que la capa de aplicación necesita del mundo exterior.
package ports

import "time"

// Clock es la única fuente de "ahora" de la aplicación.
type Clock interface {
	Now() time.Time
	Location() *time.Location
}

// IDGenerator genera IDs con prefijo (cli_, mov_, inv_, clo_, apr_, ntf_).
type IDGenerator interface {
	New(prefix string) string
}
