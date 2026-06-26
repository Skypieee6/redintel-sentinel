import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { KeyRound, Plus, Copy, Building2, Check } from 'lucide-react';
import { api, apiError } from '@/lib/api';
import { useOrg } from '@/context/OrgContext';
import { PageHeader } from '@/components/Layout';
import { Badge, Button, Card, EmptyState, Field, Input, LoadingState } from '@/components/ui';
import { Modal } from '@/components/Modal';
import { formatDate } from '@/lib/utils';

export default function Settings() {
  const { activeOrg, activeOrgId, refetch } = useOrg();
  const qc = useQueryClient();

  const [orgName, setOrgName] = useState(activeOrg?.name || '');
  const [orgMsg, setOrgMsg] = useState('');

  const renameOrg = useMutation({
    mutationFn: () => api.orgs.update(activeOrgId, orgName),
    onSuccess: () => {
      setOrgMsg('Organization updated.');
      refetch();
      setTimeout(() => setOrgMsg(''), 2500);
    },
  });

  const keys = useQuery({ queryKey: ['apikeys'], queryFn: api.auth.listApiKeys });

  const [showKey, setShowKey] = useState(false);
  const [keyName, setKeyName] = useState('');
  const [newSecret, setNewSecret] = useState('');
  const [copied, setCopied] = useState(false);
  const [err, setErr] = useState('');

  const createKey = useMutation({
    mutationFn: () => api.auth.createApiKey(keyName),
    onSuccess: (k) => {
      setNewSecret(k.secret || '');
      setKeyName('');
      qc.invalidateQueries({ queryKey: ['apikeys'] });
    },
    onError: (e) => setErr(apiError(e)),
  });

  const revokeKey = useMutation({
    mutationFn: (id: string) => api.auth.revokeApiKey(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['apikeys'] }),
  });

  return (
    <div>
      <PageHeader title="Settings" subtitle="Manage organization and API access." />

      <div className="space-y-6">
        <Card className="p-5">
          <div className="mb-4 flex items-center gap-2">
            <Building2 className="h-5 w-5 text-zinc-400" />
            <h3 className="font-heading text-lg font-semibold text-zinc-100">Organization</h3>
          </div>
          {!activeOrg ? (
            <EmptyState title="No organization selected" />
          ) : (
            <div className="flex max-w-md flex-col gap-3">
              <Field label="Organization name">
                <Input value={orgName} onChange={(e) => setOrgName(e.target.value)} data-testid="settings-org-name" />
              </Field>
              <p className="font-mono text-xs text-zinc-500">slug: {activeOrg.slug}</p>
              <div className="flex items-center gap-3">
                <Button onClick={() => renameOrg.mutate()} loading={renameOrg.isPending} disabled={!orgName.trim()} data-testid="save-org">Save changes</Button>
                {orgMsg && <span className="text-sm text-emerald-400">{orgMsg}</span>}
              </div>
            </div>
          )}
        </Card>

        <Card className="p-5">
          <div className="mb-4 flex items-center justify-between">
            <div className="flex items-center gap-2">
              <KeyRound className="h-5 w-5 text-zinc-400" />
              <h3 className="font-heading text-lg font-semibold text-zinc-100">API Keys</h3>
            </div>
            <Button variant="secondary" icon={<Plus className="h-4 w-4" />} onClick={() => { setErr(''); setNewSecret(''); setShowKey(true); }} data-testid="create-key-button">
              New Key
            </Button>
          </div>
          {keys.isLoading ? (
            <LoadingState />
          ) : (keys.data || []).length === 0 ? (
            <EmptyState title="No API keys" description="Create a key for programmatic access (X-API-Key header)." />
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-ink-700 text-xs uppercase tracking-wider text-zinc-500">
                    <th className="px-2 py-2 text-left font-semibold">Name</th>
                    <th className="px-2 py-2 text-left font-semibold">Prefix</th>
                    <th className="px-2 py-2 text-left font-semibold">Created</th>
                    <th className="px-2 py-2 text-left font-semibold">Status</th>
                    <th className="px-2 py-2 text-right font-semibold">Action</th>
                  </tr>
                </thead>
                <tbody data-testid="apikeys-table">
                  {keys.data!.map((k) => (
                    <tr key={k.id} className="border-b border-ink-800/50 hover:bg-ink-800/40">
                      <td className="px-2 py-2.5 text-zinc-200">{k.name}</td>
                      <td className="px-2 py-2.5 font-mono text-zinc-400">{k.prefix}…</td>
                      <td className="px-2 py-2.5 text-zinc-400">{formatDate(k.created_at)}</td>
                      <td className="px-2 py-2.5"><Badge tone={k.revoked ? 'critical' : 'healthy'}>{k.revoked ? 'revoked' : 'active'}</Badge></td>
                      <td className="px-2 py-2.5 text-right">
                        {!k.revoked && (
                          <Button variant="ghost" onClick={() => revokeKey.mutate(k.id)} data-testid={`revoke-key-${k.id}`}>Revoke</Button>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </Card>
      </div>

      <Modal
        open={showKey}
        onClose={() => setShowKey(false)}
        title="Create API Key"
        footer={
          newSecret ? (
            <Button onClick={() => setShowKey(false)}>Done</Button>
          ) : (
            <>
              <Button variant="secondary" onClick={() => setShowKey(false)}>Cancel</Button>
              <Button onClick={() => createKey.mutate()} loading={createKey.isPending} disabled={!keyName.trim()} data-testid="submit-create-key">Create</Button>
            </>
          )
        }
      >
        {newSecret ? (
          <div className="space-y-3">
            <p className="text-sm text-zinc-300">Copy this key now — it will not be shown again.</p>
            <div className="flex items-center gap-2 rounded-md border border-ink-700 bg-ink-950 p-3">
              <code className="flex-1 break-all font-mono text-xs text-emerald-400" data-testid="new-key-secret">{newSecret}</code>
              <button
                onClick={() => { navigator.clipboard.writeText(newSecret); setCopied(true); setTimeout(() => setCopied(false), 1500); }}
                className="text-zinc-400 hover:text-zinc-100"
                data-testid="copy-key"
              >
                {copied ? <Check className="h-4 w-4 text-emerald-400" /> : <Copy className="h-4 w-4" />}
              </button>
            </div>
          </div>
        ) : (
          <div className="space-y-4">
            <Field label="Key name">
              <Input value={keyName} onChange={(e) => setKeyName(e.target.value)} placeholder="CI pipeline" data-testid="key-name-input" autoFocus />
            </Field>
            {err && <p className="text-sm text-red-400">{err}</p>}
          </div>
        )}
      </Modal>
    </div>
  );
}
