package middleware

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/nwasiq/fieldops/backend/internal/models"
	"github.com/nwasiq/fieldops/backend/internal/services"
)

// Auditor describes how one mutating route is audited: what to call the
// action, which resource it touches, and how to snapshot that resource.
type Auditor struct {
	Action       string
	ResourceType string
	// Fetch snapshots the resource by id. It returns nil when the resource does
	// not exist, which is the before-state of a create. Nil Fetch means the
	// route records no field diff (a login).
	Fetch func(ctx context.Context, id uint) (services.Snapshot, error)
}

// Registry maps METHOD + route pattern to its Auditor. Every mutating route
// must have an entry; the lint compares these lines against routes.go.
type Registry struct {
	auditors map[string]Auditor
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{auditors: map[string]Auditor{}}
}

// Register records how a route is audited. The path is the full Gin pattern.
func (r *Registry) Register(method, path string, auditor Auditor) {
	r.auditors[method+" "+path] = auditor
}

// Lookup returns the auditor for a matched route.
func (r *Registry) Lookup(method, path string) (Auditor, bool) {
	auditor, ok := r.auditors[method+" "+path]
	return auditor, ok
}

// Registered lists every "METHOD path" key, for tests and tooling.
func (r *Registry) Registered() []string {
	keys := make([]string, 0, len(r.auditors))
	for key := range r.auditors {
		keys = append(keys, key)
	}
	return keys
}

// RegisterAuditors declares the audit behaviour of every mutating route.
func RegisterAuditors(registry *Registry, audit *services.AuditService) {
	registry.Register("POST", "/api/auth/login", Auditor{Action: models.AuditActionLogin, ResourceType: models.ResourceUser})
	registry.Register("POST", "/api/users", Auditor{Action: models.AuditActionCreate, ResourceType: models.ResourceUser, Fetch: audit.SnapshotUser})
	registry.Register("PUT", "/api/users/:id", Auditor{Action: models.AuditActionUpdate, ResourceType: models.ResourceUser, Fetch: audit.SnapshotUser})
	registry.Register("DELETE", "/api/users/:id", Auditor{Action: models.AuditActionDelete, ResourceType: models.ResourceUser, Fetch: audit.SnapshotUser})
	registry.Register("POST", "/api/sites", Auditor{Action: models.AuditActionCreate, ResourceType: models.ResourceSite, Fetch: audit.SnapshotSite})
	registry.Register("PUT", "/api/sites/:id", Auditor{Action: models.AuditActionUpdate, ResourceType: models.ResourceSite, Fetch: audit.SnapshotSite})
	registry.Register("DELETE", "/api/sites/:id", Auditor{Action: models.AuditActionDelete, ResourceType: models.ResourceSite, Fetch: audit.SnapshotSite})
	registry.Register("POST", "/api/visits", Auditor{Action: models.AuditActionCreate, ResourceType: models.ResourceVisit, Fetch: audit.SnapshotVisit})
	registry.Register("PUT", "/api/visits/:id", Auditor{Action: models.AuditActionUpdate, ResourceType: models.ResourceVisit, Fetch: audit.SnapshotVisit})
	registry.Register("POST", "/api/visits/:id/cancel", Auditor{Action: models.AuditActionCancel, ResourceType: models.ResourceVisit, Fetch: audit.SnapshotVisit})
	registry.Register("POST", "/api/visits/:id/clock-in", Auditor{Action: models.AuditActionClockIn, ResourceType: models.ResourceVisit, Fetch: audit.SnapshotVisit})
	registry.Register("POST", "/api/visits/:id/clock-out", Auditor{Action: models.AuditActionClockOut, ResourceType: models.ResourceVisit, Fetch: audit.SnapshotVisit})
}

// Audit wraps mutating routes: it snapshots the resource before the handler,
// runs the handler, snapshots again and writes the diff as an audit row (§6.1).
// Nothing is written when the handler did not succeed, because nothing
// changed. A failure to write the row is logged, never surfaced: the change
// has already been committed and the response already decided.
func Audit(registry *Registry, audit *services.AuditService, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		auditor, ok := registry.Lookup(c.Request.Method, c.FullPath())
		if !ok {
			c.Next()
			return
		}
		ctx := c.Request.Context()
		resourceID, hasID := pathID(c)
		var before services.Snapshot
		if hasID && auditor.Fetch != nil {
			snapshot, err := auditor.Fetch(ctx, resourceID)
			if err != nil {
				AbortInternal(c, logger, err)
				return
			}
			before = snapshot
		}

		c.Next()

		status := c.Writer.Status()
		if status < 200 || status >= 300 {
			return
		}
		if !hasID {
			resourceID, hasID = auditResource(c)
		}
		actorID, hasActor := UserID(c)
		if !hasID || !hasActor {
			logger.Error("audit row skipped: no resource or actor",
				"method", c.Request.Method, "path", c.Request.URL.Path)
			return
		}
		var after services.Snapshot
		if auditor.Fetch != nil {
			snapshot, err := auditor.Fetch(ctx, resourceID)
			if err != nil {
				logger.Error("audit after-snapshot failed", "path", c.Request.URL.Path, "error", err.Error())
				return
			}
			after = snapshot
		}
		entry := services.Entry{
			ActorID:      actorID,
			Action:       auditor.Action,
			ResourceType: auditor.ResourceType,
			ResourceID:   resourceID,
			Changes:      services.Diff(before, after),
		}
		if err := audit.Record(ctx, entry); err != nil {
			logger.Error("audit row not written", "path", c.Request.URL.Path, "error", err.Error())
		}
	}
}

func pathID(c *gin.Context) (uint, bool) {
	raw := c.Param("id")
	if raw == "" {
		return 0, false
	}
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, false
	}
	return uint(id), true
}
