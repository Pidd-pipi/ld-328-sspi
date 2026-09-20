import { createBrowserRouter, Navigate } from 'react-router-dom';
import Shell from '../components/Shell';
import Login from '../pages/Login';
import Dashboard from '../pages/Dashboard';
import FoodManage from '../pages/FoodManage';
import ConsumptionManage from '../pages/ConsumptionManage';
import Statistics from '../pages/Statistics';
import FamilyManage from '../pages/FamilyManage';
import Recommendations from '../pages/Recommendations';
import Profile from '../pages/Profile';
import { RequireAdmin, RequireAuth } from './guards';

export const router = createBrowserRouter([
  { path: '/login', element: <Login /> },
  {
    path: '/',
    element: (
      <RequireAuth>
        <Shell />
      </RequireAuth>
    ),
    children: [
      { index: true, element: <Navigate to="/dashboard" replace /> },
      { path: 'dashboard', element: <Dashboard /> },
      { path: 'foods', element: <FoodManage /> },
      { path: 'consumptions', element: <ConsumptionManage /> },
      { path: 'statistics', element: <Statistics /> },
      { path: 'family', element: <RequireAdmin><FamilyManage /></RequireAdmin> },
      { path: 'recommendations', element: <Recommendations /> },
      { path: 'profile', element: <Profile /> },
    ],
  },
  { path: '*', element: <Navigate to="/" replace /> },
]);
