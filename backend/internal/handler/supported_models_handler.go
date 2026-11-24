package handler

import (
	"context"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type supportedModelsLister interface {
	ListForUser(ctx context.Context, userID int64) ([]service.SupportedModelItem, error)
}

// SupportedModelsHandler serves GET /api/v1/supported-models for logged-in users.
// The catalog is independent of model_plaza_enabled / available_channels_enabled
// and never calls upstream or SyncUpstreamModelCatalog.
type SupportedModelsHandler struct {
	svc supportedModelsLister
}

// NewSupportedModelsHandler wires the user-facing supported-models endpoint.
func NewSupportedModelsHandler(
	apiKeyService *service.APIKeyService,
	gatewayService *service.GatewayService,
) *SupportedModelsHandler {
	return &SupportedModelsHandler{
		svc: service.NewSupportedModelsService(apiKeyService, gatewayService),
	}
}

// userSupportedModel is the only response shape this endpoint exposes.
type supportedModelsListItem struct {
	Name      string   `json:"name"`
	Platforms []string `json:"platforms,omitempty"`
}

// List GET /api/v1/supported-models
func (h *SupportedModelsHandler) List(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	if h == nil || h.svc == nil {
		response.InternalError(c, "Supported models service is not configured")
		return
	}

	models, err := h.svc.ListForUser(c.Request.Context(), subject.UserID)
	if err != nil {
		response.InternalError(c, "Failed to list supported models")
		return
	}
	if models == nil {
		models = []service.SupportedModelItem{}
	}

	out := make([]supportedModelsListItem, 0, len(models))
	for _, model := range models {
		out = append(out, supportedModelsListItem{
			Name:      model.Name,
			Platforms: append([]string(nil), model.Platforms...),
		})
	}
	response.Success(c, out)
}
