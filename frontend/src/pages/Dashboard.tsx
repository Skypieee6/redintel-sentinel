import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, Cell } from 'recharts';
import { Boxes, FolderKanban, Users, Activity, ArrowRight } from 'lucide-react';
import { api } from '@/lib/api';
import { useOrg } from '@/context/OrgContext';
import { PageHeader } from '@/components/Layout';
import { Card, EmptyState, ErrorState, LoadingState, Badge, Button, Skeleton } from '@/components/ui';
import { relativeTime } from '@/lib/utils';
import type { AssetType } from '@/lib/types';

const TYPE_COLORS: Record<string, string> = {
  domain: '#3B82F6',
  subdomain: '#22D3EE',
  ip: '#10B981',
  cidr: '#A78BFA',
  asn: '#F59E0B',
  dns_record: '#EC4899',
  certificate: '#F43F5E',
  technology: '#84CC16',
};

const CHART_MARGIN = { top: 8, right: 8, left: -16, bottom: 8 };
const X_AXIS_LINE = { stroke: '#27272A' };
const TOOLTIP_CURSOR = { fill: 'rgba(255,255,255,0.04)' };
const TOOLTIP_CONTENT_STYLE = { background: '#18181B', border: '1px solid #27272A', borderRadius: 8, color: '#FAFAFA' };
const SKELETON_KEYS = ['stat-1', 'stat-2', 'stat-3', 'stat-4'];

function StatCard({ icon, label, value, accent }: { icon: React.ReactNode; label: string; value: React.ReactNode; accent: string }) {
  return (
    <Card className="p-5">
      <div className="flex items-start justify-between">
        <div>
          <p className="text-xs font-semibold uppercase tracking-[0.15em] text-zinc-500">{label}</p>
          <p className="mt-2 font-heading text-3xl font-bold tracking-tight text-zinc-100">{value}</p>
        </div>
        <div className={`flex h-10 w-10 items-center justify-center rounded-md ${accent}`}>{icon}</div>
      </div>
    </Card>
  );
}

export default function Dashboard() {
  const { activeOrg, activeOrgId, loading: orgLoading } = useOrg();
  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['dashboard', activeOrgId],
    queryFn: () => api.dashboard.get(activeOrgId),
    enabled: !!activeOrgId,
  });

  if (orgLoading) return <LoadingState />;
  if (!activeOrg) {
    return (
      <EmptyState
        title="No organization yet"
        description="Create your first organization to start tracking assets."
        action={
          <Link to="/organizations">
            <Button data-testid="dashboard-create-org">Go to Organizations</Button>
          </Link>
        }
      />
    );
  }

  const chartData = (data?.assets_by_type || []).map((a) => ({ name: a.type, count: a.count }));

  return (
    <div>
      <PageHeader title="Dashboard" subtitle={`Attack surface overview for ${activeOrg.name}`} />

      {isError ? (
        <ErrorState message="Failed to load dashboard metrics." onRetry={refetch} />
      ) : isLoading ? (
        <div className="space-y-6">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
            {SKELETON_KEYS.map((k) => (
              <Skeleton key={k} className="h-28" />
            ))}
          </div>
          <Skeleton className="h-80" />
        </div>
      ) : (
        <div className="space-y-6">
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <StatCard label="Total Assets" value={data!.total_assets} icon={<Boxes className="h-5 w-5 text-blue-400" />} accent="bg-brand/10" />
            <StatCard
              label="Active Projects"
              value={data!.project_statistics.active}
              icon={<FolderKanban className="h-5 w-5 text-emerald-400" />}
              accent="bg-emerald-500/10"
            />
            <StatCard
              label="Team Members"
              value={data!.team_statistics.members}
              icon={<Users className="h-5 w-5 text-violet-400" />}
              accent="bg-violet-500/10"
            />
            <StatCard
              label="Pending Invites"
              value={data!.team_statistics.pending_invites}
              icon={<Activity className="h-5 w-5 text-amber-400" />}
              accent="bg-amber-500/10"
            />
          </div>

          <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
            {/* Chart */}
            <Card className="p-5 lg:col-span-2">
              <h3 className="mb-4 font-heading text-lg font-semibold text-zinc-100">Assets by Type</h3>
              {chartData.length === 0 ? (
                <EmptyState title="No assets yet" description="Add assets to a project to see the breakdown." />
              ) : (
                <div className="min-h-[300px]" data-testid="assets-by-type-chart">
                  <ResponsiveContainer width="100%" height={300}>
                    <BarChart data={chartData} margin={CHART_MARGIN}>
                      <XAxis dataKey="name" stroke="#71717A" fontSize={12} tickLine={false} axisLine={X_AXIS_LINE} />
                      <YAxis stroke="#71717A" fontSize={12} tickLine={false} axisLine={false} allowDecimals={false} />
                      <Tooltip
                        cursor={TOOLTIP_CURSOR}
                        contentStyle={TOOLTIP_CONTENT_STYLE}
                      />
                      <Bar dataKey="count" radius={[4, 4, 0, 0]}>
                        {chartData.map((d) => (
                          <Cell key={d.name} fill={TYPE_COLORS[d.name] || '#3B82F6'} />
                        ))}
                      </Bar>
                    </BarChart>
                  </ResponsiveContainer>
                </div>
              )}
            </Card>

            {/* Recent changes */}
            <Card className="p-5">
              <h3 className="mb-4 font-heading text-lg font-semibold text-zinc-100">Recent Changes</h3>
              {data!.recent_changes.length === 0 ? (
                <p className="py-8 text-center text-sm text-zinc-500">No recent activity.</p>
              ) : (
                <ul className="space-y-3" data-testid="recent-changes">
                  {data!.recent_changes.slice(0, 8).map((a) => (
                    <li key={a.id} className="flex items-center justify-between gap-2 border-b border-ink-800 pb-2 last:border-0">
                      <div className="min-w-0">
                        <p className="truncate font-mono text-sm text-zinc-200">{a.value}</p>
                        <Badge tone="brand" className="mt-1">{a.type}</Badge>
                      </div>
                      <span className="shrink-0 text-xs text-zinc-500">{relativeTime(a.updated_at)}</span>
                    </li>
                  ))}
                </ul>
              )}
            </Card>
          </div>

          {/* Project statistics */}
          <Card className="p-5">
            <div className="mb-4 flex items-center justify-between">
              <h3 className="font-heading text-lg font-semibold text-zinc-100">Project Statistics</h3>
              <Link to="/projects" className="inline-flex items-center gap-1 text-sm text-blue-400 hover:text-blue-300">
                View all <ArrowRight className="h-3.5 w-3.5" />
              </Link>
            </div>
            {(data!.project_statistics.by_project || []).length === 0 ? (
              <p className="py-6 text-center text-sm text-zinc-500">No projects yet.</p>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b border-ink-700 text-xs uppercase tracking-wider text-zinc-500">
                      <th className="px-2 py-2 text-left font-semibold">Project</th>
                      <th className="px-2 py-2 text-left font-semibold">Status</th>
                      <th className="px-2 py-2 text-right font-semibold">Assets</th>
                    </tr>
                  </thead>
                  <tbody>
                    {(data!.project_statistics.by_project || []).map((p) => (
                      <tr key={p.project_id} className="border-b border-ink-800/50 transition-colors hover:bg-ink-800/40">
                        <td className="px-2 py-2.5 text-zinc-200">{p.name}</td>
                        <td className="px-2 py-2.5">
                          <Badge tone={p.status === 'active' ? 'healthy' : 'neutral'}>{p.status}</Badge>
                        </td>
                        <td className="px-2 py-2.5 text-right font-mono text-zinc-300">{p.assets}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </Card>
        </div>
      )}
    </div>
  );
}

export type { AssetType };
