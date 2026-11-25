package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// SupportedModelItem is a sanitized, user-facing model listing entry.
// It contains the request-side model ID, platform labels, generation kinds,
// and the inferred model vendor. It never includes account, credential, or
// group identifiers.
type SupportedModelItem struct {
	Name      string   `json:"name"`
	Platforms []string `json:"platforms,omitempty"`
	Kinds     []string `json:"kinds"`
	Company   string   `json:"company"`
}

// userGroupLister is satisfied by APIKeyService.GetAvailableGroups, which
// already applies bindable-group and subscription permission checks.
type userGroupLister interface {
	GetAvailableGroups(ctx context.Context, userID int64) ([]Group, error)
}

// groupModelAdvertiser is satisfied by GatewayService.ListAdvertisedModelsForGroup.
type groupModelAdvertiser interface {
	ListAdvertisedModelsForGroup(ctx context.Context, group *Group) []AdvertisedModel
}

// SupportedModelsService aggregates GET /v1/models request-side IDs across
// the groups the current user can bind, without feature-flag gates or
// upstream calls.
type SupportedModelsService struct {
	groups  userGroupLister
	listing groupModelAdvertiser
}

func NewSupportedModelsService(groups userGroupLister, listing groupModelAdvertiser) *SupportedModelsService {
	return &SupportedModelsService{groups: groups, listing: listing}
}

// ListForUser returns the deduplicated request-side model IDs the current user
// can use, based on already-saved account/group configuration. An empty list
// is a valid result when the user has no bindable groups, or none of those
// groups advertise any models after allowlist filtering.
func (s *SupportedModelsService) ListForUser(ctx context.Context, userID int64) ([]SupportedModelItem, error) {
	if s == nil || s.groups == nil || s.listing == nil {
		return nil, fmt.Errorf("supported models service is not configured")
	}
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user id")
	}

	groups, err := s.groups.GetAvailableGroups(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list available groups: %w", err)
	}
	if len(groups) == 0 {
		return []SupportedModelItem{}, nil
	}

	builder := newAdvertisedModelBuilder()
	for i := range groups {
		for _, model := range s.listing.ListAdvertisedModelsForGroup(ctx, &groups[i]) {
			if len(model.Platforms) == 0 {
				builder.add(model.Name, strings.TrimSpace(groups[i].Platform))
				continue
			}
			for _, platform := range model.Platforms {
				builder.add(model.Name, platform)
			}
		}
	}

	items := builder.result()
	out := make([]SupportedModelItem, 0, len(items))
	for _, item := range items {
		platforms := append([]string(nil), item.Platforms...)
		sort.Strings(platforms)
		kinds, company := classifySupportedModel(item.Name)
		out = append(out, SupportedModelItem{
			Name:      item.Name,
			Platforms: platforms,
			Kinds:     kinds,
			Company:   company,
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out, nil
}
