// SwiftLog API Client
import type { Project, LogGroup, LogRun, LogLine, PaginatedResponse } from '@/types';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

class APIClient {
  private baseURL: string;
  private token: string | null = null;

  constructor(baseURL: string) {
    this.baseURL = baseURL;
    if (typeof window !== 'undefined') {
      this.token = localStorage.getItem('swiftlog_token');
    }
  }

  setToken(token: string) {
    this.token = token;
    if (typeof window !== 'undefined') {
      localStorage.setItem('swiftlog_token', token);
    }
  }

  private async request<T>(endpoint: string, options?: RequestInit): Promise<T> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };

    if (options?.headers) {
      Object.assign(headers, options.headers);
    }

    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    const response = await fetch(`${this.baseURL}${endpoint}`, {
      ...options,
      headers,
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Unknown error' }));
      throw new Error(error.error || error.message || `HTTP ${response.status}`);
    }

    return response.json();
  }

  // Projects
  async getProjects(): Promise<Project[]> {
    return this.request<Project[]>('/projects');
  }

  async getProject(id: string): Promise<Project> {
    return this.request<Project>(`/projects/${id}`);
  }

  // Groups
  async getProjectGroups(projectId: string): Promise<LogGroup[]> {
    return this.request<LogGroup[]>(`/projects/${projectId}/groups`);
  }

  async getGroup(id: string): Promise<LogGroup> {
    return this.request<LogGroup>(`/groups/${id}`);
  }

  // Runs
  async getGroupRuns(
    groupId: string,
    params?: { limit?: number; offset?: number }
  ): Promise<PaginatedResponse<LogRun>> {
    const query = new URLSearchParams();
    if (params?.limit) query.set('limit', params.limit.toString());
    if (params?.offset) query.set('offset', params.offset.toString());

    const queryString = query.toString();
    const url = `/groups/${groupId}/runs${queryString ? `?${queryString}` : ''}`;

    return this.request<PaginatedResponse<LogRun>>(url);
  }

  async getRun(id: string): Promise<LogRun> {
    return this.request<LogRun>(`/runs/${id}`);
  }

  async getRunLogs(id: string): Promise<LogLine[]> {
    return this.request<LogLine[]>(`/runs/${id}/logs`);
  }

  async getAIReport(id: string): Promise<{ report: string; status: string }> {
    return this.request(`/runs/${id}/ai-report`);
  }

  async triggerAIAnalysis(id: string): Promise<void> {
    return this.request(`/runs/${id}/analyze`, { method: 'POST' });
  }

  // Auth methods
  async login(username: string, password: string): Promise<{ token: string; user: any }> {
    const response = await this.request<{ token: string; user: any }>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    });
    this.setToken(response.token);
    return response;
  }

  async register(username: string, password: string): Promise<{ token: string; user: any }> {
    const response = await this.request<{ token: string; user: any }>('/auth/register', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    });
    this.setToken(response.token);
    return response;
  }

  async getCurrentUser(): Promise<any> {
    return this.request('/auth/me');
  }

  async listTokens(): Promise<{ tokens: any[] }> {
    return this.request('/auth/tokens');
  }

  async createToken(name: string): Promise<{ token: string; token_info: any }> {
    return this.request('/auth/tokens', {
      method: 'POST',
      body: JSON.stringify({ name }),
    });
  }

  async deleteToken(id: string): Promise<void> {
    return this.request(`/auth/tokens/${id}`, { method: 'DELETE' });
  }

  async listUsers(): Promise<{ users: any[] }> {
    return this.request('/auth/users');
  }

  // Status
  async getStatistics(): Promise<{
    run_statistics: {
      running: number;
      completed: number;
      failed: number;
      aborted: number;
      total: number;
    };
    ai_statistics: {
      pending: number;
      processing: number;
      completed: number;
      failed: number;
      total: number;
    };
    queue_length: number;
  }> {
    return this.request('/status/statistics');
  }

  async getRecentRuns(limit?: number): Promise<PaginatedResponse<LogRun>> {
    const query = limit ? `?limit=${limit}` : '';
    return this.request(`/status/recent${query}`);
  }

  logout() {
    if (typeof window !== 'undefined') {
      localStorage.removeItem('swiftlog_token');
    }
    this.token = null;
  }
}

export const api = new APIClient(API_BASE_URL);
