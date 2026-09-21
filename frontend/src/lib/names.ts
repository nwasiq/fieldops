interface Named {
  first_name: string;
  last_name: string;
}

export function fullName(person: Named): string {
  return `${person.first_name} ${person.last_name}`.trim();
}

export function technicianName(technician: Named | null): string {
  return technician ? fullName(technician) : 'Unassigned';
}

export function recordedByName(person: Named | null): string {
  return person ? fullName(person) : 'Unknown user';
}
