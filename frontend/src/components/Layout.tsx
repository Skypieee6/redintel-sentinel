import { type ReactNode } from 'react';
import { NavLink, useNavigate } from 'react-router-dom';
import {
  LayoutDashboard,
  Building2,
  FolderKanban,
  Radar,
  FileBarChart,
  Settings,
  UserCircle,
  ShieldCheck,
  LogOut,
  ChevronsUpDown,
} from 'lucide-react';
import { useAuth } from '@/context/AuthContext';
import { useOrg } from '@/context/OrgContext';
import { cn } from '@/lib/utils';
import { Select } from './ui';

const nav = [
  { to: '/', label: 'Dashboard', icon: LayoutDashboard, end: true },
  { to: '/organizations', label: 'Organizations', icon: Building2 },
  { to: '/projects', label: 'Projects', icon: FolderKanban },
  { to: '/assets', label: 'Asset Inventory', icon: Radar },
  { to: '/reports', label: 'Reports', icon: FileBarChart },
  { to: '/settings', label: 'Settings', icon: Settings },
  { to: '/profile', label: 'Profile', icon: UserCircle },
];

export function Layout({ children }: { children: ReactNode }) {
  const { user, logout } = useAuth();
  const { orgs, activeOrgId, setActiveOrgId } = useOrg();
  const navigate = useNavigate();

  const handleLogout = async () => {
    await logout();
    navigate('/login');
  };

  return (
    <div className="flex min-h-screen bg-ink-950">
      {/* Sidebar */}
      <aside className="fixed inset-y-0 left-0 hidden w-64 flex-col border-r border-ink-700 bg-ink-900 lg:flex">
        <div className="flex items-center gap-2.5 border-b border-ink-700 px-5 py-5">
          <div className="flex h-9 w-9 items-center justify-center rounded-md bg-brand/15 text-brand">
            <ShieldCheck className="h-5 w-5" />
          </div>
          <div className="leading-tight">
            <p className="font-heading text-sm font-bold tracking-tight text-zinc-100">RedIntel</p>
            <p className="text-[10px] uppercase tracking-[0.25em] text-zinc-500">Sentinel</p>
          </div>
        </div>

        <nav className="flex-1 space-y-1 px-3 py-4">
          {nav.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.end}
              data-testid={`nav-${item.label.toLowerCase().replace(/\s+/g, '-')}`}
              className={({ isActive }) =>
                cn(
                  'flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors duration-150',
                  isActive
                    ? 'bg-brand/10 text-blue-400'
                    : 'text-zinc-400 hover:bg-ink-800 hover:text-zinc-100'
                )
              }
            >
              <item.icon className="h-4 w-4" />
              {item.label}
            </NavLink>
          ))}
        </nav>

        <div className="border-t border-ink-700 p-3">
          <button
            onClick={handleLogout}
            data-testid="logout-button"
            className="flex w-full items-center gap-3 rounded-md px-3 py-2 text-sm font-medium text-zinc-400 transition-colors hover:bg-ink-800 hover:text-red-400"
          >
            <LogOut className="h-4 w-4" />
            Sign out
          </button>
        </div>
      </aside>

      {/* Main */}
      <div className="flex min-h-screen flex-1 flex-col lg:pl-64">
        <header className="sticky top-0 z-30 flex items-center justify-between gap-4 border-b border-ink-700 bg-ink-950/80 px-4 py-3 backdrop-blur sm:px-6">
          <div className="flex items-center gap-2">
            <div className="relative flex items-center">
              <ChevronsUpDown className="pointer-events-none absolute left-2.5 h-4 w-4 text-zinc-500" />
              <Select
                value={activeOrgId}
                onChange={(e) => setActiveOrgId(e.target.value)}
                data-testid="org-switcher"
                className="min-w-[200px] pl-8"
              >
                {orgs.length === 0 && <option value="">No organizations</option>}
                {orgs.map((o) => (
                  <option key={o.id} value={o.id}>
                    {o.name}
                  </option>
                ))}
              </Select>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <div className="hidden text-right sm:block">
              <p className="text-sm font-medium text-zinc-200">{user?.full_name || user?.email}</p>
              <p className="text-xs text-zinc-500">{user?.is_superadmin ? 'Superadmin' : user?.email}</p>
            </div>
            <div className="flex h-9 w-9 items-center justify-center rounded-full bg-ink-800 text-sm font-semibold uppercase text-zinc-300">
              {(user?.full_name || user?.email || '?').charAt(0)}
            </div>
          </div>
        </header>

        <main className="flex-1 px-4 py-6 sm:px-6 lg:px-8">
          <div className="mx-auto max-w-7xl animate-fade-in">{children}</div>
        </main>
      </div>
    </div>
  );
}

export function PageHeader({ title, subtitle, action }: { title: string; subtitle?: string; action?: ReactNode }) {
  return (
    <div className="mb-6 flex flex-wrap items-end justify-between gap-4">
      <div>
        <h1 className="font-heading text-2xl font-bold tracking-tight text-zinc-100 sm:text-3xl">{title}</h1>
        {subtitle && <p className="mt-1 text-sm text-zinc-500">{subtitle}</p>}
      </div>
      {action}
    </div>
  );
}
