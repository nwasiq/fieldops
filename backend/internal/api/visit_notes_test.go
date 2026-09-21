package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/nwasiq/fieldops/backend/internal/models"
	"github.com/nwasiq/fieldops/backend/internal/services"
)

// rule: §3.7 — notes are trimmed on write, blank notes clear the column, and the visit comes back with its associations
func TestVisitNotes_ServiceTrimsAndClears(t *testing.T) {
	s := newTestServer(t)
	tom := s.createUser("tom.field@example.com", models.RoleTechnician, "")
	site := s.createSite("Depot")
	visit := s.createVisit(site.ID, &tom.ID, "2026-07-14T08:00:00Z", 60)
	if visit.Notes != nil {
		t.Fatalf("new visit notes = %q, want none", *visit.Notes)
	}
	ctx := context.Background()

	updated, err := s.Visits.UpdateNotes(ctx, visit.ID, "  Key safe by the side door.\n")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Notes == nil || *updated.Notes != "Key safe by the side door." {
		t.Errorf("notes = %v, want the trimmed text", updated.Notes)
	}
	if updated.Site.Name != "Depot" || updated.Technician == nil || updated.Technician.ID != tom.ID || updated.Status != models.VisitStatusScheduled {
		t.Errorf("returned visit = %+v, want associations and status intact", updated)
	}
	var stored models.Visit
	if err := s.db.First(&stored, visit.ID).Error; err != nil {
		t.Fatal(err)
	}
	if stored.Notes == nil || *stored.Notes != "Key safe by the side door." {
		t.Errorf("stored notes = %v", stored.Notes)
	}

	cleared, err := s.Visits.UpdateNotes(ctx, visit.ID, "   \n")
	if err != nil {
		t.Fatal(err)
	}
	if cleared.Notes != nil {
		t.Errorf("blank write left notes = %q, want null", *cleared.Notes)
	}
	if _, err := s.Visits.UpdateNotes(ctx, 999, "x"); services.KindOf(err) != services.KindNotFound {
		t.Errorf("unknown visit: err = %v, want not found", err)
	}
}

// rule: §3.7 — the assigned technician, an admin or a dispatcher may write notes; another technician gets 403; closed visits stay writable
func TestVisitNotes_PutByAssignedTechnicianAndStaff(t *testing.T) {
	s := newTestServer(t)
	tom := s.createUser("tom.field@example.com", models.RoleTechnician, "")
	tara := s.createUser("tara.wrench@example.com", models.RoleTechnician, "")
	dispatcher := s.createUser("dev.dispatcher@example.com", models.RoleDispatcher, "")
	site := s.createSite("Depot")
	visit := s.createVisit(site.ID, &tom.ID, "2026-07-14T08:00:00Z", 60)
	path := "/api/visits/" + itoa(visit.ID) + "/notes"

	var got visitView
	s.decode(s.do(http.MethodPatch, path, s.login(tom.Email), map[string]any{"notes": "Meter cupboard is behind reception."}), http.StatusOK, &got)
	if got.Notes == nil || *got.Notes != "Meter cupboard is behind reception." || got.Status != models.VisitStatusScheduled {
		t.Errorf("technician write = %+v", got)
	}
	s.expectError(s.do(http.MethodPatch, path, s.login(tara.Email), map[string]any{"notes": "not mine"}), http.StatusForbidden)

	dispatcherToken := s.login(dispatcher.Email)
	s.decode(s.do(http.MethodPatch, path, dispatcherToken, map[string]any{"notes": "Access via the loading bay."}), http.StatusOK, &got)
	if got.Notes == nil || *got.Notes != "Access via the loading bay." {
		t.Errorf("dispatcher write = %+v", got)
	}
	s.decode(s.do(http.MethodGet, "/api/visits/"+itoa(visit.ID), dispatcherToken, nil), http.StatusOK, &got)
	if got.Notes == nil || *got.Notes != "Access via the loading bay." {
		t.Errorf("visit read back = %+v", got)
	}

	s.decode(s.do(http.MethodPost, "/api/visits/"+itoa(visit.ID)+"/cancel", dispatcherToken, nil), http.StatusOK, nil)
	s.decode(s.do(http.MethodPatch, path, dispatcherToken, map[string]any{"notes": "Customer rebooked for Friday."}), http.StatusOK, &got)
	if got.Status != models.VisitStatusCancelled || got.Notes == nil || *got.Notes != "Customer rebooked for Friday." {
		t.Errorf("write on a cancelled visit = %+v", got)
	}

	s.expectError(s.do(http.MethodPatch, path, dispatcherToken, map[string]any{}), http.StatusBadRequest)
	s.expectError(s.do(http.MethodPatch, "/api/visits/999/notes", dispatcherToken, map[string]any{"notes": "x"}), http.StatusNotFound)

	var list visitList
	s.decode(s.do(http.MethodGet, "/api/visits?from=2026-07-14&to=2026-07-14", dispatcherToken, nil), http.StatusOK, &list)
	if len(list.Visits) != 1 || list.Visits[0].Notes == nil {
		t.Errorf("list = %+v, want the visit with its notes", list.Visits)
	}
}
