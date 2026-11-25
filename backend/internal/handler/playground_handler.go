package handler

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	pkghttputil "github.com/Wei-Shaw/sub2api/internal/pkg/httputil"
	"github.com/Wei-Shaw/sub2api/internal/pkg/requestmodel"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type PlaygroundHandler struct {
	svc     *service.PlaygroundService
	gateway *GatewayHandler
	openAI  *OpenAIGatewayHandler
	async   *AsyncImageHandler
	tasks   *service.ImageTaskService
	cfg     *config.Config
}

func NewPlaygroundHandler(
	svc *service.PlaygroundService,
	gateway *GatewayHandler,
	openAI *OpenAIGatewayHandler,
	async *AsyncImageHandler,
	tasks *service.ImageTaskService,
	cfg *config.Config,
) *PlaygroundHandler {
	return &PlaygroundHandler{
		svc:     svc,
		gateway: gateway,
		openAI:  openAI,
		async:   async,
		tasks:   tasks,
		cfg:     cfg,
	}
}

func NewPlaygroundHandlerFromDeps(
	apiKeyService *service.APIKeyService,
	subscriptionService *service.SubscriptionService,
	userService *service.UserService,
	usageService *service.UsageService,
	gatewayService *service.GatewayService,
	gatewayHandler *GatewayHandler,
	openAIGatewayHandler *OpenAIGatewayHandler,
	asyncImageHandler *AsyncImageHandler,
	imageTaskService *service.ImageTaskService,
	cfg *config.Config,
) *PlaygroundHandler {
	return NewPlaygroundHandler(
		service.NewPlaygroundService(apiKeyService, subscriptionService, userService, usageService, gatewayService, imageTaskService, cfg),
		gatewayHandler,
		openAIGatewayHandler,
		asyncImageHandler,
		imageTaskService,
		cfg,
	)
}

// Session GET /api/v1/playground/session
func (h *PlaygroundHandler) Session(c *gin.Context) {
	userID, ok := playgroundUserID(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	if h == nil || h.svc == nil {
		response.InternalError(c, "Playground is not configured")
		return
	}
	session, err := h.svc.Session(c.Request.Context(), userID)
	if err != nil {
		abortPlaygroundError(c, err)
		return
	}
	response.Success(c, session)
}

// RequestCost GET /api/v1/playground/requests/:request_id
func (h *PlaygroundHandler) RequestCost(c *gin.Context) {
	userID, ok := playgroundUserID(c)
	if !ok {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	if h == nil || h.svc == nil {
		response.InternalError(c, "Playground is not configured")
		return
	}
	cost, err := h.svc.RequestCost(c.Request.Context(), userID, c.Param("request_id"))
	if err != nil {
		abortPlaygroundError(c, err)
		return
	}
	response.Success(c, cost)
}

// PrepareText is JWT middleware that authorizes a chat request and rewrites
// the path so existing ChatCompletions handlers and billing can run.
func (h *PlaygroundHandler) PrepareText() gin.HandlerFunc {
	return h.prepare(service.SupportedModelKindText, "/v1/chat/completions")
}

// PrepareImage authorizes an async image submit and rewrites the path to the
// existing /v1/images/generations/async gateway route.
func (h *PlaygroundHandler) PrepareImage() gin.HandlerFunc {
	return h.prepare(service.SupportedModelKindImage, "/v1/images/generations/async")
}

func (h *PlaygroundHandler) prepare(kind, gatewayPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := playgroundUserID(c)
		if !ok {
			abortPlaygroundError(c, servicePlaygroundUnauthorized())
			return
		}
		if h == nil || h.svc == nil {
			response.InternalError(c, "Playground is not configured")
			c.Abort()
			return
		}
		body, err := pkghttputil.ReadRequestBodyWithPrealloc(c.Request)
		if err != nil {
			if middleware2.RejectRequestBodyMemoryOverload(c, err) {
				return
			}
			middleware2.AbortWithError(c, http.StatusBadRequest, "INVALID_REQUEST", "Failed to read request body")
			return
		}
		model, groupID, rewritten, err := rewritePlaygroundGatewayBody(body, kind)
		if err != nil {
			abortPlaygroundError(c, err)
			return
		}
		prepared, err := h.svc.Prepare(c.Request.Context(), userID, model, kind, groupID)
		if err != nil {
			abortPlaygroundError(c, err)
			return
		}
		requestmodel.ResetRequestBody(c.Request, rewritten)
		if c.Request.URL != nil {
			c.Request.URL.Path = gatewayPath
		}
		middleware2.SetGatewayAuthContext(c, prepared.APIKey, prepared.Subscription)
		c.Next()
	}
}

// GetImageTask GET /api/v1/playground/images/tasks/:task_id
func (h *PlaygroundHandler) GetImageTask(c *gin.Context) {
	userID, ok := playgroundUserID(c)
	if !ok {
		abortPlaygroundError(c, servicePlaygroundUnauthorized())
		return
	}
	if h == nil || h.tasks == nil || !h.tasks.Pollable() {
		imageTaskJSONError(c, http.StatusNotFound, "not_found_error", "async image tasks are not enabled")
		return
	}
	task, err := h.tasks.GetForUser(c.Request.Context(), userID, c.Param("task_id"))
	if err != nil {
		imageTaskError(c, err)
		return
	}
	c.Header("Cache-Control", "no-store")
	if task.Status == service.ImageTaskStatusProcessing {
		c.Header("Retry-After", "3")
	}
	c.JSON(http.StatusOK, task)
}

// DownloadImage GET /api/v1/playground/images/tasks/:task_id/content/:index
func (h *PlaygroundHandler) DownloadImage(c *gin.Context) {
	userID, ok := playgroundUserID(c)
	if !ok {
		abortPlaygroundError(c, servicePlaygroundUnauthorized())
		return
	}
	if h == nil || h.tasks == nil || !h.tasks.Pollable() {
		imageTaskJSONError(c, http.StatusNotFound, "not_found_error", "async image tasks are not enabled")
		return
	}
	index, err := strconv.Atoi(c.Param("index"))
	if err != nil || index < 0 {
		imageTaskJSONError(c, http.StatusBadRequest, "invalid_request_error", "invalid image index")
		return
	}
	task, err := h.tasks.GetForUser(c.Request.Context(), userID, c.Param("task_id"))
	if err != nil {
		imageTaskError(c, err)
		return
	}
	if task.Status != service.ImageTaskStatusCompleted {
		imageTaskJSONError(c, http.StatusConflict, "task_not_ready", "image task is not completed")
		return
	}
	item, err := playgroundImageOutput(task.Result, index)
	if err != nil {
		abortPlaygroundError(c, err)
		return
	}
	filename := "playground-" + strings.TrimSpace(task.ID) + "-" + strconv.Itoa(index)
	if item.B64JSON != "" {
		data, decodeErr := decodeImageBase64(item.B64JSON)
		if decodeErr != nil {
			imageTaskJSONError(c, http.StatusBadGateway, "api_error", "invalid image data")
			return
		}
		ctype, ext := detectImageContentType(data)
		writePlaygroundImageDownload(c, filename+ext, ctype, data)
		return
	}
	if strings.HasPrefix(strings.ToLower(item.URL), "data:image/") {
		data, decodeErr := decodeImageBase64(item.URL)
		if decodeErr != nil {
			imageTaskJSONError(c, http.StatusBadGateway, "api_error", "invalid image data")
			return
		}
		ctype, ext := detectImageContentType(data)
		writePlaygroundImageDownload(c, filename+ext, ctype, data)
		return
	}
	if item.URL == "" {
		imageTaskJSONError(c, http.StatusNotFound, "not_found_error", "image content is not available")
		return
	}
	data, ctype, fetchErr := service.FetchPlaygroundImageURL(c.Request.Context(), item.URL)
	if fetchErr != nil {
		imageTaskJSONError(c, http.StatusBadGateway, "api_error", "failed to download image")
		return
	}
	if !strings.HasPrefix(ctype, "image/") {
		ctype, _ = detectImageContentType(data)
	}
	_, ext := detectImageContentType(data)
	if ctype == "application/octet-stream" {
		ctype = "image/png"
	}
	writePlaygroundImageDownload(c, filename+ext, ctype, data)
}

func rewritePlaygroundGatewayBody(body []byte, kind string) (model string, groupID *int64, rewritten []byte, err error) {
	if len(bytes.TrimSpace(body)) == 0 {
		return "", nil, nil, playgroundBadRequest("Request body is empty")
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", nil, nil, playgroundBadRequest("Request body must be JSON")
	}
	model = strings.TrimSpace(jsonRawString(payload["model"]))
	if raw, ok := payload["group_id"]; ok {
		var id int64
		if json.Unmarshal(raw, &id) == nil && id > 0 {
			groupID = &id
		}
		delete(payload, "group_id")
	}
	if kind == service.SupportedModelKindText {
		var docs []service.PlaygroundDocument
		if raw, ok := payload["documents"]; ok && len(bytes.TrimSpace(raw)) > 0 && string(bytes.TrimSpace(raw)) != "null" {
			if err := json.Unmarshal(raw, &docs); err != nil {
				return "", nil, nil, playgroundBadRequest("documents must be an array")
			}
		}
		delete(payload, "documents")
		if len(docs) > 0 {
			extracted, extractErr := service.ExtractPlaygroundDocuments(docs)
			if extractErr != nil {
				return "", nil, nil, extractErr
			}
			prompt := service.FormatPlaygroundDocumentPrompt(extracted)
			if prompt != "" {
				messages, msgErr := attachPlaygroundDocuments(payload["messages"], prompt)
				if msgErr != nil {
					return "", nil, nil, msgErr
				}
				payload["messages"] = messages
			}
		}
	} else {
		delete(payload, "documents")
	}
	rewritten, err = json.Marshal(payload)
	if err != nil {
		return "", nil, nil, playgroundBadRequest("Failed to prepare request")
	}
	return model, groupID, rewritten, nil
}

func attachPlaygroundDocuments(raw json.RawMessage, prompt string) (json.RawMessage, error) {
	var messages []map[string]any
	if len(bytes.TrimSpace(raw)) > 0 && string(bytes.TrimSpace(raw)) != "null" {
		if err := json.Unmarshal(raw, &messages); err != nil {
			return nil, playgroundBadRequest("messages must be an array")
		}
	}
	block := map[string]any{"role": "user", "content": prompt}
	if len(messages) == 0 {
		messages = []map[string]any{block}
	} else {
		last := messages[len(messages)-1]
		if strings.EqualFold(strings.TrimSpace(asString(last["role"])), "user") {
			last["content"] = mergePlaygroundUserContent(last["content"], prompt)
			messages[len(messages)-1] = last
		} else {
			messages = append(messages, block)
		}
	}
	out, err := json.Marshal(messages)
	if err != nil {
		return nil, playgroundBadRequest("Failed to attach documents")
	}
	return out, nil
}

func mergePlaygroundUserContent(content any, prompt string) any {
	switch value := content.(type) {
	case string:
		if strings.TrimSpace(value) == "" {
			return prompt
		}
		return prompt + "\n\n" + value
	case []any:
		return append([]any{map[string]any{"type": "text", "text": prompt}}, value...)
	default:
		if content == nil {
			return prompt
		}
		return []any{
			map[string]any{"type": "text", "text": prompt},
			content,
		}
	}
}

type playgroundImageItem struct {
	URL     string
	B64JSON string
}

func playgroundImageOutput(result json.RawMessage, index int) (playgroundImageItem, error) {
	if len(result) == 0 || !json.Valid(result) {
		return playgroundImageItem{}, playgroundBadRequest("image result is empty")
	}
	var envelope struct {
		Data []struct {
			URL     string `json:"url"`
			B64JSON string `json:"b64_json"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &envelope); err != nil {
		return playgroundImageItem{}, playgroundBadRequest("image result is invalid")
	}
	if index < 0 || index >= len(envelope.Data) {
		return playgroundImageItem{}, playgroundErrorStatus(404, "NOT_FOUND", "image index not found")
	}
	item := envelope.Data[index]
	return playgroundImageItem{URL: strings.TrimSpace(item.URL), B64JSON: strings.TrimSpace(item.B64JSON)}, nil
}

func decodeImageBase64(value string) ([]byte, error) {
	value = strings.TrimSpace(value)
	if i := strings.Index(value, "base64,"); i >= 0 {
		value = value[i+len("base64,"):]
	}
	// Reject oversized payloads before allocating the decoded buffer.
	// Base64 expands by 4/3, so this is a conservative pre-check.
	maxBytes := service.PlaygroundMaxImageBytes()
	if int64(len(value)) > maxBytes*4/3+128 {
		return nil, playgroundDownloadError()
	}
	var data []byte
	var err error
	if data, err = base64.StdEncoding.DecodeString(value); err != nil {
		data, err = base64.RawStdEncoding.DecodeString(value)
	}
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, playgroundDownloadError()
	}
	return data, nil
}

func detectImageContentType(data []byte) (string, string) {
	switch {
	case bytes.HasPrefix(data, []byte("\x89PNG\r\n\x1a\n")):
		return "image/png", ".png"
	case bytes.HasPrefix(data, []byte{0xff, 0xd8, 0xff}):
		return "image/jpeg", ".jpg"
	case bytes.HasPrefix(data, []byte("RIFF")) && bytes.Contains(data[:min(12, len(data))], []byte("WEBP")):
		return "image/webp", ".webp"
	case bytes.HasPrefix(data, []byte("GIF87a")) || bytes.HasPrefix(data, []byte("GIF89a")):
		return "image/gif", ".gif"
	default:
		return "image/png", ".png"
	}
}

func writePlaygroundImageDownload(c *gin.Context, filename, contentType string, data []byte) {
	c.Header("Cache-Control", "no-store")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Data(http.StatusOK, contentType, data)
}

func playgroundUserID(c *gin.Context) (int64, bool) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		return 0, false
	}
	return subject.UserID, true
}

func abortPlaygroundError(c *gin.Context, err error) {
	var playgroundErr *service.ErrPlayground
	if errors.As(err, &playgroundErr) && playgroundErr != nil {
		if playgroundErr.Status >= 500 {
			response.InternalError(c, playgroundErr.Message)
			c.Abort()
			return
		}
		middleware2.AbortWithError(c, playgroundErr.Status, playgroundErr.Code, playgroundErr.Message)
		return
	}
	response.ErrorFrom(c, err)
	c.Abort()
}

func servicePlaygroundUnauthorized() error {
	return playgroundErrorStatus(401, "UNAUTHORIZED", "User not authenticated")
}

func playgroundBadRequest(message string) error {
	return playgroundErrorStatus(400, "INVALID_REQUEST", message)
}

func playgroundErrorStatus(status int, code, message string) error {
	return &service.ErrPlayground{Status: status, Code: code, Message: message}
}

func playgroundDownloadError() error {
	return playgroundErrorStatus(502, "IMAGE_DOWNLOAD_FAILED", "failed to download image")
}

func jsonRawString(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	return strings.Trim(string(raw), `"`)
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}
