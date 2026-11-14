import { describe, it, expect } from 'vitest';
import {
  getClosingIds,
  getClosingDetail,
  updateMovementClassification,
  confirmRegimeTransition,
  approveClosing,
  approveBatch,
} from '../api';

describe('closing-review api', () => {
  it('getClosingIds works with URLSearchParams', async () => {
    const sp = new URLSearchParams({ severity: 'RED' });
    const res = await getClosingIds('2026-10', sp);
    expect(res.ids.length).toBeGreaterThan(0);
    expect(res.total).toBe(res.ids.length);
  });

  it('getClosingIds works with Record and strips page/pageSize/undefined', async () => {
    const res = await getClosingIds('2026-10', {
      severity: 'RED',
      page: 1,
      pageSize: 50,
      search: undefined,
      category: '',
    });
    expect(res.ids.length).toBeGreaterThan(0);
  });

  it('getClosingIds works with empty params', async () => {
    const res = await getClosingIds('2026-10');
    expect(res.ids.length).toBe(10);
    expect(res.total).toBe(10);
  });

  it('getClosingDetail returns details for a valid closing', async () => {
    const detail = await getClosingDetail('cls_01');
    expect(detail.closing.id).toBe('cls_01');
    expect(detail.client.name).toBe('Ignacio Sosa');
    expect(detail.periodMovements.length).toBeGreaterThan(0);
    expect(detail.periodInvoices.length).toBeGreaterThan(0);
  });

  it('updateMovementClassification updates classification with custom and default comment', async () => {
    const resCustom = await updateMovementClassification('mov_01', 'OTHER', 'Gasto personal');
    expect(resCustom.movementId).toBe('mov_01');
    expect(resCustom.nature).toBe('OTHER');
    expect(resCustom.comment).toBe('Gasto personal');

    const resDefault = await updateMovementClassification('mov_01', 'INCOME');
    expect(resDefault.nature).toBe('INCOME');
    expect(resDefault.comment).toBe('Modificado por el contador');
  });

  it('confirmRegimeTransition calls the regime-change endpoint', async () => {
    const res = await confirmRegimeTransition('cli_01', '2026-10-14', true, 'Nota explicativa');
    expect(res).toBeDefined();

    const resDefault = await confirmRegimeTransition('cli_01', '2026-10-14');
    expect(resDefault).toBeDefined();
  });

  it('approveClosing calls approve endpoint with version and optional clientNote', async () => {
    const res = await approveClosing('cls_01', 1, 'Nota para el cliente');
    expect(res.id).toBe('cls_01');
    expect(res.status).toBe('NOTIFIED');

    const resWithoutNote = await approveClosing('cls_01', 1);
    expect(resWithoutNote.id).toBe('cls_01');
  });

  it('approveBatch calls approve-batch endpoint with items', async () => {
    const res = await approveBatch('2026-10', [
      { id: 'cls_01', version: 1 },
      { id: 'cls_02', version: 1 },
    ]);
    expect(res.approved).toBe(2);
    expect(res.failed).toBe(0);
    expect(res.results.length).toBe(2);
  });
});
