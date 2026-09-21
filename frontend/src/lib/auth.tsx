import { createContext, useCallback, useContext, useMemo, useState, type ReactNode } from 'react';
import type { User } from '../types';
import { clearSession, getStoredUser, storeSession } from './session';

interface AuthState {
  user: User | null;
  signIn: (token: string, user: User) => void;
  signOut: () => void;
}

const AuthContext = createContext<AuthState | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(getStoredUser);

  const signIn = useCallback((token: string, nextUser: User) => {
    storeSession(token, nextUser);
    setUser(nextUser);
  }, []);

  const signOut = useCallback(() => {
    clearSession();
    setUser(null);
  }, []);

  const value = useMemo(() => ({ user, signIn, signOut }), [user, signIn, signOut]);
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthState {
  const state = useContext(AuthContext);
  if (!state) throw new Error('useAuth must be used inside <AuthProvider>');
  return state;
}

/** For pages behind <RequireAuth>, where a missing user is a routing bug rather than a state. */
export function useCurrentUser(): User {
  const { user } = useAuth();
  if (!user) throw new Error('useCurrentUser called outside an authenticated route');
  return user;
}
