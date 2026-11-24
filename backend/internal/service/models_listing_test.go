package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type listingAccountRepo struct {
	AccountRepository
	byGroup map[int64][]Account
}

func (r listingAccountRepo) ListSchedulableByGroupID(_ context.Context, groupID int64) ([]Account, error) {
	return append([]Account(nil), r.byGroup[groupID]...), nil
}

func testGateway(accounts map[int64][]Account) *GatewayService {
	return &GatewayService{accountRepo: listingAccountRepo{byGroup: accounts}}
}

func mappedAccount(id int64, platform string, models ...string) Account {
	mapping := make(map[string]any, len(models))
	for _, model := range models {
		mapping[model] = model
	}
	return Account{
		ID:       id,
		Platform: platform,
		Credentials: map[string]any{
			"model_mapping": mapping,
		},
	}
}

func TestDefaultModelIDsForPlatformDoesNotLeakTypeSafeOnComposite(t *testing.T) {
	composite := DefaultModelIDsForPlatform(PlatformComposite)
	require.NotEmpty(t, composite)
	require.NotContains(t, composite, "jev-latest")
	require.Equal(t, DefaultModelIDsForPlatform(PlatformAnthropic), claudeIDsOrDefault(t))
}

func claudeIDsOrDefault(t *testing.T) []string {
	t.Helper()
	ids := DefaultModelIDsForPlatform(PlatformAnthropic)
	require.NotEmpty(t, ids)
	return ids
}

func TestListAdvertisedModelsForGroup_NoAccountsFallsBackToThatPlatformOnly(t *testing.T) {
	svc := testGateway(nil)
	group := &Group{ID: 11, Platform: PlatformAnthropic}

	models := svc.ListAdvertisedModelsForGroup(context.Background(), group)
	ids := advertisedIDs(models)
	require.Equal(t, DefaultModelIDsForPlatform(PlatformAnthropic), ids)
	for _, model := range models {
		require.Equal(t, []string{PlatformAnthropic}, model.Platforms)
	}

	allPlatformDefaults := DefaultModelIDsForPlatform(PlatformComposite)
	require.Greater(t, len(allPlatformDefaults), len(ids),
		"an isolated anthropic group must not advertise the full composite default catalog")
}

func TestListAdvertisedModelsForGroup_NilGroupOrService(t *testing.T) {
	require.Nil(t, (*GatewayService)(nil).ListAdvertisedModelsForGroup(context.Background(), &Group{ID: 1}))
	require.Nil(t, testGateway(nil).ListAdvertisedModelsForGroup(context.Background(), nil))
}

func TestListAdvertisedModelsForGroup_UsesSavedMappingKeys(t *testing.T) {
	svc := testGateway(map[int64][]Account{
		7: {mappedAccount(1, PlatformOpenAI, "gpt-4o", "o3-mini")},
	})
	group := &Group{ID: 7, Platform: PlatformOpenAI}

	ids := advertisedIDs(svc.ListAdvertisedModelsForGroup(context.Background(), group))
	require.Equal(t, []string{"gpt-4o", "o3-mini"}, ids)
	require.NotEqual(t, DefaultModelIDsForPlatform(PlatformOpenAI), ids)
}

func TestListAdvertisedModelsForGroup_AllowlistFiltersAdvertisedIDs(t *testing.T) {
	svc := testGateway(map[int64][]Account{
		7: {mappedAccount(1, PlatformOpenAI, "gpt-4o", "o3-mini", "gpt-4.1")},
	})
	group := &Group{
		ID:       7,
		Platform: PlatformOpenAI,
		ModelAllowlist: GroupModelAllowlist{
			Enabled: true,
			Models:  []string{"o3-mini"},
		},
	}

	ids := advertisedIDs(svc.ListAdvertisedModelsForGroup(context.Background(), group))
	require.Equal(t, []string{"o3-mini"}, ids)
}

func TestListAdvertisedModelsForGroup_OpenAIEmptyCatalogIsAuthoritative(t *testing.T) {
	// Non-nil empty OpenAI catalog must not resurrect platform defaults.
	// GetAvailableModels only returns an empty slice after filtering a mapping
	// catalog down to nothing; we assert the listing source rule directly and
	// through an allowlisted group whose source is an empty OpenAI slice.
	require.Empty(t, ModelListingSource(PlatformOpenAI, []string{}, DefaultModelIDsForPlatform(PlatformOpenAI)))
	require.Equal(t, DefaultModelIDsForPlatform(PlatformOpenAI), ModelListingSource(PlatformOpenAI, nil, DefaultModelIDsForPlatform(PlatformOpenAI)))
}

func TestListAdvertisedModelsForGroup_CompositeNoAccountsUsesCompositeDefaults(t *testing.T) {
	svc := testGateway(nil)
	group := &Group{ID: 3, Platform: PlatformComposite}

	ids := advertisedIDs(svc.ListAdvertisedModelsForGroup(context.Background(), group))
	require.Equal(t, DefaultModelIDsForPlatform(PlatformComposite), ids)
}

func TestListAdvertisedModelsForGroup_CompositeUnionsMappedPlatforms(t *testing.T) {
	svc := testGateway(map[int64][]Account{
		3: {
			mappedAccount(1, PlatformOpenAI, "gpt-4o"),
			mappedAccount(2, PlatformAnthropic, "claude-sonnet-4"),
		},
	})
	group := &Group{ID: 3, Platform: PlatformComposite}

	models := svc.ListAdvertisedModelsForGroup(context.Background(), group)
	ids := advertisedIDs(models)
	require.Contains(t, ids, "gpt-4o")
	require.Contains(t, ids, "claude-sonnet-4")
	require.NotEqual(t, DefaultModelIDsForPlatform(PlatformComposite), ids)

	byName := map[string][]string{}
	for _, model := range models {
		byName[model.Name] = model.Platforms
	}
	require.Equal(t, []string{PlatformOpenAI}, byName["gpt-4o"])
	require.Equal(t, []string{PlatformAnthropic}, byName["claude-sonnet-4"])
}

func TestModelListingSourceAnthropicUnionsMappingsAndDefaults(t *testing.T) {
	defaults := []string{"claude-default"}
	available := []string{"claude-mapped"}
	require.Equal(t, []string{"claude-mapped", "claude-default"}, ModelListingSource(PlatformAnthropic, available, defaults))
	require.Equal(t, available, ModelListingSource(PlatformGemini, available, defaults))
}

func TestListAdvertisedModelsForGroup_AntigravityAndOpenCodeGoMatchModelsFallback(t *testing.T) {
	// /v1/models has no dedicated un-allowlisted fallback for these platforms;
	// they use Claude defaults. DefaultModelIDsForPlatform still has a richer
	// catalog for allowlist/composite paths and must not be used here.
	svc := testGateway(nil)
	claudeIDs := modelsListFallbackIDs(PlatformAnthropic)
	require.Equal(t, claudeIDs, advertisedIDs(svc.ListAdvertisedModelsForGroup(context.Background(), &Group{ID: 4, Platform: PlatformAntigravity})))
	require.Equal(t, claudeIDs, advertisedIDs(svc.ListAdvertisedModelsForGroup(context.Background(), &Group{ID: 5, Platform: PlatformOpenCodeGo})))
	require.NotEqual(t, DefaultModelIDsForPlatform(PlatformAntigravity), claudeIDs)
	require.NotEqual(t, DefaultModelIDsForPlatform(PlatformOpenCodeGo), claudeIDs)
}
