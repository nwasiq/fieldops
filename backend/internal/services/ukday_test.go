package services

import (
	"testing"
	"time"
)

// rule: §3.3 — a visit belongs to the Europe/London day of its scheduled start, whatever the host zone
func TestUKDay_BSTBoundary(t *testing.T) {
	cases := []struct {
		instant string
		want    string
	}{
		{"2026-07-14T23:30:00Z", "2026-07-15"}, // BST: 00:30 next day in London
		{"2026-07-15T00:15:00Z", "2026-07-15"}, // BST: 01:15 same day
		{"2026-01-14T23:30:00Z", "2026-01-14"}, // GMT: still the 14th
		{"2026-01-15T00:15:00Z", "2026-01-15"},
	}
	for _, tc := range cases {
		instant, err := time.Parse(time.RFC3339, tc.instant)
		if err != nil {
			t.Fatal(err)
		}
		// Re-express in a distant zone to prove the input's location is irrelevant.
		tokyo := instant.In(time.FixedZone("JST", 9*3600))
		if got := UKDay(tokyo); got != tc.want {
			t.Errorf("UKDay(%s) = %s, want %s", tc.instant, got, tc.want)
		}
	}
}

// rule: §3.3 — UK day ranges follow the zone database through clock changes
func TestUKDayRange_ClockChanges(t *testing.T) {
	cases := []struct {
		day   string
		hours float64
		start string
	}{
		{"2026-03-29", 23, "2026-03-29T00:00:00Z"}, // clocks go forward
		{"2026-10-25", 25, "2026-10-24T23:00:00Z"}, // clocks go back
		{"2026-07-14", 24, "2026-07-13T23:00:00Z"}, // BST
		{"2026-01-14", 24, "2026-01-14T00:00:00Z"}, // GMT
	}
	for _, tc := range cases {
		start, end, err := UKDayRange(tc.day)
		if err != nil {
			t.Fatalf("UKDayRange(%s): %v", tc.day, err)
		}
		if got := start.Format(time.RFC3339); got != tc.start {
			t.Errorf("UKDayRange(%s) start = %s, want %s", tc.day, got, tc.start)
		}
		if got := end.Sub(start).Hours(); got != tc.hours {
			t.Errorf("UKDayRange(%s) spans %v hours, want %v", tc.day, got, tc.hours)
		}
		if start.Location() != time.UTC {
			t.Errorf("UKDayRange(%s) start is in %s, want UTC", tc.day, start.Location())
		}
	}
}

func TestUKDayStart_RejectsBadInput(t *testing.T) {
	for _, bad := range []string{"", "2026-1-5", "14/07/2026", "2026-07-14T00:00:00Z", "yesterday"} {
		if _, err := UKDayStart(bad); err == nil {
			t.Errorf("UKDayStart(%q) accepted, want error", bad)
		}
	}
}

// rule: §3.2 — from..to is inclusive and from must not be after to
func TestUKDaySpan(t *testing.T) {
	start, end, err := UKDaySpan("2026-07-14", "2026-07-16")
	if err != nil {
		t.Fatal(err)
	}
	if got := end.Sub(start).Hours(); got != 72 {
		t.Errorf("three-day span is %v hours, want 72", got)
	}
	if _, _, err := UKDaySpan("2026-07-16", "2026-07-14"); err == nil {
		t.Error("reversed span accepted, want error")
	}
}

func TestUKClock(t *testing.T) {
	got, err := UKClock("2026-07-14", 9, 0)
	if err != nil {
		t.Fatal(err)
	}
	if want := "2026-07-14T08:00:00Z"; got.Format(time.RFC3339) != want {
		t.Errorf("09:00 BST = %s, want %s", got.Format(time.RFC3339), want)
	}
	next, err := UKDayAdd("2026-10-31", 1)
	if err != nil {
		t.Fatal(err)
	}
	if next != "2026-11-01" {
		t.Errorf("UKDayAdd = %s, want 2026-11-01", next)
	}
}
