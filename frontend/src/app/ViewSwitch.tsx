import { Link, useLocation } from 'react-router-dom';

export function ViewSwitch() {
  const location = useLocation();
  const isClient = location.pathname.startsWith('/cliente');

  return (
    <div
      role="group"
      aria-label="Selector de vista"
      className="inline-flex items-center rounded-lg bg-[#262626] p-0.5 text-xs font-medium"
      data-testid="view-switch"
    >
      <Link
        to="/contador"
        className={`rounded-md px-2.5 py-1 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-app-brand ${
          !isClient ? 'bg-[#383838] font-semibold text-white shadow-sm' : 'text-[#A3A3A3] hover:text-white'
        }`}
        aria-current={!isClient ? 'page' : undefined}
      >
        Contador
      </Link>
      <Link
        to="/cliente/cli_demo_15"
        className={`rounded-md px-2.5 py-1 transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-app-brand ${
          isClient ? 'bg-[#383838] font-semibold text-white shadow-sm' : 'text-[#A3A3A3] hover:text-white'
        }`}
        aria-current={isClient ? 'page' : undefined}
      >
        Cliente (demo)
      </Link>
    </div>
  );
}

