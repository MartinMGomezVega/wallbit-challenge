import { useEffect } from 'react';
import { CheckCircle2, AlertCircle, X } from 'lucide-react';

export interface ToastProps {
  message: string;
  type?: 'success' | 'error';
  onClose: () => void;
  duration?: number;
}

export function Toast({ message, type = 'success', onClose, duration = 4000 }: ToastProps) {
  useEffect(() => {
    if (duration > 0) {
      const timer = setTimeout(onClose, duration);
      return () => clearTimeout(timer);
    }
  }, [onClose, duration]);

  const isSuccess = type === 'success';

  return (
    <div
      role={isSuccess ? 'status' : 'alert'}
      aria-live="polite"
      className="fixed bottom-6 right-6 z-50 flex items-center gap-3 rounded-card border shadow-lg px-4 py-3 text-xs font-medium bg-white"
      style={{
        borderColor: isSuccess ? 'var(--color-severity-green-border)' : 'var(--color-severity-red-border)',
      }}
    >
      {isSuccess ? (
        <CheckCircle2 className="h-4 w-4 text-severity-green-icon flex-shrink-0" />
      ) : (
        <AlertCircle className="h-4 w-4 text-severity-red-icon flex-shrink-0" />
      )}
      <span className={isSuccess ? 'text-severity-green-text' : 'text-severity-red-text'}>
        {message}
      </span>
      <button
        type="button"
        onClick={onClose}
        aria-label="Cerrar notificación"
        className="ml-2 rounded p-1 text-app-text-muted hover:bg-app-surface-muted hover:text-app-text focus:outline-none focus:ring-1 focus:ring-app-focus"
      >
        <X className="h-3.5 w-3.5" />
      </button>
    </div>
  );
}

