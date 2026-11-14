// Package config lee la configuración de la aplicación desde variables de entorno
// (architecture.md §2.4).
package config

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"
)

// DefaultAppNow es el "hoy" fijo de la demo.
const DefaultAppNow = "2026-11-05T09:00:00-03:00"

// Config es la configuración de la aplicación.
type Config struct {
	Port           string
	AppNow         *time.Time // nil = reloj real
	SeedClients    int
	Seed           uint64
	ClosingWorkers int
	ARCAClient     string
	FXClient       string
	MockLatency    time.Duration
	CORSOrigin     string
}

// Getenv permite inyectar el lookup de variables (tests).
type Getenv func(key string) (string, bool)

// Load lee la configuración del entorno del proceso.
func Load() (Config, error) { return LoadFrom(os.LookupEnv) }

// LoadFrom lee la configuración usando el lookup dado.
func LoadFrom(get Getenv) (Config, error) {
	str := func(k, def string) string {
		if v, ok := get(k); ok {
			return v
		}
		return def
	}
	c := Config{
		Port:       str("PORT", "8080"),
		ARCAClient: str("ARCA_CLIENT", "mock"),
		FXClient:   str("FX_CLIENT", "mock"),
		CORSOrigin: str("CORS_ORIGIN", "http://localhost:5173"),
	}
	if now := str("APP_NOW", DefaultAppNow); now != "" {
		t, err := time.Parse(time.RFC3339, now)
		if err != nil {
			return Config{}, fmt.Errorf("APP_NOW inválido (RFC3339): %w", err)
		}
		c.AppNow = &t
	}
	var err error
	if c.SeedClients, err = intVar(str("SEED_CLIENTS", "300"), "SEED_CLIENTS", 0); err != nil {
		return Config{}, err
	}
	seed, err := strconv.ParseUint(str("SEED", "42"), 10, 64)
	if err != nil {
		return Config{}, fmt.Errorf("SEED inválido: %w", err)
	}
	c.Seed = seed
	if c.ClosingWorkers, err = intVar(str("CLOSING_WORKERS", strconv.Itoa(runtime.NumCPU())), "CLOSING_WORKERS", 1); err != nil {
		return Config{}, err
	}
	ms, err := intVar(str("MOCK_LATENCY_MS", "0"), "MOCK_LATENCY_MS", 0)
	if err != nil {
		return Config{}, err
	}
	c.MockLatency = time.Duration(ms) * time.Millisecond
	if c.ARCAClient != "mock" {
		return Config{}, fmt.Errorf("ARCA_CLIENT=%q no soportado (solo mock)", c.ARCAClient)
	}
	if c.FXClient != "mock" {
		return Config{}, fmt.Errorf("FX_CLIENT=%q no soportado (solo mock)", c.FXClient)
	}
	return c, nil
}

func intVar(v, name string, min int) (int, error) {
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("%s inválido: %w", name, err)
	}
	if n < min {
		return 0, fmt.Errorf("%s debe ser >= %d", name, min)
	}
	return n, nil
}
