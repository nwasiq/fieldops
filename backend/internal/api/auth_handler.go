package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nwasiq/fieldops/backend/internal/middleware"
)

type loginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type loginResponse struct {
	Token string   `json:"token"`
	User  userView `json:"user"`
}

// Login exchanges credentials for a token (§1.2).
func (h *Handlers) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "email and password are required")
		return
	}
	token, user, err := h.auth.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		fail(c, h.logger, err)
		return
	}
	middleware.SetAuthenticatedUser(c, user.ID, user.Role)
	middleware.SetAuditResource(c, user.ID)
	respond(c, http.StatusOK, loginResponse{Token: token, User: newUserView(user, false)})
}
