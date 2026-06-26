import { Navigate } from 'react-router-dom';
import { type ReactNode } from 'react';
import { useAuth } from '@/context/AuthContext';
import { LoadingState } from './ui';

export function ProtectedRoute({ children }: { children: ReactNode }) {
  const { user, loading } = useAuth();
  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-ink-950">
        <LoadingState label="Authenticating…" />
      </div>
    );
  }
  if (!user) return <Navigate to="/login" replace />;
  return <>{children}</>;
}
