interface ConfirmStepProps {
  canConfirm: boolean;
  loading: boolean;
  onConfirm: () => void;
  error?: string | null;
}

export function ConfirmStep({ canConfirm, loading, onConfirm, error }: ConfirmStepProps) {
  return (
    <div className="space-y-2 pt-2">
      {error && (
        <div
          role="alert"
          className="rounded-xl bg-red-50 p-2.5 text-xs text-red-700 border border-red-200"
        >
          {error}
        </div>
      )}

      <button
        type="button"
        disabled={!canConfirm || loading}
        onClick={onConfirm}
        className="flex h-11 w-full items-center justify-center rounded-xl bg-neutral-900 px-4 text-sm font-semibold text-white shadow transition-all hover:bg-neutral-800 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-black disabled:cursor-not-allowed disabled:opacity-40"
      >
        {loading ? 'Confirmando...' : 'Listo, ya cargué todo'}
      </button>

      {!canConfirm && (
        <p className="text-center text-[11px] text-neutral-500">
          Respondé la pregunta sobre cobros externos para continuar.
        </p>
      )}
    </div>
  );
}

