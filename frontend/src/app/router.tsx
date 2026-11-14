import { Navigate, createBrowserRouter, type RouteObject } from 'react-router-dom';
import { Layout, homeFor } from './Layout';
import { getRole } from './roleStore';
import { DashboardPage } from '../features/closings-dashboard';
import { SummaryPage } from '../features/summary-dashboard';
import { ParametersPage } from '../features/parameters';
import { ClientPage } from '../features/client-portal';

function RoleRedirect() {
  return <Navigate to={homeFor(getRole())} replace />;
}

export const routes: RouteObject[] = [
  {
    element: <Layout />,
    children: [
      { path: '/', element: <RoleRedirect /> },
      { path: '/contador', element: <DashboardPage /> },
      { path: '/contador/resumen', element: <SummaryPage /> },
      { path: '/contador/parametros', element: <ParametersPage /> },
      { path: '/cliente', element: <ClientPage /> },
      { path: '*', element: <RoleRedirect /> },
    ],
  },
];

export const createAppRouter = () => createBrowserRouter(routes);

