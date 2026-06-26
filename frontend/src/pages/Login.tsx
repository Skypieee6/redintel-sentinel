import { useState, type ReactNode } from 'react';
import { Link, Navigate, useNavigate } from 'react-router-dom';
import { ShieldCheck, Lock } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import { apiError } from '@/lib/api';
import { Button, Field, Input } from '@/components/ui';

export function AuthLayout({ children, title, subtitle }: { children: ReactNode; title: string; subtitle: string }) {
  return (
    <div className="grid min-h-screen lg:grid-cols-2">
      <div
        className="relative hidden flex-col justify-between bg-ink-900 p-12 lg:flex"
        style={{
          backgroundImage:
            'linear-gradient(to bottom, rgba(9,10,11,0.7), rgba(9,10,11,0.95)), url(https://images.unsplash.com/photo-1526289034009-0240ddb68ce3?auto=format&fit=crop&w=1200&q=80)',
          backgroundSize: 'cover',
          backgroundPosition: 'center',
        }}
      >
        <div className="flex items-center gap-2.5">
          <div className="flex h-9 w-9 items-center justify-center rounded-md bg-brand/20 text-brand">
            <ShieldCheck className="h-5 w-5" />
          </div>
          <span className="font-heading text-lg font-bold tracking-tight">RedIntel Sentinel</span>
        </div>
        <div className="space-y-4">
          <h2 className="font-heading text-3xl font-bold leading-tight tracking-tight">
            Map, monitor and master your external attack surface.
          </h2>
          <p className="max-w-md text-sm text-zinc-400">
            Enterprise Attack Surface Management for authorized, defensive security assessments. Inventory assets,
            track changes, and report with confidence.
          </p>
        </div>
        <p className="text-xs text-zinc-600">For authorized security assessments only.</p>
      </div>

      <div className="flex items-center justify-center px-6 py-12">
        <div className="w-full max-w-sm animate-fade-in">
          <div className="mb-8 flex items-center gap-2 lg:hidden">
            <ShieldCheck className="h-6 w-6 text-brand" />
            <span className="font-heading text-lg font-bold">RedIntel Sentinel</span>
          </div>
          <h1 className="font-heading text-2xl font-bold tracking-tight text-zinc-100">{title}</h1>
          <p className="mb-6 mt-1 text-sm text-zinc-500">{subtitle}</p>
          {children}
        </div>
      </div>
    </div>
  );
}

export default function Login() {
  const { user, login } = useAuth();
  const navigate = useNavigate();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  if (user) return <Navigate to="/" replace />;

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      await login(email, password);
      navigate('/');
    } catch (err) {
      setError(apiError(err));
    } finally {
      setLoading(false);
    }
  };

  return (
    <AuthLayout title="Sign in" subtitle="Access your security operations console.">
      <form onSubmit={submit} className="space-y-4" data-testid="login-form">
        <Field label="Email">
          <Input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="you@company.com"
            required
            data-testid="login-email"
          />
        </Field>
        <Field label="Password">
          <Input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="••••••••"
            required
            data-testid="login-password"
          />
        </Field>
        {error && (
          <p className="rounded-md border border-red-500/30 bg-red-500/10 px-3 py-2 text-sm text-red-400" data-testid="login-error">
            {error}
          </p>
        )}
        <Button type="submit" loading={loading} className="w-full" icon={<Lock className="h-4 w-4" />} data-testid="login-submit">
          Sign in
        </Button>
      </form>
      <p className="mt-6 text-center text-sm text-zinc-500">
        No account?{' '}
        <Link to="/register" className="font-medium text-blue-400 hover:text-blue-300" data-testid="go-register">
          Create one
        </Link>
      </p>
    </AuthLayout>
  );
}
