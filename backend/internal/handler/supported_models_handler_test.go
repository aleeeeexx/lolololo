package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type stubSupportedModelsLister struct {
	models []service.SupportedModelItem
	err    error
	userID int64
	calls  int
}

func (s *stubSupportedModelsLister) ListForUser(_ context.Context, userID int64) ([]service.SupportedModelItem, error) {
	s.calls++
	s.userID = userID
	if s.err != nil {
		return nil, s.err
	}
	return append([]service.SupportedModelItem(nil), s.models...), nil
}

func performSupportedModels(h *SupportedModelsHandler, userID int64) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/supported-models", nil)
	if userID > 0 {
		c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: userID})
	}
	h.List(c)
	return w
}

func TestSupportedModelsList_Unauthenticated401(t *testing.T) {
	lister := &stubSupportedModelsLister{models: []service.SupportedModelItem{{Name: "should-not-appear"}}}
	w := performSupportedModels(&SupportedModelsHandler{svc: lister}, 0)
	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.Zero(t, lister.calls)
	require.NotContains(t, w.Body.String(), "should-not-appear")
}

func TestSupportedModelsList_ReturnsSanitizedModels(t *testing.T) {
	lister := &stubSupportedModelsLister{models: []service.SupportedModelItem{
		{Name: "gpt-4o", Platforms: []string{service.PlatformOpenAI}},
		{Name: "claude-sonnet-4", Platforms: []string{service.PlatformAnthropic}},
	}}
	w := performSupportedModels(&SupportedModelsHandler{svc: lister}, 42)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, int64(42), lister.userID)

	var payload response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &payload))
	require.Equal(t, 0, payload.Code)

	raw, err := json.Marshal(payload.Data)
	require.NoError(t, err)
	require.JSONEq(t, `[
		{"name":"gpt-4o","platforms":["openai"]},
		{"name":"claude-sonnet-4","platforms":["anthropic"]}
	]`, string(raw))

	body := strings.ToLower(w.Body.String())
	for _, leaked := range []string{
		"account_id", "accountid", "api_key", "apikey", "secret", "password",
		"credential", "token", "model_mapping", "group_id",
	} {
		require.NotContains(t, body, leaked)
	}
}

func TestSupportedModelsList_EmptyWhenNoAdvertisedModels(t *testing.T) {
	w := performSupportedModels(&SupportedModelsHandler{svc: &stubSupportedModelsLister{}}, 7)
	require.Equal(t, http.StatusOK, w.Code)

	var payload response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &payload))
	raw, err := json.Marshal(payload.Data)
	require.NoError(t, err)
	require.JSONEq(t, `[]`, string(raw))
}

func TestSupportedModelsList_InternalError(t *testing.T) {
	w := performSupportedModels(&SupportedModelsHandler{svc: &stubSupportedModelsLister{err: errors.New("boom")}}, 7)
	require.Equal(t, http.StatusInternalServerError, w.Code)
	require.NotContains(t, w.Body.String(), "boom")
}
