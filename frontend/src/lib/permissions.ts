import type { Role, User, Visit, VisitStatus } from '../types';

const STAFF_ROLES: readonly Role[] = ['admin', 'dispatcher'];

export function isStaff(user: User): boolean {
  return STAFF_ROLES.includes(user.role);
}

export function canViewReports(user: User): boolean {
  return isStaff(user);
}

export function canCancelVisits(user: User): boolean {
  return isStaff(user);
}

export function canClock(user: User, visit: Visit): boolean {
  return isStaff(user) || (visit.technician_id !== null && visit.technician_id === user.id);
}

export function clockInAllowed(status: VisitStatus): boolean {
  return status === 'scheduled';
}

export function clockOutAllowed(status: VisitStatus): boolean {
  return status === 'in_progress';
}

export function cancelAllowed(status: VisitStatus): boolean {
  return status === 'scheduled' || status === 'in_progress';
}
