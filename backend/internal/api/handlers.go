package api

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nwasiq/fieldops/backend/internal/middleware"
	"github.com/nwasiq/fieldops/backend/internal/services"
)

// Handlers holds every HTTP handler and the services they delegate to.
type Handlers struct {
	logger  *slog.Logger
	auth    *services.AuthService
	users   *services.UserService
	sites   *services.SiteService
	visits  *services.VisitService
	reports *services.ReportService
}

// NewHandlers wires handlers to services.
func NewHandlers(logger *slog.Logger, auth *services.AuthService, users *services.UserService, sites *services.SiteService, visits *services.VisitService, reports *services.ReportService) *Handlers {
	return &Handlers{logger: logger, auth: auth, users: users, sites: sites, visits: visits, reports: reports}
}

// Health answers the unauthenticated liveness probe.
func (h *Handlers) Health(c *gin.Context) {
	respond(c, http.StatusOK, gin.H{"status": "ok"})
}

func actorFrom(c *gin.Context) services.Actor {
	id, _ := middleware.UserID(c)
	return services.Actor{ID: id, Role: middleware.Role(c)}
}
