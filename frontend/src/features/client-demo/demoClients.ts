export interface DemoClient {
  id: string;
  name: string;
  regime: 'Monotributo' | 'Responsable Inscripto';
  scenario: string;
  shortScenario: string;
}

export const DEMO_CLIENTS: DemoClient[] = [
  {
    id: 'cli_demo_15',
    name: 'Paula Rojas',
    regime: 'Monotributo',
    scenario: 'Check-in pendiente: carga de comprobantes en vivo',
    shortScenario: 'Check-in pendiente',
  },
  {
    id: 'cli_demo_01',
    name: 'Lucía Fernández',
    regime: 'Monotributo',
    scenario: 'Octubre aprobado: tarjeta validada (Monotributo)',
    shortScenario: 'Octubre aprobado (Monotributo)',
  },
  {
    id: 'cli_demo_09',
    name: 'Camila Torres',
    regime: 'Responsable Inscripto',
    scenario: 'Octubre aprobado: tarjeta validada (Responsable Inscripto)',
    shortScenario: 'Octubre aprobado (RI)',
  },
  {
    id: 'cli_demo_02',
    name: 'Ignacio Sosa',
    regime: 'Monotributo',
    scenario: 'En revisión (octubre) + septiembre validado. Al aprobar ve aviso de exclusión',
    shortScenario: 'En revisión (exclusión)',
  },
  {
    id: 'cli_demo_03',
    name: 'Carla Benítez',
    regime: 'Monotributo',
    scenario: 'En revisión (octubre) + septiembre validado',
    shortScenario: 'En revisión (faltan facturas)',
  },
];
