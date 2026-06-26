export type Role = 'admin' | 'manager' | 'analyst' | 'viewer';

export type AssetType =
  | 'domain'
  | 'subdomain'
  | 'ip'
  | 'cidr'
  | 'asn'
  | 'dns_record'
  | 'certificate'
  | 'technology';

export interface User {
  id: string;
  email: string;
  full_name: string;
  is_active: boolean;
  is_superadmin: boolean;
  created_at?: string;
}

export interface TokenPair {
  access_token: string;
  refresh_token: string;
  token_type: string;
  expires_at: string;
}

export interface Organization {
  id: string;
  name: string;
  slug: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface Membership {
  id: string;
  org_id: string;
  user_id: string;
  email?: string;
  role: Role;
  created_at: string;
}

export interface Project {
  id: string;
  org_id: string;
  name: string;
  description: string;
  owner_id: string;
  status: 'active' | 'archived';
  created_at: string;
  updated_at: string;
}

export interface Asset {
  id: string;
  org_id: string;
  project_id: string;
  type: AssetType;
  value: string;
  tags: string[] | null;
  attributes: Record<string, unknown>;
  status: 'active' | 'archived';
  first_seen: string;
  last_seen: string;
  created_at: string;
  updated_at: string;
}

export interface AssetPage {
  assets: Asset[];
  total: number;
  limit: number;
  offset: number;
}

export interface AssetTypeCount {
  type: AssetType;
  count: number;
}

export interface ProjectAssetCount {
  project_id: string;
  name: string;
  status: string;
  assets: number;
}

export interface DashboardSummary {
  total_assets: number;
  assets_by_type: AssetTypeCount[];
  recent_changes: Asset[];
  project_statistics: {
    total: number;
    active: number;
    archived: number;
    by_project: ProjectAssetCount[] | null;
  };
  team_statistics: {
    teams: number;
    members: number;
    pending_invites: number;
  };
}

export interface APIKey {
  id: string;
  name: string;
  prefix: string;
  secret?: string;
  last_used_at?: string;
  revoked: boolean;
  created_at: string;
}
