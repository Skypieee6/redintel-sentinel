import { useState } from 'react';
import { useMutation } from '@tanstack/react-query';
import { Save, Lock } from 'lucide-react';
import { api, apiError } from '@/lib/api';
import { useAuth } from '@/context/AuthContext';
import { PageHeader } from '@/components/Layout';
import { Badge, Button, Card, Field, Input } from '@/components/ui';

export default function Profile() {
  const { user, setUser } = useAuth();
  const [fullName, setFullName] = useState(user?.full_name || '');
  const [profileMsg, setProfileMsg] = useState('');

  const [oldPw, setOldPw] = useState('');
  const [newPw, setNewPw] = useState('');
  const [pwMsg, setPwMsg] = useState('');
  const [pwErr, setPwErr] = useState('');

  const updateProfile = useMutation({
    mutationFn: () => api.auth.updateProfile(fullName),
    onSuccess: (u) => {
      setUser(u);
      setProfileMsg('Profile updated.');
      setTimeout(() => setProfileMsg(''), 2500);
    },
  });

  const changePassword = useMutation({
    mutationFn: () => api.auth.changePassword(oldPw, newPw),
    onSuccess: () => {
      setPwMsg('Password changed successfully.');
      setOldPw('');
      setNewPw('');
      setPwErr('');
      setTimeout(() => setPwMsg(''), 3000);
    },
    onError: (e) => setPwErr(apiError(e)),
  });

  return (
    <div>
      <PageHeader title="User Profile" subtitle="Your account details and security." />

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <Card className="p-5">
          <h3 className="mb-4 font-heading text-lg font-semibold text-zinc-100">Profile</h3>
          <div className="space-y-4">
            <div className="flex items-center gap-3">
              <div className="flex h-12 w-12 items-center justify-center rounded-full bg-ink-800 text-lg font-semibold uppercase text-zinc-300">
                {(user?.full_name || user?.email || '?').charAt(0)}
              </div>
              <div>
                <p className="font-medium text-zinc-100">{user?.email}</p>
                {user?.is_superadmin && <Badge tone="brand">Superadmin</Badge>}
              </div>
            </div>
            <Field label="Full name">
              <Input value={fullName} onChange={(e) => setFullName(e.target.value)} data-testid="profile-name" />
            </Field>
            <div className="flex items-center gap-3">
              <Button icon={<Save className="h-4 w-4" />} onClick={() => updateProfile.mutate()} loading={updateProfile.isPending} data-testid="save-profile">
                Save
              </Button>
              {profileMsg && <span className="text-sm text-emerald-400">{profileMsg}</span>}
            </div>
          </div>
        </Card>

        <Card className="p-5">
          <h3 className="mb-4 font-heading text-lg font-semibold text-zinc-100">Change Password</h3>
          <div className="space-y-4">
            <Field label="Current password">
              <Input type="password" value={oldPw} onChange={(e) => setOldPw(e.target.value)} placeholder="••••••••" data-testid="current-password" />
            </Field>
            <Field label="New password" hint="Minimum 8 characters.">
              <Input type="password" value={newPw} onChange={(e) => setNewPw(e.target.value)} placeholder="••••••••" data-testid="new-password" />
            </Field>
            {pwErr && <p className="text-sm text-red-400">{pwErr}</p>}
            <div className="flex items-center gap-3">
              <Button
                icon={<Lock className="h-4 w-4" />}
                onClick={() => changePassword.mutate()}
                loading={changePassword.isPending}
                disabled={!oldPw || newPw.length < 8}
                data-testid="change-password"
              >
                Update password
              </Button>
              {pwMsg && <span className="text-sm text-emerald-400">{pwMsg}</span>}
            </div>
          </div>
        </Card>
      </div>
    </div>
  );
}
