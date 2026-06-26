import axios, { type AxiosRequestConfig } from 'axios';
import type {
  APIKey,
  Asset,
  AssetPage,
  DashboardSummary,
  DiscoveryInputType,
  DiscoveryJob,
  DiscoveryJobPage,
  Membership,
  Organization,
  Project,
  Role,
  TokenPair,
  User,
} from './types';

const BASE = `${import.meta.env.REACT_APP_BACKEND_URL}/api/v1`;

const ACCESS_KEY = 'rin_access';
const REFRESH_KEY = 'rin_refresh';

export const tokenStore = {
  get access() {
    return localStorage.getItem(ACCESS_KEY) || '';
  },
  get refresh() {
    return localStorage.getItem(REFRESH_KEY) || '';
  },
  set(tokens: TokenPair) {
    localStorage.setItem(ACCESS_KEY, tokens.access_token);
    localStorage.setItem(REFRESH_KEY, tokens.refresh_token);
  },
  clear() {
    localStorage.removeItem(ACCESS_KEY);
    localStorage.removeItem(REFRESH_KEY);
  },
};

const http = axios.create({ baseURL: BASE, headers: { 'Content-Type': 'application/json' } });

http.interceptors.request.use((config) => {
  const t = tokenStore.access;
  if (t) config.headers.Authorization = `Bearer ${t}`;
  return config;
});

let refreshing: Promise<string> | null = null;

async function doRefresh(): Promise<string> {
  const refresh = tokenStore.refresh;
  if (!refresh) throw new Error('no refresh token');
  const res = await axios.post(`${BASE}/auth/refresh`, { refresh_token: refresh });
  const tokens: TokenPair = res.data.data.tokens;
  tokenStore.set(tokens);
  return tokens.access_token;
}

http.interceptors.response.use(
  (r) => r,
  async (error) => {
    const original = error.config as AxiosRequestConfig & { _retry?: boolean };
    const status = error.response?.status;
    const isAuthCall = original?.url?.includes('/auth/login') || original?.url?.includes('/auth/refresh');
    if (status === 401 && !original._retry && !isAuthCall && tokenStore.refresh) {
      original._retry = true;
      try {
        if (!refreshing) refreshing = doRefresh().finally(() => (refreshing = null));
        const token = await refreshing;
        original.headers = { ...original.headers, Authorization: `Bearer ${token}` };
        return http(original);
      } catch {
        tokenStore.clear();
      }
    }
    return Promise.reject(error);
  }
);

/** Extract a user-friendly message from an axios error. */
export function apiError(err: unknown): string {
  const e = err as { response?: { data?: { error?: { message?: string } } }; message?: string };
  return e.response?.data?.error?.message || e.message || 'Something went wrong';
}

const unwrap = <T,>(d: { data: T }): T => d.data;

export const api = {
  auth: {
    async register(payload: { email: string; password: string; full_name: string }) {
      const { data } = await http.post('/auth/register', payload);
      return data.data as { user: User; tokens: TokenPair };
    },
    async login(payload: { email: string; password: string }) {
      const { data } = await http.post('/auth/login', payload);
      return data.data as { user: User; tokens: TokenPair };
    },
    async logout() {
      await http.post('/auth/logout', { refresh_token: tokenStore.refresh });
    },
    async me() {
      const { data } = await http.get('/auth/me');
      return unwrap<User>(data);
    },
    async updateProfile(full_name: string) {
      const { data } = await http.put('/auth/me', { full_name });
      return unwrap<User>(data);
    },
    async changePassword(old_password: string, new_password: string) {
      await http.post('/auth/change-password', { old_password, new_password });
    },
    async listApiKeys() {
      const { data } = await http.get('/auth/api-keys');
      return (unwrap<APIKey[]>(data) || []) as APIKey[];
    },
    async createApiKey(name: string) {
      const { data } = await http.post('/auth/api-keys', { name });
      return unwrap<APIKey>(data);
    },
    async revokeApiKey(id: string) {
      await http.delete(`/auth/api-keys/${id}`);
    },
  },
  orgs: {
    async list() {
      const { data } = await http.get('/orgs');
      return (unwrap<Organization[]>(data) || []) as Organization[];
    },
    async create(name: string) {
      const { data } = await http.post('/orgs', { name });
      return unwrap<Organization>(data);
    },
    async update(orgId: string, name: string) {
      const { data } = await http.put(`/orgs/${orgId}`, { name });
      return unwrap<Organization>(data);
    },
    async remove(orgId: string) {
      await http.delete(`/orgs/${orgId}`);
    },
    async members(orgId: string) {
      const { data } = await http.get(`/orgs/${orgId}/members`);
      return (unwrap<Membership[]>(data) || []) as Membership[];
    },
    async setMemberRole(orgId: string, email: string, role: Role) {
      await http.put(`/orgs/${orgId}/members`, { email, role });
    },
    async invite(orgId: string, email: string, role: Role) {
      const { data } = await http.post(`/orgs/${orgId}/invitations`, { email, role });
      return data.data;
    },
  },
  projects: {
    async list(orgId: string) {
      const { data } = await http.get(`/orgs/${orgId}/projects`);
      return (unwrap<Project[]>(data) || []) as Project[];
    },
    async create(orgId: string, payload: { name: string; description?: string }) {
      const { data } = await http.post(`/orgs/${orgId}/projects`, payload);
      return unwrap<Project>(data);
    },
    async get(orgId: string, projectId: string) {
      const { data } = await http.get(`/orgs/${orgId}/projects/${projectId}`);
      return unwrap<Project>(data);
    },
    async update(orgId: string, projectId: string, payload: { name?: string; description?: string; status?: string }) {
      const { data } = await http.put(`/orgs/${orgId}/projects/${projectId}`, payload);
      return unwrap<Project>(data);
    },
    async remove(orgId: string, projectId: string) {
      await http.delete(`/orgs/${orgId}/projects/${projectId}`);
    },
    async archive(orgId: string, projectId: string) {
      await http.post(`/orgs/${orgId}/projects/${projectId}/archive`);
    },
    async unarchive(orgId: string, projectId: string) {
      await http.post(`/orgs/${orgId}/projects/${projectId}/unarchive`);
    },
  },
  assets: {
    async list(
      orgId: string,
      projectId: string,
      params: { type?: string; q?: string; tag?: string; status?: string; limit?: number; offset?: number }
    ) {
      const { data } = await http.get(`/orgs/${orgId}/projects/${projectId}/assets`, { params });
      return unwrap<AssetPage>(data);
    },
    async get(orgId: string, projectId: string, assetId: string) {
      const { data } = await http.get(`/orgs/${orgId}/projects/${projectId}/assets/${assetId}`);
      return unwrap<Asset>(data);
    },
    async create(
      orgId: string,
      projectId: string,
      payload: { type: string; value: string; tags?: string[]; attributes?: Record<string, unknown> }
    ) {
      const { data } = await http.post(`/orgs/${orgId}/projects/${projectId}/assets`, payload);
      return unwrap<Asset>(data);
    },
    async update(orgId: string, projectId: string, assetId: string, payload: Partial<Asset>) {
      const { data } = await http.put(`/orgs/${orgId}/projects/${projectId}/assets/${assetId}`, payload);
      return unwrap<Asset>(data);
    },
    async remove(orgId: string, projectId: string, assetId: string) {
      await http.delete(`/orgs/${orgId}/projects/${projectId}/assets/${assetId}`);
    },
  },
  dashboard: {
    async get(orgId: string) {
      const { data } = await http.get(`/orgs/${orgId}/dashboard`);
      return unwrap<DashboardSummary>(data);
    },
  },
  discovery: {
    async list(orgId: string, projectId: string, params: { limit?: number; offset?: number } = {}) {
      const { data } = await http.get(`/orgs/${orgId}/projects/${projectId}/discovery`, { params });
      return unwrap<DiscoveryJobPage>(data);
    },
    async start(orgId: string, projectId: string, payload: { input_type: DiscoveryInputType; input_value: string }) {
      const { data } = await http.post(`/orgs/${orgId}/projects/${projectId}/discovery`, payload);
      return unwrap<DiscoveryJob>(data);
    },
    async get(orgId: string, projectId: string, jobId: string) {
      const { data } = await http.get(`/orgs/${orgId}/projects/${projectId}/discovery/${jobId}`);
      return unwrap<DiscoveryJob>(data);
    },
  },
  reports: {
    async download(orgId: string, projectId: string, format: string) {
      const res = await http.get(`/orgs/${orgId}/projects/${projectId}/report`, {
        params: { format },
        responseType: 'blob',
      });
      const disposition = res.headers['content-disposition'] || '';
      const match = /filename=([^;]+)/.exec(disposition);
      const filename = match ? match[1].trim() : `report.${format}`;
      const url = URL.createObjectURL(res.data as Blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      a.remove();
      URL.revokeObjectURL(url);
    },
  },
};
