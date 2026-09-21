package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/nwasiq/fieldops/backend/internal/database"
	"github.com/nwasiq/fieldops/backend/internal/models"
	"github.com/nwasiq/fieldops/backend/internal/services"
)

const (
	testKey      = "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY="
	testSecret   = "test-secret-that-is-at-least-32-chars"
	testPassword = "fieldops-test"
)

type testServer struct {
	*Server
	db *gorm.DB
	t  *testing.T
}

func newTestServer(t *testing.T) *testServer {
	t.Helper()
	gin.SetMode(gin.TestMode)
	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	db, err := database.ConnectSQLite(name)
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatal(err)
	}
	cipher, err := services.NewCipher(testKey)
	if err != nil {
		t.Fatal(err)
	}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	server := NewServer(db, Config{JWTSecret: testSecret, Cipher: cipher, Logger: logger, HashCost: bcrypt.MinCost})
	return &testServer{Server: server, db: db, t: t}
}

func (s *testServer) createUser(email, role, licence string) *models.User {
	s.t.Helper()
	password := testPassword
	parts := strings.SplitN(strings.SplitN(email, "@", 2)[0], ".", 2)
	first, last := strings.Title(parts[0]), "User"
	if len(parts) == 2 {
		last = strings.Title(parts[1])
	}
	user, err := s.Users.Create(context.Background(), services.UserInput{
		Email: email, FirstName: first, LastName: last, Role: role,
		LicenceNumber: &licence, Password: &password,
	})
	if err != nil {
		s.t.Fatalf("create user %s: %v", email, err)
	}
	return user
}

func (s *testServer) createSite(name string) *models.Site {
	s.t.Helper()
	site, err := s.Sites.Create(context.Background(), services.SiteInput{Name: name, Address: name + " address", ContactPhone: "0100"})
	if err != nil {
		s.t.Fatalf("create site %s: %v", name, err)
	}
	return site
}

func (s *testServer) createVisit(siteID uint, techID *uint, start string, minutes int) *models.Visit {
	s.t.Helper()
	at, err := time.Parse(time.RFC3339, start)
	if err != nil {
		s.t.Fatal(err)
	}
	visit, err := s.Visits.Create(context.Background(), services.VisitInput{
		SiteID: siteID, TechnicianID: techID, ScheduledStart: at, ScheduledEnd: at.Add(time.Duration(minutes) * time.Minute),
	})
	if err != nil {
		s.t.Fatalf("create visit: %v", err)
	}
	return visit
}

func (s *testServer) login(email string) string {
	s.t.Helper()
	token, _, err := s.Auth.Login(context.Background(), email, testPassword)
	if err != nil {
		s.t.Fatalf("login %s: %v", email, err)
	}
	return token
}

func (s *testServer) do(method, path, token string, body any) *httptest.ResponseRecorder {
	s.t.Helper()
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			s.t.Fatal(err)
		}
		reader = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	s.Engine.ServeHTTP(rec, req)
	return rec
}

type envelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   string          `json:"error"`
}

func (s *testServer) decode(rec *httptest.ResponseRecorder, wantStatus int, out any) envelope {
	s.t.Helper()
	if rec.Code != wantStatus {
		s.t.Fatalf("status %d, want %d: %s", rec.Code, wantStatus, rec.Body.String())
	}
	var env envelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		s.t.Fatalf("decode envelope: %v: %s", err, rec.Body.String())
	}
	if env.Success != (wantStatus < 400) {
		s.t.Fatalf("success=%v for status %d: %s", env.Success, wantStatus, rec.Body.String())
	}
	if out != nil && env.Data != nil {
		if err := json.Unmarshal(env.Data, out); err != nil {
			s.t.Fatalf("decode data: %v: %s", err, env.Data)
		}
	}
	return env
}

func (s *testServer) expectError(rec *httptest.ResponseRecorder, wantStatus int) string {
	s.t.Helper()
	env := s.decode(rec, wantStatus, nil)
	if env.Error == "" {
		s.t.Fatalf("status %d carried no error message: %s", wantStatus, rec.Body.String())
	}
	return env.Error
}

func (s *testServer) auditRows(resourceType string, resourceID uint) []models.AuditLog {
	s.t.Helper()
	var rows []models.AuditLog
	err := s.db.Where("resource_type = ? AND resource_id = ?", resourceType, resourceID).Order("id").Find(&rows).Error
	if err != nil {
		s.t.Fatal(err)
	}
	return rows
}

func changesOf(t *testing.T, row models.AuditLog) map[string]services.Change {
	t.Helper()
	var changes map[string]services.Change
	if err := json.Unmarshal([]byte(row.Changes), &changes); err != nil {
		t.Fatalf("decode changes %q: %v", row.Changes, err)
	}
	return changes
}

func ptr[T any](v T) *T { return &v }
