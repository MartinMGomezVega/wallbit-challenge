import * as DropdownMenu from '@radix-ui/react-dropdown-menu';
import { Bell } from 'lucide-react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { fetchMyNotifications, markNotificationRead } from './api';
import { NotificationItem } from './schemas';

function formatDate(isoStr: string): string {
  try {
    const d = new Date(isoStr);
    return d.toLocaleDateString('es-AR', {
      day: '2-digit',
      month: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    });
  } catch {
    return isoStr;
  }
}

export function NotificationsBell() {
  const queryClient = useQueryClient();

  const { data: notifications = [] } = useQuery<NotificationItem[]>({
    queryKey: ['my-notifications'],
    queryFn: fetchMyNotifications,
    refetchInterval: 30_000,
  });

  const markMutation = useMutation({
    mutationFn: markNotificationRead,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['my-notifications'] });
    },
  });

  const unreadCount = notifications.filter((n) => !n.readAt).length;

  const handleOpenChange = (open: boolean) => {
    if (open && unreadCount > 0) {
      const unreadList = notifications.filter((n) => !n.readAt);
      unreadList.forEach((n) => {
        markMutation.mutate(n.id);
      });
    }
  };

  return (
    <DropdownMenu.Root onOpenChange={handleOpenChange}>
      <DropdownMenu.Trigger
        className="relative rounded-control p-1.5 text-neutral-300 transition-colors hover:bg-neutral-800 hover:text-white focus:outline-none focus-visible:ring-2 focus-visible:ring-app-brand"
        aria-label={`Notificaciones${unreadCount > 0 ? ` (${unreadCount} no leídas)` : ''}`}
      >
        <Bell className="h-5 w-5" aria-hidden="true" />
        {unreadCount > 0 && (
          <span
            data-testid="notifications-badge"
            className="absolute -top-1 -right-1 flex h-4 min-w-[1rem] items-center justify-center rounded-full bg-severity-red-icon px-1 text-[10px] font-bold text-white tabular-nums shadow-sm"
          >
            {unreadCount > 9 ? '9+' : unreadCount}
          </span>
        )}
      </DropdownMenu.Trigger>
      <DropdownMenu.Portal>
        <DropdownMenu.Content
          align="end"
          sideOffset={8}
          className="z-50 w-80 max-h-96 overflow-y-auto rounded-card border border-app-border bg-app-surface p-2 shadow-xl"
        >
          <div className="flex items-center justify-between px-2 py-1.5 border-b border-app-border">
            <span className="text-sm font-semibold text-app-text">Notificaciones</span>
            {unreadCount > 0 && (
              <span className="text-xs text-app-text-muted">
                {unreadCount} nueva{unreadCount === 1 ? '' : 's'}
              </span>
            )}
          </div>

          {notifications.length === 0 ? (
            <div className="py-6 text-center text-sm text-app-text-muted">
              No tenés notificaciones
            </div>
          ) : (
            <div className="divide-y divide-app-border">
              {notifications.map((n) => {
                const isUnread = !n.readAt;
                return (
                  <div
                    key={n.id}
                    className={`p-2.5 transition-colors ${
                      isUnread ? 'bg-app-surface-muted/60' : 'bg-app-surface'
                    }`}
                  >
                    <div className="flex items-start justify-between gap-2">
                      <div className="flex items-center gap-1.5">
                        {isUnread && (
                          <span
                            data-testid="unread-dot"
                            className="h-2 w-2 rounded-full bg-app-primary flex-shrink-0"
                            aria-hidden="true"
                          />
                        )}
                        <span className="text-sm font-medium text-app-text leading-tight">
                          {n.title}
                        </span>
                      </div>
                      <span className="text-[11px] text-app-text-muted whitespace-nowrap">
                        {formatDate(n.createdAt)}
                      </span>
                    </div>
                    <p className="mt-1 text-xs text-app-text-muted leading-relaxed">
                      {n.body}
                    </p>
                  </div>
                );
              })}
            </div>
          )}
        </DropdownMenu.Content>
      </DropdownMenu.Portal>
    </DropdownMenu.Root>
  );
}
