// Package client modela los clientes, CUIT, regímenes fiscales e historial
// (domain-context.md §1, §2.3, architecture.md §2.2).
package client

import (
	"errors"
	"fmt"
	"regexp"
)

var cuitRegex = regexp.MustCompile(`^\d{2}-\d{8}-\d{1}$`)

var (
	ErrInvalidCUITFormat   = errors.New("formato de CUIT inválido (esperado NN-NNNNNNNN-N)")
	ErrInvalidCUITChecksum = errors.New("dígito verificador de CUIT inválido")
)

// CUIT representa una Clave Única de Identificación Tributaria argentina.
type CUIT string

// ParseCUIT valida formato NN-NNNNNNNN-N y dígito verificador módulo 11.
func ParseCUIT(s string) (CUIT, error) {
	if !cuitRegex.MatchString(s) {
		return "", ErrInvalidCUITFormat
	}
	digits := make([]int, 11)
	idx := 0
	for _, r := range s {
		if r >= '0' && r <= '9' {
			digits[idx] = int(r - '0')
			idx++
		}
	}
	// Multiplicadores AFIP: 5, 4, 3, 2, 7, 6, 5, 4, 3, 2
	multipliers := []int{5, 4, 3, 2, 7, 6, 5, 4, 3, 2}
	sum := 0
	for i := 0; i < 10; i++ {
		sum += digits[i] * multipliers[i]
	}
	rem := sum % 11
	check := 11 - rem
	if check == 11 {
		check = 0
	}
	if check == 10 || check != digits[10] {
		return "", ErrInvalidCUITChecksum
	}
	return CUIT(s), nil
}

// String devuelve la representación textual del CUIT.
func (c CUIT) String() string {
	return string(c)
}

// Terminacion devuelve el último dígito ANTES del verificador (domain-context.md §2.3).
// Ej: 20-12345678-9 -> 8.
func (c CUIT) Terminacion() int {
	s := string(c)
	if len(s) < 3 {
		return 0
	}
	// En NN-NNNNNNNN-N, el penúltimo dígito está en len(s)-3
	return int(s[len(s)-3] - '0')
}

// Masked devuelve el CUIT con los primeros 4 dígitos del cuerpo enmascarados
// (specs/011-client-portal/spec.md criterio 2: "20-****5678-9").
func (c CUIT) Masked() string {
	s := string(c)
	if len(s) != 13 {
		return s
	}
	return fmt.Sprintf("%s-****%s-%s", s[0:2], s[7:11], s[12:13])
}
