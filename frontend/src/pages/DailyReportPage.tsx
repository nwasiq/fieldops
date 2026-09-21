import { useEffect, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { ErrorMessage } from '../components/ErrorMessage';
import { apiClient, describeError } from '../lib/api';
import { addDaysUK, todayUK } from '../lib/dateFormat';
import type { DailyReport } from '../types';

export function DailyReportPage() {
  const [params, setParams] = useSearchParams();
  const date = params.get('date') ?? todayUK();

  const [report, setReport] = useState<DailyReport | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let stale = false;
    setLoading(true);
    setError(null);
    apiClient
      .getDailyReport(date)
      .then((response) => {
        if (!stale) setReport(response);
      })
      .catch((err: unknown) => {
        if (!stale) setError(describeError(err));
      })
      .finally(() => {
        if (!stale) setLoading(false);
      });
    return () => {
      stale = true;
    };
  }, [date]);

  const goTo = (day: string) => {
    if (day) setParams({ date: day }, { replace: true });
  };

  return (
    <>
      <h1>Daily report</h1>
      <form className="toolbar" onSubmit={(e) => e.preventDefault()}>
        <button type="button" className="btn" onClick={() => goTo(addDaysUK(date, -1))}>
          ‹ Previous day
        </button>
        <label className="field">
          Date
          <input type="date" value={date} onChange={(e) => goTo(e.target.value)} />
        </label>
        <button type="button" className="btn" onClick={() => goTo(addDaysUK(date, 1))}>
          Next day ›
        </button>
      </form>

      <ErrorMessage message={error} />

      <div aria-busy={loading}>
        {loading && !report && <p className="muted">Loading report…</p>}
        {report && (
          <>
            <div className="stats">
              <div className="stat">
                <span className="stat-value">{report.scheduled}</span>
                <span className="stat-label">Scheduled</span>
              </div>
              <div className="stat">
                <span className="stat-value">{report.completed}</span>
                <span className="stat-label">Completed</span>
              </div>
              <div className="stat">
                <span className="stat-value">{report.cancelled}</span>
                <span className="stat-label">Cancelled</span>
              </div>
            </div>

            <h2>Technicians</h2>
            {report.technicians.length === 0 ? (
              <p className="muted">No technician activity on this day.</p>
            ) : (
              <table className="list">
                <thead>
                  <tr>
                    <th>Technician</th>
                    <th>Visits</th>
                    <th>Completed</th>
                    <th>Hours clocked</th>
                  </tr>
                </thead>
                <tbody>
                  {report.technicians.map((row) => (
                    <tr key={row.technician_id}>
                      <td data-label="Technician">{row.name}</td>
                      <td data-label="Visits">{row.visits}</td>
                      <td data-label="Completed">{row.completed}</td>
                      <td data-label="Hours clocked">{row.hours_clocked.toFixed(2)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </>
        )}
      </div>
    </>
  );
}
