package api

import (
	"net/http"
	"testing"
	"time"

	"github.com/nwasiq/fieldops/backend/internal/models"
)

// rule: §5.1 — the daily report counts visits by day of scheduled start and sums clocked hours
func TestReport_Daily(t *testing.T) {
	s := newTestServer(t)
	admin := s.createUser("ada.admin@example.com", models.RoleAdmin, "")
	tom := s.createUser("tom.field@example.com", models.RoleTechnician, "")
	ida := s.createUser("ida.inactive@example.com", models.RoleTechnician, "")
	site := s.createSite("Depot")
	adminToken := s.login(admin.Email)

	tomVisit := s.createVisit(site.ID, &tom.ID, "2026-07-14T08:00:00Z", 120)
	idaVisit := s.createVisit(site.ID, &ida.ID, "2026-07-15T10:00:00Z", 60)
	cancelled := s.createVisit(site.ID, &tom.ID, "2026-07-15T11:00:00Z", 60)
	s.createVisit(site.ID, nil, "2026-07-15T12:00:00Z", 60)

	clock := func(visitID uint, email, in, out string) {
		token := s.login(email)
		at, _ := time.Parse(time.RFC3339, in)
		s.Visits.SetNow(func() time.Time { return at })
		s.decode(s.do(http.MethodPost, "/api/visits/"+itoa(visitID)+"/clock-in", token, nil), http.StatusOK, nil)
		at, _ = time.Parse(time.RFC3339, out)
		s.Visits.SetNow(func() time.Time { return at })
		s.decode(s.do(http.MethodPost, "/api/visits/"+itoa(visitID)+"/clock-out", token, nil), http.StatusOK, nil)
	}
	clock(tomVisit.ID, tom.Email, "2026-07-14T08:00:00Z", "2026-07-14T09:30:00Z")
	s.decode(s.do(http.MethodPost, "/api/visits/"+itoa(idaVisit.ID)+"/clock-in", s.login(ida.Email), nil), http.StatusOK, nil)
	s.decode(s.do(http.MethodPost, "/api/visits/"+itoa(cancelled.ID)+"/cancel", adminToken, nil), http.StatusOK, nil)

	s.decode(s.do(http.MethodPut, "/api/users/"+itoa(ida.ID), adminToken, map[string]any{
		"email": ida.Email, "first_name": "Ida", "last_name": "Inactive", "role": models.RoleTechnician, "is_active": false,
	}), http.StatusOK, nil)

	var report dailyReportView
	s.decode(s.do(http.MethodGet, "/api/reports/daily?date=2026-07-15", adminToken, nil), http.StatusOK, &report)
	if report.Date != "2026-07-15" || report.Scheduled != 3 || report.Completed != 0 || report.Cancelled != 1 {
		t.Errorf("15 July totals = %+v", report)
	}
	byName := map[string]technicianDayView{}
	for _, line := range report.Technicians {
		byName[line.Name] = line
	}
	if len(byName) != 2 {
		t.Fatalf("technicians = %+v, want Ida and Tom", report.Technicians)
	}
	if i := byName["Ida Inactive"]; i.Visits != 1 || i.Completed != 0 || i.HoursClocked != 0 {
		t.Errorf("deactivated Ida = %+v, want 1 visit, open clock-in counts no hours", i)
	}
	if tm := byName["Tom Field"]; tm.Visits != 1 || tm.Completed != 0 || tm.HoursClocked != 0 {
		t.Errorf("Tom on the 15th = %+v, want the cancelled visit only", tm)
	}

	s.decode(s.do(http.MethodGet, "/api/reports/daily?date=2026-07-14", adminToken, nil), http.StatusOK, &report)
	if report.Scheduled != 1 || report.Completed != 1 || len(report.Technicians) != 1 || report.Technicians[0].HoursClocked != 1.5 {
		t.Errorf("14 July = %+v, want Tom's single 1.5 h visit", report)
	}
	s.expectError(s.do(http.MethodGet, "/api/reports/daily", adminToken, nil), http.StatusBadRequest)
	s.expectError(s.do(http.MethodGet, "/api/reports/daily?date=15-07-2026", adminToken, nil), http.StatusBadRequest)
}
