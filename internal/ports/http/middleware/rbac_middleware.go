package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"

	rbacclient "github.com/example/user-service/internal/ports/rbac"
	res "github.com/example/user-service/pkg/http"
)

type RBACMiddleware struct {
	client rbacclient.Client
}

func NewRBACMiddleware(client rbacclient.Client) *RBACMiddleware {
	return &RBACMiddleware{client: client}
}

func (m *RBACMiddleware) RequireRole(role string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID, _ := c.Get("user_id").(string)
			if userID == "" {
				return res.ErrorJSON(c, http.StatusForbidden, "forbidden", "missing user", requestIDFromCtx(c), nil)
			}
			allowed := false
			if cached, ok := c.Get("role").(string); ok {
				allowed = cached == role
			} else if m.client != nil {
				ok, err := m.client.CheckRole(c.Request().Context(), userID, role)
				if err == nil {
					allowed = ok
				}
			}
			if !allowed {
				return res.ErrorJSON(c, http.StatusForbidden, "forbidden", "role required", requestIDFromCtx(c), nil)
			}
			return next(c)
		}
	}
}

func (m *RBACMiddleware) RequirePermission(permission string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID, _ := c.Get("user_id").(string)
			if userID == "" {
				return res.ErrorJSON(c, http.StatusForbidden, "forbidden", "missing user", requestIDFromCtx(c), nil)
			}
			allowed := false
			if cached, ok := c.Get("permissions").([]string); ok {
				for _, p := range cached {
					if p == permission {
						allowed = true
						break
					}
				}
			}
			if !allowed && m.client != nil {
				ok, err := m.client.CheckPermission(c.Request().Context(), userID, permission)
				if err == nil {
					allowed = ok
				}
			}
			if !allowed {
				return res.ErrorJSON(c, http.StatusForbidden, "forbidden", "permission required", requestIDFromCtx(c), nil)
			}
			return next(c)
		}
	}
}
