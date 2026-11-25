package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
)

const (
	// PlaygroundAPIKeyNamePrefix identifies server-managed keys used by the
	// logged-in "使用" page. The raw credential is never returned to the
	// frontend; the key exists only so existing gateway billing and usage
	// records can attach to a real api_key_id.
	PlaygroundAPIKeyNamePrefix = "__playground__:"

	PlaygroundMaxDocuments     = 3
	PlaygroundMaxDocumentBytes = 5 << 20
	PlaygroundMaxDocumentRunes = 100_000
)

// ErrPlayground is a user-facing playground authorization or validation error.
type ErrPlayground struct {
	Status  int
	Code    string
	Message string
}

func (e *ErrPlayground) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func playgroundError(status int, code, message string) *ErrPlayground {
	return &ErrPlayground{Status: status, Code: code, Message: message}
}

func IsPlaygroundAPIKeyName(name string) bool {
	return strings.HasPrefix(name, PlaygroundAPIKeyNamePrefix)
}

func PlaygroundAPIKeyName(groupID int64) string {
	return fmt.Sprintf("%sg%d", PlaygroundAPIKeyNamePrefix, groupID)
}

// PlaygroundService prepares a JWT user's request to enter the existing
// gateway and billing path without the browser presenting an API key.
type PlaygroundService struct {
	keys          *APIKeyService
	groups        userGroupLister
	listing       groupModelAdvertiser
	subscriptions *SubscriptionService
	users         *UserService
	usage         *UsageService
	tasks         *ImageTaskService
	cfg           *config.Config
}

func NewPlaygroundService(
	keys *APIKeyService,
	subscriptions *SubscriptionService,
	users *UserService,
	usage *UsageService,
	gateway *GatewayService,
	tasks *ImageTaskService,
	cfg *config.Config,
) *PlaygroundService {
	var groups userGroupLister
	var listing groupModelAdvertiser
	if keys != nil {
		groups = keys
	}
	if gateway != nil {
		listing = gateway
	}
	return &PlaygroundService{
		keys:          keys,
		groups:        groups,
		listing:       listing,
		subscriptions: subscriptions,
		users:         users,
		usage:         usage,
		tasks:         tasks,
		cfg:           cfg,
	}
}

// PlaygroundSession is the quota snapshot shown on the 使用 page.
type PlaygroundSession struct {
	Balance          float64                  `json:"balance"`
	FrozenBalance    float64                  `json:"frozen_balance"`
	AvailableBalance float64                  `json:"available_balance"`
	SimpleMode       bool                     `json:"simple_mode"`
	Subscriptions    []PlaygroundSubscription `json:"subscriptions"`
	Groups           []PlaygroundGroup        `json:"groups"`
	Limits           PlaygroundDocumentLimits `json:"limits"`
}

// PlaygroundGroup is a bindable group the 使用 page can charge against.
type PlaygroundGroup struct {
	ID                   int64    `json:"id"`
	Name                 string   `json:"name"`
	Platform             string   `json:"platform"`
	RateMultiplier       float64  `json:"rate_multiplier"`
	SubscriptionType     bool     `json:"subscription_type"`
	AllowImageGeneration bool     `json:"allow_image_generation"`
	SupportsImage        bool     `json:"supports_image"`
	HasRemainingQuota    bool     `json:"has_remaining_quota"`
	Models               []string `json:"models"`
}

type PlaygroundSubscription struct {
	ID              int64   `json:"id"`
	GroupID         int64   `json:"group_id"`
	GroupName       string  `json:"group_name"`
	Status          string  `json:"status"`
	DailyUsedUSD    float64 `json:"daily_used_usd"`
	DailyLimitUSD   float64 `json:"daily_limit_usd"`
	WeeklyUsedUSD   float64 `json:"weekly_used_usd"`
	WeeklyLimitUSD  float64 `json:"weekly_limit_usd"`
	MonthlyUsedUSD  float64 `json:"monthly_used_usd"`
	MonthlyLimitUSD float64 `json:"monthly_limit_usd"`
}

type PlaygroundDocumentLimits struct {
	MaxDocuments     int   `json:"max_documents"`
	MaxDocumentBytes int64 `json:"max_document_bytes"`
	MaxDocumentRunes int   `json:"max_document_runes"`
}

func playgroundDocumentLimits() PlaygroundDocumentLimits {
	return PlaygroundDocumentLimits{
		MaxDocuments:     PlaygroundMaxDocuments,
		MaxDocumentBytes: PlaygroundMaxDocumentBytes,
		MaxDocumentRunes: PlaygroundMaxDocumentRunes,
	}
}

func (s *PlaygroundService) Session(ctx context.Context, userID int64) (*PlaygroundSession, error) {
	if s == nil || s.users == nil {
		return nil, playgroundError(500, "INTERNAL_ERROR", "playground is not configured")
	}
	if userID <= 0 {
		return nil, playgroundError(401, "UNAUTHORIZED", "User not authenticated")
	}
	user, err := s.users.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	available := user.Balance - user.FrozenBalance
	if available < 0 {
		available = 0
	}
	out := &PlaygroundSession{
		Balance:          user.Balance,
		FrozenBalance:    user.FrozenBalance,
		AvailableBalance: available,
		SimpleMode:       s.cfg != nil && s.cfg.RunMode == config.RunModeSimple,
		Subscriptions:    []PlaygroundSubscription{},
		Groups:           []PlaygroundGroup{},
		Limits:           playgroundDocumentLimits(),
	}
	if s.subscriptions == nil {
		groups, err := s.listPlaygroundGroups(ctx, userID, out.Subscriptions)
		if err != nil {
			return nil, err
		}
		out.Groups = groups
		return out, nil
	}
	subs, err := s.subscriptions.ListActiveUserSubscriptions(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, sub := range subs {
		item := PlaygroundSubscription{
			ID:             sub.ID,
			GroupID:        sub.GroupID,
			Status:         sub.Status,
			DailyUsedUSD:   sub.DailyUsageUSD,
			WeeklyUsedUSD:  sub.WeeklyUsageUSD,
			MonthlyUsedUSD: sub.MonthlyUsageUSD,
		}
		if sub.Group != nil {
			item.GroupName = sub.Group.Name
			if sub.Group.DailyLimitUSD != nil {
				item.DailyLimitUSD = *sub.Group.DailyLimitUSD
			}
			if sub.Group.WeeklyLimitUSD != nil {
				item.WeeklyLimitUSD = *sub.Group.WeeklyLimitUSD
			}
			if sub.Group.MonthlyLimitUSD != nil {
				item.MonthlyLimitUSD = *sub.Group.MonthlyLimitUSD
			}
		}
		out.Subscriptions = append(out.Subscriptions, item)
	}
	groups, err := s.listPlaygroundGroups(ctx, userID, out.Subscriptions)
	if err != nil {
		return nil, err
	}
	out.Groups = groups
	return out, nil
}

// PlaygroundRequestCost is the user-facing cost of one playground request.
type PlaygroundRequestCost struct {
	Found      bool    `json:"found"`
	RequestID  string  `json:"request_id"`
	Model      string  `json:"model,omitempty"`
	ActualCost float64 `json:"actual_cost"`
	TotalCost  float64 `json:"total_cost"`
}

func (s *PlaygroundService) RequestCost(ctx context.Context, userID int64, requestID string) (*PlaygroundRequestCost, error) {
	if s == nil || s.usage == nil {
		return nil, playgroundError(500, "INTERNAL_ERROR", "playground is not configured")
	}
	if userID <= 0 {
		return nil, playgroundError(401, "UNAUTHORIZED", "User not authenticated")
	}
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return nil, playgroundError(400, "INVALID_REQUEST", "request_id is required")
	}
	logs, _, err := s.usage.ListWithFilters(ctx, paginationParamsOne(), usageFiltersForPlaygroundCost(userID, requestID))
	if err != nil {
		return nil, err
	}
	out := &PlaygroundRequestCost{RequestID: requestID}
	if len(logs) == 0 {
		return out, nil
	}
	log := logs[0]
	out.Found = true
	out.Model = log.RequestedModel
	if strings.TrimSpace(out.Model) == "" {
		out.Model = log.Model
	}
	out.ActualCost = log.ActualCost
	out.TotalCost = log.TotalCost
	return out, nil
}

// PlaygroundPrepareResult is the authorized gateway context for one request.
type PlaygroundPrepareResult struct {
	APIKey       *APIKey
	Subscription *UserSubscription
	Group        *Group
	Platform     string
	Model        string
}

func (s *PlaygroundService) Prepare(ctx context.Context, userID int64, model, kind string, groupID *int64) (*PlaygroundPrepareResult, error) {
	if s == nil || s.keys == nil || s.groups == nil || s.listing == nil {
		return nil, playgroundError(500, "INTERNAL_ERROR", "playground is not configured")
	}
	if userID <= 0 {
		return nil, playgroundError(401, "UNAUTHORIZED", "User not authenticated")
	}
	model = strings.TrimSpace(model)
	if model == "" {
		return nil, playgroundError(400, "INVALID_REQUEST", "model is required")
	}
	if err := validatePlaygroundModelKind(model, kind); err != nil {
		return nil, err
	}
	group, err := s.resolveGroup(ctx, userID, model, kind, groupID)
	if err != nil {
		return nil, err
	}
	apiKey, err := s.keys.EnsurePlaygroundAPIKey(ctx, userID, group.ID)
	if err != nil {
		return nil, err
	}
	if apiKey == nil || apiKey.User == nil || apiKey.ID <= 0 {
		return nil, playgroundError(500, "INTERNAL_ERROR", "failed to prepare playground session")
	}
	if !apiKey.User.IsActive() {
		return nil, playgroundError(401, "USER_INACTIVE", "User account is not active")
	}
	if apiKey.GroupID == nil || *apiKey.GroupID != group.ID {
		return nil, playgroundError(500, "INTERNAL_ERROR", "playground api key is bound to a different group")
	}
	if apiKey.Group == nil {
		apiKey.Group = group
	}
	if apiKey.Group.ID != group.ID {
		return nil, playgroundError(500, "INTERNAL_ERROR", "playground api key is bound to a different group")
	}
	if err := s.enforceBilling(ctx, apiKey); err != nil {
		return nil, err
	}
	sub, err := s.loadSubscription(ctx, apiKey)
	if err != nil {
		return nil, err
	}
	return &PlaygroundPrepareResult{
		APIKey:       apiKey,
		Subscription: sub,
		Group:        apiKey.Group,
		Platform:     strings.TrimSpace(apiKey.Group.Platform),
		Model:        model,
	}, nil
}

func (s *PlaygroundService) LoadSubscription(ctx context.Context, apiKey *APIKey) (*UserSubscription, error) {
	return s.loadSubscription(ctx, apiKey)
}

func paginationParamsOne() pagination.PaginationParams {
	return pagination.PaginationParams{Page: 1, PageSize: 1, SortOrder: pagination.SortOrderDesc}
}

func usageFiltersForPlaygroundCost(userID int64, requestID string) usagestats.UsageLogFilters {
	return usagestats.UsageLogFilters{UserID: userID, RequestID: requestID}
}

func (s *PlaygroundService) resolveGroup(ctx context.Context, userID int64, model, kind string, groupID *int64) (*Group, error) {
	groups, err := s.groups.GetAvailableGroups(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list available groups: %w", err)
	}
	if len(groups) == 0 {
		return nil, playgroundError(403, "NO_AVAILABLE_GROUP", "No available group for this account")
	}
	if groupID != nil && *groupID > 0 {
		for i := range groups {
			if groups[i].ID == *groupID {
				if !groupAdvertisesModel(s.listing, ctx, &groups[i], model) {
					return nil, playgroundError(403, "MODEL_NOT_ALLOWED", "This model is not available in the selected group")
				}
				if kind == SupportedModelKindImage && !groupSupportsPlaygroundImage(&groups[i]) {
					return nil, playgroundError(404, "IMAGES_NOT_SUPPORTED", "Images API is not supported for this group")
				}
				g := groups[i]
				return &g, nil
			}
		}
		return nil, playgroundError(403, "GROUP_NOT_ALLOWED", "The selected group is not available")
	}
	var candidates []Group
	for i := range groups {
		if !groupAdvertisesModel(s.listing, ctx, &groups[i], model) {
			continue
		}
		if kind == SupportedModelKindImage && !groupSupportsPlaygroundImage(&groups[i]) {
			continue
		}
		candidates = append(candidates, groups[i])
	}
	if len(candidates) == 0 {
		if kind == SupportedModelKindImage {
			return nil, playgroundError(404, "IMAGES_NOT_SUPPORTED", "No available group supports this image model")
		}
		return nil, playgroundError(403, "MODEL_NOT_ALLOWED", "This model is not available for your account")
	}
	if len(candidates) > 1 {
		return nil, playgroundError(400, "GROUP_REQUIRED", "group_id is required when multiple groups can serve this model")
	}
	g := candidates[0]
	return &g, nil
}

func groupAdvertisesModel(listing groupModelAdvertiser, ctx context.Context, group *Group, model string) bool {
	if listing == nil || group == nil {
		return false
	}
	want := strings.TrimSpace(model)
	if want == "" {
		return false
	}
	for _, item := range listing.ListAdvertisedModelsForGroup(ctx, group) {
		if strings.TrimSpace(item.Name) == want {
			return true
		}
	}
	return false
}

func groupSupportsPlaygroundImage(group *Group) bool {
	if group == nil || !GroupAllowsImageGeneration(group) {
		return false
	}
	switch strings.TrimSpace(group.Platform) {
	case PlatformOpenAI, PlatformGrok, PlatformComposite:
		return true
	default:
		return false
	}
}

func validatePlaygroundModelKind(model, kind string) error {
	kinds, _ := classifySupportedModel(model)
	switch kind {
	case SupportedModelKindImage:
		if !hasSupportedKind(kinds, SupportedModelKindImage) {
			return playgroundError(400, "INVALID_MODEL_KIND", "Selected model is not an image generation model")
		}
	default:
		if hasSupportedKind(kinds, SupportedModelKindImage) || hasSupportedKind(kinds, SupportedModelKindVideo) {
			return playgroundError(400, "INVALID_MODEL_KIND", "Selected model cannot be used for text chat")
		}
	}
	return nil
}

func hasSupportedKind(kinds []string, want string) bool {
	for _, kind := range kinds {
		if kind == want {
			return true
		}
	}
	return false
}

func (s *PlaygroundService) enforceBilling(ctx context.Context, apiKey *APIKey) error {
	if s != nil && s.cfg != nil && s.cfg.RunMode == config.RunModeSimple {
		return nil
	}
	if apiKey == nil {
		return playgroundError(500, "INTERNAL_ERROR", "playground session is incomplete")
	}
	switch apiKey.Status {
	case StatusAPIKeyQuotaExhausted:
		return playgroundError(429, "API_KEY_QUOTA_EXHAUSTED", "API key 额度已用完")
	case StatusAPIKeyExpired:
		return playgroundError(403, "API_KEY_EXPIRED", "API key 已过期")
	}
	if !apiKey.IsActive() && apiKey.Status != StatusAPIKeyQuotaExhausted && apiKey.Status != StatusAPIKeyExpired {
		return playgroundError(401, "API_KEY_DISABLED", "API key is disabled")
	}
	if apiKey.IsExpired() {
		return playgroundError(403, "API_KEY_EXPIRED", "API key 已过期")
	}
	if apiKey.IsQuotaExhausted() {
		return playgroundError(429, "API_KEY_QUOTA_EXHAUSTED", "API key 额度已用完")
	}
	sub, err := s.loadSubscription(ctx, apiKey)
	if err != nil {
		return err
	}
	if sub != nil && s.subscriptions != nil {
		needsMaintenance, validateErr := s.subscriptions.ValidateAndCheckLimits(sub, apiKey.Group)
		if needsMaintenance {
			refreshed, maintenanceErr := s.subscriptions.EnsureWindowMaintenance(ctx, sub)
			if maintenanceErr != nil {
				return playgroundError(500, "SUBSCRIPTION_MAINTENANCE_FAILED", "Failed to maintain subscription usage windows")
			}
			sub = refreshed
			_, validateErr = s.subscriptions.ValidateAndCheckLimits(sub, apiKey.Group)
		}
		if validateErr != nil {
			code := "SUBSCRIPTION_INVALID"
			status := 403
			if errors.Is(validateErr, ErrDailyLimitExceeded) ||
				errors.Is(validateErr, ErrWeeklyLimitExceeded) ||
				errors.Is(validateErr, ErrMonthlyLimitExceeded) {
				code = "USAGE_LIMIT_EXCEEDED"
				status = 429
			}
			return playgroundError(status, code, validateErr.Error())
		}
		return nil
	}
	if apiKey.User != nil && apiKey.User.Balance <= 0 {
		return playgroundError(403, "INSUFFICIENT_BALANCE", "Insufficient account balance")
	}
	return nil
}

func (s *PlaygroundService) loadSubscription(ctx context.Context, apiKey *APIKey) (*UserSubscription, error) {
	if apiKey == nil || apiKey.Group == nil || !apiKey.Group.IsSubscriptionType() || s == nil || s.subscriptions == nil || s.cfg != nil && s.cfg.RunMode == config.RunModeSimple {
		return nil, nil
	}
	if apiKey.User == nil {
		return nil, playgroundError(401, "USER_NOT_FOUND", "User associated with API key not found")
	}
	sub, err := s.subscriptions.GetActiveSubscription(ctx, apiKey.User.ID, apiKey.Group.ID)
	if err != nil {
		return nil, playgroundError(403, "SUBSCRIPTION_NOT_FOUND", "No active subscription found for this group")
	}
	return sub, nil
}

func (s *PlaygroundService) listPlaygroundGroups(ctx context.Context, userID int64, subs []PlaygroundSubscription) ([]PlaygroundGroup, error) {
	if s == nil || s.groups == nil {
		return []PlaygroundGroup{}, nil
	}
	groups, err := s.groups.GetAvailableGroups(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list available groups: %w", err)
	}
	rates := map[int64]float64{}
	if s.keys != nil {
		if custom, rateErr := s.keys.GetUserGroupRates(ctx, userID); rateErr == nil && custom != nil {
			rates = custom
		}
	}
	remaining := map[int64]bool{}
	for _, sub := range subs {
		remaining[sub.GroupID] = playgroundSubscriptionHasRemaining(sub)
	}
	out := make([]PlaygroundGroup, 0, len(groups))
	for i := range groups {
		group := groups[i]
		item := PlaygroundGroup{
			ID:                   group.ID,
			Name:                 group.Name,
			Platform:             group.Platform,
			RateMultiplier:       group.RateMultiplier,
			SubscriptionType:     group.IsSubscriptionType(),
			AllowImageGeneration: group.AllowImageGeneration,
			SupportsImage:        groupSupportsPlaygroundImage(&group),
			Models:               []string{},
		}
		if rate, ok := rates[group.ID]; ok && rate > 0 {
			item.RateMultiplier = rate
		}
		if group.IsSubscriptionType() {
			item.HasRemainingQuota = remaining[group.ID]
		}
		if s.listing != nil {
			seen := map[string]struct{}{}
			for _, model := range s.listing.ListAdvertisedModelsForGroup(ctx, &group) {
				name := strings.TrimSpace(model.Name)
				if name == "" {
					continue
				}
				if _, ok := seen[name]; ok {
					continue
				}
				seen[name] = struct{}{}
				item.Models = append(item.Models, name)
			}
		}
		out = append(out, item)
	}
	return out, nil
}

func playgroundSubscriptionHasRemaining(sub PlaygroundSubscription) bool {
	if sub.DailyLimitUSD > 0 && sub.DailyUsedUSD >= sub.DailyLimitUSD {
		return false
	}
	if sub.WeeklyLimitUSD > 0 && sub.WeeklyUsedUSD >= sub.WeeklyLimitUSD {
		return false
	}
	if sub.MonthlyLimitUSD > 0 && sub.MonthlyUsedUSD >= sub.MonthlyLimitUSD {
		return false
	}
	return true
}
