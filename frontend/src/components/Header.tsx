import { Link, NavLink, useNavigate } from 'react-router-dom';
import { useCurrentUser, useAuth } from '../lib/auth';
import { fullName } from '../lib/names';
import { canViewReports } from '../lib/permissions';

export function Header() {
  const user = useCurrentUser();
  const { signOut } = useAuth();
  const navigate = useNavigate();

  const logout = () => {
    signOut();
    navigate('/login', { replace: true });
  };

  return (
    <header className="topbar">
      <div className="topbar-inner">
        <Link to="/visits" className="brand">
          fieldops
        </Link>
        <nav className="topnav" aria-label="Main">
          <NavLink to="/visits">Visits</NavLink>
          {canViewReports(user) && <NavLink to="/reports/daily">Daily report</NavLink>}
        </nav>
        <div className="whoami">
          <span className="whoami-name">{fullName(user)}</span>
          <span className="whoami-role">{user.role}</span>
          <button type="button" className="btn btn-link" onClick={logout}>
            Log out
          </button>
        </div>
      </div>
    </header>
  );
}
