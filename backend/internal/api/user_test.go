package api

import (
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/nwasiq/fieldops/backend/internal/models"
	"github.com/nwasiq/fieldops/backend/internal/services"
)

func itoa(id uint) string { return strconv.FormatUint(uint64(id), 10) }

// rule: §1.3 — the licence number is ciphertext in the row and is decrypted only for an admin detail read
func TestUser_LicenceEncryptedAtRest(t *testing.T) {
	s := newTestServer(t)
	admin := s.createUser("ada.admin@example.com", models.RoleAdmin, "")
	dispatcher := s.createUser("dev.dispatcher@example.com", models.RoleDispatcher, "")
	adminToken, dispatcherToken := s.login(admin.Email), s.login(dispatcher.Email)

	var created map[string]any
	s.decode(s.do(http.MethodPost, "/api/users", adminToken, map[string]any{
		"email": "tom.field@example.com", "first_name": "Tom", "last_name": "Field",
		"role": models.RoleTechnician, "password": "longenough", "licence_number": "GS-284611",
	}), http.StatusCreated, &created)
	if _, present := created["licence_number"]; present {
		t.Error("create response echoes the licence number")
	}
	id := uint(created["id"].(float64))

	var stored string
	if err := s.db.Raw("SELECT licence_number FROM users WHERE id = ?", id).Scan(&stored).Error; err != nil {
		t.Fatal(err)
	}
	if !services.IsEncrypted(stored) || strings.Contains(stored, "GS-284611") {
		t.Fatalf("row holds %q, want ciphertext", stored)
	}

	var asAdmin map[string]any
	s.decode(s.do(http.MethodGet, "/api/users/"+itoa(id), adminToken, nil), http.StatusOK, &asAdmin)
	if asAdmin["licence_number"] != "GS-284611" {
		t.Errorf("admin sees licence %v, want GS-284611", asAdmin["licence_number"])
	}
	if _, present := asAdmin["password_hash"]; present {
		t.Error("detail exposes password_hash")
	}

	var asDispatcher map[string]any
	s.decode(s.do(http.MethodGet, "/api/users/"+itoa(id), dispatcherToken, nil), http.StatusOK, &asDispatcher)
	if _, present := asDispatcher["licence_number"]; present {
		t.Errorf("dispatcher sees licence_number: %v", asDispatcher)
	}

	var list struct {
		Users []map[string]any `json:"users"`
	}
	s.decode(s.do(http.MethodGet, "/api/users?role=technician", adminToken, nil), http.StatusOK, &list)
	if len(list.Users) != 1 {
		t.Fatalf("technician list = %v", list.Users)
	}
	if _, present := list.Users[0]["licence_number"]; present {
		t.Error("list exposes licence_number")
	}
}

// rule: §1.2 — a soft-deleted user leaves the list; filters and pagination are honoured
func TestUser_ListExcludesDeletedAndFilters(t *testing.T) {
	s := newTestServer(t)
	admin := s.createUser("ada.admin@example.com", models.RoleAdmin, "")
	gone := s.createUser("dana.departed@example.com", models.RoleTechnician, "")
	s.createUser("tom.field@example.com", models.RoleTechnician, "")
	token := s.login(admin.Email)

	s.decode(s.do(http.MethodDelete, "/api/users/"+itoa(gone.ID), token, nil), http.StatusOK, nil)
	s.expectError(s.do(http.MethodGet, "/api/users/"+itoa(gone.ID), token, nil), http.StatusNotFound)

	var list struct {
		Users    []map[string]any `json:"users"`
		Total    int              `json:"total"`
		PageSize int              `json:"page_size"`
	}
	s.decode(s.do(http.MethodGet, "/api/users?page_size=500", token, nil), http.StatusOK, &list)
	if list.Total != 2 || len(list.Users) != 2 || list.PageSize != services.MaxPageSize {
		t.Errorf("list = total %d, %d rows, page_size %d", list.Total, len(list.Users), list.PageSize)
	}
	for _, u := range list.Users {
		if u["email"] == gone.Email {
			t.Error("soft-deleted user is listed")
		}
	}
	s.decode(s.do(http.MethodGet, "/api/users?active=false", token, nil), http.StatusOK, &list)
	if list.Total != 0 {
		t.Errorf("active=false total = %d, want 0", list.Total)
	}
	s.expectError(s.do(http.MethodGet, "/api/users?active=maybe", token, nil), http.StatusBadRequest)
	s.expectError(s.do(http.MethodDelete, "/api/users/"+itoa(admin.ID), token, nil), http.StatusConflict)
}

func TestUser_Validation(t *testing.T) {
	s := newTestServer(t)
	admin := s.createUser("ada.admin@example.com", models.RoleAdmin, "")
	token := s.login(admin.Email)
	bad := []map[string]any{
		{"email": "x@example.com", "first_name": "A", "last_name": "B", "role": "superuser", "password": "longenough"},
		{"email": "x@example.com", "first_name": "A", "last_name": "B", "role": models.RoleAdmin, "password": "short"},
		{"email": "x@example.com", "first_name": "A", "last_name": "B", "role": models.RoleAdmin},
		{"email": "", "first_name": "A", "last_name": "B", "role": models.RoleAdmin, "password": "longenough"},
	}
	for _, body := range bad {
		s.expectError(s.do(http.MethodPost, "/api/users", token, body), http.StatusBadRequest)
	}
	dup := map[string]any{"email": "ADA.ADMIN@example.com", "first_name": "A", "last_name": "B", "role": models.RoleAdmin, "password": "longenough"}
	s.expectError(s.do(http.MethodPost, "/api/users", token, dup), http.StatusConflict)
	s.expectError(s.do(http.MethodGet, "/api/users/abc", token, nil), http.StatusBadRequest)
	s.expectError(s.do(http.MethodGet, "/api/users/999", token, nil), http.StatusNotFound)
}
