import { useEffect, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { FileJson, FileSpreadsheet, FileText, FileCode, Download } from 'lucide-react';
import { api, apiError } from '@/lib/api';
import { useOrg } from '@/context/OrgContext';
import { PageHeader } from '@/components/Layout';
import { Button, Card, EmptyState, Field, LoadingState, Select } from '@/components/ui';

const FORMATS = [
  { key: 'json', label: 'JSON', icon: FileJson, desc: 'Machine-readable export for pipelines.' },
  { key: 'csv', label: 'CSV', icon: FileSpreadsheet, desc: 'Spreadsheet-friendly tabular data.' },
  { key: 'markdown', label: 'Markdown', icon: FileText, desc: 'Readable report for wikis & tickets.' },
  { key: 'html', label: 'HTML', icon: FileCode, desc: 'Styled, shareable web report.' },
];

export default function Reports() {
  const { activeOrgId, loading: orgLoading } = useOrg();
  const [params] = useSearchParams();
  const [projectId, setProjectId] = useState(params.get('project') || '');
  const [busy, setBusy] = useState('');
  const [err, setErr] = useState('');

  const projects = useQuery({
    queryKey: ['projects', activeOrgId],
    queryFn: () => api.projects.list(activeOrgId),
    enabled: !!activeOrgId,
  });

  useEffect(() => {
    if (!projectId && projects.data && projects.data.length > 0) setProjectId(projects.data[0].id);
  }, [projects.data, projectId]);

  const download = async (format: string) => {
    setErr('');
    setBusy(format);
    try {
      await api.reports.download(activeOrgId, projectId, format);
    } catch (e) {
      setErr(apiError(e));
    } finally {
      setBusy('');
    }
  };

  if (orgLoading || projects.isLoading) return <LoadingState />;
  if ((projects.data || []).length === 0) {
    return <EmptyState title="No projects" description="Create a project and add assets to generate reports." />;
  }

  return (
    <div>
      <PageHeader title="Reports" subtitle="Export the asset inventory for a project." />

      <Card className="mb-6 max-w-md p-4">
        <Field label="Project">
          <Select value={projectId} onChange={(e) => setProjectId(e.target.value)} data-testid="report-project-select">
            {projects.data!.map((p) => <option key={p.id} value={p.id}>{p.name}</option>)}
          </Select>
        </Field>
      </Card>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {FORMATS.map((f) => (
          <Card key={f.key} className="flex flex-col p-5">
            <div className="mb-3 flex h-10 w-10 items-center justify-center rounded-md bg-brand/10 text-blue-400">
              <f.icon className="h-5 w-5" />
            </div>
            <h3 className="font-heading text-lg font-semibold text-zinc-100">{f.label}</h3>
            <p className="mt-1 flex-1 text-sm text-zinc-500">{f.desc}</p>
            <Button
              className="mt-4 w-full"
              variant="secondary"
              icon={<Download className="h-4 w-4" />}
              loading={busy === f.key}
              disabled={!projectId}
              onClick={() => download(f.key)}
              data-testid={`download-${f.key}`}
            >
              Download {f.label}
            </Button>
          </Card>
        ))}
      </div>
      {err && <p className="mt-4 text-sm text-red-400" data-testid="report-error">{err}</p>}
    </div>
  );
}
