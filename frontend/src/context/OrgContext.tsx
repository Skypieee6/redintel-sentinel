import { createContext, useContext, useEffect, useState, type ReactNode } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';
import type { Organization } from '@/lib/types';

interface OrgState {
  orgs: Organization[];
  activeOrg: Organization | null;
  activeOrgId: string;
  setActiveOrgId: (id: string) => void;
  loading: boolean;
  refetch: () => void;
}

const OrgContext = createContext<OrgState | undefined>(undefined);
const ORG_KEY = 'rin_active_org';

export function OrgProvider({ children }: { children: ReactNode }) {
  const { data: orgs = [], isLoading, refetch } = useQuery({
    queryKey: ['orgs'],
    queryFn: api.orgs.list,
  });
  const [activeOrgId, setActive] = useState<string>(() => localStorage.getItem(ORG_KEY) || '');

  useEffect(() => {
    if (orgs.length === 0) return;
    const exists = orgs.some((o) => o.id === activeOrgId);
    if (!activeOrgId || !exists) {
      setActive(orgs[0].id);
      localStorage.setItem(ORG_KEY, orgs[0].id);
    }
  }, [orgs, activeOrgId]);

  const setActiveOrgId = (id: string) => {
    setActive(id);
    localStorage.setItem(ORG_KEY, id);
  };

  const activeOrg = orgs.find((o) => o.id === activeOrgId) || null;

  return (
    <OrgContext.Provider value={{ orgs, activeOrg, activeOrgId, setActiveOrgId, loading: isLoading, refetch }}>
      {children}
    </OrgContext.Provider>
  );
}

export function useOrg() {
  const ctx = useContext(OrgContext);
  if (!ctx) throw new Error('useOrg must be used within OrgProvider');
  return ctx;
}
