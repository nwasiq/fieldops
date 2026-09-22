import { useCallback, useEffect, useState, type FormEvent } from 'react';
import { Link, useParams } from 'react-router-dom';
import { ErrorMessage } from '../components/ErrorMessage';
import { RecentNotes } from '../components/RecentNotes';
import { StatusBadge } from '../components/StatusBadge';
import { apiClient, describeError } from '../lib/api';
import { useCurrentUser } from '../lib/auth';
import { formatDateTimeUK, formatRangeUK, ukDatePart } from '../lib/dateFormat';
import { recordedByName, technicianName } from '../lib/names';
import {
  canCancelVisits,
  canClock,
  canEditNotes,
  cancelAllowed,
  clockInAllowed,
  clockOutAllowed,
} from '../lib/permissions';
import type { ClockKind, Visit } from '../types';

const KIND_LABELS: Record<ClockKind, string> = { in: 'Clock in', out: 'Clock out' };

export function VisitDetailPage() {
  const { id } = useParams();
  const visitId = Number(id);
  const user = useCurrentUser();

  const [visit, setVisit] = useState<Visit | null>(null);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [notesDraft, setNotesDraft] = useState('');

  const load = useCallback(async () => {
    try {
      setVisit(await apiClient.getVisit(visitId));
      setLoadError(null);
    } catch (err) {
      setLoadError(describeError(err));
    }
  }, [visitId]);

  useEffect(() => {
    void load();
  }, [load]);

  const savedNotes = visit?.notes ?? '';
  useEffect(() => {
    setNotesDraft(savedNotes);
  }, [savedNotes]);

  const perform = async (action: () => Promise<Visit>) => {
    setBusy(true);
    setActionError(null);
    try {
      await action();
      await load();
    } catch (err) {
      setActionError(describeError(err));
    } finally {
      setBusy(false);
    }
  };

  const cancel = () => {
    if (!window.confirm('Cancel this visit? This cannot be undone.')) return;
    void perform(() => apiClient.cancelVisit(visitId));
  };

  const saveNotes = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    void perform(() => apiClient.updateVisitNotes(visitId, notesDraft));
  };

  if (loadError && !visit) {
    return (
      <>
        <h1>Visit</h1>
        <ErrorMessage message={loadError} />
        <Link to="/visits">Back to visits</Link>
      </>
    );
  }
  if (!visit) return <p className="muted">Loading visit…</p>;

  const showClock = canClock(user, visit);
  const showCancel = canCancelVisits(user);
  const showNotesEditor = canEditNotes(user, visit);
  const notesDirty = notesDraft !== savedNotes;

  return (
    <>
      <p>
        <Link to="/visits">← Visits</Link>
      </p>
      <h1 className="title-row">
        {visit.site.name} <StatusBadge status={visit.status} />
      </h1>

      <dl className="facts card">
        <dt>Site</dt>
        <dd>{visit.site.name}</dd>
        <dt>Technician</dt>
        <dd>{technicianName(visit.technician)}</dd>
        <dt>Scheduled</dt>
        <dd>{formatRangeUK(visit.scheduled_start, visit.scheduled_end)}</dd>
        <dt>Status</dt>
        <dd>
          <StatusBadge status={visit.status} />
        </dd>
        {visit.cancelled_at && (
          <>
            <dt>Cancelled</dt>
            <dd>{formatDateTimeUK(visit.cancelled_at)}</dd>
          </>
        )}
      </dl>

      {(showClock || showCancel) && (
        <section className="actions">
          {showClock && (
            <>
              <button
                type="button"
                className="btn btn-primary"
                disabled={busy || !clockInAllowed(visit.status)}
                onClick={() => void perform(() => apiClient.clockIn(visitId))}
              >
                Clock in
              </button>
              <button
                type="button"
                className="btn btn-primary"
                disabled={busy || !clockOutAllowed(visit.status)}
                onClick={() => void perform(() => apiClient.clockOut(visitId))}
              >
                Clock out
              </button>
            </>
          )}
          {showCancel && (
            <button
              type="button"
              className="btn btn-danger"
              disabled={busy || !cancelAllowed(visit.status)}
              onClick={cancel}
            >
              Cancel visit
            </button>
          )}
        </section>
      )}
      <ErrorMessage message={actionError} />

      <h2 id="notes-heading">Notes</h2>
      {showNotesEditor ? (
        <form className="notes-form" onSubmit={saveNotes}>
          <textarea
            aria-labelledby="notes-heading"
            rows={4}
            placeholder="Access instructions, what was found on site…"
            value={notesDraft}
            disabled={busy}
            onChange={(e) => setNotesDraft(e.target.value)}
          />
          <button type="submit" className="btn btn-primary" disabled={busy || !notesDirty}>
            Save notes
          </button>
        </form>
      ) : visit.notes ? (
        <p className="notes card">{visit.notes}</p>
      ) : (
        <p className="muted">No notes.</p>
      )}

      {visit.technician_id !== null && (
        <RecentNotes
          technicianId={visit.technician_id}
          day={ukDatePart(visit.scheduled_start)}
          excludeVisitId={visit.id}
        />
      )}

      <h2>Clock events</h2>
      {visit.clock_events.length === 0 ? (
        <p className="muted">No clock events yet.</p>
      ) : (
        <table className="list">
          <thead>
            <tr>
              <th>Event</th>
              <th>Time</th>
              <th>Recorded by</th>
            </tr>
          </thead>
          <tbody>
            {visit.clock_events.map((event) => (
              <tr key={event.id}>
                <td data-label="Event">{KIND_LABELS[event.kind]}</td>
                <td data-label="Time">{formatDateTimeUK(event.occurred_at)}</td>
                <td data-label="Recorded by">{recordedByName(event.recorded_by)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </>
  );
}
