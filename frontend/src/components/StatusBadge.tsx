import type { VisitStatus } from '../types';

const LABELS: Record<VisitStatus, string> = {
  scheduled: 'Scheduled',
  in_progress: 'In progress',
  completed: 'Completed',
  cancelled: 'Cancelled',
};

export function StatusBadge({ status }: { status: VisitStatus }) {
  return <span className={`badge badge-${status}`}>{LABELS[status]}</span>;
}
