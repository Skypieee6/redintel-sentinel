import { useMemo } from 'react';
import { Link, useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { ArrowLeft, Globe, FileText, ShieldCheck, AlertTriangle } from 'lucide-react';
import { api } from '@/lib/api';
import { useOrg } from '@/context/OrgContext';
import { PageHeader } from '@/components/Layout';
import { Badge, Card, EmptyState, ErrorState, LoadingState } from '@/components/ui';
import { DiscoveryStatusBadge } from '@/components/DiscoveryStatusBadge';
import { formatDate } from '@/lib/utils';
import type { AssetType, DiscoveryResult } from '@/lib/types';

const TYPE_META: Partial<Record<AssetType, { label: string; icon: typeof Globe }>> = {
  subdomain: { label: 'Subdomains', icon: Globe },
  dns_record: { label: 'DNS Records', icon: FileText },
  certificate: { label: 'Certificates', icon: ShieldCheck },
};

function Stat({ label, value, accent }: { label: string; value: string | number; accent?: string }) {
  return (
    <div className="rounded-md border border-ink-700 bg-ink-900 p-4">
      <p className="text-xs font-semibold uppercase tracking-[0.15em] text-zinc-500">{label}</p>
      <p className={`mt-1 font-heading text-2xl font-bold ${accent || 'text-zinc-100'}`}>{value}</p>
    </div>
  );
}

export default function DiscoveryDetail() {
  const { projectId = '', jobId = '' } = useParams();
  const { activeOrgId } = useOrg();

  const { data: job, isLoading, isError, refetch } = useQuery({
    queryKey: ['discovery-job', activeOrgId, projectId, jobId],
    queryFn: () => api.discovery.get(activeOrgId, projectId, jobId),
    enabled: !!activeOrgId && !!projectId && !!jobId,
    refetchInterval: (query) => {
      const s = query.state.data?.status;
      return s === 'pending' || s === 'running' ? 2000 : false;
    },
  });

  const grouped = useMemo(() => {
    const map = new Map<AssetType, DiscoveryResult[]>();
    (job?.results || []).forEach((r) => {
      const arr = map.get(r.type) || [];
      arr.push(r);
      map.set(r.type, arr);
    });
    return map;
  }, [job]);

  if (isLoading) return <LoadingState />;
  if (isError || !job) return <ErrorState message="Failed to load discovery job." onRetry={refetch} />;

  return (
    <div>
      <Link to="/discovery" className="mb-4 inline-flex items-center gap-1.5 text-sm text-zinc-400 hover:text-zinc-200" data-testid="back-to-discovery">
        <ArrowLeft className="h-4 w-4" /> Back to Discovery
      </Link>

      <PageHeader
        title={job.input_value}
        subtitle={`${job.input_type} · started ${formatDate(job.started_at || job.created_at)}`}
        action={<DiscoveryStatusBadge status={job.status} />}
      />

      <div className="mb-6 grid grid-cols-2 gap-3 sm:grid-cols-4" data-testid="discovery-stats">
        <Stat label="Assets Found" value={job.assets_found} />
        <Stat label="Newly Added" value={job.assets_created} accent="text-emerald-400" />
        <Stat label="Completed" value={job.completed_at ? formatDate(job.completed_at) : '—'} />
        <Stat label="Status" value={job.status} />
      </div>

      {job.status === 'failed' && (
        <Card className="mb-6 border-red-500/30 bg-red-500/5 p-4" data-testid="discovery-failure">
          <div className="flex items-start gap-2 text-sm text-red-300">
            <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0" />
            <span>{job.error || 'Discovery failed.'}</span>
          </div>
        </Card>
      )}

      {(job.status === 'pending' || job.status === 'running') ? (
        <LoadingState label="Discovery in progress — results will appear as they are found…" />
      ) : (job.results || []).length === 0 ? (
        <EmptyState title="No assets discovered" description="The passive sources returned no findings for this seed." />
      ) : (
        <div className="space-y-6">
          {Array.from(grouped.entries()).map(([type, results]) => {
            const meta = TYPE_META[type] || { label: type, icon: FileText };
            const Icon = meta.icon;
            return (
              <Card key={type} className="overflow-hidden" >
                <div className="flex items-center gap-2 border-b border-ink-700 px-4 py-3">
                  <Icon className="h-4 w-4 text-zinc-400" />
                  <h3 className="font-heading text-sm font-semibold text-zinc-200">{meta.label}</h3>
                  <Badge tone="neutral">{results.length}</Badge>
                </div>
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b border-ink-700 text-xs uppercase tracking-wider text-zinc-500">
                        <th className="px-4 py-2.5 text-left font-semibold">Value</th>
                        <th className="px-4 py-2.5 text-left font-semibold">Source</th>
                        <th className="px-4 py-2.5 text-left font-semibold">State</th>
                      </tr>
                    </thead>
                    <tbody data-testid={`discovery-results-${type}`}>
                      {results.map((r) => (
                        <tr key={r.id} className="border-b border-ink-800/50">
                          <td className="px-4 py-2.5">
                            {r.asset_id ? (
                              <Link to={`/assets/${job.project_id}/${r.asset_id}`} className="font-mono text-zinc-200 hover:text-blue-400">
                                {r.value}
                              </Link>
                            ) : (
                              <span className="font-mono text-zinc-200">{r.value}</span>
                            )}
                          </td>
                          <td className="px-4 py-2.5"><Badge tone="neutral">{r.source}</Badge></td>
                          <td className="px-4 py-2.5">
                            {r.is_new ? <Badge tone="healthy">new</Badge> : <Badge tone="neutral">known</Badge>}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </Card>
            );
          })}
        </div>
      )}
    </div>
  );
}
