import React from 'react';
import { Clock } from 'lucide-react';
import { Tooltip } from './Tooltip';

export interface ComingSoonButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  children: React.ReactNode;
  tooltipContent?: React.ReactNode;
  size?: 'sm' | 'md';
}

export const ComingSoonButton = React.forwardRef<HTMLButtonElement, ComingSoonButtonProps>(
  (
    {
      children,
      tooltipContent = 'Próximamente',
      size = 'md',
      className = '',
      onClick: _onClick,
      onKeyDown,
      ...props
    },
    ref
  ) => {
    const sizeClasses =
      size === 'sm'
        ? 'h-8 px-3 text-xs gap-1.5'
        : 'h-10 px-4 text-sm gap-2';

    return (
      <Tooltip content={tooltipContent}>
        <button
          ref={ref}
          type="button"
          aria-disabled="true"
          onClick={(e) => {
            e.preventDefault();
            e.stopPropagation();
          }}
          onKeyDown={(e) => {
            if (e.key === 'Enter' || e.key === ' ') {
              e.preventDefault();
              e.stopPropagation();
            }
            onKeyDown?.(e);
          }}
          className={`inline-flex items-center justify-center font-medium rounded-control border border-app-border bg-app-surface text-app-text opacity-55 cursor-not-allowed focus:outline-none focus:ring-2 focus:ring-app-focus focus:ring-offset-2 transition-opacity ${sizeClasses} ${className}`}
          {...props}
        >
          <Clock className="w-3.5 h-3.5 flex-shrink-0" aria-hidden="true" />
          <span>{children}</span>
        </button>
      </Tooltip>
    );
  }
);

ComingSoonButton.displayName = 'ComingSoonButton';

