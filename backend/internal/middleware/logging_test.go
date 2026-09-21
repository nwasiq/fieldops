package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/nwasiq/fieldops/backend/internal/models"
)

func TestRequestLogger_OneJSONLinePerRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var out bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&out, nil))
	engine := gin.New()
	engine.Use(RequestLogger(logger))
	engine.GET("/api/thing", func(c *gin.Context) {
		SetAuthenticatedUser(c, 42, models.RoleAdmin)
		c.Status(http.StatusNoContent)
	})

	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/thing", nil))

	if rec.Header().Get("X-Request-ID") == "" {
		t.Error("no X-Request-ID header echoed")
	}
	lines := bytes.Split(bytes.TrimSpace(out.Bytes()), []byte("\n"))
	if len(lines) != 1 {
		t.Fatalf("logged %d lines, want 1: %s", len(lines), out.String())
	}
	var line map[string]any
	if err := json.Unmarshal(lines[0], &line); err != nil {
		t.Fatalf("log line is not JSON: %v: %s", err, lines[0])
	}
	for _, key := range []string{"request_id", "method", "path", "status", "duration_ms", "user_id"} {
		if _, ok := line[key]; !ok {
			t.Errorf("log line lacks %s: %v", key, line)
		}
	}
	if line["status"] != float64(http.StatusNoContent) || line["user_id"] != float64(42) || line["path"] != "/api/thing" {
		t.Errorf("log line = %v", line)
	}
}

func TestRecover_WrapsPanicAsEnvelope500(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var out bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&out, nil))
	engine := gin.New()
	engine.Use(Recover(logger))
	engine.GET("/boom", func(c *gin.Context) { panic("kaboom") })

	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/boom", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status %d", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil || body["success"] != false || body["error"] == "" {
		t.Errorf("body = %s", rec.Body.String())
	}
	if !bytes.Contains(out.Bytes(), []byte(`"path":"/boom"`)) {
		t.Errorf("panic log lacks the path: %s", out.String())
	}
}

// rule: §1.1 — the role gate refuses any role not listed
func TestRequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/admin", func(c *gin.Context) { SetAuthenticatedUser(c, 1, models.RoleTechnician); c.Next() }, RequireRole(models.RoleAdmin), func(c *gin.Context) { c.Status(http.StatusOK) })
	engine.GET("/either", func(c *gin.Context) { SetAuthenticatedUser(c, 1, models.RoleDispatcher); c.Next() }, RequireRole(models.RoleAdmin, models.RoleDispatcher), func(c *gin.Context) { c.Status(http.StatusOK) })

	for path, want := range map[string]int{"/admin": http.StatusForbidden, "/either": http.StatusOK} {
		rec := httptest.NewRecorder()
		engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != want {
			t.Errorf("%s: status %d, want %d", path, rec.Code, want)
		}
	}
}
