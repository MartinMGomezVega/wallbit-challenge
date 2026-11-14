import { OctagonAlert, TriangleAlert, CircleCheck } from 'lucide-react';
import { Tooltip } from './Tooltip';

export type SeverityType = 'RED' | 'YELLOW' | 'GREEN' | 0 | 1 | 2;

export interface SeverityMeta {
  code: 'RED' | 'YELLOW' | 'GREEN';
  label: string;
  shortLabel: string;
  iconClass: string;
  bgClass: string;
  borderClass: string;
  textClass: string;
}

export function parseSeverity(s: SeverityType): 'RED' | 'YELLOW' | 'GREEN' {
  if (s === 'RED' || s === 2) return 'RED';
  if (s === 'YELLOW' || s === 1) return 'YELLOW';
  return 'GREEN';
}

export const SEVERITY_CONFIG: Record<'RED' | 'YELLOW' | 'GREEN', SeverityMeta> = {
  RED: {
    code: 'RED',
    label: 'Rojo: requiere acción',
    shortLabel: 'Requiere acción',
    iconClass: 'text-severity-red-icon',
    bgClass: 'bg-severity-red-bg',
    borderClass: 'border-severity-red-border',
    textClass: 'text-severity-red-text',
  },
  YELLOW: {
    code: 'YELLOW',
    label: 'Amarillo: revisar',
    shortLabel: 'Revisar',
    iconClass: 'text-severity-yellow-icon',
    bgClass: 'bg-severity-yellow-bg',
    borderClass: 'border-severity-yellow-border',
    textClass: 'text-severity-yellow-text',
  },
  GREEN: {
    code: 'GREEN',
    label: 'Verde: todo ok',
    shortLabel: 'Todo ok',
    iconClass: 'text-severity-green-icon',
    bgClass: 'bg-severity-green-bg',
    borderClass: 'border-severity-green-border',
    textClass: 'text-severity-green-text',
  },
};

export interface SeverityIndicatorProps {
  severity: SeverityType;
  mode?: 'table' | 'icon' | 'dot' | 'badge';
  withTooltip?: boolean;
  className?: string;
}

export function SeverityIndicator({
  severity,
  mode = 'table',
  withTooltip = true,
  className = '',
}: SeverityIndicatorProps) {
  const norm = parseSeverity(severity);
  const cfg = SEVERITY_CONFIG[norm];

  const renderIcon = () => {
    switch (norm) {
      case 'RED':
        return <OctagonAlert className={`h-4 w-4 ${cfg.iconClass}`} aria-hidden="true" />;
      case 'YELLOW':
        return <TriangleAlert className={`h-4 w-4 ${cfg.iconClass}`} aria-hidden="true" />;
      case 'GREEN':
        return <CircleCheck className={`h-4 w-4 ${cfg.iconClass}`} aria-hidden="true" />;
    }
  };

  const renderDot = () => {
    const dotColor =
      norm === 'RED'
        ? 'bg-severity-red-icon'
        : norm === 'YELLOW'
        ? 'bg-severity-yellow-icon'
        : 'bg-severity-green-icon';
    return <span className={`h-2.5 w-2.5 rounded-full ${dotColor}`} aria-hidden="true" />;
  };

  let content: JSX.Element;

  if (mode === 'dot') {
    content = (
      <span className={`inline-flex items-center ${className}`} aria-label={cfg.label}>
        {renderDot()}
      </span>
    );
  } else if (mode === 'icon') {
    content = (
      <span className={`inline-flex items-center ${className}`} aria-label={cfg.label}>
        {renderIcon()}
      </span>
    );
  } else if (mode === 'badge') {
    content = (
      <span
        className={`inline-flex items-center gap-1.5 rounded-control border px-2 py-0.5 text-xs font-medium ${cfg.bgClass} ${cfg.borderClass} ${cfg.textClass} ${className}`}
        aria-label={cfg.label}
      >
        {renderIcon()}
        <span>{cfg.shortLabel}</span>
      </span>
    );
  } else {
    // Mode 'table': dot + icon in 40px width container
    content = (
      <div
        className={`flex w-10 items-center justify-center gap-1.5 ${className}`}
        aria-label={cfg.label}
      >
        {renderDot()}
        {renderIcon()}
      </div>
    );
  }

  if (withTooltip && mode !== 'badge') {
    return <Tooltip content={cfg.label}>{content}</Tooltip>;
  }

  return content;
}

