package handler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestPlaygroundSessionUnauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/playground/session", nil)
	(&PlaygroundHandler{svc: service.NewPlaygroundService(nil, nil, nil, nil, nil, nil, nil)}).Session(c)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestPlaygroundRequestCostUnauthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/playground/requests/abc", nil)
	c.Params = gin.Params{{Key: "request_id", Value: "abc"}}
	(&PlaygroundHandler{svc: service.NewPlaygroundService(nil, nil, nil, nil, nil, nil, nil)}).RequestCost(c)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRewritePlaygroundGatewayBodyAttachesDocuments(t *testing.T) {
	body := []byte(`{"model":"gpt-4o","group_id":9,"stream":true,"documents":[{"filename":"note.txt","text":"secret doc"}],"messages":[{"role":"user","content":"summarize this"}]}`)
	model, groupID, rewritten, err := rewritePlaygroundGatewayBody(body, service.SupportedModelKindText)
	require.NoError(t, err)
	require.Equal(t, "gpt-4o", model)
	require.NotNil(t, groupID)
	require.Equal(t, int64(9), *groupID)
	require.NotContains(t, string(rewritten), `"group_id"`)
	require.NotContains(t, string(rewritten), `"documents"`)
	require.Contains(t, string(rewritten), "secret doc")
	require.Contains(t, string(rewritten), "summarize this")
}

func TestRewritePlaygroundGatewayBodyRejectsTooManyDocuments(t *testing.T) {
	body := []byte(`{"model":"gpt-4o","messages":[],"documents":[{"filename":"a.txt","text":"1"},{"filename":"b.txt","text":"2"},{"filename":"c.txt","text":"3"},{"filename":"d.txt","text":"4"}]}`)
	_, _, _, err := rewritePlaygroundGatewayBody(body, service.SupportedModelKindText)
	require.Error(t, err)
	var playgroundErr *service.ErrPlayground
	require.ErrorAs(t, err, &playgroundErr)
	require.Equal(t, "DOCUMENT_LIMIT", playgroundErr.Code)
}

func TestPlaygroundGetImageTaskIsolatesByUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &imageTaskMemoryStore{}
	tasks := service.NewImageTaskServiceWithOptions(store, time.Hour, time.Minute)
	created, err := tasks.Create(context.Background(), service.ImageTaskOwner{UserID: 7, APIKeyID: 3})
	require.NoError(t, err)
	h := &PlaygroundHandler{tasks: tasks}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/playground/images/tasks/"+created.ID, nil)
	c.Params = gin.Params{{Key: "task_id", Value: created.ID}}
	c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 8})
	h.GetImageTask(c)
	require.Equal(t, http.StatusNotFound, w.Code)

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/playground/images/tasks/"+created.ID, nil)
	c.Params = gin.Params{{Key: "task_id", Value: created.ID}}
	c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 7})
	h.GetImageTask(c)
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), created.ID)
}

func TestPlaygroundDownloadImageFromB64(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &imageTaskMemoryStore{}
	tasks := service.NewImageTaskServiceWithOptions(store, time.Hour, time.Minute)
	created, err := tasks.Create(context.Background(), service.ImageTaskOwner{UserID: 7, APIKeyID: 3})
	require.NoError(t, err)
	png := []byte("\x89PNG\r\n\x1a\n")
	result, _ := json.Marshal(map[string]any{
		"data": []map[string]string{{"b64_json": base64.StdEncoding.EncodeToString(png)}},
	})
	require.NoError(t, tasks.Complete(context.Background(), created.ID, http.StatusOK, result))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/playground/images/tasks/"+created.ID+"/content/0", nil)
	c.Params = gin.Params{{Key: "task_id", Value: created.ID}, {Key: "index", Value: "0"}}
	c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 7})
	(&PlaygroundHandler{tasks: tasks}).DownloadImage(c)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "image/png", w.Header().Get("Content-Type"))
	require.Contains(t, w.Header().Get("Content-Disposition"), "attachment;")
	require.Equal(t, png, w.Body.Bytes())
}

func TestPlaygroundDownloadImageRejectsOtherUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := &imageTaskMemoryStore{}
	tasks := service.NewImageTaskServiceWithOptions(store, time.Hour, time.Minute)
	created, err := tasks.Create(context.Background(), service.ImageTaskOwner{UserID: 7, APIKeyID: 3})
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/playground/images/tasks/"+created.ID+"/content/0", nil)
	c.Params = gin.Params{{Key: "task_id", Value: created.ID}, {Key: "index", Value: "0"}}
	c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 99})
	(&PlaygroundHandler{tasks: tasks}).DownloadImage(c)
	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestPlaygroundSessionRequiresConfiguredService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/playground/session", nil)
	c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 7})
	(&PlaygroundHandler{}).Session(c)
	require.Equal(t, http.StatusInternalServerError, w.Code)
	require.Equal(t, "no-store", w.Header().Get("Cache-Control"))
	var payload response.Response
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &payload))
}

// imageTaskMemoryStore is defined in the service tests, so the handler tests
// use a local copy with the same behavior.
type imageTaskMemoryStore struct {
	task *service.ImageTaskRecord
}

func (s *imageTaskMemoryStore) Save(_ context.Context, task *service.ImageTaskRecord, _ time.Duration) error {
	copy := *task
	s.task = &copy
	return nil
}

func (s *imageTaskMemoryStore) Get(_ context.Context, _ string) (*service.ImageTaskRecord, error) {
	if s.task == nil {
		return nil, service.ErrImageTaskNotFound
	}
	copy := *s.task
	return &copy, nil
}
