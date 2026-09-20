import { createContext, createElement, useContext, useMemo, useState, type ReactNode } from 'react';
import type { User } from '../types';
import { UserRole } from '../constants/user';

interface AuthState {
  user: User | null;
  token: string;
  setAuth: (token: string, user: User) => void;
  updateUser: (user: User) => void;
  logout: () => void;
  isAdmin: () => boolean;
}

const AuthContext = createContext<AuthState | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [token, setToken] = useState<string>(() => localStorage.getItem('cyfreshfood_token') || '');
  const [user, setUser] = useState<User | null>(() => {
    const raw = localStorage.getItem('cyfreshfood_user');
    return raw ? (JSON.parse(raw) as User) : null;
  });

  const value = useMemo<AuthState>(
    () => ({
      user,
      token,
      setAuth: (t, u) => {
        localStorage.setItem('cyfreshfood_token', t);
        localStorage.setItem('cyfreshfood_user', JSON.stringify(u));
        setToken(t);
        setUser(u);
      },
      updateUser: (u) => {
        localStorage.setItem('cyfreshfood_user', JSON.stringify(u));
        setUser(u);
      },
      logout: () => {
        localStorage.removeItem('cyfreshfood_token');
        localStorage.removeItem('cyfreshfood_user');
        setToken('');
        setUser(null);
      },
      isAdmin: () => user?.role === UserRole.ADMIN,
    }),
    [user, token]
  );

  return createElement(AuthContext.Provider, { value }, children);
}

export function useAuth(): AuthState {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
