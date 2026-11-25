//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type modelsCacheAccountRepo struct {
	AccountRepository
	nextID   int64
	byID     map[int64]*Account
	bindings map[int64][]int64
	shadows  map[int64][]*Account
}

func newModelsCacheAccountRepo(accounts ...*Account) *modelsCacheAccountRepo {
	repo := &modelsCacheAccountRepo{
		nextID:   1,
		byID:     map[int64]*Account{},
		bindings: map[int64][]int64{},
		shadows:  map[int64][]*Account{},
	}
	for _, account := range accounts {
		clone := *account
		if clone.ID == 0 {
			clone.ID = repo.nextID
			repo.nextID++
		}
		if clone.ID >= repo.nextID {
			repo.nextID = clone.ID + 1
		}
		repo.byID[clone.ID] = &clone
		repo.bindings[clone.ID] = append([]int64(nil), clone.GroupIDs...)
	}
	return repo
}

func (r *modelsCacheAccountRepo) snapshot(account *Account) *Account {
	if account == nil {
		return nil
	}
	clone := *account
	clone.GroupIDs = append([]int64(nil), r.bindings[account.ID]...)
	if account.Credentials != nil {
		clone.Credentials = map[string]any{}
		for key, value := range account.Credentials {
			clone.Credentials[key] = value
		}
	}
	return &clone
}

func (r *modelsCacheAccountRepo) Create(_ context.Context, account *Account) error {
	if account.ID == 0 {
		account.ID = r.nextID
		r.nextID++
	}
	r.byID[account.ID] = r.snapshot(account)
	r.byID[account.ID].GroupIDs = append([]int64(nil), account.GroupIDs...)
	r.bindings[account.ID] = append([]int64(nil), account.GroupIDs...)
	return nil
}

func (r *modelsCacheAccountRepo) GetByID(_ context.Context, id int64) (*Account, error) {
	account, ok := r.byID[id]
	if !ok {
		return nil, ErrAccountNotFound
	}
	return r.snapshot(account), nil
}

func (r *modelsCacheAccountRepo) GetByIDs(_ context.Context, ids []int64) ([]*Account, error) {
	out := make([]*Account, 0, len(ids))
	for _, id := range ids {
		account, ok := r.byID[id]
		if !ok {
			continue
		}
		out = append(out, r.snapshot(account))
	}
	return out, nil
}

func (r *modelsCacheAccountRepo) Update(_ context.Context, account *Account) error {
	if _, ok := r.byID[account.ID]; !ok {
		return ErrAccountNotFound
	}
	stored := r.snapshot(account)
	stored.GroupIDs = append([]int64(nil), r.bindings[account.ID]...)
	r.byID[account.ID] = stored
	return nil
}

func (r *modelsCacheAccountRepo) BindGroups(_ context.Context, accountID int64, groupIDs []int64) error {
	account, ok := r.byID[accountID]
	if !ok {
		return ErrAccountNotFound
	}
	r.bindings[accountID] = append([]int64(nil), groupIDs...)
	account.GroupIDs = append([]int64(nil), groupIDs...)
	return nil
}

func (r *modelsCacheAccountRepo) Delete(_ context.Context, id int64) error {
	if _, ok := r.byID[id]; !ok {
		return ErrAccountNotFound
	}
	delete(r.byID, id)
	delete(r.bindings, id)
	return nil
}

func (r *modelsCacheAccountRepo) BulkUpdate(_ context.Context, ids []int64, updates AccountBulkUpdate) (int64, error) {
	var count int64
	for _, id := range ids {
		account, ok := r.byID[id]
		if !ok {
			continue
		}
		if len(updates.Credentials) > 0 {
			if account.Credentials == nil {
				account.Credentials = map[string]any{}
			}
			for key, value := range updates.Credentials {
				account.Credentials[key] = value
			}
		}
		count++
	}
	return count, nil
}

func (r *modelsCacheAccountRepo) ListShadowsByParent(_ context.Context, parentID int64) ([]*Account, error) {
	shadows := r.shadows[parentID]
	out := make([]*Account, 0, len(shadows))
	for _, shadow := range shadows {
		out = append(out, r.snapshot(shadow))
	}
	return out, nil
}

type modelsCacheGroupRepo struct {
	GroupRepository
	byID map[int64]*Group
}

func (r *modelsCacheGroupRepo) GetByID(_ context.Context, id int64) (*Group, error) {
	group, ok := r.byID[id]
	if !ok {
		return nil, ErrGroupNotFound
	}
	clone := *group
	return &clone, nil
}

func testModelsCacheAdmin(repo *modelsCacheAccountRepo, groups ...int64) (*adminServiceImpl, *modelsCacheInvalidatorSpy) {
	groupByID := map[int64]*Group{}
	for _, id := range groups {
		groupByID[id] = &Group{ID: id, Name: "g", Platform: PlatformOpenAI, Status: StatusActive}
	}
	spy := &modelsCacheInvalidatorSpy{}
	return &adminServiceImpl{
		accountRepo:            repo,
		groupRepo:              &modelsCacheGroupRepo{byID: groupByID},
		modelsCacheInvalidator: spy,
	}, spy
}

func requireInvalidatedGroups(t *testing.T, spy *modelsCacheInvalidatorSpy, want ...int64) {
	t.Helper()
	require.ElementsMatch(t, want, spy.groupIDs)
}

func TestCreateAccountInvalidatesBoundGroupModelsCache(t *testing.T) {
	repo := newModelsCacheAccountRepo()
	svc, spy := testModelsCacheAdmin(repo, 11, 12)

	account, err := svc.CreateAccount(context.Background(), &CreateAccountInput{
		Name:                  "test",
		Platform:              PlatformOpenAI,
		Type:                  AccountTypeAPIKey,
		Credentials:           map[string]any{"api_key": "k", "model_mapping": map[string]any{"gpt-4o": "gpt-4o"}},
		GroupIDs:              []int64{11, 12},
		SkipDefaultGroupBind:  true,
		SkipMixedChannelCheck: true,
	})
	require.NoError(t, err)
	require.NotZero(t, account.ID)
	requireInvalidatedGroups(t, spy, 11, 12)
}

func TestUpdateAccountInvalidatesOldAndNewGroupModelsCache(t *testing.T) {
	repo := newModelsCacheAccountRepo(&Account{
		ID:          5,
		Name:        "test",
		Platform:    PlatformOpenAI,
		Type:        AccountTypeAPIKey,
		GroupIDs:    []int64{11},
		Credentials: map[string]any{"api_key": "k", "model_mapping": map[string]any{"old": "old"}},
	})
	svc, spy := testModelsCacheAdmin(repo, 11, 12)
	newGroups := []int64{12}

	_, err := svc.UpdateAccount(context.Background(), 5, &UpdateAccountInput{
		Credentials:           map[string]any{"model_mapping": map[string]any{"new": "new"}},
		GroupIDs:              &newGroups,
		SkipMixedChannelCheck: true,
	})
	require.NoError(t, err)
	requireInvalidatedGroups(t, spy, 11, 12)
}

func TestBulkUpdateAccountsInvalidatesBoundGroupModelsCache(t *testing.T) {
	repo := newModelsCacheAccountRepo(&Account{
		ID:          5,
		Name:        "test",
		Platform:    PlatformOpenAI,
		Type:        AccountTypeAPIKey,
		GroupIDs:    []int64{11},
		Credentials: map[string]any{"api_key": "k"},
	})
	svc, spy := testModelsCacheAdmin(repo, 11, 13)
	newGroups := []int64{13}

	_, err := svc.BulkUpdateAccounts(context.Background(), &BulkUpdateAccountsInput{
		AccountIDs:            []int64{5},
		Credentials:           map[string]any{"model_mapping": map[string]any{"bulk": "bulk"}},
		GroupIDs:              &newGroups,
		SkipMixedChannelCheck: true,
	})
	require.NoError(t, err)
	requireInvalidatedGroups(t, spy, 11, 13)
}

func TestDeleteAccountInvalidatesBoundGroupModelsCache(t *testing.T) {
	parent := &Account{ID: 5, Name: "parent", Platform: PlatformOpenAI, Type: AccountTypeOAuth, GroupIDs: []int64{11}}
	shadow := &Account{ID: 6, Name: "shadow", Platform: PlatformOpenAI, Type: AccountTypeOAuth, GroupIDs: []int64{12}}
	repo := newModelsCacheAccountRepo(parent, shadow)
	repo.shadows[5] = []*Account{shadow}
	svc, spy := testModelsCacheAdmin(repo, 11, 12)

	require.NoError(t, svc.DeleteAccount(context.Background(), 5))
	requireInvalidatedGroups(t, spy, 11, 12)
}

func TestInvalidateAvailableModelsCacheForAccountUsesBoundGroups(t *testing.T) {
	svc, spy := testModelsCacheAdmin(newModelsCacheAccountRepo(), 11, 12)
	svc.InvalidateAvailableModelsCacheForAccount(&Account{ID: 1, GroupIDs: []int64{11, 11, 0, 12}})
	requireInvalidatedGroups(t, spy, 11, 12)
}
