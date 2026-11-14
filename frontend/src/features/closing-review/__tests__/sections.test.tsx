import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { SummarySection } from '../sections/SummarySection';
import { SignalsSection } from '../sections/SignalsSection';
import { MovementsSection } from '../sections/MovementsSection';
import { InvoicesSection } from '../sections/InvoicesSection';
import { HistorySection } from '../sections/HistorySection';
import { ClientMessageSection } from '../sections/ClientMessageSection';
import { ClosingDetailDTO } from '../schemas';
import * as api from '../api';

const mockDetailMono: ClosingDetailDTO = {
  closing: {
    id: 'cls_01',
    clientId: 'cli_01',
    period: '2026-10',
    status: 'READY_FOR_REVIEW',
    version: 1,
    resultHash: 'hash-mono-1234567890abcdef',
    severity: 'RED',
    description: 'Excede tope K',
    signals: [
      {
        code: 'MONO_EXCLUSION_K_CAP_EXCEEDED',
        severity: 'RED',
        message: 'Superó el tope de la categoría K el 14/10/2026',
        data: { exclusionDate: '2026-10-14', excessARS: 25000000 },
      },
    ],
    mono: {
      billedInPeriodUSD: { amount: 750000, currency: 'USD' },
      billedInPeriodARS: { amount: 1125000000, currency: 'ARS' },
      currentCategory: 'K',
      requiredCategory: undefined,
      billed12mARS: { amount: 12840000000, currency: 'ARS' },
      kCapARS: { amount: 12660000000, currency: 'ARS' },
      pctOfKCapBp: 10140,
      fee: { amount: 161444604, currency: 'ARS' },
      feeDueDate: '2026-11-20T00:00:00Z',
      exclusionReason: 'K_CAP_EXCEEDED',
      exclusionEffectiveDate: '2026-10-14',
    },
    ri: null,
    computedAt: '2026-11-01T03:00:00Z',
    paidAt: null,
    approvedAt: null,
    approvedBy: null,
  },
  client: {
    id: 'cli_01',
    name: 'Ignacio Sosa',
    cuit: '20-28222221-1',
    regime: 'MONOTRIBUTO',
    monotributoCategory: 'K',
  },
  periodMovements: [
    {
      movement: {
        id: 'mov_01',
        clientId: 'cli_01',
        type: 'INCOMING_WIRE',
        amount: { amount: 750000, currency: 'USD' },
        counterparty: 'Acme Corp',
        description: 'Monthly consulting retainer',
        occurredAt: '2026-10-15T14:30:00Z',
      },
      classification: {
        nature: 'INCOME',
        rule: 'DEFAULT_INCOME',
        confidence: 0.98,
        comment: 'Ingreso habitual',
        overridden: false,
      },
    },
  ],
  periodInvoices: [
    {
      invoice: {
        id: 'inv_01',
        clientId: 'cli_01',
        pointOfSale: 1,
        number: 104,
        type: 'Factura E',
        concept: 'Servicios de software',
        issueDate: '2026-10-15T15:00:00Z',
        totalAmount: { amount: 750000, currency: 'USD' },
        currency: 'USD',
        cae: '74125896325874',
        caeDueDate: '2026-10-25T00:00:00Z',
      },
      amountArs: { amount: 1125000000, currency: 'ARS' },
      fxRateUsed: 1500.0,
    },
  ],
  reconciliation: {
    matchedPairs: [
      {
        movement: {
          id: 'mov_01',
          clientId: 'cli_01',
          type: 'INCOMING_WIRE',
          amount: { amount: 750000, currency: 'USD' },
          counterparty: 'Acme Corp',
          description: 'Monthly consulting retainer',
          occurredAt: '2026-10-15T14:30:00Z',
        },
        invoice: {
          id: 'inv_01',
          clientId: 'cli_01',
          pointOfSale: 1,
          number: 104,
          type: 'Factura E',
          concept: 'Servicios de software',
          issueDate: '2026-10-15T15:00:00Z',
          totalAmount: { amount: 750000, currency: 'USD' },
          currency: 'USD',
          cae: '74125896325874',
          caeDueDate: '2026-10-25T00:00:00Z',
        },
        diffAmount: 0,
      },
    ],
    unmatchedIncome: [],
    unmatchedInvoices: [],
  },
  clientInput: {},
};

const mockDetailRI: ClosingDetailDTO = {
  closing: {
    id: 'cls_03',
    clientId: 'cli_03',
    period: '2026-10',
    status: 'APPROVED',
    version: 2,
    resultHash: 'hash-ri-9876543210abcdef',
    severity: 'GREEN',
    description: 'Todo conciliado',
    signals: [],
    mono: null,
    ri: {
      billedInPeriodUSD: { amount: 200000, currency: 'USD' },
      billedInPeriodARS: { amount: 300000000, currency: 'ARS' },
      debitoFiscalARS: { amount: 4200000, currency: 'ARS' },
      creditoFiscalARS: { amount: 800000, currency: 'ARS' },
      saldoTecnicoARS: { amount: 3400000, currency: 'ARS' },
      retencionesPercepcionesARS: { amount: 200000, currency: 'ARS' },
      ivaToPay: { amount: 3200000, currency: 'ARS' },
      vatDueDate: '2026-11-18T00:00:00Z',
      autonomousCategory: 'Cat. III',
      autonomousFee: { amount: 450000, currency: 'ARS' },
      selfEmployedDueDate: '2026-11-05T00:00:00Z',
      incomeTaxAdvance: { amount: 1200000, currency: 'ARS' },
      incomeTaxAdvanceDueDate: '2026-11-15T00:00:00Z',
      totalToPay: { amount: 4850000, currency: 'ARS' },
    },
    computedAt: '2026-11-01T03:00:00Z',
    paidAt: '2026-11-04T12:00:00Z',
    approvedAt: '2026-11-02T10:00:00Z',
    approvedBy: 'contador@estudio.com',
  },
  client: {
    id: 'cli_03',
    name: 'Ana Clara Díaz',
    cuit: '27-33333333-3',
    regime: 'RI',
    monotributoCategory: '',
  },
  periodMovements: [],
  periodInvoices: [],
  reconciliation: {
    matchedPairs: [],
    unmatchedIncome: [],
    unmatchedInvoices: [],
  },
  clientInput: {},
};

describe('Review Sections', () => {
  describe('SummarySection', () => {
    it('renderiza resumen de Monotributo con facturado, acumulado 12m y cuota', () => {
      render(<SummarySection detail={mockDetailMono} />);
      expect(screen.getByText(/facturado en el mes/i)).toBeInTheDocument();
      expect(screen.getByText(/acumulado 12 meses/i)).toBeInTheDocument();
      expect(screen.getByText(/categoría actual: k/i)).toBeInTheDocument();
    });

    it('renderiza sugerencia de recategorización si requiredCategory difiere', () => {
      const detailRecat: ClosingDetailDTO = {
        ...mockDetailMono,
        closing: {
          ...mockDetailMono.closing,
          mono: {
            ...mockDetailMono.closing.mono!,
            requiredCategory: 'H',
          },
        },
      };
      render(<SummarySection detail={detailRecat} />);
      expect(screen.getByText(/requiere pase a: h/i)).toBeInTheDocument();
    });

    it('renderiza desglose tributario para Responsable Inscripto', () => {
      render(<SummarySection detail={mockDetailRI} />);
      expect(screen.getByText('IVA')).toBeInTheDocument();
      expect(screen.getByText('Autónomos')).toBeInTheDocument();
      expect(screen.getByText('Anticipo Ganancias')).toBeInTheDocument();
      expect(screen.getByText(/total a ingresar/i)).toBeInTheDocument();
    });
  });

  describe('SignalsSection', () => {
    it('muestra señales con código, severidad y datos adicionales', () => {
      render(<SignalsSection detail={mockDetailMono} />);
      expect(screen.getByText('MONO_EXCLUSION_K_CAP_EXCEEDED')).toBeInTheDocument();
      expect(screen.getByText(/superó el tope de la categoría k/i)).toBeInTheDocument();
      expect(screen.getByText(/exclusiondate:/i)).toBeInTheDocument();
    });

    it('muestra señales de severidad YELLOW con estilo de revisión preventiva', () => {
      const detailYellow: ClosingDetailDTO = {
        ...mockDetailMono,
        closing: {
          ...mockDetailMono.closing,
          severity: 'YELLOW',
          signals: [
            {
              code: 'MONO_NEAR_K_CAP',
              severity: 'YELLOW',
              message: 'Cerca del tope de Monotributo',
            },
          ],
        },
      };
      render(<SignalsSection detail={detailYellow} />);
      expect(screen.getByText('MONO_NEAR_K_CAP')).toBeInTheDocument();
      expect(screen.getByText('Revisión preventiva')).toBeInTheDocument();
    });

    it('muestra estado sin alertas cuando la lista de señales está vacía', () => {
      render(<SignalsSection detail={mockDetailRI} />);
      expect(screen.getByText('Sin señales de alerta')).toBeInTheDocument();
    });
  });

  describe('MovementsSection', () => {
    it('renderiza tabla de movimientos con datos y clasificación', () => {
      render(<MovementsSection detail={mockDetailMono} onRefreshDetail={vi.fn()} />);
      expect(screen.getByText('Acme Corp')).toBeInTheDocument();
      expect(screen.getByText('DEFAULT_INCOME')).toBeInTheDocument();
      expect(screen.getByText('Ingreso')).toBeInTheDocument();
    });

    it('permite reclasificar movimiento mediante PATCH y refresca', async () => {
      const user = userEvent.setup();
      const onRefresh = vi.fn();
      const spy = vi.spyOn(api, 'updateMovementClassification').mockResolvedValueOnce({
        id: 'cls_mov_01',
        movementId: 'mov_01',
        nature: 'OTHER',
        rule: 'MANUAL_OVERRIDE',
        comment: 'Reclasificado a No ingreso por el contador',
        overridden: true,
        updatedAt: '2026-11-01T04:00:00Z',
        updatedBy: 'contador@estudio.com',
      });

      render(<MovementsSection detail={mockDetailMono} onRefreshDetail={onRefresh} />);

      const toggleBtn = screen.getByRole('button', { name: /cambiar a no ingreso/i });
      await user.click(toggleBtn);

      expect(spy).toHaveBeenCalledWith('mov_01', 'OTHER', expect.any(String));
      expect(onRefresh).toHaveBeenCalled();
    });

    it('muestra mensaje de error si falla la reclasificación', async () => {
      const user = userEvent.setup();
      vi.spyOn(api, 'updateMovementClassification').mockRejectedValueOnce(
        new Error('Fallo de red en reclasificación')
      );

      render(<MovementsSection detail={mockDetailMono} onRefreshDetail={vi.fn()} />);

      const toggleBtn = screen.getByRole('button', { name: /cambiar a no ingreso/i });
      await user.click(toggleBtn);

      expect(await screen.findByRole('alert')).toHaveTextContent('Fallo de red en reclasificación');
    });

    it('permite reclasificar movimiento de OTHER a INCOME', async () => {
      const user = userEvent.setup();
      const onRefresh = vi.fn();
      const spy = vi.spyOn(api, 'updateMovementClassification').mockResolvedValueOnce({
        id: 'cls_mov_01',
        movementId: 'mov_01',
        nature: 'INCOME',
        rule: 'MANUAL_OVERRIDE',
        comment: 'Reclasificado a Ingreso por el contador',
        overridden: true,
        updatedAt: '2026-11-01T04:00:00Z',
        updatedBy: 'contador@estudio.com',
      });

      const firstMovement = mockDetailMono.periodMovements[0]!;
      const detailOther: ClosingDetailDTO = {
        ...mockDetailMono,
        periodMovements: [
          {
            ...firstMovement,
            classification: {
              ...firstMovement.classification,
              nature: 'OTHER',
            },
          },
        ],
      };

      render(<MovementsSection detail={detailOther} onRefreshDetail={onRefresh} />);

      const toggleBtn = screen.getByRole('button', { name: /cambiar a ingreso/i });
      await user.click(toggleBtn);

      expect(spy).toHaveBeenCalledWith('mov_01', 'INCOME', expect.any(String));
      expect(onRefresh).toHaveBeenCalled();
    });

    it('muestra estado vacío si no hay movimientos', () => {
      render(<MovementsSection detail={mockDetailRI} onRefreshDetail={vi.fn()} />);
      expect(screen.getByText(/no se registraron movimientos/i)).toBeInTheDocument();
    });
  });

  describe('InvoicesSection', () => {
    it('renderiza facturas emitidas y pares conciliados', () => {
      render(<InvoicesSection detail={mockDetailMono} />);
      expect(screen.getByText(/facturas del período/i)).toBeInTheDocument();
      expect(screen.getByText('0001-00000104')).toBeInTheDocument();
      expect(screen.getByText(/pares conciliados/i)).toBeInTheDocument();
    });

    it('muestra estado vacío si no hay facturas emitidas', () => {
      render(<InvoicesSection detail={mockDetailRI} />);
      expect(screen.getByText(/no se emitieron facturas en este período/i)).toBeInTheDocument();
    });

    it('renderiza cobros sin factura y facturas sin cobro cuando existen', () => {
      const detailWithUnmatched: ClosingDetailDTO = {
        ...mockDetailMono,
        reconciliation: {
          matchedPairs: [],
          unmatchedIncome: [
            {
              id: 'mov_unmatched_1',
              clientId: 'cli_01',
              type: 'INCOMING_WIRE',
              amount: { amount: 50000, currency: 'USD' },
              counterparty: 'Client X',
              description: 'Payment without invoice',
              occurredAt: '2026-10-18T10:00:00Z',
            },
          ],
          unmatchedInvoices: [
            {
              id: 'inv_unmatched_1',
              clientId: 'cli_01',
              pointOfSale: 1,
              number: 105,
              type: 'Factura E',
              concept: 'Pending payment service',
              issueDate: '2026-10-20T10:00:00Z',
              totalAmount: { amount: 80000, currency: 'USD' },
              currency: 'USD',
              cae: '12345678901234',
              caeDueDate: '2026-10-30T00:00:00Z',
            },
          ],
        },
      };

      render(<InvoicesSection detail={detailWithUnmatched} />);
      expect(screen.getByText('Facturas sin cobro (1)')).toBeInTheDocument();
      expect(screen.getByText('Cobros sin factura (1)')).toBeInTheDocument();
      expect(screen.getByText(/Client X/)).toBeInTheDocument();
      expect(screen.getByText(/Factura Factura E 00000105/)).toBeInTheDocument();
    });
  });

  describe('HistorySection', () => {
    it('renderiza información de versión, hash, cómputo y estado de aprobación', () => {
      render(<HistorySection detail={mockDetailMono} />);
      expect(screen.getByText('v1')).toBeInTheDocument();
      expect(screen.getByText('hash-mono-1234567890abcdef')).toBeInTheDocument();
      expect(screen.getByText('Pendiente de aprobación')).toBeInTheDocument();
      expect(screen.getByText('Sin registrar')).toBeInTheDocument();
    });

    it('renderiza información de aprobación y pago cuando están registrados', () => {
      render(<HistorySection detail={mockDetailRI} />);
      expect(screen.getByText('v2')).toBeInTheDocument();
      expect(screen.getByText(/contador@estudio.com/i)).toBeInTheDocument();
      expect(screen.getByText('04/11/2026')).toBeInTheDocument();
    });
  });

  describe('ClientMessageSection', () => {
    it('renderiza borrador de aviso con montos a pagar y vencimientos', () => {
      render(<ClientMessageSection detail={mockDetailMono} />);
      expect(screen.getByText(/borrador de aviso al cliente/i)).toBeInTheDocument();
      expect(screen.getByText(/hola ignacio/i)).toBeInTheDocument();
    });

    it('renderiza borrador para responsable inscripto', () => {
      render(<ClientMessageSection detail={mockDetailRI} />);
      expect(screen.getByText(/hola ana/i)).toBeInTheDocument();
      expect(screen.getByText(/\$ 48\.500,00/i)).toBeInTheDocument();
    });

    it('renderiza estado notificado y mensaje borrador con nota', () => {
      const notifiedDetail: ClosingDetailDTO = {
        ...mockDetailMono,
        closing: {
          ...mockDetailMono.closing,
          status: 'NOTIFIED',
        },
        draftMessage: {
          title: 'Tu cierre de octubre 2026 está listo',
          body: 'Facturaste US$ 1.500,00. Cuota: $ 30.000,00.',
        },
      };

      const { rerender } = render(
        <ClientMessageSection detail={notifiedDetail} clientNote="Revisar deducciones" />
      );

      expect(screen.getByText('Notificado')).toBeInTheDocument();
      expect(screen.getByTestId('client-message-title')).toHaveTextContent('Tu cierre de octubre 2026 está listo');
      expect(screen.getByTestId('client-message-body')).toHaveTextContent(/nota de tu contador: revisar deducciones/i);

      // Sin nota
      rerender(<ClientMessageSection detail={notifiedDetail} />);
      expect(screen.getByTestId('client-message-body')).not.toHaveTextContent(/nota de tu contador/i);
    });

    it('muestra lo publicado (Published) con fecha cuando el cierre fue notificado', () => {
      const publishedDetail: ClosingDetailDTO = {
        ...mockDetailMono,
        closing: {
          ...mockDetailMono.closing,
          status: 'NOTIFIED',
          published: {
            version: 1,
            title: 'Cierre de octubre 2026 publicado',
            body: 'Tu cuota de octubre es $ 26.600,00 con vencimiento 20/11/2026.',
            figures: {
              billedUSD: { amount: 150000, currency: 'USD' },
              billedARS: { amount: 180000000, currency: 'ARS' },
              items: [],
              total: { amount: 2660000, currency: 'ARS' },
            },
            publishedAt: '2026-11-05T10:00:00Z',
          },
        },
      };

      render(<ClientMessageSection detail={publishedDetail} />);
      expect(screen.getByTestId('published-client-message')).toBeInTheDocument();
      expect(screen.getByTestId('published-message-title')).toHaveTextContent('Cierre de octubre 2026 publicado');
      expect(screen.getByTestId('published-message-body')).toHaveTextContent(/26\.600,00/);
      expect(screen.getByTestId('published-at')).toHaveTextContent('05/11/2026');
    });
  });
});
