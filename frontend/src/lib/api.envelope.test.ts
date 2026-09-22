import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ApiClient, ApiError } from './api';
import { getToken } from './session';

type Body = { success: true; data: unknown } | { success: false; error?: string };

function jsonResponse(status: number, body: Body): Response {
  return {
    ok: status >= 200 && status < 300,
    status,
    json: () => Promise.resolve(body),
  } as unknown as Response;
}

describe('ApiClient.request envelope handling', () => {
  const fetchMock = vi.fn<(input: string, init: RequestInit) => Promise<Response>>();
  const onUnauthorized = vi.fn();
  let client: ApiClient;

  beforeEach(() => {
    vi.stubGlobal('fetch', fetchMock);
    fetchMock.mockReset();
    onUnauthorized.mockReset();
    client = new ApiClient({ onUnauthorized });
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('unwraps {success, data} and returns data', async () => {
    const site = { id: 1, name: 'Depot', address: '1 Dock Rd', contact_phone: '0161 000 0000' };
    const page = { sites: [site], total: 1, page: 1, page_size: 20, total_pages: 1 };
    fetchMock.mockResolvedValue(jsonResponse(200, { success: true, data: page }));

    await expect(client.getSites()).resolves.toEqual(page);
    expect(fetchMock).toHaveBeenCalledWith('/api/sites', expect.objectContaining({ headers: expect.any(Object) }));
  });

  it('attaches the stored JWT as a bearer token and sends JSON bodies', async () => {
    localStorage.setItem('fieldops_token', 'jwt-abc');
    fetchMock.mockResolvedValue(jsonResponse(200, { success: true, data: { token: 't', user: {} } }));

    await client.login('a@b.c', 'secret');

    const [url, init] = fetchMock.mock.calls[0];
    expect(url).toBe('/api/auth/login');
    expect(init.method).toBe('POST');
    expect(init.body).toBe(JSON.stringify({ email: 'a@b.c', password: 'secret' }));
    expect(init.headers).toMatchObject({
      Authorization: 'Bearer jwt-abc',
      'Content-Type': 'application/json',
    });
  });

  it('throws ApiError carrying the status and the server message on {success:false}', async () => {
    fetchMock.mockResolvedValue(
      jsonResponse(409, { success: false, error: 'Clock-out without a clock-in is rejected' }),
    );

    const failure = client.clockOut(7);
    await expect(failure).rejects.toBeInstanceOf(ApiError);
    await expect(failure).rejects.toMatchObject({
      status: 409,
      message: 'Clock-out without a clock-in is rejected',
    });
    expect(onUnauthorized).not.toHaveBeenCalled();
  });

  it('falls back to a generic message when the error body is not an envelope', async () => {
    fetchMock.mockResolvedValue({
      ok: false,
      status: 502,
      json: () => Promise.reject(new Error('not json')),
    } as unknown as Response);

    await expect(client.getSites()).rejects.toMatchObject({ status: 502, message: 'Request failed (502)' });
  });

  it('clears the token and signals unauthorized on a 401 to an authenticated request', async () => {
    localStorage.setItem('fieldops_token', 'expired');
    fetchMock.mockResolvedValue(jsonResponse(401, { success: false, error: 'token expired' }));

    await expect(client.getVisit(1)).rejects.toMatchObject({ status: 401, message: 'token expired' });
    expect(getToken()).toBeNull();
    expect(onUnauthorized).toHaveBeenCalledTimes(1);
  });

  it('treats a 401 with no token as a plain failure (a wrong password on the login form)', async () => {
    fetchMock.mockResolvedValue(jsonResponse(401, { success: false, error: 'invalid credentials' }));

    await expect(client.login('a@b.c', 'wrong')).rejects.toMatchObject({ status: 401, message: 'invalid credentials' });
    expect(onUnauthorized).not.toHaveBeenCalled();
  });

  it('serialises list params server-side and omits the ones that are unset', async () => {
    fetchMock.mockResolvedValue(
      jsonResponse(200, { success: true, data: { visits: [], total: 0, page: 1, page_size: 25, total_pages: 0 } }),
    );

    await client.getVisits({ from: '2026-09-21', to: '2026-09-22', page: 2, page_size: 25, status: undefined });

    expect(fetchMock.mock.calls[0][0]).toBe('/api/visits?from=2026-09-21&to=2026-09-22&page=2&page_size=25');
  });
});
