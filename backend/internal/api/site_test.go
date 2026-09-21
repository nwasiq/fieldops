package api

import (
	"net/http"
	"testing"

	"github.com/nwasiq/fieldops/backend/internal/models"
)

// rule: §2.1 — a deleted site's past visits stay readable with the site's name
func TestSite_DeleteKeepsVisitReadable(t *testing.T) {
	s := newTestServer(t)
	dispatcher := s.createUser("dev.dispatcher@example.com", models.RoleDispatcher, "")
	token := s.login(dispatcher.Email)
	site := s.createSite("Old Depot")
	visit := s.createVisit(site.ID, nil, "2026-07-14T08:00:00Z", 60)

	s.decode(s.do(http.MethodDelete, "/api/sites/"+itoa(site.ID), token, nil), http.StatusOK, nil)
	s.expectError(s.do(http.MethodGet, "/api/sites/"+itoa(site.ID), token, nil), http.StatusNotFound)

	var got visitView
	s.decode(s.do(http.MethodGet, "/api/visits/"+itoa(visit.ID), token, nil), http.StatusOK, &got)
	if got.Site.ID != site.ID || got.Site.Name != "Old Depot" {
		t.Errorf("visit site = %+v, want the deleted site's id and name", got.Site)
	}
	s.expectError(s.do(http.MethodPost, "/api/visits", token, map[string]any{
		"site_id": site.ID, "scheduled_start": "2026-07-15T08:00:00Z", "scheduled_end": "2026-07-15T09:00:00Z",
	}), http.StatusBadRequest)
}

func TestSite_CRUD(t *testing.T) {
	s := newTestServer(t)
	dispatcher := s.createUser("dev.dispatcher@example.com", models.RoleDispatcher, "")
	token := s.login(dispatcher.Email)

	var created siteView
	s.decode(s.do(http.MethodPost, "/api/sites", token, map[string]any{"name": "Depot", "address": "1 Road", "contact_phone": "0100"}), http.StatusCreated, &created)
	s.expectError(s.do(http.MethodPost, "/api/sites", token, map[string]any{"name": "", "address": "1 Road"}), http.StatusBadRequest)

	var updated siteView
	s.decode(s.do(http.MethodPut, "/api/sites/"+itoa(created.ID), token, map[string]any{"name": "Depot North", "address": "1 Road"}), http.StatusOK, &updated)
	if updated.Name != "Depot North" || updated.ContactPhone != "" {
		t.Errorf("updated = %+v", updated)
	}

	var list struct {
		Sites      []siteView `json:"sites"`
		Total      int        `json:"total"`
		Page       int        `json:"page"`
		PageSize   int        `json:"page_size"`
		TotalPages int        `json:"total_pages"`
	}
	s.decode(s.do(http.MethodGet, "/api/sites", token, nil), http.StatusOK, &list)
	if list.Total != 1 || list.Page != 1 || list.PageSize != 20 || list.TotalPages != 1 {
		t.Errorf("list = %+v", list)
	}
	s.expectError(s.do(http.MethodPut, "/api/sites/999", token, map[string]any{"name": "X", "address": "Y"}), http.StatusNotFound)
	s.expectError(s.do(http.MethodDelete, "/api/sites/999", token, nil), http.StatusNotFound)
}
