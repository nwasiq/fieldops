package api

import (
	"net/http"
	"testing"

	"github.com/nwasiq/fieldops/backend/internal/models"
)

func TestHealth(t *testing.T) {
	s := newTestServer(t)
	var data map[string]string
	s.decode(s.do(http.MethodGet, "/api/health", "", nil), http.StatusOK, &data)
	if data["status"] != "ok" {
		t.Errorf("health = %v", data)
	}
}

// rule: §1.2 — login returns a token and the user; the envelope never carries the password hash
func TestLogin_Success(t *testing.T) {
	s := newTestServer(t)
	s.createUser("ada.admin@example.com", models.RoleAdmin, "")
	var data struct {
		Token string         `json:"token"`
		User  map[string]any `json:"user"`
	}
	rec := s.do(http.MethodPost, "/api/auth/login", "", map[string]string{"email": "ada.admin@example.com", "password": testPassword})
	s.decode(rec, http.StatusOK, &data)
	if data.Token == "" || data.User["email"] != "ada.admin@example.com" || data.User["role"] != models.RoleAdmin {
		t.Errorf("login response = %+v", data)
	}
	for _, forbidden := range []string{"password_hash", "licence_number"} {
		if _, present := data.User[forbidden]; present {
			t.Errorf("login response exposes %s", forbidden)
		}
	}
	s.decode(s.do(http.MethodGet, "/api/sites", data.Token, nil), http.StatusOK, nil)
}

// rule: §1.2 — wrong password, deactivated and soft-deleted users all get 401
func TestLogin_Refusals(t *testing.T) {
	s := newTestServer(t)
	admin := s.createUser("ada.admin@example.com", models.RoleAdmin, "")
	inactive := s.createUser("ida.inactive@example.com", models.RoleTechnician, "")
	deleted := s.createUser("dana.departed@example.com", models.RoleTechnician, "")

	adminToken := s.login(admin.Email)
	s.decode(s.do(http.MethodPut, "/api/users/"+itoa(inactive.ID), adminToken, map[string]any{
		"email": inactive.Email, "first_name": "Ida", "last_name": "Inactive", "role": models.RoleTechnician, "is_active": false,
	}), http.StatusOK, nil)
	s.decode(s.do(http.MethodDelete, "/api/users/"+itoa(deleted.ID), adminToken, nil), http.StatusOK, nil)

	cases := map[string]map[string]string{
		"wrong password": {"email": admin.Email, "password": "nope"},
		"unknown email":  {"email": "nobody@example.com", "password": testPassword},
		"deactivated":    {"email": inactive.Email, "password": testPassword},
		"soft-deleted":   {"email": deleted.Email, "password": testPassword},
	}
	for name, body := range cases {
		if rec := s.do(http.MethodPost, "/api/auth/login", "", body); rec.Code != http.StatusUnauthorized {
			t.Errorf("%s: status %d, want 401: %s", name, rec.Code, rec.Body.String())
		}
	}
	s.expectError(s.do(http.MethodPost, "/api/auth/login", "", map[string]string{"email": admin.Email}), http.StatusBadRequest)
}

// rule: §1.2 — a token stops working the moment its user is deactivated or deleted
func TestAuth_TokenDiesWithUser(t *testing.T) {
	s := newTestServer(t)
	admin := s.createUser("ada.admin@example.com", models.RoleAdmin, "")
	tech := s.createUser("tom.field@example.com", models.RoleTechnician, "")
	techToken := s.login(tech.Email)
	s.decode(s.do(http.MethodGet, "/api/sites", techToken, nil), http.StatusOK, nil)

	s.decode(s.do(http.MethodDelete, "/api/users/"+itoa(tech.ID), s.login(admin.Email), nil), http.StatusOK, nil)
	s.expectError(s.do(http.MethodGet, "/api/sites", techToken, nil), http.StatusUnauthorized)

	s.expectError(s.do(http.MethodGet, "/api/sites", "", nil), http.StatusUnauthorized)
	s.expectError(s.do(http.MethodGet, "/api/sites", "not-a-token", nil), http.StatusUnauthorized)
}

// rule: §1.1 — admin writes users; dispatcher reads them; technicians get 403 on both
func TestRoleGates(t *testing.T) {
	s := newTestServer(t)
	admin := s.createUser("ada.admin@example.com", models.RoleAdmin, "")
	dispatcher := s.createUser("dev.dispatcher@example.com", models.RoleDispatcher, "")
	tech := s.createUser("tom.field@example.com", models.RoleTechnician, "")
	site := s.createSite("Depot")
	visit := s.createVisit(site.ID, &tech.ID, "2026-07-14T08:00:00Z", 60)
	tokens := map[string]string{
		models.RoleAdmin:      s.login(admin.Email),
		models.RoleDispatcher: s.login(dispatcher.Email),
		models.RoleTechnician: s.login(tech.Email),
	}
	newUser := map[string]any{"email": "new@example.com", "first_name": "N", "last_name": "U", "role": models.RoleTechnician, "password": "longenough"}
	siteBody := map[string]any{"name": "Depot 2", "address": "Somewhere"}
	visitBody := map[string]any{"site_id": site.ID, "scheduled_start": "2026-07-15T08:00:00Z", "scheduled_end": "2026-07-15T09:00:00Z"}

	cases := []struct {
		method, path string
		body         any
		allowed      map[string]int
	}{
		{http.MethodGet, "/api/users", nil, map[string]int{models.RoleAdmin: 200, models.RoleDispatcher: 200, models.RoleTechnician: 403}},
		{http.MethodPost, "/api/users", newUser, map[string]int{models.RoleAdmin: 201, models.RoleDispatcher: 403, models.RoleTechnician: 403}},
		{http.MethodGet, "/api/users/" + itoa(tech.ID), nil, map[string]int{models.RoleAdmin: 200, models.RoleDispatcher: 200, models.RoleTechnician: 403}},
		{http.MethodPost, "/api/sites", siteBody, map[string]int{models.RoleAdmin: 201, models.RoleDispatcher: 201, models.RoleTechnician: 403}},
		{http.MethodGet, "/api/sites", nil, map[string]int{models.RoleAdmin: 200, models.RoleDispatcher: 200, models.RoleTechnician: 200}},
		{http.MethodPost, "/api/visits", visitBody, map[string]int{models.RoleAdmin: 201, models.RoleDispatcher: 201, models.RoleTechnician: 403}},
		{http.MethodPut, "/api/visits/" + itoa(visit.ID), visitBody, map[string]int{models.RoleAdmin: 200, models.RoleDispatcher: 200, models.RoleTechnician: 403}},
		{http.MethodGet, "/api/reports/daily?date=2026-07-14", nil, map[string]int{models.RoleAdmin: 200, models.RoleDispatcher: 200, models.RoleTechnician: 403}},
	}
	for _, tc := range cases {
		for role, want := range tc.allowed {
			body := tc.body
			if tc.method == http.MethodPost && tc.path == "/api/users" {
				body = map[string]any{"email": role + "-new@example.com", "first_name": "N", "last_name": "U", "role": models.RoleTechnician, "password": "longenough"}
			}
			rec := s.do(tc.method, tc.path, tokens[role], body)
			if rec.Code != want {
				t.Errorf("%s %s as %s: status %d, want %d: %s", tc.method, tc.path, role, rec.Code, want, rec.Body.String())
			}
		}
	}
}
