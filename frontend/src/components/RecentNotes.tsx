import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { formatTimeUK } from '../lib/dateFormat';
import { getToken } from '../lib/session';
import type { Visit, VisitListResponse } from '../types';

interface RecentNotesProps {
  technicianId: number;
  day: string;
  excludeVisitId: number;
}

type ListEnvelope = { success: true; data: VisitListResponse } | { success: false; error?: string };

/** The technician's other visits on the same London day that carry notes. */
export function RecentNotes({ technicianId, day, excludeVisitId }: RecentNotesProps) {
  const [visits, setVisits] = useState<Visit[] | null>(null);
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    let stale = false;
    setVisits(null);
    setFailed(false);
    const load = async (): Promise<Visit[]> => {
      const response = await fetch(`/api/visits?from=${day}&to=${day}&technician_id=${technicianId}`, {
        headers: { Accept: 'application/json', Authorization: `Bearer ${getToken() ?? ''}` },
      });
      const body = (await response.json()) as ListEnvelope;
      if (!response.ok || !body.success) throw new Error('recent notes request failed');
      return body.data.visits.filter(
        (v) => v.id !== excludeVisitId && v.scheduled_start.split('T')[0] === day && v.notes,
      );
    };
    load()
      .then((rows) => {
        if (!stale) setVisits(rows);
      })
      .catch(() => {
        if (!stale) setFailed(true);
      });
    return () => {
      stale = true;
    };
  }, [technicianId, day, excludeVisitId]);

  return (
    <section className="recent-notes">
      <h2>Recent notes</h2>
      {failed && <p className="muted">Recent notes could not be loaded.</p>}
      {visits && visits.length === 0 && (
        <p className="muted">No other notes from this technician on this day.</p>
      )}
      {visits && visits.length > 0 && (
        <ul className="note-list">
          {visits.map((v) => (
            <li key={v.id} className="card">
              <p className="note-meta">
                <Link to={`/visits/${v.id}`}>{v.site.name}</Link> · {formatTimeUK(v.scheduled_start)}
              </p>
              <p className="notes">{v.notes}</p>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
