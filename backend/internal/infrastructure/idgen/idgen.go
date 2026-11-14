// Package idgen implementa ports.IDGenerator.
package idgen

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
)

// Sequential genera prefix_000001, prefix_000002, ... por prefijo (determinístico).
type Sequential struct {
	mu   sync.Mutex
	next map[string]int
}

// NewSequential crea un generador secuencial.
func NewSequential() *Sequential { return &Sequential{next: map[string]int{}} }

// New devuelve el siguiente ID del prefijo.
func (s *Sequential) New(prefix string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.next[prefix]++
	return fmt.Sprintf("%s_%06d", prefix, s.next[prefix])
}

// Random genera IDs aleatorios con crypto/rand.
type Random struct{}

// New devuelve prefix_<16 hex>.
func (Random) New(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b) // crypto/rand.Read nunca falla en plataformas soportadas
	return prefix + "_" + hex.EncodeToString(b)
}
