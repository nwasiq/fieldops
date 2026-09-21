import { Navigate, Route, Routes } from 'react-router-dom';
import { Layout } from './components/Layout';
import { RequireAuth } from './components/RequireAuth';
import { DailyReportPage } from './pages/DailyReportPage';
import { LoginPage } from './pages/LoginPage';
import { VisitDetailPage } from './pages/VisitDetailPage';
import { VisitsPage } from './pages/VisitsPage';
import type { Role } from './types';

const REPORT_ROLES: readonly Role[] = ['admin', 'dispatcher'];

export function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route element={<RequireAuth />}>
        <Route element={<Layout />}>
          <Route path="/" element={<Navigate to="/visits" replace />} />
          <Route path="/visits" element={<VisitsPage />} />
          <Route path="/visits/:id" element={<VisitDetailPage />} />
          <Route element={<RequireAuth roles={REPORT_ROLES} />}>
            <Route path="/reports/daily" element={<DailyReportPage />} />
          </Route>
        </Route>
      </Route>
      <Route path="*" element={<Navigate to="/visits" replace />} />
    </Routes>
  );
}
