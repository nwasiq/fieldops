import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { makeVisit, makeVisitList } from '../test/fixtures';
import { RecentNotes } from './RecentNotes';

function jsonResponse(body: unknown): Response {
  return { ok: true, status: 200, json: () => Promise.resolve(body) } as unknown as Response;
}

describe('RecentNotes', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  // rule: §3.7 — the panel lists the technician's other visits that carry notes on the day
  it("lists the technician's other noted visits for the day and skips the current one", async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      jsonResponse({
        success: true,
        data: makeVisitList({
          visits: [
            makeVisit({ id: 7, notes: 'Key safe by the side door.' }),
            makeVisit({
              id: 8,
              site: { id: 3, name: 'Northgate Shopping Centre' },
              scheduled_start: '2026-09-21T11:00:00Z',
              scheduled_end: '2026-09-21T12:30:00Z',
              notes: 'Meter cupboard key is with reception.',
            }),
            makeVisit({
              id: 9,
              site: { id: 4, name: 'Harbour View Apartments' },
              scheduled_start: '2026-09-21T14:00:00Z',
              scheduled_end: '2026-09-21T15:00:00Z',
              notes: null,
            }),
          ],
          total: 3,
        }),
      }),
    );
    vi.stubGlobal('fetch', fetchMock);
    localStorage.setItem('fieldops_token', 'jwt-abc');

    render(
      <MemoryRouter>
        <RecentNotes technicianId={3} day="2026-09-21" excludeVisitId={7} />
      </MemoryRouter>,
    );

    expect(await screen.findByText('Meter cupboard key is with reception.')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Northgate Shopping Centre' })).toHaveAttribute('href', '/visits/8');
    expect(screen.getByText(/12:00/)).toBeInTheDocument();
    expect(screen.queryByText('Key safe by the side door.')).not.toBeInTheDocument();
    expect(screen.queryByText('Harbour View Apartments')).not.toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/visits?from=2026-09-21&to=2026-09-21&technician_id=3',
      expect.objectContaining({ headers: expect.objectContaining({ Authorization: 'Bearer jwt-abc' }) }),
    );
  });

  it('shows an empty state when nothing else is noted that day', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        jsonResponse({ success: true, data: makeVisitList({ visits: [makeVisit({ id: 7, notes: 'Mine.' })] }) }),
      ),
    );

    render(
      <MemoryRouter>
        <RecentNotes technicianId={3} day="2026-09-21" excludeVisitId={7} />
      </MemoryRouter>,
    );

    expect(await screen.findByText('No other notes from this technician on this day.')).toBeInTheDocument();
  });
});
