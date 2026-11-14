package usecase

import (
	"context"
	"math/rand/v2"
	"sort"
	"time"

	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/application/ports"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/closing"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/notification"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/period"
	"github.com/MartinMGomezVega/wallbit-challenge/backend/internal/domain/signal"
)

// SeedSeptemberClosings aprueba y notifica todos los cierres de 2026-09 con datos determinísticos.
func SeedSeptemberClosings(
	ctx context.Context,
	closingsRepo ports.ClosingRepository,
	clientsRepo ports.ClientRepository,
	idGen ports.IDGenerator,
	seed uint64,
) error {
	pSep := period.MustParse("2026-09")
	list, err := closingsRepo.ListByPeriod(ctx, pSep)
	if err != nil {
		return err
	}
	if len(list) == 0 {
		return nil
	}

	// Orden determinístico por clientID
	sort.Slice(list, func(i, j int) bool {
		return list[i].ClientID < list[j].ClientID
	})

	pcg := rand.NewPCG(seed, 0x5e97)
	rng := rand.New(pcg)

	for _, c := range list {
		cli, err := clientsRepo.Get(ctx, c.ClientID)
		if err != nil {
			return err
		}

		// ~85% publicadas hasta el día 10
		beforeDay10 := rng.Float64() < 0.85
		var day int
		if beforeDay10 {
			day = 1 + rng.IntN(10) // 1..10
		} else {
			day = 11 + rng.IntN(4) // 11..14
		}
		hour := 9 + rng.IntN(9)
		min := rng.IntN(60)
		approvedAt := time.Date(2026, 10, day, hour, min, 0, 0, period.Location())

		// Aprobación con approvedBy "accountant" y nota "Cierre de septiembre (dato semilla)"
		approval := closing.Approval{
			ID:         idGen.New("apr"),
			Version:    c.Version,
			ResultHash: c.ResultHash,
			ApprovedBy: "accountant",
			ApprovedAt: approvedAt,
			Note:       "Cierre de septiembre (dato semilla)",
			Batch:      rng.Float64() < 0.5,
		}

		// clientNote breve para rojos y amarillos
		var clientNote string
		switch c.Severity {
		case signal.SeverityRed:
			clientNote = "Cierre de septiembre: detectamos situaciones que requieren revisión antes del próximo vencimiento."
		case signal.SeverityYellow:
			clientNote = "Cierre de septiembre: parámetros bajo seguimiento preventivo."
		default:
			clientNote = ""
		}

		msg := notification.BuildClosingMessage(c, cli, clientNote)
		pub := BuildPublished(c, cli, msg.Title, msg.Body, approvedAt, "C.P. Mariana Quiroga", "Matrícula de ejemplo")

		_ = c.Approve(approval)
		_ = c.Publish(pub)
		c.FirstPublishedAt = &approvedAt
		c.Status = closing.StatusNotified

		// paidAt antes del vencimiento (20/10) para ~90% de los notificados;
		// el resto sin pagar o pagado tarde
		paidRoll := rng.Float64()
		if paidRoll < 0.90 {
			// Pagado antes o en la fecha de vencimiento (20/10)
			minPayDay := day
			if minPayDay < 2 {
				minPayDay = 2
			}
			maxPayDay := 20
			payDay := minPayDay
			if maxPayDay > minPayDay {
				payDay = minPayDay + rng.IntN(maxPayDay-minPayDay+1)
			}
			if payDay > 20 {
				payDay = 20
			}
			paidAt := time.Date(2026, 10, payDay, 10+rng.IntN(8), rng.IntN(60), 0, 0, period.Location())
			c.PaidAt = &paidAt
		} else if paidRoll < 0.95 {
			// Pagado tarde (entre 21 y 25 de octubre)
			lateDay := 21 + rng.IntN(5)
			paidAt := time.Date(2026, 10, lateDay, 14, 0, 0, 0, period.Location())
			c.PaidAt = &paidAt
		} else {
			// Sin pagar
			c.PaidAt = nil
		}

		if err := closingsRepo.Save(ctx, c); err != nil {
			return err
		}
	}

	return nil
}

// SeedOctoberApprovedClosings aprueba y notifica los cierres de octubre 2026 para cli_demo_01 y cli_demo_09
// reutilizando ApproveClosing (no escribir estados a mano), con el reloj fijado en 2026-11-06 (spec 017 criterio 3).
func SeedOctoberApprovedClosings(
	ctx context.Context,
	apprUC *ApproveClosing,
	closingsRepo ports.ClosingRepository,
) error {
	pOct := period.MustParse("2026-10")
	clientIDs := []string{"cli_demo_01", "cli_demo_09"}

	for _, cid := range clientIDs {
		c, found, err := closingsRepo.ByClientPeriod(ctx, cid, pOct)
		if err != nil {
			return err
		}
		if !found {
			continue
		}
		if c.Status == closing.StatusApproved || c.Status == closing.StatusNotified {
			continue
		}
		_, err = apprUC.Execute(ctx, ApproveClosingRequest{
			ClosingID:  c.ID,
			Version:    c.Version,
			ClientNote: "Cierre de octubre (dato semilla)",
			Batch:      false,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
