import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Building2, Plus, UserPlus, Mail } from 'lucide-react';
import { api, apiError } from '@/lib/api';
import { useOrg } from '@/context/OrgContext';
import { PageHeader } from '@/components/Layout';
import { Badge, Button, Card, EmptyState, ErrorState, Field, Input, LoadingState, Select } from '@/components/ui';
import { Modal } from '@/components/Modal';
import type { Role } from '@/lib/types';
import { formatDate } from '@/lib/utils';

const ROLES: Role[] = ['admin', 'manager', 'analyst', 'viewer'];

export default function Organizations() {
  const { orgs, activeOrg, activeOrgId, loading, refetch, setActiveOrgId } = useOrg();
  const qc = useQueryClient();
  const [showCreate, setShowCreate] = useState(false);
  const [name, setName] = useState('');
  const [showInvite, setShowInvite] = useState(false);
  const [inviteEmail, setInviteEmail] = useState('');
  const [inviteRole, setInviteRole] = useState<Role>('analyst');
  const [inviteToken, setInviteToken] = useState('');
  const [err, setErr] = useState('');

  const members = useQuery({
    queryKey: ['members', activeOrgId],
    queryFn: () => api.orgs.members(activeOrgId),
    enabled: !!activeOrgId,
  });

  const createOrg = useMutation({
    mutationFn: () => api.orgs.create(name),
    onSuccess: (org) => {
      setShowCreate(false);
      setName('');
      refetch();
      setActiveOrgId(org.id);
    },
    onError: (e) => setErr(apiError(e)),
  });

  const invite = useMutation({
    mutationFn: () => api.orgs.invite(activeOrgId, inviteEmail, inviteRole),
    onSuccess: (inv: { token?: string }) => {
      setInviteToken(inv.token || '');
      setInviteEmail('');
      qc.invalidateQueries({ queryKey: ['members', activeOrgId] });
    },
    onError: (e) => setErr(apiError(e)),
  });

  if (loading) return <LoadingState />;

  return (
    <div>
      <PageHeader
        title="Organizations"
        subtitle="Manage tenants, members and roles."
        action={
          <Button icon={<Plus className="h-4 w-4" />} onClick={() => { setErr(''); setShowCreate(true); }} data-testid="create-org-button">
            New Organization
          </Button>
        }
      />

      {orgs.length === 0 ? (
        <EmptyState
          title="No organizations"
          description="Create an organization to begin."
          action={<Button onClick={() => setShowCreate(true)} data-testid="empty-create-org">Create organization</Button>}
        />
      ) : (
        <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
          <div className="space-y-3 lg:col-span-1">
            {orgs.map((o) => (
              <Card
                key={o.id}
                className={`cursor-pointer p-4 ${o.id === activeOrgId ? 'border-brand/50 ring-1 ring-brand/30' : ''}`}
              >
                <button className="flex w-full items-center gap-3 text-left" onClick={() => setActiveOrgId(o.id)} data-testid={`org-card-${o.slug}`}>
                  <div className="flex h-9 w-9 items-center justify-center rounded-md bg-ink-800 text-zinc-300">
                    <Building2 className="h-4 w-4" />
                  </div>
                  <div className="min-w-0">
                    <p className="truncate font-medium text-zinc-100">{o.name}</p>
                    <p className="truncate font-mono text-xs text-zinc-500">{o.slug}</p>
                  </div>
                </button>
              </Card>
            ))}
          </div>

          <Card className="p-5 lg:col-span-2">
            <div className="mb-4 flex items-center justify-between">
              <div>
                <h3 className="font-heading text-lg font-semibold text-zinc-100">{activeOrg?.name} · Members</h3>
                <p className="text-xs text-zinc-500">Created {formatDate(activeOrg?.created_at)}</p>
              </div>
              <Button variant="secondary" icon={<UserPlus className="h-4 w-4" />} onClick={() => { setErr(''); setInviteToken(''); setShowInvite(true); }} data-testid="invite-member-button">
                Invite
              </Button>
            </div>

            {members.isError ? (
              <ErrorState message="Could not load members." onRetry={members.refetch} />
            ) : members.isLoading ? (
              <LoadingState />
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b border-ink-700 text-xs uppercase tracking-wider text-zinc-500">
                      <th className="px-2 py-2 text-left font-semibold">Email</th>
                      <th className="px-2 py-2 text-left font-semibold">Role</th>
                      <th className="px-2 py-2 text-left font-semibold">Joined</th>
                    </tr>
                  </thead>
                  <tbody data-testid="members-table">
                    {(members.data || []).map((m) => (
                      <tr key={m.id} className="border-b border-ink-800/50 hover:bg-ink-800/40">
                        <td className="px-2 py-2.5 text-zinc-200">{m.email}</td>
                        <td className="px-2 py-2.5">
                          <Badge tone={m.role === 'admin' ? 'brand' : 'neutral'}>{m.role}</Badge>
                        </td>
                        <td className="px-2 py-2.5 text-zinc-400">{formatDate(m.created_at)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </Card>
        </div>
      )}

      <Modal
        open={showCreate}
        onClose={() => setShowCreate(false)}
        title="New Organization"
        footer={
          <>
            <Button variant="secondary" onClick={() => setShowCreate(false)}>Cancel</Button>
            <Button onClick={() => createOrg.mutate()} loading={createOrg.isPending} disabled={!name.trim()} data-testid="submit-create-org">Create</Button>
          </>
        }
      >
        <Field label="Organization name">
          <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="Acme Security" data-testid="org-name-input" autoFocus />
        </Field>
        {err && <p className="mt-3 text-sm text-red-400">{err}</p>}
      </Modal>

      <Modal
        open={showInvite}
        onClose={() => setShowInvite(false)}
        title="Invite member"
        footer={
          <>
            <Button variant="secondary" onClick={() => setShowInvite(false)}>Close</Button>
            <Button onClick={() => invite.mutate()} loading={invite.isPending} disabled={!inviteEmail.trim()} icon={<Mail className="h-4 w-4" />} data-testid="submit-invite">
              Send invite
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <Field label="Email">
            <Input type="email" value={inviteEmail} onChange={(e) => setInviteEmail(e.target.value)} placeholder="teammate@company.com" data-testid="invite-email-input" />
          </Field>
          <Field label="Role">
            <Select value={inviteRole} onChange={(e) => setInviteRole(e.target.value as Role)} data-testid="invite-role-select">
              {ROLES.map((r) => <option key={r} value={r}>{r}</option>)}
            </Select>
          </Field>
          {inviteToken && (
            <div className="rounded-md border border-emerald-500/30 bg-emerald-500/5 p-3">
              <p className="text-xs text-emerald-400">Invitation created. Share this token:</p>
              <p className="mt-1 break-all font-mono text-xs text-zinc-300">{inviteToken}</p>
            </div>
          )}
          {err && <p className="text-sm text-red-400">{err}</p>}
        </div>
      </Modal>
    </div>
  );
}
