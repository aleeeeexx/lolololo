package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type stubGroupLister struct {
	groups []Group
	err    error
	userID int64
	calls  int
}

func (s *stubGroupLister) GetAvailableGroups(_ context.Context, userID int64) ([]Group, error) {
	s.calls++
	s.userID = userID
	if s.err != nil {
		return nil, s.err
	}
	return append([]Group(nil), s.groups...), nil
}

type stubAdvertiser struct {
	byGroup map[int64][]AdvertisedModel
	calls   []int64
}

func (s *stubAdvertiser) ListAdvertisedModelsForGroup(_ context.Context, group *Group) []AdvertisedModel {
	if group == nil {
		return nil
	}
	s.calls = append(s.calls, group.ID)
	return append([]AdvertisedModel(nil), s.byGroup[group.ID]...)
}

type explodingAdvertiser struct {
	t *testing.T
}

func (s explodingAdvertiser) ListAdvertisedModelsForGroup(context.Context, *Group) []AdvertisedModel {
	s.t.Fatal("listing should not run when the user has no bindable groups")
	return nil
}

func TestSupportedModelsService_EmptyWhenNoBindableGroups(t *testing.T) {
	groups := &stubGroupLister{}
	svc := NewSupportedModelsService(groups, explodingAdvertiser{t: t})

	models, err := svc.ListForUser(context.Background(), 42)
	require.NoError(t, err)
	require.Equal(t, []SupportedModelItem{}, models)
	require.Equal(t, 1, groups.calls)
	require.Equal(t, int64(42), groups.userID)
}

func TestSupportedModelsService_OnlyListsAuthorizedGroups(t *testing.T) {
	groups := &stubGroupLister{groups: []Group{
		{ID: 1, Platform: PlatformOpenAI},
		{ID: 3, Platform: PlatformAnthropic},
	}}
	listing := &stubAdvertiser{byGroup: map[int64][]AdvertisedModel{
		1: {{Name: "gpt-4o", Platforms: []string{PlatformOpenAI}}},
		2: {{Name: "secret-hidden-model", Platforms: []string{PlatformOpenAI}}},
		3: {{Name: "claude-sonnet-4", Platforms: []string{PlatformAnthropic}}, {Name: "gpt-4o", Platforms: []string{PlatformAnthropic}}},
	}}
	svc := NewSupportedModelsService(groups, listing)

	models, err := svc.ListForUser(context.Background(), 7)
	require.NoError(t, err)
	require.Equal(t, []SupportedModelItem{
		{Name: "claude-sonnet-4", Platforms: []string{PlatformAnthropic}},
		{Name: "gpt-4o", Platforms: []string{PlatformAnthropic, PlatformOpenAI}},
	}, models)
	require.Equal(t, []int64{1, 3}, listing.calls)
	for _, model := range models {
		require.NotEqual(t, "secret-hidden-model", model.Name)
	}
}

func TestSupportedModelsService_ResponseContainsOnlySafeFields(t *testing.T) {
	svc := NewSupportedModelsService(
		&stubGroupLister{groups: []Group{{ID: 9, Platform: PlatformOpenAI, Name: "hidden-group"}}},
		&stubAdvertiser{byGroup: map[int64][]AdvertisedModel{
			9: {{Name: "gpt-4o", Platforms: []string{PlatformOpenAI}}},
		}},
	)
	models, err := svc.ListForUser(context.Background(), 1)
	require.NoError(t, err)

	raw, err := json.Marshal(models)
	require.NoError(t, err)
	body := strings.ToLower(string(raw))
	require.JSONEq(t, `[{"name":"gpt-4o","platforms":["openai"]}]`, string(raw))
	for _, leaked := range []string{
		"account_id", "accountid", "api_key", "apikey", "secret", "password",
		"credential", "token", "hidden-group", `"id":9`,
	} {
		require.NotContains(t, body, leaked)
	}
}

func TestSupportedModelsService_HasNoSettingOrUpstreamDependencies(t *testing.T) {
	// The service only depends on bindable-group lookup and saved listing.
	// Feature flags and upstream sync are intentionally not in the struct.
	svc := NewSupportedModelsService(&stubGroupLister{}, &stubAdvertiser{})
	require.NotNil(t, svc.groups)
	require.NotNil(t, svc.listing)
}
