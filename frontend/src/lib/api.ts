import type {
  DailyReport,
  LoginResponse,
  UserListParams,
  Visit,
  VisitListParams,
  VisitListResponse,
  UserListResponse,
  SiteListResponse,
} from '../types';
import { clearSession, getToken } from './session';

type Envelope<T> = { success: true; data: T } | { success: false; error?: string };

export class ApiError extends Error {
  readonly status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
  }
}

export function describeError(err: unknown): string {
  if (err instanceof Error && err.message) return err.message;
  return 'Something went wrong. Please try again.';
}

interface ApiClientOptions {
  baseUrl?: string;
  onUnauthorized?: () => void;
}

type QueryValue = string | number | boolean | undefined;

function queryString(params: Record<string, QueryValue>): string {
  const search = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === '') continue;
    search.set(key, String(value));
  }
  const encoded = search.toString();
  return encoded ? `?${encoded}` : '';
}

function redirectToLogin(): void {
  if (window.location.pathname !== '/login') window.location.assign('/login');
}

export class ApiClient {
  private readonly baseUrl: string;
  private readonly onUnauthorized: () => void;

  constructor(options: ApiClientOptions = {}) {
    this.baseUrl = options.baseUrl ?? '/api';
    this.onUnauthorized = options.onUnauthorized ?? redirectToLogin;
  }

  private async request<T>(path: string, init: RequestInit = {}): Promise<T> {
    const headers: Record<string, string> = { Accept: 'application/json' };
    if (init.body !== undefined) headers['Content-Type'] = 'application/json';
    const token = getToken();
    if (token) headers['Authorization'] = `Bearer ${token}`;

    const response = await fetch(`${this.baseUrl}${path}`, { ...init, headers });

    let body: Envelope<T> | null = null;
    try {
      body = (await response.json()) as Envelope<T>;
    } catch {
      body = null;
    }

    // A rejected token means the session is over; a 401 on the login form itself
    // is just a wrong password and must not bounce the user around.
    if (response.status === 401 && token) {
      clearSession();
      this.onUnauthorized();
    }

    if (!response.ok || body === null || body.success !== true) {
      const message =
        body !== null && body.success === false && body.error
          ? body.error
          : `Request failed (${response.status})`;
      throw new ApiError(response.status, message);
    }
    return body.data;
  }

  private post<T>(path: string, payload?: unknown): Promise<T> {
    return this.request<T>(path, {
      method: 'POST',
      body: payload === undefined ? undefined : JSON.stringify(payload),
    });
  }

  login(email: string, password: string): Promise<LoginResponse> {
    return this.post<LoginResponse>('/auth/login', { email, password });
  }

  getVisits(params: VisitListParams): Promise<VisitListResponse> {
    return this.request<VisitListResponse>(`/visits${queryString({ ...params })}`);
  }

  getVisit(id: number): Promise<Visit> {
    return this.request<Visit>(`/visits/${id}`);
  }

  clockIn(id: number): Promise<Visit> {
    return this.post<Visit>(`/visits/${id}/clock-in`);
  }

  clockOut(id: number): Promise<Visit> {
    return this.post<Visit>(`/visits/${id}/clock-out`);
  }

  cancelVisit(id: number): Promise<Visit> {
    return this.post<Visit>(`/visits/${id}/cancel`);
  }

  getUsers(params: UserListParams = {}): Promise<UserListResponse> {
    return this.request<UserListResponse>(`/users${queryString({ ...params })}`);
  }

  getSites(): Promise<SiteListResponse> {
    return this.request<SiteListResponse>('/sites');
  }

  getDailyReport(date: string): Promise<DailyReport> {
    return this.request<DailyReport>(`/reports/daily${queryString({ date })}`);
  }
}

export const apiClient = new ApiClient();
