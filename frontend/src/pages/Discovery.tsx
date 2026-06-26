import { useEffect, useState } from 'react';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { ScanSearch, Radar } from 'lucide-react';
import { api, apiError } from '@/lib/api';
import { useOrg } from '@/context/OrgContext';
import { PageHeader } from '@/components/Layout';
import { Badge, Button, Card, EmptyState, ErrorState, Field, Input, LoadingState, Select } from '@/components/ui';
import { Modal } from '@/components/Modal';
import { DiscoveryStatusBadge } from '@/components/DiscoveryStatusBadge';
import { relativeTime } from '@/lib/utils';
import type { DiscoveryInputType } from '@/lib/types';

const INPUT_TYPES: { value: DiscoveryInputType; hint: string }[] = [
  { value: 'domain', hint: 'example.com' },
  { value: 'subdomain', hint: 'app.example.com' },
  { value: 'cidr', hint: '192.0.2.0/24' },
  { value: 'asn', hint: 'AS15169' },
];

export default function Discovery() {
  const { activeOrgId, loading: orgLoading } = useOrg();
  const qc = useQueryClient();
  const navigate = useNavigate();
  const [params] = useSearchParams();

  const projects = useQuery({
    queryKey: ['projects', activeOrgId],
    queryFn: () => api.projects.list(activeOrgId),
    enabled: !!activeOrgId,
  });

  const [projectId, setProjectId] = useState(params.get('project') || '');

  useEffect(() => {
    if (!projectId && projects.data && projects.data.length > 0) {
      setProjectId(projects.data[0].id);
    }
  }, [projects.data, projectId]);

  const jobs = useQuery({
    queryKey: ['discovery', activeOrgId, projectId],
    queryFn: () => api.discovery.list(activeOrgId, projectId, { limit: 50 }),
    enabled: !!activeOrgId && !!projectId,
    // Poll while any job is still in flight so the table updates live.
    refetchInterval: (query) => {
      const data = query.state.data;
      const active = data?.jobs.some((j) => j.status === 'pending' || j.status === 'running');
      return active ? 2500 : false;
    },
  });

  // Start discovery modal
  const [show, setShow] = useState(false);
  const [inputType, setInputType] = useState<DiscoveryInputType>('domain');
  const [inputValue, setInputValue] = useState('');
  const [err, setErr] = useState('');

  const start = useMutation({
    mutationFn: () => api.discovery.start(activeOrgId, projectId, { input_type: inputType, input_value: inputValue.trim() }),
    onSuccess: (job) => {
      setShow(false);
      setInputValue('');
      qc.invalidateQueries({ queryKey: ['discovery', activeOrgId, projectId] });
      navigate(`/discovery/${projectId}/${job.id}`);
    },
    onError: (e) => setErr(apiError(e)),
  });

  if (orgLoading || projects.isLoading) return <LoadingState />;
  if ((projects.data || []).length === 0) {
    return (
      <EmptyState
        title="No projects to scan"
        description="Create a project before running passive discovery."
        action={<Link to="/projects"><Button>Go to Projects</Button></Link>}
      />
    );
  }

  return (
    <div>
      <PageHeader
        title="Passive Discovery"
        subtitle="Authorized, defensive enumeration via public DNS & Certificate Transparency. No intrusive scanning."
        action={
          <Button icon={<ScanSearch className="h-4 w-4" />} onClick={() => { setErr(''); setShow(true); }} disabled={!projectId} data-testid="start-discovery-button">
            Start Discovery
          </Button>
        }
      />

      <Card className="mb-4 p-4">
        <div className="max-w-sm">
          <Field label="Project">
            <Select value={projectId} onChange={(e) => setProjectId(e.target.value)} data-testid="discovery-project-select">
              {projects.data!.map((p) => <option key={p.id} value={p.id}>{p.name}</option>)}
            </Select>
          </Field>
        </div>
      </Card>

      {jobs.isError ? (
        <ErrorState message="Failed to load discovery history." onRetry={jobs.refetch} />
      ) : jobs.isLoading ? (
        <LoadingState />
      ) : !jobs.data || jobs.data.jobs.length === 0 ? (
        <EmptyState
          title="No discovery runs yet"
          description="Start a passive discovery job to enumerate subdomains, DNS records and certificates."
          action={<Button onClick={() => { setErr(''); setShow(true); }} data-testid="empty-start-discovery">Start Discovery</Button>}
        />
      ) : (
        <Card className="overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-ink-700 text-xs uppercase tracking-wider text-zinc-500">
                  <th className="px-4 py-3 text-left font-semibold">Seed</th>
                  <th className="px-4 py-3 text-left font-semibold">Status</th>
                  <th className="px-4 py-3 text-left font-semibold">Found</th>
                  <th className="px-4 py-3 text-left font-semibold">New</th>
                  <th className="px-4 py-3 text-left font-semibold">Started</th>
                </tr>
              </thead>
              <tbody data-testid="discovery-jobs-table">
                {jobs.data.jobs.map((j) => (
                  <tr
                    key={j.id}
                    className="cursor-pointer border-b border-ink-800/50 transition-colors hover:bg-ink-800/40"
                    onClick={() => navigate(`/discovery/${j.project_id}/${j.id}`)}
                    data-testid={`discovery-job-row-${j.id}`}
                  >
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2">
                        <Radar className="h-4 w-4 text-zinc-500" />
                        <Badge tone="brand">{j.input_type}</Badge>
                        <span className="font-mono text-zinc-200">{j.input_value}</span>
                      </div>
                    </td>
                    <td className="px-4 py-3"><DiscoveryStatusBadge status={j.status} /></td>
                    <td className="px-4 py-3 text-zinc-300">{j.assets_found}</td>
                    <td className="px-4 py-3 text-emerald-400">{j.assets_created}</td>
                    <td className="px-4 py-3 text-zinc-400">{relativeTime(j.started_at || j.created_at)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Card>
      )}

      <Modal
        open={show}
        onClose={() => setShow(false)}
        title="Start Passive Discovery"
        footer={
          <>
            <Button variant="secondary" onClick={() => setShow(false)}>Cancel</Button>
            <Button onClick={() => start.mutate()} loading={start.isPending} disabled={!inputValue.trim()} data-testid="submit-start-discovery">Run</Button>
          </>
        }
      >
        <div className="space-y-4">
          <p className="rounded-md border border-amber-500/20 bg-amber-500/5 px-3 py-2 text-xs text-amber-300">
            Only run discovery against assets you own or are explicitly authorized to assess.
          </p>
          <Field label="Seed Type">
            <Select value={inputType} onChange={(e) => setInputType(e.target.value as DiscoveryInputType)} data-testid="discovery-input-type">
              {INPUT_TYPES.map((t) => <option key={t.value} value={t.value}>{t.value}</option>)}
            </Select>
          </Field>
          <Field label="Seed Value" hint={`e.g. ${INPUT_TYPES.find((t) => t.value === inputType)?.hint}`}>
            <Input value={inputValue} onChange={(e) => setInputValue(e.target.value)} placeholder={INPUT_TYPES.find((t) => t.value === inputType)?.hint} data-testid="discovery-input-value" autoFocus />
          </Field>
          {err && <p className="text-sm text-red-400" data-testid="discovery-error">{err}</p>}
        </div>
      </Modal>
    </div>
  );
}
