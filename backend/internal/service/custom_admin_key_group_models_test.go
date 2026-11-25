//go:build unit

package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type scopedGroupModelsRepo struct {
	GroupRepository
	config GroupModelAllowlist
	writes int
	err    error
}

func (r *scopedGroupModelsRepo) ScopedAdminUpdateGroupModels(_ context.Context, id int64, cfg GroupModelAllowlist) (*Group, error) {
	r.writes++
	r.config = cfg
	return &Group{ID: id, ModelAllowlist: cfg}, r.err
}

func TestScopedAdminGroupModelsNormalizeAndInvalidateOnlyOnSuccess(t *testing.T) {
	repo := &scopedGroupModelsRepo{}
	cache := &authCacheInvalidatorStub{}
	modelsCache := &modelsCacheInvalidatorSpy{}
	s := &adminServiceImpl{groupRepo: repo, authCacheInvalidator: cache, modelsCacheInvalidator: modelsCache}
	cfg := GroupModelAllowlist{Enabled: true, Models: []string{" second ", "first", "second"}}
	g, err := s.ScopedAdminUpdateGroupModels(context.Background(), 2, cfg)
	require.NoError(t, err)
	require.Equal(t, []string{"second", "first"}, g.ModelAllowlist.Models)
	require.Equal(t, []int64{2}, cache.groupIDs)
	require.Equal(t, []int64{2}, modelsCache.groupIDs, "保存分组模型后必须失效模型列表缓存")
	repo.err = errors.New("提交失败")
	_, err = s.ScopedAdminUpdateGroupModels(context.Background(), 3, cfg)
	require.Error(t, err)
	require.Equal(t, []int64{2}, cache.groupIDs, "仓储失败不得提前清理缓存")
	require.Equal(t, []int64{2}, modelsCache.groupIDs, "仓储失败不得提前失效模型列表缓存")
	writes := repo.writes
	_, err = s.ScopedAdminUpdateGroupModels(context.Background(), 2, GroupModelAllowlist{Enabled: true})
	require.Error(t, err, "上游禁止启用空白名单，不能变成仅展示配置")
	for _, models := range [][]string{{""}, {"   "}, {strings.Repeat("x", 513)}, make([]string, 1025)} {
		_, err = s.ScopedAdminUpdateGroupModels(context.Background(), 2, GroupModelAllowlist{Models: models})
		require.Error(t, err)
	}
	require.Equal(t, writes, repo.writes, "无效模型必须在写库前拒绝")
}

type modelsCacheInvalidatorSpy struct {
	groupIDs []int64
}

func (s *modelsCacheInvalidatorSpy) InvalidateAvailableModelsCache(groupID *int64, platform string) {
	if s == nil || groupID == nil {
		return
	}
	s.groupIDs = append(s.groupIDs, *groupID)
}
