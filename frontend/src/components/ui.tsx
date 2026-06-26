import { type ButtonHTMLAttributes, type InputHTMLAttributes, type ReactNode, type SelectHTMLAttributes, forwardRef } from 'react';
import { Loader2, AlertTriangle, Inbox } from 'lucide-react';
import { cn } from '@/lib/utils';

/* ---------------- Button ---------------- */
type Variant = 'primary' | 'secondary' | 'ghost' | 'danger';
interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant;
  loading?: boolean;
  icon?: ReactNode;
}
const variants: Record<Variant, string> = {
  primary: 'bg-brand hover:bg-brand-hover text-white focus:ring-brand/50',
  secondary: 'bg-ink-800 hover:bg-ink-700 text-zinc-100 border border-ink-700 focus:ring-ink-700',
  ghost: 'bg-transparent hover:bg-ink-800 text-zinc-300 focus:ring-ink-700',
  danger: 'bg-red-500/90 hover:bg-red-500 text-white focus:ring-red-500/40',
};
export function Button({ variant = 'primary', loading, icon, className, children, disabled, ...rest }: ButtonProps) {
  return (
    <button
      className={cn(
        'inline-flex items-center justify-center gap-2 rounded-md px-4 py-2 text-sm font-medium transition-colors duration-150 focus:outline-none focus:ring-2 disabled:opacity-50 disabled:cursor-not-allowed',
        variants[variant],
        className
      )}
      disabled={disabled || loading}
      {...rest}
    >
      {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : icon}
      {children}
    </button>
  );
}

/* ---------------- Card ---------------- */
export function Card({ className, children }: { className?: string; children: ReactNode }) {
  return (
    <div className={cn('rounded-md border border-ink-700 bg-ink-900 transition-colors hover:border-ink-700/80', className)}>
      {children}
    </div>
  );
}

/* ---------------- Input + Field ---------------- */
export const Input = forwardRef<HTMLInputElement, InputHTMLAttributes<HTMLInputElement>>(
  ({ className, ...rest }, ref) => (
    <input
      ref={ref}
      className={cn(
        'w-full rounded-md border border-ink-700 bg-ink-950 px-3 py-2 text-sm text-zinc-100 placeholder:text-zinc-500 transition-colors focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand/40',
        className
      )}
      {...rest}
    />
  )
);
Input.displayName = 'Input';

export const Select = forwardRef<HTMLSelectElement, SelectHTMLAttributes<HTMLSelectElement>>(
  ({ className, children, ...rest }, ref) => (
    <select
      ref={ref}
      className={cn(
        'w-full rounded-md border border-ink-700 bg-ink-950 px-3 py-2 text-sm text-zinc-100 transition-colors focus:border-brand focus:outline-none focus:ring-1 focus:ring-brand/40',
        className
      )}
      {...rest}
    >
      {children}
    </select>
  )
);
Select.displayName = 'Select';

export function Field({ label, children, hint }: { label: string; children: ReactNode; hint?: string }) {
  return (
    <label className="block space-y-1.5">
      <span className="text-xs font-semibold uppercase tracking-[0.15em] text-zinc-500">{label}</span>
      {children}
      {hint && <span className="block text-xs text-zinc-500">{hint}</span>}
    </label>
  );
}

/* ---------------- Badge ---------------- */
type Tone = 'brand' | 'healthy' | 'warning' | 'critical' | 'neutral';
const tones: Record<Tone, string> = {
  brand: 'bg-brand/10 text-blue-400 border-brand/20',
  healthy: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20',
  warning: 'bg-amber-500/10 text-amber-400 border-amber-500/20',
  critical: 'bg-red-500/10 text-red-400 border-red-500/20',
  neutral: 'bg-ink-800 text-zinc-400 border-ink-700',
};
export function Badge({ tone = 'neutral', children, className }: { tone?: Tone; children: ReactNode; className?: string }) {
  return (
    <span className={cn('inline-flex items-center gap-1.5 rounded border px-2 py-0.5 text-xs font-medium', tones[tone], className)}>
      {children}
    </span>
  );
}

/* ---------------- States ---------------- */
export function Spinner({ className }: { className?: string }) {
  return <Loader2 className={cn('h-5 w-5 animate-spin text-zinc-500', className)} />;
}

export function LoadingState({ label = 'Loading…' }: { label?: string }) {
  return (
    <div className="flex flex-col items-center justify-center gap-3 py-16 text-zinc-500" data-testid="loading-state">
      <Spinner className="h-6 w-6" />
      <p className="text-sm">{label}</p>
    </div>
  );
}

export function ErrorState({ message, onRetry }: { message: string; onRetry?: () => void }) {
  return (
    <div
      className="flex flex-col items-center justify-center gap-3 rounded-md border border-red-500/30 bg-red-500/5 py-12 text-center"
      data-testid="error-state"
    >
      <AlertTriangle className="h-7 w-7 text-red-400" />
      <p className="max-w-md text-sm text-zinc-300">{message}</p>
      {onRetry && (
        <Button variant="secondary" onClick={onRetry} data-testid="retry-button">
          Try again
        </Button>
      )}
    </div>
  );
}

export function EmptyState({
  title,
  description,
  action,
}: {
  title: string;
  description?: string;
  action?: ReactNode;
}) {
  return (
    <div className="flex flex-col items-center justify-center gap-3 py-16 text-center" data-testid="empty-state">
      <div className="rounded-full border border-ink-700 bg-ink-800 p-3">
        <Inbox className="h-6 w-6 text-zinc-500" />
      </div>
      <h3 className="font-heading text-lg font-semibold text-zinc-200">{title}</h3>
      {description && <p className="max-w-sm text-sm text-zinc-500">{description}</p>}
      {action}
    </div>
  );
}

/* ---------------- Skeleton ---------------- */
export function Skeleton({ className }: { className?: string }) {
  return <div className={cn('animate-pulse rounded-md bg-ink-800', className)} />;
}
