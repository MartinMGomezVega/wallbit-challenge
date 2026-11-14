package invoice

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
)

// ErrInvalidCSVHeaders indica que el archivo no tiene los encabezados esperados de Mis Comprobantes.
var ErrInvalidCSVHeaders = errors.New("encabezados inválidos o faltantes en el CSV de Mis Comprobantes")

// LineIssue describe una advertencia o error en una fila del CSV.
type LineIssue struct {
	Line    int    `json:"line"`
	Message string `json:"message"`
}

// ParsedInvoice representa una factura parseada exitosamente de una fila del CSV.
type ParsedInvoice struct {
	Type         InvoiceType
	PointOfSale  int
	Number       int
	IssueDate    time.Time
	Currency     money.Currency
	NetAmount    money.Money
	VatAmount    money.Money
	TotalAmount  money.Money
	FXRate       *money.Rate
	ReceiverName string
	ReceiverDoc  string
	LineNumber   int
}

// ParseMisComprobantes parsea un CSV exportado de "Mis Comprobantes" de ARCA (criterios 5–7).
func ParseMisComprobantes(r io.Reader) ([]ParsedInvoice, []LineIssue, []LineIssue, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, nil, nil, err
	}

	// Quitar UTF-8 BOM si está presente
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))

	if len(data) == 0 {
		return nil, nil, nil, ErrInvalidCSVHeaders
	}

	// Autodetectar delimitador (',' vs ';') mirando la primera línea
	firstLineEnd := bytes.IndexByte(data, '\n')
	var firstLine []byte
	if firstLineEnd == -1 {
		firstLine = data
	} else {
		firstLine = data[:firstLineEnd]
	}

	delimiter := rune(',')
	if bytes.Count(firstLine, []byte(";")) > bytes.Count(firstLine, []byte(",")) {
		delimiter = ';'
	}

	csvReader := csv.NewReader(bytes.NewReader(data))
	csvReader.Comma = delimiter
	csvReader.LazyQuotes = true
	csvReader.TrimLeadingSpace = true

	headers, err := csvReader.Read()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%w: %v", ErrInvalidCSVHeaders, err)
	}

	colMap := make(map[string]int)
	for i, h := range headers {
		colMap[normalizeHeader(h)] = i
	}

	requiredHeaders := []string{"fecha", "tipo", "puntodeventa", "numerodesde", "moneda", "imptotal"}
	for _, req := range requiredHeaders {
		if _, ok := colMap[req]; !ok {
			return nil, nil, nil, fmt.Errorf("%w: falta columna requerida %q", ErrInvalidCSVHeaders, req)
		}
	}

	var rows []ParsedInvoice
	var warnings []LineIssue
	var lineErrors []LineIssue

	lineNumber := 1 // header es la línea 1

	for {
		record, err := csvReader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		lineNumber++
		if err != nil {
			lineErrors = append(lineErrors, LineIssue{Line: lineNumber, Message: fmt.Sprintf("Error leyendo línea: %v", err)})
			continue
		}

		// Omitir líneas vacías
		isEmpty := true
		for _, field := range record {
			if strings.TrimSpace(field) != "" {
				isEmpty = false
				break
			}
		}
		if isEmpty {
			continue
		}

		getVal := func(key string) string {
			idx, ok := colMap[key]
			if !ok || idx >= len(record) {
				return ""
			}
			return strings.TrimSpace(record[idx])
		}

		// 1. Tipo de comprobante
		rawType := getVal("tipo")
		invType, isCreditDebit, okType := parseInvoiceType(rawType)
		if isCreditDebit {
			warnings = append(warnings, LineIssue{
				Line:    lineNumber,
				Message: "Notas de crédito/débito fuera de alcance",
			})
			continue
		}
		if !okType {
			lineErrors = append(lineErrors, LineIssue{
				Line:    lineNumber,
				Message: fmt.Sprintf("Tipo de comprobante desconocido o no soportado: %q", rawType),
			})
			continue
		}

		// 2. Fecha (dd/mm/yyyy)
		rawDate := getVal("fecha")
		issueDate, err := time.ParseInLocation("02/01/2006", rawDate, period.Location())
		if err != nil {
			// Fallback yyyy-mm-dd
			issueDate, err = time.ParseInLocation("2006-01-02", rawDate, period.Location())
		}
		if err != nil {
			lineErrors = append(lineErrors, LineIssue{
				Line:    lineNumber,
				Message: fmt.Sprintf("Fecha inválida %q (se espera formato dd/mm/yyyy)", rawDate),
			})
			continue
		}

		// 3. Punto de venta y número
		rawPV := getVal("puntodeventa")
		pv, err := strconv.Atoi(rawPV)
		if err != nil || pv <= 0 {
			lineErrors = append(lineErrors, LineIssue{
				Line:    lineNumber,
				Message: fmt.Sprintf("Punto de venta inválido: %q", rawPV),
			})
			continue
		}

		rawNum := getVal("numerodesde")
		num, err := strconv.Atoi(rawNum)
		if err != nil || num <= 0 {
			lineErrors = append(lineErrors, LineIssue{
				Line:    lineNumber,
				Message: fmt.Sprintf("Número de comprobante inválido: %q", rawNum),
			})
			continue
		}

		// 4. Moneda
		rawCurrency := getVal("moneda")
		curr, okCurr := parseCurrency(rawCurrency)
		if !okCurr {
			lineErrors = append(lineErrors, LineIssue{
				Line:    lineNumber,
				Message: fmt.Sprintf("Moneda desconocida o no soportada: %q", rawCurrency),
			})
			continue
		}

		// 5. Importes
		rawTotal := getVal("imptotal")
		total, err := money.ParseLocalized(rawTotal, curr)
		if err != nil || !total.IsPositive() {
			lineErrors = append(lineErrors, LineIssue{
				Line:    lineNumber,
				Message: fmt.Sprintf("Importe total inválido o no positivo: %q", rawTotal),
			})
			continue
		}

		rawIVA := getVal("iva")
		iva, err := money.ParseLocalized(rawIVA, curr)
		if err != nil {
			iva = money.Zero(curr)
		}

		// Neto: suma gravado + no gravado + exento si existen, de lo contrario total - iva
		rawNetoGrav := getVal("impnetogravado")
		rawNetoNoGrav := getVal("impnetonogravado")
		rawOpExentas := getVal("impopexentas")
		var net money.Money
		if rawNetoGrav != "" || rawNetoNoGrav != "" || rawOpExentas != "" {
			g, _ := money.ParseLocalized(rawNetoGrav, curr)
			ng, _ := money.ParseLocalized(rawNetoNoGrav, curr)
			ex, _ := money.ParseLocalized(rawOpExentas, curr)
			net = money.New(g.Amount+ng.Amount+ex.Amount, curr)
		} else {
			net = money.New(total.Amount-iva.Amount, curr)
		}

		// 6. Tipo de cambio
		rawTC := getVal("tipocambio")
		var fxRate *money.Rate
		if rawTC != "" && rawTC != "1" && rawTC != "0" {
			r, err := money.RateFromDecimal(rawTC)
			if err == nil && r > 0 {
				fxRate = &r
			}
		}

		// 7. Receptor
		receiverName := getVal("denominacionreceptor")
		receiverDoc := getVal("nrodocreceptor")

		rows = append(rows, ParsedInvoice{
			Type:         invType,
			PointOfSale:  pv,
			Number:       num,
			IssueDate:    issueDate,
			Currency:     curr,
			NetAmount:    net,
			VatAmount:    iva,
			TotalAmount:  total,
			FXRate:       fxRate,
			ReceiverName: receiverName,
			ReceiverDoc:  receiverDoc,
			LineNumber:   lineNumber,
		})
	}

	return rows, warnings, lineErrors, nil
}

func normalizeHeader(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var sb strings.Builder
	for _, r := range s {
		switch r {
		case 'á', 'à', 'ä':
			sb.WriteRune('a')
		case 'é', 'è', 'ë':
			sb.WriteRune('e')
		case 'í', 'ì', 'ï':
			sb.WriteRune('i')
		case 'ó', 'ò', 'ö':
			sb.WriteRune('o')
		case 'ú', 'ù', 'ü':
			sb.WriteRune('u')
		default:
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				sb.WriteRune(r)
			}
		}
	}
	return sb.String()
}

func parseInvoiceType(s string) (invType InvoiceType, isCreditDebit bool, ok bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", false, false
	}

	// Extraer código numérico inicial si existe (ej. "19 - Factura E", "1 - Factura A")
	fields := strings.Fields(s)
	codeStr := fields[0]
	codeStr = strings.TrimRight(codeStr, " -:")

	switch codeStr {
	case "1":
		return TypeA, false, true
	case "6":
		return TypeB, false, true
	case "11":
		return TypeC, false, true
	case "19":
		return TypeE, false, true
	case "2", "3", "7", "8", "12", "13", "20", "21":
		return "", true, false
	}

	// Fallback por texto si no empieza con número
	upper := strings.ToUpper(s)
	if strings.Contains(upper, "NOTA DE CRÉDITO") || strings.Contains(upper, "NOTA DE CREDITO") ||
		strings.Contains(upper, "NOTA DE DÉBITO") || strings.Contains(upper, "NOTA DE DEBITO") {
		return "", true, false
	}
	if strings.Contains(upper, "FACTURA E") || upper == "E" {
		return TypeE, false, true
	}
	if strings.Contains(upper, "FACTURA A") || upper == "A" {
		return TypeA, false, true
	}
	if strings.Contains(upper, "FACTURA B") || upper == "B" {
		return TypeB, false, true
	}
	if strings.Contains(upper, "FACTURA C") || upper == "C" {
		return TypeC, false, true
	}

	return "", false, false
}

func parseCurrency(s string) (money.Currency, bool) {
	upper := strings.ToUpper(strings.TrimSpace(s))
	switch upper {
	case "DOL", "USD", "U$S", "US$":
		return money.USD, true
	case "$", "PES", "ARS", "PESOS":
		return money.ARS, true
	default:
		return "", false
	}
}

