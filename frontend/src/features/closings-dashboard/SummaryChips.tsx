import { OctagonAlert, TriangleAlert, CircleCheck } from 'lucide-react';
import { ClosingsSummaryResponse } from './schemas';

export interface SummaryChipsProps {
  summary?: ClosingsSummaryResponse;
  selectedSeverities: string[];
  onToggleSeverity: (severity: string) => void;
  onReviewSeverity?: (severity: string) => void;
  className?: string;
}

export function SummaryChips({
  summary,
  selectedSeverities,
  onToggleSeverity,
  onReviewSeverity,
  className = '',
}: SummaryChipsProps) {
  const redCount = summary?.bySeverity['RED'] ?? 0;
  const yellowCount = summary?.bySeverity['YELLOW'] ?? 0;
  const greenCount = summary?.bySeverity['GREEN'] ?? 0;

  const isRedActive = selectedSeverities.includes('RED');
  const isYellowActive = selectedSeverities.includes('YELLOW');
  const isGreenActive = selectedSeverities.includes('GREEN');

  return (
    <div className={`flex flex-wrap items-center gap-3 ${className}`}>
      {/* Chip Rojo */}
      <div
        className={`inline-flex h-9 items-center rounded-full bg-white transition-colors duration-150 ${
          isRedActive
            ? 'border-2 border-severity-red-icon bg-severity-red-bg'
            : 'border border-severity-red-border hover:bg-severity-red-bg/50'
        }`}
      >
        <button
          type="button"
          onClick={() => onToggleSeverity('RED')}
          aria-pressed={isRedActive}
          className="flex h-full items-center gap-2 pl-3 pr-2 text-xs font-medium text-severity-red-text focus:outline-none focus:ring-2 focus:ring-app-focus rounded-l-full"
        >
          <OctagonAlert className="h-4 w-4 text-severity-red-icon" aria-hidden="true" />
          <span>Rojos</span>
          <span className="font-semibold tabular-nums">{redCount}</span>
        </button>
        {onReviewSeverity && (
          <>
            <span className="h-4 w-[1px] bg-severity-red-border" aria-hidden="true" />
            <button
              type="button"
              onClick={() => onReviewSeverity('RED')}
              className="px-2.5 text-xs text-severity-red-text hover:underline focus:outline-none focus:ring-2 focus:ring-app-focus rounded-r-full"
            >
              Revisar
            </button>
          </>
        )}
      </div>

      {/* Chip Amarillo */}
      <div
        className={`inline-flex h-9 items-center rounded-full bg-white transition-colors duration-150 ${
          isYellowActive
            ? 'border-2 border-severity-yellow-icon bg-severity-yellow-bg'
            : 'border border-severity-yellow-border hover:bg-severity-yellow-bg/50'
        }`}
      >
        <button
          type="button"
          onClick={() => onToggleSeverity('YELLOW')}
          aria-pressed={isYellowActive}
          className="flex h-full items-center gap-2 pl-3 pr-2 text-xs font-medium text-severity-yellow-text focus:outline-none focus:ring-2 focus:ring-app-focus rounded-l-full"
        >
          <TriangleAlert className="h-4 w-4 text-severity-yellow-icon" aria-hidden="true" />
          <span>Amarillos</span>
          <span className="font-semibold tabular-nums">{yellowCount}</span>
        </button>
        {onReviewSeverity && (
          <>
            <span className="h-4 w-[1px] bg-severity-yellow-border" aria-hidden="true" />
            <button
              type="button"
              onClick={() => onReviewSeverity('YELLOW')}
              className="px-2.5 text-xs text-severity-yellow-text hover:underline focus:outline-none focus:ring-2 focus:ring-app-focus rounded-r-full"
            >
              Revisar
            </button>
          </>
        )}
      </div>

      {/* Chip Verde */}
      <div
        className={`inline-flex h-9 items-center rounded-full bg-white transition-colors duration-150 ${
          isGreenActive
            ? 'border-2 border-severity-green-icon bg-severity-green-bg'
            : 'border border-severity-green-border hover:bg-severity-green-bg/50'
        }`}
      >
        <button
          type="button"
          onClick={() => onToggleSeverity('GREEN')}
          aria-pressed={isGreenActive}
          className="flex h-full items-center gap-2 px-3 text-xs font-medium text-severity-green-text focus:outline-none focus:ring-2 focus:ring-app-focus rounded-full"
        >
          <CircleCheck className="h-4 w-4 text-severity-green-icon" aria-hidden="true" />
          <span>Verdes</span>
          <span className="font-semibold tabular-nums">{greenCount}</span>
        </button>
      </div>
    </div>
  );
}

