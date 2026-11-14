import { Link, NavLink, Outlet, useSearchParams } from 'react-router-dom';
import { WallbitLogo } from '../shared/ui/WallbitLogo';
import { PeriodSelect } from './PeriodSelect';

export function Layout() {
  const [searchParams] = useSearchParams();
  const periodParam = searchParams.get('period');
  const periodQuery = periodParam ? `?period=${encodeURIComponent(periodParam)}` : '';

  return (
    <div className="min-h-screen bg-app-bg text-app-text">
      <header className="sticky top-0 z-30 h-14 border-b border-neutral-800 bg-app-header-bg text-app-header-text">
        <div className="mx-auto flex h-full max-w-[1400px] items-center justify-between px-6">
          <div className="flex h-full items-center gap-8">
            <Link
              to={`/contador${periodQuery}`}
              className="flex items-center rounded focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-app-brand"
            >
              <WallbitLogo variant="white" className="h-6 w-auto" />
            </Link>

            <nav className="flex h-full items-center gap-6 text-sm text-[#A3A3A3]" aria-label="Navegación principal">
              <NavLink
                to={`/contador${periodQuery}`}
                end
                className={({ isActive }) =>
                  `flex h-full items-center border-b-2 pt-0.5 transition-colors hover:text-white ${
                    isActive ? 'border-[#0D99FF] font-medium text-white' : 'border-transparent'
                  }`
                }
              >
                Cierres
              </NavLink>
              <NavLink
                to={`/contador/resumen${periodQuery}`}
                className={({ isActive }) =>
                  `flex h-full items-center border-b-2 pt-0.5 transition-colors hover:text-white ${
                    isActive ? 'border-[#0D99FF] font-medium text-white' : 'border-transparent'
                  }`
                }
              >
                Resumen
              </NavLink>
              <NavLink
                to="/contador/parametros"
                className={({ isActive }) =>
                  `flex h-full items-center border-b-2 pt-0.5 transition-colors hover:text-white ${
                    isActive ? 'border-[#0D99FF] font-medium text-white' : 'border-transparent'
                  }`
                }
              >
                Parámetros
              </NavLink>
            </nav>
          </div>
          <div className="flex items-center gap-4">
            <PeriodSelect />
            <div
              className="flex items-center gap-2 rounded-md px-2 py-1"
              title="Sesión de demo"
              data-testid="accountant-identity"
            >
              <span className="flex h-7 w-7 items-center justify-center rounded-full bg-[#262626] text-xs font-semibold text-white">
                C
              </span>
              <span className="text-sm text-[#D4D4D4]">Contador</span>
            </div>
          </div>
        </div>
      </header>
      <main className="mx-auto max-w-[1400px] px-6 py-6">
        <Outlet />
      </main>
    </div>
  );
}
