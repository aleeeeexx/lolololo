package service

import (
	"context"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/pkg/antigravity"
	"github.com/Wei-Shaw/sub2api/internal/pkg/claude"
	"github.com/Wei-Shaw/sub2api/internal/pkg/geminicli"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/Wei-Shaw/sub2api/internal/pkg/xai"
)

// AdvertisedModel is a request-side model ID plus the platform(s) it was
// advertised from. It never includes account IDs, credentials, or group
// inventory.
type AdvertisedModel struct {
	Name      string
	Platforms []string
}

// DefaultModelIDsForPlatform returns the static fallback IDs used by GET /v1/models
// when a group has no advertised mapping catalog. The composite fallback unions
// concrete platform defaults (except TypeSafe). This must not be used as a
// site-wide catalog when the user has no applicable group.
func DefaultModelIDsForPlatform(platform string) []string {
	switch platform {
	case PlatformOpenAI:
		return openai.DefaultModelIDs()
	case PlatformGemini:
		ids := make([]string, 0, len(geminicli.DefaultModels))
		for _, model := range geminicli.DefaultModels {
			ids = append(ids, model.ID)
		}
		return ids
	case PlatformAntigravity:
		models := antigravity.DefaultModels()
		ids := make([]string, 0, len(models))
		for _, model := range models {
			ids = append(ids, model.ID)
		}
		return ids
	case PlatformAnthropic:
		return claude.DefaultModelIDs()
	case PlatformGrok:
		return xai.DefaultModelIDs()
	case PlatformOpenCodeGo:
		return DefaultOpenCodeGoModelIDs()
	case PlatformTypeSafe:
		return []string{"jev-latest"}
	case PlatformComposite:
		ids := make([]string, 0)
		seen := make(map[string]struct{})
		for _, concretePlatform := range domain.CompositePrecedencePlatformIDs() {
			// TypeSafe is deliberately skipped: jev-latest only works through
			// /v1/systemone, so the static fallback never advertises it to LLM
			// clients. CompositeAvailableModels lists it when the group can serve it.
			if concretePlatform == PlatformTypeSafe {
				continue
			}
			for _, id := range DefaultModelIDsForPlatform(concretePlatform) {
				if _, ok := seen[id]; ok {
					continue
				}
				seen[id] = struct{}{}
				ids = append(ids, id)
			}
		}
		return ids
	default:
		ids := make([]string, 0, len(claude.DefaultModels))
		for _, model := range claude.DefaultModels {
			ids = append(ids, model.ID)
		}
		return ids
	}
}

// ModelListingSource 汇总模型列表过滤的候选来源：账号映射键（availableModels）
// 与平台默认列表（fallbackModels）。账号映射为空时回落默认列表；Anthropic
// 平台两者取并集，其余平台以账号映射键为准。
// OpenAI 的空切片是权威结果（availableModels != nil），不得回落到默认列表。
func ModelListingSource(platform string, availableModels, fallbackModels []string) []string {
	if platform == PlatformOpenAI && availableModels != nil {
		return availableModels
	}
	if len(availableModels) == 0 {
		return fallbackModels
	}
	if platform == PlatformAnthropic {
		return MergeModelIDs(availableModels, fallbackModels)
	}
	return availableModels
}

// MergeModelIDs concatenates model IDs, trimming blanks and dropping duplicates
// while preserving first-seen order.
func MergeModelIDs(primary, secondary []string) []string {
	seen := make(map[string]struct{}, len(primary)+len(secondary))
	merged := make([]string, 0, len(primary)+len(secondary))
	for _, models := range [][]string{primary, secondary} {
		for _, model := range models {
			model = strings.TrimSpace(model)
			if model == "" {
				continue
			}
			if _, ok := seen[model]; ok {
				continue
			}
			seen[model] = struct{}{}
			merged = append(merged, model)
		}
	}
	return merged
}

// CompositeAvailableModels lists the models a composite group can serve, using
// the same rules as GET /v1/models. includeSystemOne adds TypeSafe models,
// which only work through /v1/systemone; LLM client catalogs must exclude them.
func (s *GatewayService) CompositeAvailableModels(ctx context.Context, groupID *int64, endpoint string, includeSystemOne bool) []string {
	return advertisedIDs(s.compositeAdvertisedModels(ctx, groupID, endpoint, includeSystemOne))
}

type advertisedModelBuilder struct {
	order []string
	items map[string]*AdvertisedModel
}

func newAdvertisedModelBuilder() *advertisedModelBuilder {
	return &advertisedModelBuilder{items: make(map[string]*AdvertisedModel)}
}

func (b *advertisedModelBuilder) add(name, platform string) {
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}
	platform = strings.TrimSpace(platform)
	item, ok := b.items[name]
	if !ok {
		item = &AdvertisedModel{Name: name}
		b.items[name] = item
		b.order = append(b.order, name)
	}
	if platform == "" {
		return
	}
	for _, existing := range item.Platforms {
		if existing == platform {
			return
		}
	}
	item.Platforms = append(item.Platforms, platform)
}

func (b *advertisedModelBuilder) result() []AdvertisedModel {
	out := make([]AdvertisedModel, 0, len(b.order))
	for _, name := range b.order {
		item := b.items[name]
		out = append(out, AdvertisedModel{
			Name:      item.Name,
			Platforms: append([]string(nil), item.Platforms...),
		})
	}
	return out
}

func withPlatforms(ids []string, platform string) []AdvertisedModel {
	b := newAdvertisedModelBuilder()
	for _, id := range ids {
		b.add(id, platform)
	}
	return b.result()
}

func advertisedIDs(models []AdvertisedModel) []string {
	ids := make([]string, 0, len(models))
	for _, model := range models {
		ids = append(ids, model.Name)
	}
	return ids
}

func (s *GatewayService) compositeAdvertisedModels(ctx context.Context, groupID *int64, endpoint string, includeSystemOne bool) []AdvertisedModel {
	b := newAdvertisedModelBuilder()
	if s == nil {
		return nil
	}
	schedulablePlatforms := s.GetSchedulablePlatforms(ctx, groupID)
	for _, platform := range domain.CompositePrecedencePlatformIDs() {
		if platform == PlatformTypeSafe && !includeSystemOne {
			continue
		}
		platformModels := s.GetAvailableModels(ctx, groupID, platform)
		if len(platformModels) == 0 && (platform != PlatformOpenAI || platformModels == nil) {
			// CN 供应商没有静态默认模型列表（DefaultModelIDsForPlatform 的
			// default 分支是 Claude 列表），composite 下只暴露账号映射键。
			if _, ok := schedulablePlatforms[platform]; ok && !IsMultiProtocolAPIKeyProvider(platform) {
				platformModels = DefaultModelIDsForPlatform(platform)
			}
		}
		for _, model := range platformModels {
			b.add(model, platform)
		}
	}
	// A route can expose a public ID that no account model mapping contains.
	// On lookup failure, retain the existing account-derived catalog only.
	if routeModels, err := s.GetCompositeRouteModels(ctx, groupID, endpoint, includeSystemOne); err == nil {
		for _, model := range routeModels {
			if _, exists := b.items[strings.TrimSpace(model)]; exists {
				continue
			}
			b.add(model, PlatformComposite)
		}
	}
	return b.result()
}

// ListAdvertisedModelsForGroup returns the request-side model IDs that
// GET /v1/models would advertise for an API key bound to this group, using
// already-saved configuration only.
//
// It does not call upstream APIs, does not run SyncUpstreamModelCatalog, and
// does not follow the live pinned Codex / OpenAI models-list discovery path
// (that path would contact upstream accounts).
//
// Fallback rules match /v1/models:
//   - no schedulable accounts / no mappings → the un-allowlisted /v1/models
//     fallback for that group only (not every platform's static catalog).
//     OpenAI/Gemini/Grok/TypeSafe use their platform defaults; Anthropic,
//     Antigravity, OpenCodeGo, and CN providers use Claude defaults, as
//     GatewayHandler.Models does;
//   - OpenAI: a non-nil empty catalog is authoritative and must not resurrect
//     retired defaults;
//   - Anthropic may union mappings with defaults and expand long-context
//     variants from saved upstream metadata;
//   - composite unions per-platform advertised IDs, then falls back to the
//     composite default union when nothing is advertised.
func (s *GatewayService) ListAdvertisedModelsForGroup(ctx context.Context, group *Group) []AdvertisedModel {
	if s == nil || group == nil {
		return nil
	}
	groupID := &group.ID
	platform := strings.TrimSpace(group.Platform)

	if platform == PlatformComposite {
		available := s.compositeAdvertisedModels(ctx, groupID, "", true)
		ids := advertisedIDs(available)
		if group.ModelAllowlistEnabled() {
			source := ids
			if len(source) == 0 {
				source = DefaultModelIDsForPlatform(PlatformComposite)
				return withPlatforms(group.ModelAllowlist.FilterForListing(source), PlatformComposite)
			}
			allowed := group.ModelAllowlist.FilterForListing(source)
			return filterAdvertisedModels(available, allowed)
		}
		if len(available) > 0 {
			return available
		}
		return withPlatforms(DefaultModelIDsForPlatform(PlatformComposite), PlatformComposite)
	}

	availableModels := s.GetAvailableModels(ctx, groupID, platform)
	if platform == PlatformAnthropic {
		source := availableModels
		if group.ModelAllowlistEnabled() {
			source = ModelListingSource(platform, availableModels, DefaultModelIDsForPlatform(platform))
		} else if len(source) == 0 {
			source = DefaultModelIDsForPlatform(platform)
		}
		expanded := s.ExpandAnthropicLongContextModels(ctx, group, source)
		if len(availableModels) > 0 || len(expanded) > len(source) || group.ModelAllowlistEnabled() {
			availableModels = expanded
		}
	}
	if group.ModelAllowlistEnabled() {
		source := ModelListingSource(platform, availableModels, DefaultModelIDsForPlatform(platform))
		return withPlatforms(group.ModelAllowlist.FilterForListing(source), platform)
	}
	if len(availableModels) > 0 || (platform == PlatformOpenAI && availableModels != nil) {
		return withPlatforms(availableModels, platform)
	}
	// Un-allowlisted fallback must follow GatewayHandler.Models, not
	// DefaultModelIDsForPlatform: Antigravity and OpenCodeGo also land on
	// Claude's default catalog when no mappings are advertised.
	return withPlatforms(modelsListFallbackIDs(platform), platform)
}

// modelsListFallbackIDs is the un-allowlisted GET /v1/models fallback catalog.
func modelsListFallbackIDs(platform string) []string {
	switch platform {
	case PlatformOpenAI:
		return openai.DefaultModelIDs()
	case PlatformGemini:
		ids := make([]string, 0, len(geminicli.DefaultModels))
		for _, model := range geminicli.DefaultModels {
			ids = append(ids, model.ID)
		}
		return ids
	case PlatformGrok:
		return xai.DefaultModelIDs()
	case PlatformTypeSafe:
		return []string{"jev-latest"}
	default:
		ids := make([]string, 0, len(claude.DefaultModels))
		for _, model := range claude.DefaultModels {
			ids = append(ids, model.ID)
		}
		return ids
	}
}

func filterAdvertisedModels(models []AdvertisedModel, allowed []string) []AdvertisedModel {
	index := make(map[string]AdvertisedModel, len(models))
	for _, model := range models {
		index[model.Name] = model
	}
	out := make([]AdvertisedModel, 0, len(allowed))
	for _, id := range allowed {
		if model, ok := index[id]; ok {
			out = append(out, model)
			continue
		}
		out = append(out, AdvertisedModel{Name: id, Platforms: []string{PlatformComposite}})
	}
	return out
}
