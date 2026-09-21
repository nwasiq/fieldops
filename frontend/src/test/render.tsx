import { render } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { App } from '../App';
import { AuthProvider } from '../lib/auth';
import { storeSession } from '../lib/session';
import type { User } from '../types';

/** Mounts the real router at `route` with `user` already signed in. */
export function renderApp(route: string, user: User | null) {
  if (user) storeSession('test-token', user);
  return render(
    <MemoryRouter initialEntries={[route]}>
      <AuthProvider>
        <App />
      </AuthProvider>
    </MemoryRouter>,
  );
}
