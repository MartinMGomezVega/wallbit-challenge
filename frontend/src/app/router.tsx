import { Navigate, createBrowserRouter, type RouteObject } from 'react-router-dom';
import { Layout } from './Layout';
import { DashboardPage } from '../features/closings-dashboard';
import { SummaryPage } from '../features/summary-dashboard';
import { ParametersPage } from '../features/parameters';
import { ClientDemoPage } from '../features/client-demo';

export const routes: RouteObject[] = [
  {
    element: <Layout />,
    children: [
      { path: '/', element: <Navigate to="/contador" replace /> },
      { path: '/contador', element: <DashboardPage /> },
      { path: '/contador/resumen', element: <SummaryPage /> },
      { path: '/contador/parametros', element: <ParametersPage /> },
      { path: '/cliente', element: <Navigate to="/cliente/cli_demo_15" replace /> },
      { path: '/cliente/:clientId', element: <ClientDemoPage /> },
      { path: '*', element: <Navigate to="/contador" replace /> },
    ],
  },
];

export const createAppRouter = () => createBrowserRouter(routes);
