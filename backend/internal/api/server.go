package api

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nwasiq/fieldops/backend/internal/middleware"
	"github.com/nwasiq/fieldops/backend/internal/services"
)

// Config is what the server needs beyond a database.
type Config struct {
	JWTSecret string
	Cipher    *services.Cipher
	Logger    *slog.Logger
	// HashCost overrides the bcrypt cost when non-zero; tests lower it.
	HashCost int
}

// Server is the wired application: every service, handler and middleware.
type Server struct {
	Engine   *gin.Engine
	Auth     *services.AuthService
	Users    *services.UserService
	Sites    *services.SiteService
	Visits   *services.VisitService
	Reports  *services.ReportService
	Audit    *services.AuditService
	Registry *middleware.Registry
}

// NewServer builds the router with every dependency wired. main and the
// tests share this so the two can never drift apart.
func NewServer(db *gorm.DB, cfg Config) *Server {
	auth := services.NewAuthService(db, cfg.JWTSecret)
	if cfg.HashCost > 0 {
		auth.SetHashCost(cfg.HashCost)
	}
	users := services.NewUserService(db, cfg.Cipher, auth)
	sites := services.NewSiteService(db)
	visits := services.NewVisitService(db)
	reports := services.NewReportService(db)
	audit := services.NewAuditService(db, cfg.Cipher)

	registry := middleware.NewRegistry()
	middleware.RegisterAuditors(registry, audit)

	engine := gin.New()
	engine.Use(middleware.RequestLogger(cfg.Logger), middleware.Recover(cfg.Logger))
	handlers := NewHandlers(cfg.Logger, auth, users, sites, visits, reports)
	RegisterRoutes(engine, handlers,
		middleware.Authenticate(auth, cfg.Logger),
		middleware.Audit(registry, audit, cfg.Logger),
	)

	return &Server{
		Engine:   engine,
		Auth:     auth,
		Users:    users,
		Sites:    sites,
		Visits:   visits,
		Reports:  reports,
		Audit:    audit,
		Registry: registry,
	}
}
