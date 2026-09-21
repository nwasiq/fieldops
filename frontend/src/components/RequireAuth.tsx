import { Navigate, Outlet, useLocation } from 'react-router-dom';
import { useAuth } from '../lib/auth';
import type { Role } from '../types';

export function RequireAuth({ roles }: { roles?: readonly Role[] }) {
  const { user } = useAuth();
  const location = useLocation();

  if (!user) return <Navigate to="/login" replace state={{ from: location }} />;
  if (roles && !roles.includes(user.role)) return <Navigate to="/visits" replace />;
  return <Outlet />;
}
