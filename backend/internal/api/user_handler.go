package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/nwasiq/fieldops/backend/internal/middleware"
	"github.com/nwasiq/fieldops/backend/internal/models"
	"github.com/nwasiq/fieldops/backend/internal/services"
)

type userRequest struct {
	Email         string  `json:"email"`
	FirstName     string  `json:"first_name"`
	LastName      string  `json:"last_name"`
	Role          string  `json:"role"`
	IsActive      *bool   `json:"is_active"`
	LicenceNumber *string `json:"licence_number"`
	Password      *string `json:"password"`
}

func (r userRequest) input() services.UserInput {
	return services.UserInput{
		Email:         r.Email,
		FirstName:     r.FirstName,
		LastName:      r.LastName,
		Role:          r.Role,
		IsActive:      r.IsActive,
		LicenceNumber: r.LicenceNumber,
		Password:      r.Password,
	}
}

// ListUsers pages through live users with optional role and active filters.
func (h *Handlers) ListUsers(c *gin.Context) {
	filter := services.UserFilter{Role: c.Query("role")}
	if raw := c.Query("active"); raw != "" {
		active, err := strconv.ParseBool(raw)
		if err != nil {
			respondError(c, http.StatusBadRequest, "active must be true or false")
			return
		}
		filter.Active = &active
	}
	result, err := h.users.List(c.Request.Context(), filter, pageFromQuery(c))
	if err != nil {
		fail(c, h.logger, err)
		return
	}
	users := make([]userView, 0, len(result.Items))
	for i := range result.Items {
		users = append(users, newUserView(&result.Items[i], false))
	}
	respond(c, http.StatusOK, userListView{Users: users, pageView: newPageView(result)})
}

// GetUser returns one user; admins also see the decrypted licence number (§1.3).
func (h *Handlers) GetUser(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	withLicence := middleware.Role(c) == models.RoleAdmin
	user, err := h.users.Get(c.Request.Context(), id, withLicence)
	if err != nil {
		fail(c, h.logger, err)
		return
	}
	respond(c, http.StatusOK, newUserView(user, withLicence))
}

// CreateUser adds a user.
func (h *Handlers) CreateUser(c *gin.Context) {
	var req userRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	user, err := h.users.Create(c.Request.Context(), req.input())
	if err != nil {
		fail(c, h.logger, err)
		return
	}
	middleware.SetAuditResource(c, user.ID)
	respond(c, http.StatusCreated, newUserView(user, false))
}

// UpdateUser replaces a user's editable fields.
func (h *Handlers) UpdateUser(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	var req userRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "invalid request body")
		return
	}
	user, err := h.users.Update(c.Request.Context(), id, req.input())
	if err != nil {
		fail(c, h.logger, err)
		return
	}
	respond(c, http.StatusOK, newUserView(user, false))
}

// DeleteUser soft-deletes a user.
func (h *Handlers) DeleteUser(c *gin.Context) {
	id, err := pathID(c)
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.users.Delete(c.Request.Context(), id, actorFrom(c)); err != nil {
		fail(c, h.logger, err)
		return
	}
	respond(c, http.StatusOK, gin.H{"id": id})
}
