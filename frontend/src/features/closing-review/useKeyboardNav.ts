import { useEffect } from 'react';

export interface KeyboardNavOptions {
  enabled?: boolean;
  onNext?: () => void;
  onPrev?: () => void;
  onClose?: () => void;
  onApprove?: () => void;
}

/**
 * Atajos de teclado para la cola de revisión (spec 009 criterio 5):
 * ← / → navegan, Esc cierra, A aprueba.
 * Se deshabilitan cuando el foco está en inputs/textarea/select.
 */
export function useKeyboardNav({
  enabled = true,
  onNext,
  onPrev,
  onClose,
  onApprove,
}: KeyboardNavOptions) {
  useEffect(() => {
    if (!enabled) return;

    function handleKeyDown(e: KeyboardEvent) {
      const target = e.target instanceof HTMLElement ? e.target : null;
      if (
        target &&
        (target.tagName === 'INPUT' ||
          target.tagName === 'TEXTAREA' ||
          target.tagName === 'SELECT' ||
          target.isContentEditable ||
          target.contentEditable === 'true' ||
          target.getAttribute('contenteditable') !== null)
      ) {
        return;
      }

      if (e.key === 'ArrowLeft') {
        e.preventDefault();
        onPrev?.();
      } else if (e.key === 'ArrowRight') {
        e.preventDefault();
        onNext?.();
      } else if (e.key === 'Escape') {
        e.preventDefault();
        onClose?.();
      } else if (e.key === 'a' || e.key === 'A') {
        if (!e.ctrlKey && !e.metaKey && !e.altKey) {
          e.preventDefault();
          onApprove?.();
        }
      }
    }

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [enabled, onNext, onPrev, onClose, onApprove]);
}
