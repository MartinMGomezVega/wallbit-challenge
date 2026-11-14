package closing

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/signal"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/tax"
)

type canonicalKV struct {
	Key   string `json:"k"`
	Value string `json:"v"`
}

type canonicalSignal struct {
	Code     signal.Code   `json:"code"`
	Severity string        `json:"severity"`
	Priority int           `json:"priority"`
	Message  string        `json:"message"`
	Data     []canonicalKV `json:"data,omitempty"`
}

type canonicalPayload struct {
	Mono    *tax.MonotributoAssessment `json:"mono,omitempty"`
	RI      *tax.RIAssessment          `json:"ri,omitempty"`
	Signals []canonicalSignal          `json:"signals,omitempty"`
}

// ComputeHash calcula el SHA-256 en formato hex del JSON canónico de la liquidación y señales.
// Garantiza que el orden de los maps en Data no afecte el hash y que cualquier cambio en montos
// produzca un hash diferente.
func ComputeHash(mono *tax.MonotributoAssessment, ri *tax.RIAssessment, signals []signal.Signal) string {
	canonSignals := make([]canonicalSignal, 0, len(signals))
	for _, s := range signals {
		var kvs []canonicalKV
		if len(s.Data) > 0 {
			keys := make([]string, 0, len(s.Data))
			for k := range s.Data {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			kvs = make([]canonicalKV, 0, len(keys))
			for _, k := range keys {
				kvs = append(kvs, canonicalKV{
					Key:   k,
					Value: s.Data[k],
				})
			}
		}

		canonSignals = append(canonSignals, canonicalSignal{
			Code:     s.Code,
			Severity: s.Severity.String(),
			Priority: s.Priority,
			Message:  s.Message,
			Data:     kvs,
		})
	}

	payload := canonicalPayload{
		Mono:    mono,
		RI:      ri,
		Signals: canonSignals,
	}

	bytes, err := json.Marshal(payload)
	if err != nil {
		// En structs sin canales ni funciones json.Marshal nunca falla
		panic(err)
	}

	sum := sha256.Sum256(bytes)
	return hex.EncodeToString(sum[:])
}

