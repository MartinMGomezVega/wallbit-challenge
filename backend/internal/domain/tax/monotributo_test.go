package tax

import (
	"fmt"
	"testing"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/invoice"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/money"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/parameters"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
)

func sampleAug2026Scale() Scale {
	return Scale{
		ValidFrom: time.Date(2026, 8, 1, 0, 0, 0, 0, period.Location()),
		Rows: []CategoryRow{
			{Category: CatA, CapARS: money.New(1200941045, money.ARS), ServicesFeeARS: money.New(4952718, money.ARS)},
			{Category: CatB, CapARS: money.New(1759518300, money.ARS), ServicesFeeARS: money.New(5637908, money.ARS)},
			{Category: CatC, CapARS: money.New(2467049400, money.ARS), ServicesFeeARS: money.New(6602000, money.ARS)},
			{Category: CatD, CapARS: money.New(3062865100, money.ARS), ServicesFeeARS: money.New(8461300, money.ARS)},
			{Category: CatE, CapARS: money.New(3602823100, money.ARS), ServicesFeeARS: money.New(11981100, money.ARS)},
			{Category: CatF, CapARS: money.New(4515165900, money.ARS), ServicesFeeARS: money.New(15078400, money.ARS)},
			{Category: CatG, CapARS: money.New(5399579900, money.ARS), ServicesFeeARS: money.New(23031300, money.ARS)},
			{Category: CatH, CapARS: money.New(8192466037, money.ARS), ServicesFeeARS: money.New(52270668, money.ARS)},
			{Category: CatI, CapARS: money.New(9169976190, money.ARS), ServicesFeeARS: money.New(96374786, money.ARS)},
			{Category: CatJ, CapARS: money.New(10501251920, money.ARS), ServicesFeeARS: money.New(116729976, money.ARS)},
			{Category: CatK, CapARS: money.New(12661083875, money.ARS), ServicesFeeARS: money.New(161444604, money.ARS)},
		},
	}
}

func sampleFeb2026Scale() Scale {
	return Scale{
		ValidFrom: time.Date(2026, 2, 1, 0, 0, 0, 0, period.Location()),
		Rows: []CategoryRow{
			{Category: CatA, CapARS: money.New(1027798813, money.ARS), ServicesFeeARS: money.New(4238674, money.ARS)},
			{Category: CatB, CapARS: money.New(1505844771, money.ARS), ServicesFeeARS: money.New(4825078, money.ARS)},
			{Category: CatC, CapARS: money.New(2111369652, money.ARS), ServicesFeeARS: money.New(5650185, money.ARS)},
			{Category: CatD, CapARS: money.New(2621285342, money.ARS), ServicesFeeARS: money.New(7241410, money.ARS)},
			{Category: CatE, CapARS: money.New(3083396437, money.ARS), ServicesFeeARS: money.New(10253797, money.ARS)},
			{Category: CatF, CapARS: money.New(3864204836, money.ARS), ServicesFeeARS: money.New(12904532, money.ARS)},
			{Category: CatG, CapARS: money.New(4621110937, money.ARS), ServicesFeeARS: money.New(19710823, money.ARS)},
			{Category: CatH, CapARS: money.New(7011340733, money.ARS), ServicesFeeARS: money.New(44734693, money.ARS)},
			{Category: CatI, CapARS: money.New(7847921162, money.ARS), ServicesFeeARS: money.New(82480226, money.ARS)},
			{Category: CatJ, CapARS: money.New(8987264030, money.ARS), ServicesFeeARS: money.New(99900765, money.ARS)},
			{Category: CatK, CapARS: money.New(10835708400, money.ARS), ServicesFeeARS: money.New(138168790, money.ARS)},
		},
	}
}

func dynamicScaleProvider(t time.Time) Scale {
	aug := time.Date(2026, 8, 1, 0, 0, 0, 0, period.Location())
	if !t.Before(aug) {
		return sampleAug2026Scale()
	}
	return sampleFeb2026Scale()
}

func TestCivilDay(t *testing.T) {
	loc := period.Location()
	t1 := time.Date(2026, 10, 14, 15, 30, 0, 0, loc)
	cd := CivilDayFromTime(t1, loc)
	if cd.Year != 2026 || cd.Month != 10 || cd.Day != 14 {
		t.Errorf("CivilDayFromTime = %+v, want 2026-10-14", cd)
	}

	cdTime := cd.Time(loc)
	if cdTime.Year() != 2026 || cdTime.Month() != 10 || cdTime.Day() != 14 || cdTime.Hour() != 0 {
		t.Errorf("cd.Time = %v", cdTime)
	}

	cdTimeUTC := cd.Time(nil)
	if cdTimeUTC.IsZero() {
		t.Error("cd.Time(nil) returned zero time")
	}

	cdNil := CivilDayFromTime(t1, nil)
	if cdNil.Year != t1.Year() {
		t.Errorf("CivilDayFromTime(t, nil) = %+v", cdNil)
	}
}

func TestDailySeries(t *testing.T) {
	loc := period.Location()
	minDate := time.Date(2024, 2, 27, 0, 0, 0, 0, loc)
	maxDate := time.Date(2024, 3, 2, 0, 0, 0, 0, loc)

	items := []DatedAmount{
		{Day: time.Date(2024, 2, 28, 12, 0, 0, 0, loc), Amount: money.New(100, money.ARS)},
		{Day: time.Date(2024, 2, 29, 12, 0, 0, 0, loc), Amount: money.New(200, money.ARS)}, // Bisiesto
		{Day: time.Date(2024, 3, 1, 12, 0, 0, 0, loc), Amount: money.New(300, money.ARS)},
	}

	series := newDailySeries(minDate, maxDate, items, loc)

	// Consulta rango vacío (to <= from)
	sumEmpty := series.Sum(time.Date(2024, 2, 28, 0, 0, 0, 0, loc), time.Date(2024, 2, 28, 0, 0, 0, 0, loc))
	if sumEmpty != 0 {
		t.Errorf("Sum empty range = %d, want 0", sumEmpty)
	}

	// Consulta día 29 de febrero inclusive (from 28 to 29)
	sumLeap := series.Sum(time.Date(2024, 2, 28, 0, 0, 0, 0, loc), time.Date(2024, 2, 29, 0, 0, 0, 0, loc))
	if sumLeap != 200 {
		t.Errorf("Sum leap day = %d, want 200", sumLeap)
	}

	// Consulta total (from 27 to 2 de marzo)
	sumAll := series.Sum(time.Date(2024, 2, 27, 0, 0, 0, 0, loc), time.Date(2024, 3, 2, 0, 0, 0, 0, loc))
	if sumAll != 600 {
		t.Errorf("Sum all = %d, want 600", sumAll)
	}

	// Clamping out of bounds
	sumClamp := series.Sum(time.Date(2020, 1, 1, 0, 0, 0, 0, loc), time.Date(2030, 1, 1, 0, 0, 0, 0, loc))
	if sumClamp != 600 {
		t.Errorf("Sum clamp = %d, want 600", sumClamp)
	}

	// Inverted min/max date in newDailySeries
	seriesInv := newDailySeries(maxDate, minDate, items, nil)
	if seriesInv == nil {
		t.Error("newDailySeries inverted returned nil")
	}
}

func TestAssessMonotributo_Errors(t *testing.T) {
	p := period.Period{Year: 2026, Month: 10}
	dueDate := time.Date(2026, 11, 20, 0, 0, 0, 0, period.Location())

	// ScaleAt nil
	_, err := AssessMonotributo(MonotributoInput{
		Period:          p,
		CurrentCategory: CatD,
		FeeDueDate:      dueDate,
	})
	if err == nil {
		t.Error("expected error when ScaleAt is nil")
	}

	// CurrentCategory empty
	_, err = AssessMonotributo(MonotributoInput{
		Period:     p,
		ScaleAt:    dynamicScaleProvider,
		FeeDueDate: dueDate,
	})
	if err == nil {
		t.Error("expected error when CurrentCategory is empty")
	}

	// FeeDueDate zero
	_, err = AssessMonotributo(MonotributoInput{
		Period:          p,
		CurrentCategory: CatD,
		ScaleAt:         dynamicScaleProvider,
	})
	if err == nil {
		t.Error("expected error when FeeDueDate is zero")
	}
}

func TestAssessMonotributo_Criterio5(t *testing.T) {
	loc := period.Location()
	p := period.Period{Year: 2026, Month: 10}
	feeDueDate := time.Date(2026, 11, 20, 0, 0, 0, 0, loc)
	params := parameters.Defaults()

	scaleAug := sampleAug2026Scale()
	capD := scaleAug.Rows[3].CapARS // 30,628,651.00 ARS
	capE := scaleAug.Rows[4].CapARS // 36,028,231.00 ARS
	capK := scaleAug.MaxCap()       // 126,610,838.75 ARS

	// 1. Acumulado exactamente igual al cap de D -> requerida D
	t.Run("Acumulado exactamente igual al cap de D", func(t *testing.T) {
		inv := PricedInvoice{
			Invoice: invoice.Invoice{
				ID:        "inv_1",
				Type:      invoice.TypeC,
				IssueDate: time.Date(2026, 10, 5, 10, 0, 0, 0, loc),
			},
			AmountARS: capD,
		}

		res, err := AssessMonotributo(MonotributoInput{
			Period:          p,
			CurrentCategory: CatA,
			Invoices:        []PricedInvoice{inv},
			ScaleAt:         dynamicScaleProvider,
			FeeDueDate:      feeDueDate,
			Params:          params,
		})
		if err != nil {
			t.Fatalf("AssessMonotributo failed: %v", err)
		}

		if res.Excluded {
			t.Error("should not be excluded")
		}
		if res.RequiredCategory == nil || *res.RequiredCategory != CatD {
			t.Errorf("RequiredCategory = %v, want %s", res.RequiredCategory, CatD)
		}
		if res.Billed12mARS != capD {
			t.Errorf("Billed12mARS = %v, want %v", res.Billed12mARS, capD)
		}
	})

	// 2. Cap de D + 1 centavo -> requerida E
	t.Run("Cap de D + 1 centavo -> requerida E", func(t *testing.T) {
		inv := PricedInvoice{
			Invoice: invoice.Invoice{
				ID:        "inv_2",
				Type:      invoice.TypeC,
				IssueDate: time.Date(2026, 10, 5, 10, 0, 0, 0, loc),
			},
			AmountARS: money.New(capD.Amount+1, money.ARS),
		}

		res, err := AssessMonotributo(MonotributoInput{
			Period:          p,
			CurrentCategory: CatA,
			Invoices:        []PricedInvoice{inv},
			ScaleAt:         dynamicScaleProvider,
			FeeDueDate:      feeDueDate,
			Params:          params,
		})
		if err != nil {
			t.Fatalf("AssessMonotributo failed: %v", err)
		}

		if res.RequiredCategory == nil || *res.RequiredCategory != CatE {
			t.Errorf("RequiredCategory = %v, want %s", res.RequiredCategory, CatE)
		}
	})

	// 3. Acumulado > cap K -> Excluded, ExclusionDate = 14/10/2026, ExcessARS, CanReturnFrom = 2030-01-01
	t.Run("Acumulado > cap K cruza el 14/10/2026", func(t *testing.T) {
		// Factura 1 el 10/10/2026 por capK - 1000 ARS
		inv1 := PricedInvoice{
			Invoice: invoice.Invoice{
				ID:        "inv_k1",
				Type:      invoice.TypeC,
				IssueDate: time.Date(2026, 10, 10, 10, 0, 0, 0, loc),
			},
			AmountARS: money.New(capK.Amount-1000, money.ARS),
		}
		// Factura 2 el 14/10/2026 por 5000 ARS -> supera cap K por 4000 ARS
		inv2 := PricedInvoice{
			Invoice: invoice.Invoice{
				ID:        "inv_k2",
				Type:      invoice.TypeC,
				IssueDate: time.Date(2026, 10, 14, 11, 0, 0, 0, loc),
			},
			AmountARS: money.New(5000, money.ARS),
		}

		res, err := AssessMonotributo(MonotributoInput{
			Period:          p,
			CurrentCategory: CatK,
			Invoices:        []PricedInvoice{inv1, inv2},
			ScaleAt:         dynamicScaleProvider,
			FeeDueDate:      feeDueDate,
			Params:          params,
		})
		if err != nil {
			t.Fatalf("AssessMonotributo failed: %v", err)
		}

		if !res.Excluded {
			t.Fatal("expected Excluded to be true")
		}
		if res.RequiredCategory != nil {
			t.Errorf("expected RequiredCategory to be nil when excluded, got %v", res.RequiredCategory)
		}
		if res.ExclusionDate == nil {
			t.Fatal("expected ExclusionDate to be set")
		}
		expectedExclDate := time.Date(2026, 10, 14, 0, 0, 0, 0, loc)
		if !res.ExclusionDate.Equal(expectedExclDate) {
			t.Errorf("ExclusionDate = %v, want %v", res.ExclusionDate, expectedExclDate)
		}

		if res.ExcessARS == nil || res.ExcessARS.Amount != 4000 {
			t.Errorf("ExcessARS = %v, want 4000 ARS", res.ExcessARS)
		}

		if res.CanReturnFrom == nil {
			t.Fatal("expected CanReturnFrom to be set")
		}
		expectedReturnDate := time.Date(2030, 1, 1, 0, 0, 0, 0, loc)
		if !res.CanReturnFrom.Equal(expectedReturnDate) {
			t.Errorf("CanReturnFrom = %v, want %v", res.CanReturnFrom, expectedReturnDate)
		}
	})

	// 4. Cruce de escala: facturas antes y después de 01/08/2026
	t.Run("Cruce de escala evaluado con escala vigente de cada día", func(t *testing.T) {
		// En julio 2026 (escala Feb 2026), Cap K es 108,357,084.00 ARS.
		// Si el 15 de julio de 2026 emite por 110,000,000 ARS, excede la escala vigente en julio.
		invJuly := PricedInvoice{
			Invoice: invoice.Invoice{
				ID:        "inv_july",
				Type:      invoice.TypeC,
				IssueDate: time.Date(2026, 7, 15, 10, 0, 0, 0, loc),
			},
			AmountARS: money.New(11000000000, money.ARS), // 110M ARS > 108.35M
		}

		res, err := AssessMonotributo(MonotributoInput{
			Period:          p,
			CurrentCategory: CatK,
			Invoices:        []PricedInvoice{invJuly},
			ScaleAt:         dynamicScaleProvider,
			FeeDueDate:      feeDueDate,
			Params:          params,
		})
		if err != nil {
			t.Fatalf("AssessMonotributo failed: %v", err)
		}

		if !res.Excluded {
			t.Fatal("expected Excluded to be true for July crossing")
		}
		expectedExclDate := time.Date(2026, 7, 15, 0, 0, 0, 0, loc)
		if !res.ExclusionDate.Equal(expectedExclDate) {
			t.Errorf("ExclusionDate = %v, want %v", res.ExclusionDate, expectedExclDate)
		}
	})

	// 5. InvoiceECountsForCap=false con solo facturas E -> acumulado 0, categoría A
	t.Run("InvoiceECountsForCap=false con facturas E", func(t *testing.T) {
		usdAmount := money.New(200000, money.USD)
		invE := PricedInvoice{
			Invoice: invoice.Invoice{
				ID:          "inv_e",
				Type:        invoice.TypeE,
				IssueDate:   time.Date(2026, 10, 10, 10, 0, 0, 0, loc),
				Currency:    money.USD,
				TotalAmount: usdAmount,
			},
			AmountARS: money.New(200000000, money.ARS),
			AmountUSD: &usdAmount,
		}

		paramsNoE := params
		paramsNoE.InvoiceECountsForCap = false

		res, err := AssessMonotributo(MonotributoInput{
			Period:          p,
			CurrentCategory: CatD,
			Invoices:        []PricedInvoice{invE},
			ScaleAt:         dynamicScaleProvider,
			FeeDueDate:      feeDueDate,
			Params:          paramsNoE,
		})
		if err != nil {
			t.Fatalf("AssessMonotributo failed: %v", err)
		}

		if res.Billed12mARS.Amount != 0 {
			t.Errorf("Billed12mARS = %v, want 0", res.Billed12mARS)
		}
		if res.RequiredCategory == nil || *res.RequiredCategory != CatA {
			t.Errorf("RequiredCategory = %v, want CatA", res.RequiredCategory)
		}
		// Pero BilledInPeriod refleja la factura
		if res.BilledInPeriodARS.Amount != 200000000 {
			t.Errorf("BilledInPeriodARS = %v, want 200000000", res.BilledInPeriodARS)
		}
		if res.BilledInPeriodUSD.Amount != 200000 {
			t.Errorf("BilledInPeriodUSD = %v, want 200000", res.BilledInPeriodUSD)
		}
	})

	// 6. Fee = cuota de la categoría actual, aunque la requerida sea mayor
	t.Run("Fee es cuota de categoría actual", func(t *testing.T) {
		inv := PricedInvoice{
			Invoice: invoice.Invoice{
				ID:        "inv_big",
				Type:      invoice.TypeC,
				IssueDate: time.Date(2026, 10, 5, 10, 0, 0, 0, loc),
			},
			AmountARS: capE, // Requiere CatE
		}

		res, err := AssessMonotributo(MonotributoInput{
			Period:          p,
			CurrentCategory: CatB, // Cliente está en CatB
			Invoices:        []PricedInvoice{inv},
			ScaleAt:         dynamicScaleProvider,
			FeeDueDate:      feeDueDate,
			Params:          params,
		})
		if err != nil {
			t.Fatalf("AssessMonotributo failed: %v", err)
		}

		expectedFeeCatB := scaleAug.Rows[1].ServicesFeeARS // CatB fee
		if res.Fee != expectedFeeCatB {
			t.Errorf("Fee = %v, want %v (CatB fee)", res.Fee, expectedFeeCatB)
		}
		if res.RequiredCategory == nil || *res.RequiredCategory != CatE {
			t.Errorf("RequiredCategory = %v, want CatE", res.RequiredCategory)
		}
	})

	// 7. Porcentajes en bp con redondeo hacia abajo
	t.Run("Porcentajes en bp redondean hacia abajo", func(t *testing.T) {
		// Supongamos billed12m = 50% exacto de CatD
		halfCapD := capD.Amount / 2
		inv := PricedInvoice{
			Invoice: invoice.Invoice{
				ID:        "inv_half",
				Type:      invoice.TypeC,
				IssueDate: time.Date(2026, 10, 5, 10, 0, 0, 0, loc),
			},
			AmountARS: money.New(halfCapD, money.ARS),
		}

		res, err := AssessMonotributo(MonotributoInput{
			Period:          p,
			CurrentCategory: CatD,
			Invoices:        []PricedInvoice{inv},
			ScaleAt:         dynamicScaleProvider,
			FeeDueDate:      feeDueDate,
			Params:          params,
		})
		if err != nil {
			t.Fatalf("AssessMonotributo failed: %v", err)
		}

		expectedBp := halfCapD * 10000 / capD.Amount
		if res.PctOfCurrentCapBp != expectedBp {
			t.Errorf("PctOfCurrentCapBp = %d, want %d", res.PctOfCurrentCapBp, expectedBp)
		}
		expectedKBp := halfCapD * 10000 / capK.Amount
		if res.PctOfKCapBp != expectedKBp {
			t.Errorf("PctOfKCapBp = %d, want %d", res.PctOfKCapBp, expectedKBp)
		}
	})
}

func TestAssessMonotributo_InboundAndCard12m(t *testing.T) {
	loc := period.Location()
	p := period.Period{Year: 2026, Month: 10}
	feeDueDate := time.Date(2026, 11, 20, 0, 0, 0, 0, loc)

	inbounds := []DatedAmount{
		{Day: time.Date(2025, 11, 15, 10, 0, 0, 0, loc), Amount: money.New(500000, money.ARS)},
		{Day: time.Date(2026, 5, 20, 10, 0, 0, 0, loc), Amount: money.New(700000, money.ARS)},
		{Day: time.Date(2025, 10, 1, 10, 0, 0, 0, loc), Amount: money.New(999999, money.ARS)}, // Fuera de los 12 meses
	}

	cardSpends := []DatedAmount{
		{Day: time.Date(2026, 1, 10, 10, 0, 0, 0, loc), Amount: money.New(150000, money.ARS)},
		{Day: time.Date(2026, 8, 12, 10, 0, 0, 0, loc), Amount: money.New(250000, money.ARS)},
	}

	res, err := AssessMonotributo(MonotributoInput{
		Period:              p,
		CurrentCategory:     CatD,
		Invoices:            nil,
		ScaleAt:             dynamicScaleProvider,
		FeeDueDate:          feeDueDate,
		Params:              parameters.Defaults(),
		InboundMovementsARS: inbounds,
		CardSpendARS:        cardSpends,
	})
	if err != nil {
		t.Fatalf("AssessMonotributo failed: %v", err)
	}

	if res.Inbound12mARS.Amount != 1200000 {
		t.Errorf("Inbound12mARS = %v, want 1,200,000 ARS", res.Inbound12mARS)
	}
	if res.CardSpend12mARS.Amount != 400000 {
		t.Errorf("CardSpend12mARS = %v, want 400,000 ARS", res.CardSpend12mARS)
	}
}

func BenchmarkAssessMonotributo(b *testing.B) {
	loc := period.Location()
	p := period.Period{Year: 2026, Month: 10}
	feeDueDate := time.Date(2026, 11, 20, 0, 0, 0, 0, loc)
	params := parameters.Defaults()

	// Crear 300 clientes ficticios, cada uno con 13 facturas mensuales
	const numClients = 300
	clientsInvoices := make([][]PricedInvoice, numClients)

	for c := 0; c < numClients; c++ {
		invs := make([]PricedInvoice, 13)
		for m := 0; m < 13; m++ {
			year := 2025
			month := 10 + m
			if month > 12 {
				year = 2026
				month -= 12
			}
			t := time.Date(year, time.Month(month), 10, 10, 0, 0, 0, loc)
			invs[m] = PricedInvoice{
				Invoice: invoice.Invoice{
					ID:        fmt.Sprintf("inv_%d_%d", c, m),
					Type:      invoice.TypeC,
					IssueDate: t,
				},
				AmountARS: money.New(int64(200000+c*1000)*100, money.ARS),
			}
		}
		clientsInvoices[c] = invs
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Liquidar los 300 clientes
		for c := 0; c < numClients; c++ {
			_, err := AssessMonotributo(MonotributoInput{
				Period:          p,
				CurrentCategory: CatD,
				Invoices:        clientsInvoices[c],
				ScaleAt:         dynamicScaleProvider,
				FeeDueDate:      feeDueDate,
				Params:          params,
			})
			if err != nil {
				b.Fatalf("AssessMonotributo failed: %v", err)
			}
		}
	}
}
