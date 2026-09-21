// Every timestamp the API sends is an RFC3339 instant. Which calendar day and
// wall-clock time that instant is on is a Europe/London question, never a
// browser-timezone one, so all rendering goes through this module.

const LONDON = 'Europe/London';
const MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
const DAY_PATTERN = /^(\d{4})-(\d{2})-(\d{2})$/;

const londonClock = new Intl.DateTimeFormat('en-GB', {
  timeZone: LONDON,
  year: 'numeric',
  month: '2-digit',
  day: '2-digit',
  hour: '2-digit',
  minute: '2-digit',
  hourCycle: 'h23',
});

interface LondonParts {
  year: number;
  month: number;
  day: number;
  hour: number;
  minute: number;
}

function inLondon(instant: Date): LondonParts {
  const parts = londonClock.formatToParts(instant);
  const read = (type: Intl.DateTimeFormatPartTypes): number =>
    Number(parts.find((part) => part.type === type)?.value);
  return {
    year: read('year'),
    month: read('month'),
    day: read('day'),
    hour: read('hour'),
    minute: read('minute'),
  };
}

function pad2(n: number): string {
  return String(n).padStart(2, '0');
}

function parseInstant(iso: string): Date | null {
  const instant = new Date(iso);
  return Number.isNaN(instant.getTime()) ? null : instant;
}

function dayOf(instant: Date): string {
  const { year, month, day } = inLondon(instant);
  return `${year}-${pad2(month)}-${pad2(day)}`;
}

/** "21 Sep 2026, 14:30" — the instant as seen on a London clock. */
export function formatDateTimeUK(iso: string): string {
  const instant = parseInstant(iso);
  if (!instant) return '—';
  const { year, month, day, hour, minute } = inLondon(instant);
  return `${day} ${MONTHS[month - 1]} ${year}, ${pad2(hour)}:${pad2(minute)}`;
}

/** "14:30" — the London wall-clock time of the instant. */
export function formatTimeUK(iso: string): string {
  const instant = parseInstant(iso);
  if (!instant) return '—';
  const { hour, minute } = inLondon(instant);
  return `${pad2(hour)}:${pad2(minute)}`;
}

/** "21 Sep 2026, 09:00 – 11:30", or both ends in full when they fall on different London days. */
export function formatRangeUK(startIso: string, endIso: string): string {
  const start = parseInstant(startIso);
  const end = parseInstant(endIso);
  if (!start || !end) return `${formatDateTimeUK(startIso)} – ${formatDateTimeUK(endIso)}`;
  const sameDay = dayOf(start) === dayOf(end);
  return `${formatDateTimeUK(startIso)} – ${sameDay ? formatTimeUK(endIso) : formatDateTimeUK(endIso)}`;
}

/** "YYYY-MM-DD" of the London day the instant falls on. */
export function ukDatePart(iso: string): string {
  const instant = parseInstant(iso);
  if (!instant) throw new RangeError(`Not a timestamp: ${iso}`);
  return dayOf(instant);
}

/** "YYYY-MM-DD" of today in London. */
export function todayUK(): string {
  return dayOf(new Date());
}

/** Shift a "YYYY-MM-DD" day by n calendar days. Pure calendar arithmetic, so DST cannot skew it. */
export function addDaysUK(day: string, n: number): string {
  const match = DAY_PATTERN.exec(day);
  if (!match) throw new RangeError(`Not a YYYY-MM-DD day: ${day}`);
  const [, year, month, date] = match;
  const shifted = new Date(Date.UTC(Number(year), Number(month) - 1, Number(date) + n));
  return `${shifted.getUTCFullYear()}-${pad2(shifted.getUTCMonth() + 1)}-${pad2(shifted.getUTCDate())}`;
}
