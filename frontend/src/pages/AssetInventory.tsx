import { useEffect, useMemo, useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Plus, Search, ChevronLeft, ChevronRight, Tag } from 'lucide-react';
import { api, apiError } from '@/lib/api';
import { useOrg } from '@/context/OrgContext';
import { PageHeader } from '@/components/Layout';
import { Badge, Button, Card, EmptyState, ErrorState, Field, Input, LoadingState, Select } from '@/components/ui';
import { Modal } from '@/components/Modal';
import { relativeTime } from '@/lib/utils';
import type { AssetType } from '@/lib/types';

const ASSET_TYPES: AssetType[] = ['domain', 'subdomain', 'ip', 'cidr', 'asn', 'dns_record', 'certificate', 'technology'];
const PAGE_SIZE = 20;

export default function AssetInventory() {
  const { activeOrgId, loading: orgLoading } = useOrg();
  const qc = useQueryClient();
  const [params, setParams] = useSearchParams();

  const projects = useQuery({
    queryKey: ['projects', activeOrgId],
    queryFn: () => api.projects.list(activeOrgId),
    enabled: !!activeOrgId,
  });

  const [projectId, setProjectId] = useState(params.get('project') || '');
  const [typeFilter, setTypeFilter] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [search, setSearch] = useState('');
  const [debounced, setDebounced] = useState('');
  const [page, setPage] = useState(0);

  useEffect(() => {
    const t = setTimeout(() => setDebounced(search), 350);
    return () => clearTimeout(t);
  }, [search]);

  useEffect(() => {
    if (!projectId && projects.data && projects.data.length > 0) {
      setProjectId(projects.data[0].id);
    }
  }, [projects.data, projectId]);

  useEffect(() => setPage(0), [projectId, typeFilter, statusFilter, debounced]);

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['assets', activeOrgId, projectId, typeFilter, statusFilter, debounced, page],
    queryFn: () =>
      api.assets.list(activeOrgId, projectId, {
        type: typeFilter || undefined,
        status: statusFilter || undefined,
        q: debounced || undefined,
        limit: PAGE_SIZE,
        offset: page * PAGE_SIZE,
      }),
    enabled: !!activeOrgId && !!projectId,
  });

  // Add asset modal
  const [show, setShow] = useState(false);
  const [type, setType] = useState<AssetType>('domain');
  const [value, setValue] = useState('');
  const [tags, setTags] = useState('');
  const [err, setErr] = useState('');

  const create = useMutation({
    mutationFn: () =>
      api.assets.create(activeOrgId, projectId, {
        type,
        value,
        tags: tags.split(',').map((t) => t.trim()).filter(Boolean),
      }),
    onSuccess: () => {
      setShow(false);
      setValue('');
      setTags('');
      qc.invalidateQueries({ queryKey: ['assets', activeOrgId, projectId] });
    },
    onError: (e) => setErr(apiError(e)),
  });

  const totalPages = useMemo(() => (data ? Math.max(1, Math.ceil(data.total / PAGE_SIZE)) : 1), [data]);

  if (orgLoading || projects.isLoading) return <LoadingState />;
  if ((projects.data || []).length === 0) {
    return (
      <EmptyState
        title="No projects to inventory"
        description="Create a project before adding assets."
        action={<Link to="/projects"><Button>Go to Projects</Button></Link>}
      />
    );
  }

  return (
    <div>
      <PageHeader
        title="Asset Inventory"
        subtitle="Search, filter and manage your tracked attack surface."
        action={
          <Button icon={<Plus className="h-4 w-4" />} onClick={() => { setErr(''); setShow(true); }} disabled={!projectId} data-testid="add-asset-button">
            Add Asset
          </Button>
        }
      />

      <Card className="mb-4 p-4">
        <div className="grid grid-cols-1 gap-3 md:grid-cols-4">
          <Field label="Project">
            <Select value={projectId} onChange={(e) => setProjectId(e.target.value)} data-testid="asset-project-select">
              {projects.data!.map((p) => <option key={p.id} value={p.id}>{p.name}</option>)}
            </Select>
          </Field>
          <Field label="Type">
            <Select value={typeFilter} onChange={(e) => setTypeFilter(e.target.value)} data-testid="asset-type-filter">
              <option value="">All types</option>
              {ASSET_TYPES.map((t) => <option key={t} value={t}>{t}</option>)}
            </Select>
          </Field>
          <Field label="Status">
            <Select value={statusFilter} onChange={(e) => setStatusFilter(e.target.value)} data-testid="asset-status-filter">
              <option value="">All</option>
              <option value="active">active</option>
              <option value="archived">archived</option>
            </Select>
          </Field>
          <Field label="Search">
            <div className="relative">
              <Search className="pointer-events-none absolute left-2.5 top-2.5 h-4 w-4 text-zinc-500" />
              <Input value={search} onChange={(e) => setSearch(e.target.value)} placeholder="value contains…" className="pl-8" data-testid="asset-search" />
            </div>
          </Field>
        </div>
      </Card>

      {isError ? (
        <ErrorState message="Failed to load assets." onRetry={refetch} />
      ) : isLoading ? (
        <LoadingState />
      ) : !data || data.assets.length === 0 ? (
        <EmptyState title="No assets found" description="Adjust filters or add your first asset." action={<Button onClick={() => setShow(true)} data-testid="empty-add-asset">Add asset</Button>} />
      ) : (
        <Card className="overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-ink-700 text-xs uppercase tracking-wider text-zinc-500">
                  <th className="px-4 py-3 text-left font-semibold">Type</th>
                  <th className="px-4 py-3 text-left font-semibold">Value</th>
                  <th className="px-4 py-3 text-left font-semibold">Tags</th>
                  <th className="px-4 py-3 text-left font-semibold">Status</th>
                  <th className="px-4 py-3 text-left font-semibold">Last Seen</th>
                </tr>
              </thead>
              <tbody data-testid="assets-table">
                {data.assets.map((a) => (
                  <tr key={a.id} className="border-b border-ink-800/50 transition-colors hover:bg-ink-800/40">
                    <td className="px-4 py-3"><Badge tone="brand">{a.type}</Badge></td>
                    <td className="px-4 py-3">
                      <Link to={`/assets/${a.project_id}/${a.id}`} className="font-mono text-zinc-200 hover:text-blue-400" data-testid={`asset-link-${a.id}`}>
                        {a.value}
                      </Link>
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex flex-wrap gap-1">
                        {(a.tags || []).length === 0 ? <span className="text-zinc-600">—</span> : (a.tags || []).map((t) => (
                          <Badge key={t} tone="neutral"><Tag className="h-3 w-3" />{t}</Badge>
                        ))}
                      </div>
                    </td>
                    <td className="px-4 py-3"><Badge tone={a.status === 'active' ? 'healthy' : 'neutral'}>{a.status}</Badge></td>
                    <td className="px-4 py-3 text-zinc-400">{relativeTime(a.last_seen)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <div className="flex items-center justify-between border-t border-ink-700 px-4 py-3 text-sm text-zinc-400">
            <span data-testid="asset-total">{data.total} asset{data.total === 1 ? '' : 's'}</span>
            <div className="flex items-center gap-2">
              <Button variant="secondary" disabled={page === 0} onClick={() => setPage((p) => p - 1)} data-testid="prev-page" icon={<ChevronLeft className="h-4 w-4" />}>Prev</Button>
              <span className="px-2">Page {page + 1} / {totalPages}</span>
              <Button variant="secondary" disabled={page + 1 >= totalPages} onClick={() => setPage((p) => p + 1)} data-testid="next-page">Next<ChevronRight className="h-4 w-4" /></Button>
            </div>
          </div>
        </Card>
      )}

      <Modal
        open={show}
        onClose={() => setShow(false)}
        title="Add Asset"
        footer={
          <>
            <Button variant="secondary" onClick={() => setShow(false)}>Cancel</Button>
            <Button onClick={() => create.mutate()} loading={create.isPending} disabled={!value.trim()} data-testid="submit-add-asset">Add</Button>
          </>
        }
      >
        <div className="space-y-4">
          <Field label="Type">
            <Select value={type} onChange={(e) => setType(e.target.value as AssetType)} data-testid="new-asset-type">
              {ASSET_TYPES.map((t) => <option key={t} value={t}>{t}</option>)}
            </Select>
          </Field>
          <Field label="Value">
            <Input value={value} onChange={(e) => setValue(e.target.value)} placeholder="example.com" data-testid="new-asset-value" autoFocus />
          </Field>
          <Field label="Tags" hint="Comma-separated, e.g. prod, external">
            <Input value={tags} onChange={(e) => setTags(e.target.value)} placeholder="prod, external" data-testid="new-asset-tags" />
          </Field>
          {err && <p className="text-sm text-red-400">{err}</p>}
        </div>
      </Modal>
    </div>
  );
}
