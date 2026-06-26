import { type ReactNode } from 'react';
import { Loader2, CheckCircle2, XCircle, Clock } from 'lucide-react';
import { Badge } from './ui';
import type { DiscoveryStatus } from '@/lib/types';

const config: Record<DiscoveryStatus, { tone: 'brand' | 'healthy' | 'critical' | 'warning' | 'neutral'; icon: ReactNode; label: string }> = {
  pending: { tone: 'neutral', icon: <Clock className="h-3 w-3" />, label: 'Pending' },
  running: { tone: 'brand', icon: <Loader2 className="h-3 w-3 animate-spin" />, label: 'Running' },
  completed: { tone: 'healthy', icon: <CheckCircle2 className="h-3 w-3" />, label: 'Completed' },
  failed: { tone: 'critical', icon: <XCircle className="h-3 w-3" />, label: 'Failed' },
};

export function DiscoveryStatusBadge({ status }: { status: DiscoveryStatus }) {
  const c = config[status] ?? config.pending;
  return (
    <Badge tone={c.tone} className="capitalize" >
      <span data-testid={`discovery-status-${status}`} className="inline-flex items-center gap-1.5">
        {c.icon}
        {c.label}
      </span>
    </Badge>
  );
}
