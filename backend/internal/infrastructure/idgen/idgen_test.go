package idgen

import (
	"strings"
	"sync"
	"testing"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
)

var (
	_ ports.IDGenerator = (*Sequential)(nil)
	_ ports.IDGenerator = Random{}
)

func TestSequential(t *testing.T) {
	s := NewSequential()
	if got := s.New("cli"); got != "cli_000001" {
		t.Fatal(got)
	}
	if got := s.New("cli"); got != "cli_000002" {
		t.Fatal(got)
	}
	if got := s.New("mov"); got != "mov_000001" {
		t.Fatal(got)
	}
}

func TestSequentialConcurrent(t *testing.T) {
	s := NewSequential()
	var wg sync.WaitGroup
	seen := sync.Map{}
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id := s.New("clo")
			if _, dup := seen.LoadOrStore(id, true); dup {
				t.Errorf("duplicado %s", id)
			}
		}()
	}
	wg.Wait()
}

func TestRandom(t *testing.T) {
	a, b := Random{}.New("apr"), Random{}.New("apr")
	if a == b || !strings.HasPrefix(a, "apr_") || len(a) != 4+16 {
		t.Fatalf("%s %s", a, b)
	}
}
