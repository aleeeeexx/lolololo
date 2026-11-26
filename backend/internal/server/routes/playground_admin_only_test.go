package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newPlaygroundAdminOnlyRouter(role string, userID int64) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	v1 := r.Group("/api/v1")
	jwtAuth := middleware.JWTAuthMiddleware(func(c *gin.Context) {
		if userID > 0 {
			c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: userID})
		}
		if role != "" {
			c.Set(string(middleware.ContextKeyUserRole), role)
		}
		c.Next()
	})
	audit := middleware.AuditLogMiddleware(func(c *gin.Context) { c.Next() })
	h := &handler.Handlers{
		Playground: handler.NewPlaygroundHandler(
			service.NewPlaygroundService(nil, nil, nil, nil, nil, nil, nil),
			nil,
			nil,
			nil,
			nil,
			nil,
		),
	}
	RegisterPlaygroundRoutes(v1, h, jwtAuth, audit, nil, middleware.NewPanelRateLimiter(nil, nil), nil, nil)
	return r
}

func TestPlaygroundRoutesRejectNonAdminUsers(t *testing.T) {
	router := newPlaygroundAdminOnlyRouter(service.RoleUser, 7)

	for _, path := range []string{
		"/api/v1/playground/session",
		"/api/v1/playground/requests/abc",
		"/api/v1/playground/images/tasks/task-1",
	} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusForbidden, w.Code, path)

		var body map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body), path)
		require.Equal(t, "FORBIDDEN", body["code"], "response %s: %s", path, w.Body.String())
		require.Equal(t, "Admin access required", body["message"], path)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/playground/chat/completions", nil)
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/playground/images/generations", nil)
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusForbidden, w.Code)
}

func TestPlaygroundRoutesAllowAdministratorsPastAdminGate(t *testing.T) {
	router := newPlaygroundAdminOnlyRouter(service.RoleAdmin, 7)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/playground/session", nil)
	router.ServeHTTP(w, req)
	require.NotEqual(t, http.StatusForbidden, w.Code)
	require.NotEqual(t, http.StatusUnauthorized, w.Code)
}
