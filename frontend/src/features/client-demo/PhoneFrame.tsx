import type { ReactNode } from 'react';
import { WallbitLogo } from '../../shared/ui/WallbitLogo';

interface PhoneFrameProps {
  clientName?: string;
  children: ReactNode;
}

export function PhoneFrame({ clientName, children }: PhoneFrameProps) {
  return (
    <div className="flex flex-col items-center">
      <div
        role="region"
        aria-label="Vista demo de la app móvil de Wallbit"
        className="relative flex h-[780px] w-[390px] flex-col overflow-hidden rounded-[36px] border-[8px] border-[#1E1E1E] bg-[#F1F1F1] text-[15px] text-[#171717] shadow-2xl"
      >
        {/* Dynamic Island / Notch */}
        <div className="relative z-20 flex h-8 w-full items-center justify-between px-6 pt-1 text-xs font-semibold text-black">
          <span>9:41</span>
          <div className="h-4 w-20 rounded-full bg-black" aria-hidden="true" />
          <div className="flex items-center gap-1 text-[11px]" aria-hidden="true">
            <span>5G</span>
            <div className="h-2.5 w-4 rounded-sm border border-black p-0.5">
              <div className="h-full w-full bg-black rounded-2xs" />
            </div>
          </div>
        </div>

        {/* Mobile App Header */}
        <header className="flex h-12 items-center justify-between border-b border-neutral-200 bg-white px-4">
          <WallbitLogo variant="dark" className="h-5 w-auto" />
          {clientName && (
            <div className="flex items-center gap-2">
              <span className="text-xs font-medium text-neutral-600 truncate max-w-[150px]">
                {clientName}
              </span>
              <div
                className="flex h-6 w-6 items-center justify-center rounded-full bg-neutral-900 text-[10px] font-bold text-white"
                aria-hidden="true"
              >
                {clientName.charAt(0)}
              </div>
            </div>
          )}
        </header>

        {/* Scrollable Screen Content */}
        <div className="flex-1 overflow-y-auto px-4 py-4 space-y-4">
          {children}
        </div>

        {/* Home Indicator */}
        <div className="flex h-5 w-full items-center justify-center pb-1">
          <div className="h-1 w-32 rounded-full bg-neutral-400" aria-hidden="true" />
        </div>
      </div>

      <p className="mt-3 text-center text-xs font-medium text-[#737373]">
        Así se vería dentro de la app de Wallbit
      </p>
    </div>
  );
}
