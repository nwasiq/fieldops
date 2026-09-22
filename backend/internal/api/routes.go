package api

import (
	"github.com/gin-gonic/gin"

	"github.com/nwasiq/fieldops/backend/internal/middleware"
	"github.com/nwasiq/fieldops/backend/internal/models"
)

// RegisterRoutes declares every route, one per line. The audit-coverage lint
// reads this file, so keep each route on its own line as group.METHOD(path, …).
func RegisterRoutes(r *gin.Engine, h *Handlers, authenticate, audit gin.HandlerFunc) {
	adminOnly := middleware.RequireRole(models.RoleAdmin)
	scheduling := middleware.RequireRole(models.RoleAdmin, models.RoleDispatcher)

	api := r.Group("/api")
	api.GET("/health", h.Health)

	auth := api.Group("/auth")
	auth.POST("/login", audit, h.Login)

	// Every route registered below this line requires a bearer token and is
	// seen by the audit middleware. Routes above it are public.
	api.Use(authenticate, audit)

	api.GET("/users", scheduling, h.ListUsers)
	api.POST("/users", adminOnly, h.CreateUser)
	users := api.Group("/users")
	users.GET("/:id", scheduling, h.GetUser)
	users.PUT("/:id", adminOnly, h.UpdateUser)
	users.DELETE("/:id", adminOnly, h.DeleteUser)

	api.GET("/sites", h.ListSites)
	api.POST("/sites", scheduling, h.CreateSite)
	sites := api.Group("/sites")
	sites.GET("/:id", h.GetSite)
	sites.PUT("/:id", scheduling, h.UpdateSite)
	sites.DELETE("/:id", scheduling, h.DeleteSite)

	api.GET("/visits", h.ListVisits)
	api.POST("/visits", scheduling, h.CreateVisit)
	visits := api.Group("/visits")
	visits.GET("/:id", h.GetVisit)
	visits.PUT("/:id", scheduling, h.UpdateVisit)
	visits.PATCH("/:id/notes", h.UpdateVisitNotes)
	visits.POST("/:id/cancel", scheduling, h.CancelVisit)
	visits.POST("/:id/clock-in", h.ClockIn)
	visits.POST("/:id/clock-out", h.ClockOut)

	reports := api.Group("/reports")
	reports.GET("/daily", scheduling, h.DailyReport)
}
