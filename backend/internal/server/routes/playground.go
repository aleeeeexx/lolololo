package routes

import (
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/handler"
	pkghttputil "github.com/Wei-Shaw/sub2api/internal/pkg/httputil"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// RegisterPlaygroundRoutes registers JWT-authenticated "直接用" page APIs.
// These routes are temporarily limited to administrators (role=admin).
// They prepare an internal API key context and then enter the existing
// gateway handlers and billing. The browser never sends an API key.
func RegisterPlaygroundRoutes(
	v1 *gin.RouterGroup,
	h *handler.Handlers,
	jwtAuth middleware.JWTAuthMiddleware,
	auditLog middleware.AuditLogMiddleware,
	settingService *service.SettingService,
	panelRateLimiter *middleware.PanelRateLimiter,
	compositeResolver *service.CompositeRouteResolver,
	cfg *config.Config,
) {
	if h == nil || h.Playground == nil {
		return
	}

	var memoryBudgetBytes, textMaxBody int64
	if cfg != nil {
		memoryBudgetBytes = cfg.Gateway.RequestBodyMemoryBudgetBytes
		textMaxBody = cfg.Gateway.TextMaxBodySize
	}
	bodyMemoryBudget := pkghttputil.NewBodyMemoryBudget(memoryBudgetBytes)
	textBodyLimit := middleware.RequestBodyLimit(textMaxBody, bodyMemoryBudget)

	group := v1.Group("/playground")
	group.Use(gin.HandlerFunc(jwtAuth))
	group.Use(middleware.AdminOnly())
	group.Use(middleware.BackendModeUserGuard(settingService))
	group.Use(panelRateLimiter.Global())
	group.Use(gin.HandlerFunc(auditLog))
	group.Use(middleware.ClientRequestID())
	group.Use(handler.InboundEndpointMiddleware())

	group.GET("/session", h.Playground.Session)
	group.GET("/requests/:request_id", panelRateLimiter.Heavy(), h.Playground.RequestCost)
	group.GET("/images/tasks/:task_id", h.Playground.GetImageTask)
	group.GET("/images/tasks/:task_id/content/:index", panelRateLimiter.Heavy(), h.Playground.DownloadImage)

	requireGroup := middleware.RequireGroupAssignment(settingService, middleware.OpenAIErrorWriter)
	groupModelAllowlist := middleware.GroupModelAllowlist()
	compositeTarget := compositeTargetPlatformMiddleware(compositeResolver)
	rewriteImagePlatform := rewriteCompositePlaygroundImagePlatform()

	chat := group.Group("")
	chat.Use(panelRateLimiter.Heavy())
	chat.Use(textBodyLimit)
	chat.Use(h.Playground.PrepareText())
	chat.Use(groupModelAllowlist)
	chat.Use(compositeTarget)
	chat.Use(requireGroup)
	chat.POST("/chat/completions", playgroundChatCompletions(h))

	images := group.Group("")
	images.Use(panelRateLimiter.Heavy())
	images.Use(textBodyLimit)
	images.Use(h.Playground.PrepareImage())
	images.Use(groupModelAllowlist)
	images.Use(compositeTarget)
	images.Use(rewriteImagePlatform)
	images.Use(requireGroup)
	images.POST("/images/generations", h.AsyncImage.Submit)
}

func playgroundChatCompletions(h *handler.Handlers) gin.HandlerFunc {
	return func(c *gin.Context) {
		if domain.UsesOpenAIGateway(getGroupPlatform(c)) {
			h.OpenAIGateway.ChatCompletions(c)
			return
		}
		h.Gateway.ChatCompletions(c)
	}
}

func rewriteCompositePlaygroundImagePlatform() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey, ok := middleware.GetAPIKeyFromContext(c)
		if ok && apiKey != nil && apiKey.Group != nil && apiKey.Group.Platform == service.PlatformComposite {
			if platform, resolved := service.ResolvedTargetPlatformFromContext(c.Request.Context()); resolved &&
				(platform == service.PlatformOpenAI || platform == service.PlatformGrok) {
				groupCopy := *apiKey.Group
				groupCopy.Platform = platform
				keyCopy := *apiKey
				keyCopy.Group = &groupCopy
				c.Set(string(middleware.ContextKeyAPIKey), &keyCopy)
			}
		}
		c.Next()
	}
}
