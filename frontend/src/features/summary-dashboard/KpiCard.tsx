import { ReactNode } from 'react';
import { Link } from 'react-router-dom';

export interface KpiCardProps {
  label: string;
  value: ReactNode;
  secondary?: ReactNode;
  to?: string;
  children?: ReactNode;
  className?: string;
}

export function KpiCard({
  label,
  value,
  secondary,
  to,
  children,
  className = '',
}: KpiCardProps) {
  const baseClasses = `flex flex-col justify-between rounded-lg border border-app-border bg-app-surface p-5 text-left ${className}`;

  const content = (
    <>
      <div>
        <div className="text-xs font-medium uppercase tracking-wide text-app-text-muted">
          {label}
        </div>
        <div className="mt-2 text-2xl font-semibold tabular-nums text-app-text">
          {value}
        </div>
      </div>
      {children}
      {secondary && (
        <div className="mt-2 text-sm text-app-text-muted">{secondary}</div>
      )}
    </>
  );

  if (to) {
    return (
      <Link
        to={to}
        className={`${baseClasses} cursor-pointer transition-colors hover:bg-app-surface-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-app-focus`}
      >
        {content}
      </Link>
    );
  }

  return <div className={baseClasses}>{content}</div>;
}

