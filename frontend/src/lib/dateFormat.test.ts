import { describe, expect, it } from 'vitest';
import { addDaysUK, formatDateTimeUK, formatRangeUK, formatTimeUK, todayUK, ukDatePart } from './dateFormat';

// These run under TZ=Asia/Tokyo and TZ=America/Los_Angeles in CI; every
// expectation below is the London answer and must not move with the host zone.

describe('ukDatePart', () => {
  it('reads the London day, not the UTC day, during BST', () => {
    expect(ukDatePart('2026-06-30T23:30:00Z')).toBe('2026-07-01');
  });

  it('reads the London day during GMT', () => {
    expect(ukDatePart('2026-01-15T23:30:00Z')).toBe('2026-01-15');
  });

  it('accepts offset timestamps', () => {
    expect(ukDatePart('2026-09-21T00:30:00+02:00')).toBe('2026-09-20');
  });

  it('rejects input that is not a timestamp', () => {
    expect(() => ukDatePart('not-a-date')).toThrow(RangeError);
  });
});

describe('formatDateTimeUK', () => {
  it('renders London wall-clock time in BST', () => {
    expect(formatDateTimeUK('2026-09-21T13:30:00Z')).toBe('21 Sep 2026, 14:30');
  });

  it('renders London wall-clock time in GMT', () => {
    expect(formatDateTimeUK('2026-01-15T23:30:00Z')).toBe('15 Jan 2026, 23:30');
  });

  it('rolls the date forward when BST crosses midnight', () => {
    expect(formatDateTimeUK('2026-06-30T23:30:00Z')).toBe('1 Jul 2026, 00:30');
  });

  it('renders a placeholder for an unparseable value', () => {
    expect(formatDateTimeUK('')).toBe('—');
  });
});

describe('formatTimeUK', () => {
  it('renders the London time only', () => {
    expect(formatTimeUK('2026-09-21T13:30:00Z')).toBe('14:30');
    expect(formatTimeUK('2026-01-15T00:05:00Z')).toBe('00:05');
  });
});

describe('formatRangeUK', () => {
  it('shows the end as a bare time on the same London day', () => {
    expect(formatRangeUK('2026-09-21T08:00:00Z', '2026-09-21T10:00:00Z')).toBe('21 Sep 2026, 09:00 – 11:00');
  });

  it('shows both ends in full when the range crosses a London midnight', () => {
    expect(formatRangeUK('2026-06-30T22:00:00Z', '2026-06-30T23:30:00Z')).toBe(
      '30 Jun 2026, 23:00 – 1 Jul 2026, 00:30',
    );
  });
});

describe('todayUK', () => {
  it('is a YYYY-MM-DD day that matches ukDatePart of now', () => {
    const today = todayUK();
    expect(today).toMatch(/^\d{4}-\d{2}-\d{2}$/);
    expect(today).toBe(ukDatePart(new Date().toISOString()));
  });
});

describe('addDaysUK', () => {
  it('shifts across month and year ends', () => {
    expect(addDaysUK('2026-02-28', 1)).toBe('2026-03-01');
    expect(addDaysUK('2026-01-01', -1)).toBe('2025-12-31');
  });

  it('shifts by exactly one calendar day across the DST changes', () => {
    expect(addDaysUK('2026-03-29', 1)).toBe('2026-03-30');
    expect(addDaysUK('2026-10-25', -1)).toBe('2026-10-24');
  });

  it('rejects anything that is not a day', () => {
    expect(() => addDaysUK('2026-09-21T00:00:00Z', 1)).toThrow(RangeError);
  });
});
