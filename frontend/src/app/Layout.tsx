import { Link, NavLink, Outlet, useNavigate } from 'react-router-dom';
import { RoleSwitcher } from './RoleSwitcher';
import { useRole, type Role } from './roleStore';
import { WallbitLogo } from '../shared/ui/WallbitLogo';
import { NotificationsBell } from '../features/client-portal/NotificationsBell';

export function homeFor(role: Role): string {
  return role.kind === 'accountant' ? '/contador' : '/cliente';
}

export function Layout() {
  const role = useRole();
  const navigate = useNavigate();

  const toolName = role.kind === 'accountant' ? 'Contabilidad' : 'Impuestos';

  return (
    <div className="min-h-screen bg-app-bg text-app-text">
      <header className="sticky top-0 z-30 h-14 border-b border-neutral-800 bg-app-header-bg text-app-header-text">
        <div className="mx-auto flex h-full max-w-[1400px] items-center justify-between px-6">
          <div className="flex h-full items-center gap-6">
            <div className="flex items-center gap-3">
              <Link
                to={homeFor(role)}
                className="flex items-center rounded focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-app-brand"
              >
                <WallbitLogo variant="white" className="h-6 w-auto" />
              </Link>
              <div className="h-4 w-px bg-white/20" aria-hidden="true" />
              <span className="text-sm font-medium text-white">{toolName}</span>
            </div>

            {role.kind === 'accountant' && (
              <nav className="flex h-full items-center gap-6 text-sm text-neutral-400" aria-label="Navegación principal">
                <NavLink
                  to="/contador"
                  end
                  className={({ isActive }) =>
                    `flex h-full items-center border-b-2 pt-0.5 transition-colors hover:text-white ${
                      isActive ? 'border-app-brand font-medium text-white' : 'border-transparent'
                    }`
                  }
                >
                  Cierres
                </NavLink>
                <NavLink
                  to="/contador/parametros"
                  className={({ isActive }) =>
                    `flex h-full items-center border-b-2 pt-0.5 transition-colors hover:text-white ${
                      isActive ? 'border-app-brand font-medium text-white' : 'border-transparent'
                    }`
                  }
                >
                  Parámetros
                </NavLink>
              </nav>
            )}
          </div>
          <div className="flex items-center gap-3">
            {role.kind === 'client' && <NotificationsBell />}
            <RoleSwitcher onChange={(r) => navigate(homeFor(r))} />
          </div>
        </div>
      </header>
      <main className="mx-auto max-w-[1400px] px-6 py-6">
        <Outlet />
      </main>
    </div>
  );
}
