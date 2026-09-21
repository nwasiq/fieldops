package middleware

import "github.com/gin-gonic/gin"

// Context keys set by the auth middleware and read by handlers and the audit
// and logging middleware.
const (
	ContextUserIDKey        = "user_id"
	ContextRoleKey          = "role"
	contextAuditResourceKey = "audit_resource_id"
)

// SetAuthenticatedUser attributes the rest of the request to a user. The auth
// middleware calls it for every bearer token; the login handler calls it on
// success so the request log and audit row name who logged in.
func SetAuthenticatedUser(c *gin.Context, id uint, role string) {
	c.Set(ContextUserIDKey, id)
	c.Set(ContextRoleKey, role)
}

// UserID returns the authenticated user's id, if any.
func UserID(c *gin.Context) (uint, bool) {
	value, ok := c.Get(ContextUserIDKey)
	if !ok {
		return 0, false
	}
	id, ok := value.(uint)
	return id, ok
}

// Role returns the authenticated user's role, or "" when unauthenticated.
func Role(c *gin.Context) string {
	value, ok := c.Get(ContextRoleKey)
	if !ok {
		return ""
	}
	role, _ := value.(string)
	return role
}

// SetAuditResource tells the audit middleware which resource a create (or any
// route without an :id) acted on, once the handler knows the id.
func SetAuditResource(c *gin.Context, id uint) {
	c.Set(contextAuditResourceKey, id)
}

func auditResource(c *gin.Context) (uint, bool) {
	value, ok := c.Get(contextAuditResourceKey)
	if !ok {
		return 0, false
	}
	id, ok := value.(uint)
	return id, ok
}
