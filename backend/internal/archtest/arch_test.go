package archtest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const module = "github.com/MartinMGomezVega/wallbit-challenge/backend"

func TestArchitecture(t *testing.T) {
	root, _ := filepath.Abs("../..")
	vs, err := Check(root, module)
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range vs {
		t.Errorf("%s: %s", v.File, v.Rule)
	}
}

func TestDetectsViolations(t *testing.T) {
	vs, err := Check("testdata/bad", "fake")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"internal/domain/x/x.go: internal/domain no puede importar fake/internal/application/ports",
		"internal/domain/x/x.go: internal/domain no puede importar fake/internal/infrastructure/memory",
		"internal/domain/x/x.go: float64 prohibido en internal/domain",
		"internal/application/y/y.go: internal/application no puede importar fake/internal/interfaces/http",
		"internal/application/y/y.go: time.Now() solo en infrastructure/clock",
	}
	got := map[string]bool{}
	for _, v := range vs {
		got[v.File+": "+v.Rule] = true
	}
	for _, w := range want {
		if !got[w] {
			t.Errorf("no detectó: %s (got %v)", w, vs)
		}
	}
	for k := range got {
		if strings.Contains(k, "ok.go") {
			t.Errorf("falso positivo: %s", k)
		}
	}
}

func TestParseError(t *testing.T) {
	dir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(dir, "internal"), 0o755)
	_ = os.WriteFile(filepath.Join(dir, "internal", "z.go"), []byte("pakage z\n"), 0o644)
	if _, err := Check(dir, "fake"); err == nil {
		t.Fatal("esperaba error de parseo")
	}
	if _, err := Check("testdata/missing", "fake"); err == nil {
		t.Fatal("esperaba error por directorio inexistente")
	}
}
