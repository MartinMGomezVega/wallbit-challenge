import { Navigate, createBrowserRouter, type RouteObject } from 'react-router-dom';
import { Layout } from './Layout';
import { DashboardPage } from '../features/closings-dashboard';
import { SummaryPage } from '../features/summary-dashboard';
import { ParametersPage } from '../features/parameters';

export const routes: RouteObject[] = [
  {
    element: <Layout />,
    children: [
      { path: '/', element: <Navigate to="/contador" replace /> },
      { path: '/contador', element: <DashboardPage /> },
      { path: '/contador/resumen', element: <SummaryPage /> },
      { path: '/contador/parametros', element: <ParametersPage /> },
      { path: '*', element: <Navigate to="/contador" replace /> },
    ],
  },
];

export const createAppRouter = () => createBrowserRouter(routes);
