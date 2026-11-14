import { ClosingDetailDTO } from '../schemas';
import { formatMoney, formatDate } from '../../../shared/format';
import { Mail, Clock, CheckCircle2 } from 'lucide-react';

export interface ClientMessageSectionProps {
  detail: ClosingDetailDTO;
  clientNote?: string;
}

export function ClientMessageSection({ detail, clientNote }: ClientMessageSectionProps) {
  const { closing, client, draftMessage } = detail;
  const isNotified = closing.status === 'NOTIFIED';
  const isMono = client.regime === 'MONOTRIBUTO';
  const toPay = isMono ? closing.mono?.fee : closing.ri?.totalToPay;
  const dueDate = isMono ? closing.mono?.feeDueDate : (closing.ri?.vatDueDate || closing.ri?.selfEmployedDueDate);

  const fullBody = clientNote?.trim()
    ? `${draftMessage?.body || ''} Nota de tu contador: ${clientNote.trim()}`
    : draftMessage?.body;

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between border-b border-app-border pb-2">
        <div className="flex items-center gap-2">
          <Mail className="h-4 w-4 text-app-primary" />
          <h3 className="text-xs font-bold uppercase tracking-wider text-app-text">
            {draftMessage ? 'Aviso al cliente' : 'Borrador de aviso al cliente'}
          </h3>
        </div>
        {isNotified ? (
          <span className="inline-flex items-center gap-1 rounded-full bg-status-notified-bg px-2.5 py-0.5 text-[11px] font-medium text-status-notified-text">
            <CheckCircle2 className="h-3 w-3" />
            <span>Notificado</span>
          </span>
        ) : (
          <span className="inline-flex items-center gap-1 rounded-full bg-app-surface-muted px-2.5 py-0.5 text-[11px] font-medium text-app-text-muted">
            <Clock className="h-3 w-3" />
            <span>{draftMessage ? 'Borrador en vivo' : 'Pendiente de aprobación'}</span>
          </span>
        )}
      </div>

      <div className="rounded-card border border-app-border bg-app-surface p-5 space-y-4 shadow-sm font-sans">
        <div className="text-xs text-app-text-muted border-b border-app-border pb-2 flex items-center justify-between">
          <div>
            <span className="font-semibold text-app-text">Para: </span>
            {client.name} &lt;{client.id}@cliente.wallbit.com&gt;
          </div>
          <span className="text-[11px] font-mono text-app-text-muted">In-app + Notificación</span>
        </div>

        {draftMessage ? (
          <div className="space-y-2">
            <div className="text-xs font-bold text-app-text" data-testid="client-message-title">
              {draftMessage.title}
            </div>

            <div
              className="text-xs text-app-text leading-relaxed whitespace-pre-wrap rounded-control bg-app-surface-muted p-4 border border-app-border"
              data-testid="client-message-body"
            >
              {fullBody}
            </div>
          </div>
        ) : (
          <div className="text-xs text-app-text space-y-2 pt-1 leading-relaxed">
            <p>Hola {client.name.split(' ')[0]},</p>
            <p>
              Te informamos que hemos finalizado la liquidación correspondiente al período{' '}
              <span className="font-semibold">{closing.period}</span>.
            </p>
            {toPay && (
              <div className="rounded-control bg-app-surface-muted p-3 border border-app-border my-2">
                <div className="text-xs text-app-text-muted">Total a ingresar:</div>
                <div className="text-base font-bold text-app-text">{formatMoney(toPay)}</div>
                {dueDate && (
                  <div className="text-[11px] text-app-text-muted mt-1">
                    Vencimiento: {formatDate(dueDate, 'full')}
                  </div>
                )}
              </div>
            )}
            <p>
              Podés revisar el detalle completo de tus facturas y movimientos desde tu portal en Wallbit.
            </p>
            {clientNote && (
              <p className="font-medium text-app-primary">
                Nota de tu contador: {clientNote}
              </p>
            )}
            <p className="text-app-text-muted text-[11px] pt-2">
              Estudio Contable Asociado — Notificación generada por Wallbit Contabilidad
            </p>
          </div>
        )}

        <div className="text-[11px] text-app-text-muted border-t border-app-border pt-2 flex items-center justify-between">
          <span>Los emails y notificaciones push salientes no contienen montos ni CUIT por privacidad.</span>
          <span className="font-medium text-app-primary">Garantía de privacidad</span>
        </div>
      </div>
    </div>
  );
}
