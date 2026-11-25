package service

import (
	"context"
	"fmt"
	"html"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

func reservedPlaygroundKeyNameError() error {
	return infraerrors.BadRequest("PLAYGROUND_KEY_RESERVED", "this key name is reserved by the playground")
}

func rejectReservedPlaygroundKeyName(name string) error {
	if IsPlaygroundAPIKeyName(name) {
		return reservedPlaygroundKeyNameError()
	}
	return nil
}

// EnsurePlaygroundAPIKey returns an active server-managed API key for the
// (user, group) pair so the 使用 page can enter existing gateway billing.
// The raw key is never exposed by playground handlers.
func (s *APIKeyService) EnsurePlaygroundAPIKey(ctx context.Context, userID, groupID int64) (*APIKey, error) {
	if s == nil {
		return nil, fmt.Errorf("api key service is not configured")
	}
	if userID <= 0 || groupID <= 0 {
		return nil, playgroundError(400, "INVALID_REQUEST", "user and group are required")
	}
	name := PlaygroundAPIKeyName(groupID)
	existing, err := s.findPlaygroundAPIKey(ctx, userID, groupID, name)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		existing, err = s.repairPlaygroundAPIKeyGroup(ctx, userID, groupID, name)
		if err != nil {
			return nil, err
		}
	}
	if existing != nil {
		if !playgroundAPIKeyBoundToGroup(*existing, groupID) {
			return nil, playgroundError(500, "INTERNAL_ERROR", "playground api key is bound to a different group")
		}
		if existing.Status != StatusActive {
			existing.Status = StatusActive
			if updateErr := s.apiKeyRepo.Update(ctx, existing, APIKeyUpdateFields{Status: true}); updateErr != nil {
				return nil, fmt.Errorf("reactivate playground api key: %w", updateErr)
			}
		}
		return s.GetByID(ctx, existing.ID)
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("get group: %w", err)
	}
	if !s.canUserBindGroup(ctx, user, group) && !group.IsSubscriptionType() {
		return nil, ErrGroupNotAllowed
	}
	key, err := s.GenerateKey()
	if err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}
	apiKey := &APIKey{
		UserID:  userID,
		Key:     key,
		Name:    html.EscapeString(name),
		GroupID: &groupID,
		Status:  StatusActive,
	}
	if err := s.apiKeyRepo.Create(ctx, apiKey); err != nil {
		if existing, findErr := s.findPlaygroundAPIKey(ctx, userID, groupID, name); findErr == nil && existing != nil {
			return s.GetByID(ctx, existing.ID)
		}
		return nil, fmt.Errorf("create playground api key: %w", err)
	}
	s.InvalidateAuthCacheByKey(ctx, apiKey.Key)
	s.compileAPIKeyIPRules(apiKey)
	return s.GetByID(ctx, apiKey.ID)
}

func (s *APIKeyService) repairPlaygroundAPIKeyGroup(ctx context.Context, userID, groupID int64, name string) (*APIKey, error) {
	existing, err := s.findRepairablePlaygroundAPIKey(ctx, userID, groupID, name)
	if err != nil || existing == nil {
		return existing, err
	}
	existing.GroupID = &groupID
	if updateErr := s.apiKeyRepo.Update(ctx, existing, APIKeyUpdateFields{GroupID: true}); updateErr != nil {
		return nil, fmt.Errorf("repair playground api key group: %w", updateErr)
	}
	s.InvalidateAuthCacheByKey(ctx, existing.Key)
	return existing, nil
}

func (s *APIKeyService) findPlaygroundAPIKey(ctx context.Context, userID, groupID int64, name string) (*APIKey, error) {
	keys, err := s.listPlaygroundAPIKeys(ctx, userID)
	if err != nil {
		return nil, err
	}
	var match *APIKey
	for i := range keys {
		key := keys[i]
		if !playgroundAPIKeyBoundToGroup(key, groupID) {
			continue
		}
		if key.Name == name {
			item := key
			return &item, nil
		}
		if match == nil || key.ID < match.ID {
			item := key
			match = &item
		}
	}
	return match, nil
}

// findRepairablePlaygroundAPIKey returns a key that was named for this group
// but has no group_id, so EnsurePlaygroundAPIKey can bind it correctly.
// Keys already bound to a different group are never reused.
func (s *APIKeyService) findRepairablePlaygroundAPIKey(ctx context.Context, userID, groupID int64, name string) (*APIKey, error) {
	keys, err := s.listPlaygroundAPIKeys(ctx, userID)
	if err != nil {
		return nil, err
	}
	var match *APIKey
	for i := range keys {
		key := keys[i]
		if key.Name != name || key.GroupID != nil {
			continue
		}
		if match == nil || key.ID < match.ID {
			item := key
			match = &item
		}
	}
	return match, nil
}

func (s *APIKeyService) listPlaygroundAPIKeys(ctx context.Context, userID int64) ([]APIKey, error) {
	lister, ok := s.apiKeyRepo.(apiKeyAllByUserIDLister)
	if !ok {
		return nil, fmt.Errorf("list api keys: repository does not support listing all keys")
	}
	keys, err := lister.ListAllByUserID(ctx, userID, APIKeyListFilters{IncludePlaygroundKeys: true})
	if err != nil {
		return nil, fmt.Errorf("list api keys: %w", err)
	}
	return keys, nil
}

func playgroundAPIKeyBoundToGroup(key APIKey, groupID int64) bool {
	return IsPlaygroundAPIKeyName(key.Name) && key.GroupID != nil && *key.GroupID == groupID
}

// ListActivePersonalBoundGroupIDs returns unique group IDs bound on the
// user's enabled personal API keys. Hidden playground system keys, disabled
// or expired keys, and keys with no group are excluded. The raw key material
// is never returned.
func (s *APIKeyService) ListActivePersonalBoundGroupIDs(ctx context.Context, userID int64) (map[int64]struct{}, error) {
	if s == nil {
		return nil, fmt.Errorf("api key service is not configured")
	}
	if userID <= 0 {
		return map[int64]struct{}{}, nil
	}
	lister, ok := s.apiKeyRepo.(apiKeyAllByUserIDLister)
	if !ok {
		return nil, fmt.Errorf("list api keys: repository does not support listing all keys")
	}
	keys, err := lister.ListAllByUserID(ctx, userID, userFacingAPIKeyListFilters(APIKeyListFilters{}))
	if err != nil {
		return nil, fmt.Errorf("list api keys: %w", err)
	}
	out := make(map[int64]struct{})
	for i := range keys {
		key := keys[i]
		if IsPlaygroundAPIKeyName(key.Name) {
			continue
		}
		if !key.IsActive() || key.IsExpired() {
			continue
		}
		if key.GroupID == nil || *key.GroupID <= 0 {
			continue
		}
		out[*key.GroupID] = struct{}{}
	}
	return out, nil
}
func normalizePlaygroundListFilters(filters APIKeyListFilters) APIKeyListFilters {
	if !filters.IncludePlaygroundKeys {
		filters.ExcludePlaygroundKeys = true
	}
	return filters
}

func isReservedPlaygroundKeyName(name string) bool {
	return IsPlaygroundAPIKeyName(strings.TrimSpace(name))
}
