import { useState } from 'react';
import * as DropdownMenu from '@radix-ui/react-dropdown-menu';
import { useQuery } from '@tanstack/react-query';
import { setRole, useRole, type Role } from './roleStore';
import { fetchDemoClients } from '../features/client-portal/api';

export type DemoClientOption = { id: string; name: string };

type Props = {
  clients?: DemoClientOption[];
  onChange?: (role: Role) => void;
};

function roleLabel(role: Role, clients: DemoClientOption[]): string {
  if (role.kind === 'accountant') return 'Contador';
  const c = clients.find((x) => x.id === role.clientId);
  return c ? `Cliente: ${c.name}` : `Cliente: ${role.clientId}`;
}

function RoleSwitcherPresentational({ clients, onChange }: { clients: DemoClientOption[]; onChange?: (role: Role) => void }) {
  const role = useRole();
  const [customId, setCustomId] = useState('');

  const choose = (r: Role) => {
    setRole(r);
    onChange?.(r);
  };

  return (
    <div className="flex items-center gap-2 text-sm">
      <span className="text-neutral-400">Ver como:</span>
      <DropdownMenu.Root>
        <DropdownMenu.Trigger
          className="rounded-control border border-neutral-700 bg-neutral-900 px-3 py-1.5 font-medium text-neutral-200 transition-colors hover:bg-neutral-800 focus:outline-none focus-visible:ring-2 focus-visible:ring-app-brand"
          aria-label="Cambiar rol de demo"
        >
          {roleLabel(role, clients)} ▾
        </DropdownMenu.Trigger>
        <DropdownMenu.Portal>
          <DropdownMenu.Content
            align="end"
            sideOffset={4}
            className="z-50 max-h-96 min-w-[16rem] overflow-y-auto rounded-card border border-app-border bg-app-surface p-1 shadow-xl"
          >
            <DropdownMenu.Item
              className="cursor-pointer rounded-control px-2.5 py-1.5 text-sm text-app-text outline-none data-[highlighted]:bg-app-primary-soft data-[highlighted]:text-app-primary"
              onSelect={() => choose({ kind: 'accountant' })}
            >
              Contador
            </DropdownMenu.Item>
            {clients.length > 0 && <DropdownMenu.Separator className="my-1 h-px bg-app-border" />}
            {clients.map((c) => (
              <DropdownMenu.Item
                key={c.id}
                className="cursor-pointer rounded-control px-2.5 py-1.5 text-sm text-app-text outline-none data-[highlighted]:bg-app-primary-soft data-[highlighted]:text-app-primary"
                onSelect={() => choose({ kind: 'client', clientId: c.id })}
              >
                {c.name}
              </DropdownMenu.Item>
            ))}
          </DropdownMenu.Content>
        </DropdownMenu.Portal>
      </DropdownMenu.Root>
      {clients.length === 0 && (
        <form
          className="flex items-center gap-1"
          onSubmit={(e) => {
            e.preventDefault();
            const id = customId.trim();
            if (id) choose({ kind: 'client', clientId: id });
          }}
        >
          <label htmlFor="demo-client-id" className="sr-only">
            ID de cliente
          </label>
          <input
            id="demo-client-id"
            value={customId}
            onChange={(e) => setCustomId(e.target.value)}
            placeholder="cli_demo_15"
            className="w-32 rounded-control border border-neutral-700 bg-neutral-900 px-2 py-1 text-sm text-neutral-200 placeholder:text-neutral-500 focus:outline-none focus-visible:ring-2 focus-visible:ring-app-brand"
          />
          <button
            type="submit"
            className="rounded-control border border-neutral-700 bg-neutral-900 px-2.5 py-1 text-sm text-neutral-200 hover:bg-neutral-800 focus:outline-none focus-visible:ring-2 focus-visible:ring-app-brand"
          >
            Ir
          </button>
        </form>
      )}
    </div>
  );
}

function RoleSwitcherQuery({ onChange }: { onChange?: (role: Role) => void }) {
  const { data: clients = [] } = useQuery({
    queryKey: ['demo-clients'],
    queryFn: fetchDemoClients,
    staleTime: 60_000,
  });

  return <RoleSwitcherPresentational clients={clients} onChange={onChange} />;
}

/** Selector de rol de demo: "Ver como: Contador | Cliente…". */
export function RoleSwitcher({ clients, onChange }: Props) {
  if (clients !== undefined) {
    return <RoleSwitcherPresentational clients={clients} onChange={onChange} />;
  }
  return <RoleSwitcherQuery onChange={onChange} />;
}
