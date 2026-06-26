import { useNavigate, useParams } from 'react-router-dom';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { ArrowLeft, Trash2, Tag } from 'lucide-react';
import { api } from '@/lib/api';
import { useOrg } from '@/context/OrgContext';
import { PageHeader } from '@/components/Layout';
import { Badge, Button, Card, ErrorState, LoadingState } from '@/components/ui';
import { formatDate } from '@/lib/utils';

export default function AssetDetails() {
  const { projectId = '', assetId = '' } = useParams();
  const { activeOrgId } = useOrg();
  const qc = useQueryClient();
  const navigate = useNavigate();

  const { data, isLoading, isError, refetch } = useQuery({
    queryKey: ['asset', activeOrgId, projectId, assetId],
    queryFn: () => api.assets.get(activeOrgId, projectId, assetId),
    enabled: !!activeOrgId && !!projectId && !!assetId,
  });

  const remove = useMutation({
    mutationFn: () => api.assets.remove(activeOrgId, projectId, assetId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['assets', activeOrgId, projectId] });
      navigate('/assets');
    },
  });

  return (
    <div>
      <Button variant="ghost" icon={<ArrowLeft className="h-4 w-4" />} onClick={() => navigate(-1)} className="mb-4" data-testid="asset-back">
        Back
      </Button>

      {isError ? (
        <ErrorState message="Failed to load asset." onRetry={refetch} />
      ) : isLoading ? (
        <LoadingState />
      ) : data ? (
        <>
          <PageHeader
            title={data.value}
            subtitle={`Asset · ${data.type}`}
            action={
              <Button variant="danger" icon={<Trash2 className="h-4 w-4" />} loading={remove.isPending} onClick={() => remove.mutate()} data-testid="delete-asset">
                Delete
              </Button>
            }
          />

          <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
            <Card className="space-y-4 p-5 lg:col-span-2">
              <h3 className="font-heading text-lg font-semibold text-zinc-100">Overview</h3>
              <dl className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <Detail label="Type"><Badge tone="brand">{data.type}</Badge></Detail>
                <Detail label="Status"><Badge tone={data.status === 'active' ? 'healthy' : 'neutral'}>{data.status}</Badge></Detail>
                <Detail label="Value"><span className="font-mono text-zinc-200">{data.value}</span></Detail>
                <Detail label="First Seen">{formatDate(data.first_seen)}</Detail>
                <Detail label="Last Seen">{formatDate(data.last_seen)}</Detail>
                <Detail label="Created">{formatDate(data.created_at)}</Detail>
              </dl>
              <div>
                <p className="mb-2 text-xs font-semibold uppercase tracking-[0.15em] text-zinc-500">Tags</p>
                <div className="flex flex-wrap gap-1.5">
                  {(data.tags || []).length === 0 ? (
                    <span className="text-sm text-zinc-600">No tags</span>
                  ) : (
                    (data.tags || []).map((t) => <Badge key={t} tone="neutral"><Tag className="h-3 w-3" />{t}</Badge>)
                  )}
                </div>
              </div>
            </Card>

            <Card className="p-5">
              <h3 className="mb-3 font-heading text-lg font-semibold text-zinc-100">Attributes</h3>
              {Object.keys(data.attributes || {}).length === 0 ? (
                <p className="text-sm text-zinc-500">No additional attributes.</p>
              ) : (
                <pre className="overflow-x-auto rounded-md bg-ink-950 p-3 font-mono text-xs text-zinc-300" data-testid="asset-attributes">
                  {JSON.stringify(data.attributes, null, 2)}
                </pre>
              )}
            </Card>
          </div>
        </>
      ) : null}
    </div>
  );
}

function Detail({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div>
      <dt className="text-xs font-semibold uppercase tracking-[0.15em] text-zinc-500">{label}</dt>
      <dd className="mt-1 text-sm text-zinc-200">{children}</dd>
    </div>
  );
}
