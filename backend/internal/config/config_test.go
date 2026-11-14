package config

import (
	"runtime"
	"testing"
	"time"
)

func env(m map[string]string) Getenv {
	return func(k string) (string, bool) { v, ok := m[k]; return v, ok }
}

func TestDefaults(t *testing.T) {
	c, err := LoadFrom(env(nil))
	if err != nil {
		t.Fatal(err)
	}
	if c.Port != "8080" || c.SeedClients != 300 || c.Seed != 42 || c.ClosingWorkers != runtime.NumCPU() ||
		c.ARCAClient != "mock" || c.FXClient != "mock" || c.MockLatency != 0 || c.CORSOrigin != "http://localhost:5173" {
		t.Fatalf("defaults inesperados: %+v", c)
	}
	want, _ := time.Parse(time.RFC3339, DefaultAppNow)
	if c.AppNow == nil || !c.AppNow.Equal(want) {
		t.Fatalf("AppNow = %v", c.AppNow)
	}
}

func TestOverride(t *testing.T) {
	c, err := LoadFrom(env(map[string]string{
		"PORT": "9090", "APP_NOW": "2026-12-01T10:00:00-03:00", "SEED_CLIENTS": "10", "SEED": "7",
		"CLOSING_WORKERS": "3", "MOCK_LATENCY_MS": "25", "CORS_ORIGIN": "http://x",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if c.Port != "9090" || c.SeedClients != 10 || c.Seed != 7 || c.ClosingWorkers != 3 ||
		c.MockLatency != 25*time.Millisecond || c.CORSOrigin != "http://x" || c.AppNow.Day() != 1 {
		t.Fatalf("override inesperado: %+v", c)
	}
}

func TestEmptyAppNowMeansRealClock(t *testing.T) {
	c, err := LoadFrom(env(map[string]string{"APP_NOW": ""}))
	if err != nil || c.AppNow != nil {
		t.Fatalf("esperaba AppNow nil, got %v %v", c.AppNow, err)
	}
}

func TestInvalid(t *testing.T) {
	cases := map[string]map[string]string{
		"app_now":  {"APP_NOW": "ayer"},
		"seed_cli": {"SEED_CLIENTS": "x"},
		"seed_neg": {"SEED_CLIENTS": "-1"},
		"seed":     {"SEED": "-3"},
		"workers":  {"CLOSING_WORKERS": "0"},
		"latency":  {"MOCK_LATENCY_MS": "abc"},
		"arca":     {"ARCA_CLIENT": "http"},
		"fx":       {"FX_CLIENT": "http"},
	}
	for name, m := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := LoadFrom(env(m)); err == nil {
				t.Fatal("esperaba error")
			}
		})
	}
}

func TestLoadUsesProcessEnv(t *testing.T) {
	t.Setenv("PORT", "1234")
	c, err := Load()
	if err != nil || c.Port != "1234" {
		t.Fatalf("%v %v", c.Port, err)
	}
}
