package seed

import (
	"fmt"
	"strconv"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/client"
)

var cuitMultipliers = [10]int{5, 4, 3, 2, 7, 6, 5, 4, 3, 2}

// MakeValidCUIT genera un CUIT válido (dígito verificador AFIP) con la terminación solicitada (0..9).
// La terminación es el último dígito del número antes del dígito verificador.
func MakeValidCUIT(prefix int, baseNum int, termination int) client.CUIT {
	term := termination % 10
	if term < 0 {
		term = -term
	}

	for offset := 0; offset < 1000; offset++ {
		numPart := (baseNum + offset) % 10000000
		numStr := fmt.Sprintf("%07d%d", numPart, term)
		prefStr := fmt.Sprintf("%02d", prefix)

		full10 := prefStr + numStr
		sum := 0
		for i := 0; i < 10; i++ {
			digit, _ := strconv.Atoi(string(full10[i]))
			sum += digit * cuitMultipliers[i]
		}

		rem := sum % 11
		var checkDigit int
		if rem == 0 {
			checkDigit = 0
		} else if rem == 1 {
			// AFIP: si rem == 1 suele cambiarse el tipo a 23 o no ser válido; probamos siguiente offset
			continue
		} else {
			checkDigit = 11 - rem
		}

		cuitStr := fmt.Sprintf("%s-%s-%d", prefStr, numStr, checkDigit)
		c, err := client.ParseCUIT(cuitStr)
		if err == nil && c.Terminacion() == term {
			return c
		}
	}

	// Fallback garantizado
	c, _ := client.ParseCUIT(fmt.Sprintf("%02d-1234567%d-0", prefix, term))
	return c
}
