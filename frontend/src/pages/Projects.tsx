import { useState } from 'react';
import { Link } from 'react-router-dom';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Plus, Archive, ArchiveRestore, Radar, FileBarChart } from 'lucide-react';
import { api, apiError } from '@/lib/api';
import { useOrg } from '@/context/OrgContext';
import { PageHeader } from '@/components/Layout';
import { Badge, Button, Card, EmptyState, ErrorState, Field, Input, LoadingState } from '@/components/ui';
import { Modal } from '@/components/Modal';
import { formatDate } from '@/lib/utils';

export default function Projects() {
  const { activeOrg, activeOrgId, loading: orgLoading } = useOrg();
  const qc = useQueryClient();
  const [show, setShow] = useState(false);
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [err, setErr] = useState('');

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['projects', activeOrgId],
    queryFn: () => api.projects.list(activeOrgId),
    enabled: !!activeOrgId,
  });

  const create = useMutation({
    mutationFn: () => api.projects.create(activeOrgId, { name, description }),
    onSuccess: () => {
      setShow(false);
      setName('');
      setDescription('');
      qc.invalidateQueries({ queryKey: ['projects', activeOrgId] });
    },
    onError: (e) => setErr(apiError(e)),
  });

  const toggleArchive = useMutation({
    mutationFn: ({ id, archived }: { id: string; archived: boolean }) =>
      archived ? api.projects.unarchive(activeOrgId, id) : api.projects.archive(activeOrgId, id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['projects', activeOrgId] }),
  });

  if (orgLoading) return <LoadingState />;
  if (!activeOrg) return <EmptyState title="Select an organization" description="Create or pick an organization first." />;

  return (
    <div>
      <PageHeader
        title="Projects"
        subtitle={`Assessment projects in ${activeOrg.name}`}
        action={<Button icon={<Plus className="h-4 w-4" />} onClick={() => { setErr(''); setShow(true); }} data-testid="create-project-button">New Project</Button>}
      />

      {isError ? (
        <ErrorState message="Failed to load projects." onRetry={refetch} />
      ) : isLoading ? (
        <LoadingState />
      ) : (data || []).length === 0 ? (
        <EmptyState title="No projects" description="Create a project to start collecting assets." action={<Button onClick={() => setShow(true)} data-testid="empty-create-project">Create project</Button>} />
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {data!.map((p) => (
            <Card key={p.id} className="flex flex-col p-5" >
              <div className="flex items-start justify-between">
                <h3 className="font-heading text-lg font-semibold text-zinc-100">{p.name}</h3>
                <Badge tone={p.status === 'active' ? 'healthy' : 'neutral'}>{p.status}</Badge>
              </div>
              <p className="mt-1 line-clamp-2 min-h-[2.5rem] text-sm text-zinc-500">{p.description || 'No description.'}</p>
              <p className="mt-2 text-xs text-zinc-600">Created {formatDate(p.created_at)}</p>
              <div className="mt-4 flex flex-wrap gap-2 border-t border-ink-800 pt-4">
                <Link to={`/assets?project=${p.id}`}>
                  <Button variant="secondary" icon={<Radar className="h-4 w-4" />} data-testid={`project-assets-${p.id}`}>Assets</Button>
                </Link>
                <Link to={`/reports?project=${p.id}`}>
                  <Button variant="ghost" icon={<FileBarChart className="h-4 w-4" />}>Report</Button>
                </Link>
                <Button
                  variant="ghost"
                  icon={p.status === 'archived' ? <ArchiveRestore className="h-4 w-4" /> : <Archive className="h-4 w-4" />}
                  onClick={() => toggleArchive.mutate({ id: p.id, archived: p.status === 'archived' })}
                  data-testid={`project-archive-${p.id}`}
                >
                  {p.status === 'archived' ? 'Restore' : 'Archive'}
                </Button>
              </div>
            </Card>
          ))}
        </div>
      )}

      <Modal
        open={show}
        onClose={() => setShow(false)}
        title="New Project"
        footer={
          <>
            <Button variant="secondary" onClick={() => setShow(false)}>Cancel</Button>
            <Button onClick={() => create.mutate()} loading={create.isPending} disabled={!name.trim()} data-testid="submit-create-project">Create</Button>
          </>
        }
      >
        <div className="space-y-4">
          <Field label="Name">
            <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="External Perimeter" data-testid="project-name-input" autoFocus />
          </Field>
          <Field label="Description">
            <Input value={description} onChange={(e) => setDescription(e.target.value)} placeholder="Authorized scope notes" data-testid="project-desc-input" />
          </Field>
          {err && <p className="text-sm text-red-400">{err}</p>}
        </div>
      </Modal>
    </div>
  );
}
