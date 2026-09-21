package services

import (
	"fmt"
	"time"
	_ "time/tzdata" // the runtime image has no zoneinfo; Europe/London must always resolve
)

// DayFormat is the wire format for a calendar day.
const DayFormat = "2006-01-02"

// ukLocation is the one timezone the business operates in. Every "which day is
// this" question is answered here, never in the host's local zone.
var ukLocation = mustLoadLocation("Europe/London")

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(fmt.Sprintf("load timezone %s: %v", name, err))
	}
	return loc
}

// UKDay returns the Europe/London calendar day an instant falls on, as
// YYYY-MM-DD (§3.3).
func UKDay(t time.Time) string {
	return t.In(ukLocation).Format(DayFormat)
}

// UKDayStart returns the instant at which the Europe/London day begins, in UTC.
// It returns an error when day is not a valid YYYY-MM-DD.
func UKDayStart(day string) (time.Time, error) {
	parsed, err := time.ParseInLocation(DayFormat, day, ukLocation)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid day %q: expected YYYY-MM-DD", day)
	}
	return parsed.UTC(), nil
}

// UKDayRange returns the half-open instant range [start, end) covering the
// Europe/London day, in UTC. The end is the start of the following day, so a
// 23- or 25-hour clock-change day is handled by the zone database.
func UKDayRange(day string) (time.Time, time.Time, error) {
	start, err := UKDayStart(day)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	next := start.In(ukLocation).AddDate(0, 0, 1)
	return start, next.UTC(), nil
}

// UKDaySpan returns the half-open instant range covering the inclusive run of
// Europe/London days from..to, in UTC (§3.2).
func UKDaySpan(from, to string) (time.Time, time.Time, error) {
	start, err := UKDayStart(from)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	_, end, err := UKDayRange(to)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if !start.Before(end) {
		return time.Time{}, time.Time{}, fmt.Errorf("from %s is after to %s", from, to)
	}
	return start, end, nil
}

// UKToday returns today's Europe/London day.
func UKToday(now time.Time) string {
	return UKDay(now)
}

// UKClock returns the instant, in UTC, of a wall-clock time on a
// Europe/London day. It is how fixtures and tests say "09:00 UK on that day"
// without caring whether the day is in GMT or BST.
func UKClock(day string, hour, minute int) (time.Time, error) {
	start, err := UKDayStart(day)
	if err != nil {
		return time.Time{}, err
	}
	local := start.In(ukLocation)
	return time.Date(local.Year(), local.Month(), local.Day(), hour, minute, 0, 0, ukLocation).UTC(), nil
}

// UKDayAdd returns the Europe/London day n days after day.
func UKDayAdd(day string, n int) (string, error) {
	start, err := UKDayStart(day)
	if err != nil {
		return "", err
	}
	return start.In(ukLocation).AddDate(0, 0, n).Format(DayFormat), nil
}
