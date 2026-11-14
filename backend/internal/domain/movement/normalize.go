package movement

import (
	"strings"
	"unicode"
)

var corporateSuffixes = map[string]bool{
	"inc":  true,
	"llc":  true,
	"ltd":  true,
	"sa":   true,
	"srl":  true,
	"gmbh": true,
	"corp": true,
	"co":   true,
}

// NormalizeName normaliza un nombre de contraparte o receptor de factura según domain-context.md §6:
// minúsculas, sin acentos, sin puntuación, sin sufijos societarios (inc|llc|ltd|sa|srl|gmbh|corp|co),
// y colapsando espacios múltiples.
func NormalizeName(s string) string {
	if s == "" {
		return ""
	}

	lower := strings.ToLower(s)
	withoutAccents := stripAccents(lower)

	var sb strings.Builder
	sb.Grow(len(withoutAccents))
	for _, r := range withoutAccents {
		if r == '.' {
			// Se descartan puntos para unir siglas: "S.A." -> "sa", "S.R.L." -> "srl", "Co." -> "co"
			continue
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			sb.WriteRune(r)
		} else {
			sb.WriteRune(' ')
		}
	}

	fields := strings.Fields(sb.String())
	if len(fields) == 0 {
		return ""
	}

	// Quitar sufijos societarios del final (soporta sufijos compuestos como "Co., Ltd.")
	for len(fields) > 1 && corporateSuffixes[fields[len(fields)-1]] {
		fields = fields[:len(fields)-1]
	}

	return strings.Join(fields, " ")
}

func stripAccents(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case 'á', 'à', 'ä', 'â', 'ã':
			b.WriteRune('a')
		case 'é', 'è', 'ë', 'ê':
			b.WriteRune('e')
		case 'í', 'ì', 'ï', 'î':
			b.WriteRune('i')
		case 'ó', 'ò', 'ö', 'ô', 'õ':
			b.WriteRune('o')
		case 'ú', 'ù', 'ü', 'û':
			b.WriteRune('u')
		case 'ñ':
			b.WriteRune('n')
		case 'ç':
			b.WriteRune('c')
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
