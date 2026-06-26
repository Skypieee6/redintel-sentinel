import { createContext, useContext, useEffect, useState, type ReactNode } from 'react';
import { api, tokenStore } from '@/lib/api';
import type { User } from '@/lib/types';

interface AuthState {
  user: User | null;
  loading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string, fullName: string) => Promise<void>;
  logout: () => Promise<void>;
  setUser: (u: User) => void;
}

const AuthContext = createContext<AuthState | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let active = true;
    async function bootstrap() {
      if (!tokenStore.access && !tokenStore.refresh) {
        setLoading(false);
        return;
      }
      try {
        const me = await api.auth.me();
        if (active) setUser(me);
      } catch {
        tokenStore.clear();
      } finally {
        if (active) setLoading(false);
      }
    }
    bootstrap();
    return () => {
      active = false;
    };
  }, []);

  const login = async (email: string, password: string) => {
    const { user: u, tokens } = await api.auth.login({ email, password });
    tokenStore.set(tokens);
    setUser(u);
  };

  const register = async (email: string, password: string, fullName: string) => {
    const { user: u, tokens } = await api.auth.register({ email, password, full_name: fullName });
    tokenStore.set(tokens);
    setUser(u);
  };

  const logout = async () => {
    try {
      await api.auth.logout();
    } catch {
      /* ignore */
    }
    tokenStore.clear();
    setUser(null);
  };

  return (
    <AuthContext.Provider value={{ user, loading, login, register, logout, setUser }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
