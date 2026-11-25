//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

func TestAPIKeyServiceListExcludesPlaygroundKeysAndIgnoresInclude(t *testing.T) {
	repo := &apiKeyRepoStub{
		allowListByUserID: true,
		listByUserIDKeys: []APIKey{
			{ID: 1, UserID: 7, Key: "sk-user", Name: "production"},
			{ID: 2, UserID: 7, Key: "sk-play", Name: PlaygroundAPIKeyName(5)},
		},
	}
	svc := &APIKeyService{apiKeyRepo: repo}

	keys, result, err := svc.List(context.Background(), 7, pagination.PaginationParams{Page: 1, PageSize: 20}, APIKeyListFilters{IncludePlaygroundKeys: true})
	require.NoError(t, err)
	require.Len(t, keys, 1)
	require.Equal(t, int64(1), keys[0].ID)
	require.Equal(t, "production", keys[0].Name)
	require.NotEmpty(t, repo.listByUserIDFilters)
	require.True(t, repo.listByUserIDFilters[0].ExcludePlaygroundKeys)
	require.False(t, repo.listByUserIDFilters[0].IncludePlaygroundKeys)
	require.Equal(t, int64(1), result.Total)
}

func TestAPIKeyServiceGetByIDForUserHidesPlaygroundKey(t *testing.T) {
	repo := &apiKeyRepoStub{
		apiKey: &APIKey{ID: 2, UserID: 7, Key: "sk-secret-playground", Name: PlaygroundAPIKeyName(5)},
	}
	svc := &APIKeyService{apiKeyRepo: repo}

	key, err := svc.GetByIDForUser(context.Background(), 2, 7)
	require.ErrorIs(t, err, ErrAPIKeyNotFound)
	require.Nil(t, key)

	visible, err := svc.GetByID(context.Background(), 2)
	require.NoError(t, err)
	require.Equal(t, "sk-secret-playground", visible.Key)
}

func TestAPIKeyServiceUpdateRejectsPlaygroundKeyRename(t *testing.T) {
	repo := &apiKeyRepoStub{
		apiKey: &APIKey{ID: 2, UserID: 7, Key: "sk-play", Name: PlaygroundAPIKeyName(5), Status: StatusActive},
	}
	svc := &APIKeyService{apiKeyRepo: repo}
	name := "my-key"

	key, err := svc.Update(context.Background(), 2, 7, UpdateAPIKeyRequest{Name: &name})
	require.ErrorIs(t, err, ErrAPIKeyNotFound)
	require.Nil(t, key)
	require.Empty(t, repo.updatedKeys)
}

func TestAPIKeyServiceDeleteRejectsPlaygroundKey(t *testing.T) {
	repo := &apiKeyRepoStub{
		apiKey: &APIKey{ID: 2, UserID: 7, Key: "sk-play", Name: PlaygroundAPIKeyName(5)},
	}
	cache := &apiKeyCacheStub{}
	svc := &APIKeyService{apiKeyRepo: repo, cache: cache}

	err := svc.Delete(context.Background(), 2, 7)
	require.ErrorIs(t, err, ErrAPIKeyNotFound)
	require.Empty(t, repo.deletedIDs)
	require.Empty(t, cache.deleteAuthKeys)
}

func TestFindPlaygroundAPIKeyRequiresMatchingGroup(t *testing.T) {
	group8 := int64(8)
	repo := &apiKeyRepoStub{
		allowListAllByUserID: true,
		listAllByUserIDKeys: []APIKey{
			{ID: 11, UserID: 7, Key: "sk-mismatch", Name: PlaygroundAPIKeyName(5), GroupID: &group8},
			{ID: 12, UserID: 7, Key: "sk-ok", Name: PlaygroundAPIKeyName(8), GroupID: &group8},
		},
	}
	svc := &APIKeyService{apiKeyRepo: repo}

	// Name says group 5, but the key is bound to group 8 — must not be reused for group 5.
	match, err := svc.findPlaygroundAPIKey(context.Background(), 7, 5, PlaygroundAPIKeyName(5))
	require.NoError(t, err)
	require.Nil(t, match)

	match, err = svc.findPlaygroundAPIKey(context.Background(), 7, 8, PlaygroundAPIKeyName(8))
	require.NoError(t, err)
	require.NotNil(t, match)
	require.Equal(t, int64(12), match.ID)
	require.NotNil(t, match.GroupID)
	require.Equal(t, group8, *match.GroupID)
}

func TestEnsurePlaygroundAPIKeyRepairsUnsetGroupAndIgnoresMismatchedGroup(t *testing.T) {
	group5 := int64(5)
	group8 := int64(8)
	repo := &apiKeyRepoStub{
		apiKey:               &APIKey{ID: 21, UserID: 7, Key: "sk-repair", Name: PlaygroundAPIKeyName(5), GroupID: nil, Status: StatusActive},
		allowListAllByUserID: true,
		listAllByUserIDKeys: []APIKey{
			{ID: 20, UserID: 7, Key: "sk-wrong", Name: PlaygroundAPIKeyName(5), GroupID: &group8, Status: StatusActive},
			{ID: 21, UserID: 7, Key: "sk-repair", Name: PlaygroundAPIKeyName(5), GroupID: nil, Status: StatusActive},
		},
	}
	svc := &APIKeyService{apiKeyRepo: repo}

	key, err := svc.EnsurePlaygroundAPIKey(context.Background(), 7, group5)
	require.NoError(t, err)
	require.NotNil(t, key)
	require.Equal(t, int64(21), key.ID)
	require.NotNil(t, key.GroupID)
	require.Equal(t, group5, *key.GroupID)
	require.Len(t, repo.updatedKeys, 1)
	require.Equal(t, group5, *repo.updatedKeys[0].GroupID)
}
