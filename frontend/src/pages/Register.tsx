import { useState } from 'react';
import { Link, Navigate, useNavigate } from 'react-router-dom';
import { UserPlus } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import { apiError } from '@/lib/api';
import { Button, Field, Input } from '@/components/ui';
import { AuthLayout } from './Login';

export default function Register() {
  const { user, register } = useAuth();
  const navigate = useNavigate();
  const [fullName, setFullName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  if (user) return <Navigate to="/" replace />;

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    if (password.length < 8) {
      setError('Password must be at least 8 characters.');
      return;
    }
    setLoading(true);
    try {
      await register(email, password, fullName);
      navigate('/');
    } catch (err) {
      setError(apiError(err));
    } finally {
      setLoading(false);
    }
  };

  return (
    <AuthLayout title="Create account" subtitle="Start managing your attack surface in minutes.">
      <form onSubmit={submit} className="space-y-4" data-testid="register-form">
        <Field label="Full name">
          <Input value={fullName} onChange={(e) => setFullName(e.target.value)} placeholder="Jane Analyst" required data-testid="register-name" />
        </Field>
        <Field label="Email">
          <Input type="email" value={email} onChange={(e) => setEmail(e.target.value)} placeholder="you@company.com" required data-testid="register-email" />
        </Field>
        <Field label="Password" hint="Minimum 8 characters.">
          <Input type="password" value={password} onChange={(e) => setPassword(e.target.value)} placeholder="••••••••" required data-testid="register-password" />
        </Field>
        {error && (
          <p className="rounded-md border border-red-500/30 bg-red-500/10 px-3 py-2 text-sm text-red-400" data-testid="register-error">
            {error}
          </p>
        )}
        <Button type="submit" loading={loading} className="w-full" icon={<UserPlus className="h-4 w-4" />} data-testid="register-submit">
          Create account
        </Button>
      </form>
      <p className="mt-6 text-center text-sm text-zinc-500">
        Already registered?{' '}
        <Link to="/login" className="font-medium text-blue-400 hover:text-blue-300" data-testid="go-login">
          Sign in
        </Link>
      </p>
    </AuthLayout>
  );
}
