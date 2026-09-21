import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { apiClient } from '../lib/api';
import { todayUK } from '../lib/dateFormat';
import { makeReport, makeUser } from '../test/fixtures';
import { renderApp } from '../test/render';

describe('DailyReportPage', () => {
  const admin = makeUser({ id: 1, role: 'admin', first_name: 'Ada', last_name: 'Min' });

  // rule: §5.1 — counts and per-technician hours for the London day
  it('renders the counts and the technician rows with hours to 2 dp', async () => {
    const getDailyReport = vi.spyOn(apiClient, 'getDailyReport').mockResolvedValue(makeReport());
    renderApp('/reports/daily?date=2026-09-21', admin);

    await screen.findByText('Tess Turner');
    expect(getDailyReport).toHaveBeenCalledWith('2026-09-21');
    const stat = (label: string) => screen.getByText(label, { selector: '.stat-label' }).previousElementSibling;
    expect(stat('Scheduled')).toHaveTextContent('12');
    expect(stat('Completed')).toHaveTextContent('9');
    expect(stat('Cancelled')).toHaveTextContent('1');

    const rows = screen.getAllByRole('row').slice(1);
    expect(rows).toHaveLength(2);
    expect(within(rows[0]).getAllByRole('cell').map((cell) => cell.textContent)).toEqual([
      'Tess Turner',
      '5',
      '4',
      '7.50',
    ]);
    expect(within(rows[1]).getAllByRole('cell').map((cell) => cell.textContent)).toEqual([
      'Omar Bell',
      '4',
      '3',
      '6.13',
    ]);
  });

  it('defaults to today in London and navigates when the date changes', async () => {
    const getDailyReport = vi.spyOn(apiClient, 'getDailyReport').mockResolvedValue(makeReport());
    renderApp('/reports/daily', admin);

    await screen.findByText('Tess Turner');
    expect(getDailyReport).toHaveBeenCalledWith(todayUK());

    fireEvent.change(screen.getByLabelText('Date'), { target: { value: '2026-03-01' } });
    await waitFor(() => expect(getDailyReport).toHaveBeenLastCalledWith('2026-03-01'));

    fireEvent.click(screen.getByRole('button', { name: /Previous day/ }));
    await waitFor(() => expect(getDailyReport).toHaveBeenLastCalledWith('2026-02-28'));
  });

  it('sends technicians back to the visits list', async () => {
    const getDailyReport = vi.spyOn(apiClient, 'getDailyReport');
    vi.spyOn(apiClient, 'getVisits').mockResolvedValue({ visits: [], total: 0, page: 1, page_size: 25, total_pages: 0 });
    renderApp('/reports/daily', makeUser({ role: 'technician' }));

    expect(await screen.findByRole('heading', { name: 'Visits' })).toBeInTheDocument();
    expect(getDailyReport).not.toHaveBeenCalled();
  });
});
