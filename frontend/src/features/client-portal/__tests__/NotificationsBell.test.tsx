import { describe, it, expect } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { server } from '../../../mocks/server';
import { NotificationsBell } from '../NotificationsBell';
import { Providers, createQueryClient } from '../../../app/providers';

describe('NotificationsBell', () => {
  it('renders unread badge and opens panel showing notification details', async () => {
    const user = userEvent.setup();
    const qc = createQueryClient();

    render(
      <Providers client={qc}>
        <NotificationsBell />
      </Providers>
    );

    // Initial render with mock notification (unread)
    const badge = await screen.findByTestId('notifications-badge');
    expect(badge).toHaveTextContent('1');

    const trigger = screen.getByRole('button', { name: /notificaciones/i });
    await user.click(trigger);

    // Dropdown content should display title and body
    expect(await screen.findByText('Cierre de Septiembre 2026 listo')).toBeInTheDocument();
    expect(screen.getByText('Tu cuota mensual es de $26.600.')).toBeInTheDocument();
    expect(screen.getByTestId('unread-dot')).toBeInTheDocument();
  });

  it('marks unread notifications as read on opening', async () => {
    const user = userEvent.setup();
    const qc = createQueryClient();
    let readCalledWithId = '';

    server.use(
      http.post('/api/v1/me/notifications/:id/read', ({ params }) => {
        readCalledWithId = params.id as string;
        return HttpResponse.json({ id: params.id, status: 'read' });
      })
    );

    render(
      <Providers client={qc}>
        <NotificationsBell />
      </Providers>
    );

    await screen.findByTestId('notifications-badge');
    const trigger = screen.getByRole('button', { name: /notificaciones/i });
    await user.click(trigger);

    await waitFor(() => {
      expect(readCalledWithId).toBe('notif_demo_1');
    });
  });

  it('shows empty state when there are no notifications', async () => {
    const user = userEvent.setup();
    const qc = createQueryClient();

    server.use(
      http.get('/api/v1/me/notifications', () => {
        return HttpResponse.json([]);
      })
    );

    render(
      <Providers client={qc}>
        <NotificationsBell />
      </Providers>
    );

    const trigger = await screen.findByRole('button', { name: 'Notificaciones' });
    expect(screen.queryByTestId('notifications-badge')).not.toBeInTheDocument();

    await user.click(trigger);

    expect(await screen.findByText('No tenés notificaciones')).toBeInTheDocument();
  });
});
