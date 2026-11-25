package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
)

func TestIsPlaygroundAPIKeyName(t *testing.T) {
	require.True(t, IsPlaygroundAPIKeyName(PlaygroundAPIKeyName(12)))
	require.Equal(t, "__playground__:g12", PlaygroundAPIKeyName(12))
	require.False(t, IsPlaygroundAPIKeyName("production"))
	require.False(t, IsPlaygroundAPIKeyName("playground"))
}

func TestValidatePlaygroundModelKind(t *testing.T) {
	require.NoError(t, validatePlaygroundModelKind("gpt-4o", SupportedModelKindText))
	require.Error(t, validatePlaygroundModelKind("dall-e-3", SupportedModelKindText))
	require.NoError(t, validatePlaygroundModelKind("dall-e-3", SupportedModelKindImage))
	require.Error(t, validatePlaygroundModelKind("gpt-4o", SupportedModelKindImage))
	require.Error(t, validatePlaygroundModelKind("sora-2", SupportedModelKindText))
}

func TestGroupSupportsPlaygroundImage(t *testing.T) {
	require.False(t, groupSupportsPlaygroundImage(nil))
	require.True(t, groupSupportsPlaygroundImage(&Group{Platform: PlatformOpenAI, AllowImageGeneration: true}))
	require.True(t, groupSupportsPlaygroundImage(&Group{Platform: PlatformGrok, AllowImageGeneration: true}))
	require.True(t, groupSupportsPlaygroundImage(&Group{Platform: PlatformComposite, AllowImageGeneration: true}))
	require.False(t, groupSupportsPlaygroundImage(&Group{Platform: PlatformOpenAI, AllowImageGeneration: false}))
	require.False(t, groupSupportsPlaygroundImage(&Group{Platform: PlatformGemini, AllowImageGeneration: true}))
	require.False(t, groupSupportsPlaygroundImage(&Group{Platform: PlatformAnthropic, AllowImageGeneration: true}))
}

func TestExtractPlaygroundDocumentsTextAndLimits(t *testing.T) {
	docs, err := ExtractPlaygroundDocuments([]PlaygroundDocument{
		{Filename: "notes.txt", Text: "hello world"},
	})
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.Equal(t, "notes.txt", docs[0].Filename)
	require.Equal(t, "hello world", docs[0].Text)

	_, err = ExtractPlaygroundDocuments([]PlaygroundDocument{
		{Filename: "a.txt", Text: "1"},
		{Filename: "b.txt", Text: "2"},
		{Filename: "c.txt", Text: "3"},
		{Filename: "d.txt", Text: "4"},
	})
	require.Error(t, err)
	var playgroundErr *ErrPlayground
	require.ErrorAs(t, err, &playgroundErr)
	require.Equal(t, "DOCUMENT_LIMIT", playgroundErr.Code)

	_, err = ExtractPlaygroundDocuments([]PlaygroundDocument{
		{Filename: "photo.png", Data: "AAAA"},
	})
	require.Error(t, err)

	long := strings.Repeat("字", PlaygroundMaxDocumentRunes+50)
	truncated, err := ExtractPlaygroundDocuments([]PlaygroundDocument{
		{Filename: "long.txt", Text: long},
	})
	require.NoError(t, err)
	require.True(t, truncated[0].Truncated)
	require.Equal(t, PlaygroundMaxDocumentRunes, utf8.RuneCountInString(truncated[0].Text))
}

func TestExtractPlaygroundDocumentsPDF(t *testing.T) {
	pdf := []byte("%PDF-1.1\n1 0 obj<< /Length 48 >>stream\nBT (Hello playground) Tj ET\nendstream\nendobj\n%%EOF\n")
	docs, err := ExtractPlaygroundDocuments([]PlaygroundDocument{
		{Filename: "note.pdf", Data: base64.StdEncoding.EncodeToString(pdf)},
	})
	require.NoError(t, err)
	require.Contains(t, docs[0].Text, "Hello playground")
}

func TestFormatPlaygroundDocumentPrompt(t *testing.T) {
	prompt := FormatPlaygroundDocumentPrompt([]PlaygroundExtractedDocument{
		{Filename: "a.txt", Text: "alpha"},
		{Filename: "b.pdf", Text: "beta", Truncated: true},
	})
	require.Contains(t, prompt, "a.txt")
	require.Contains(t, prompt, "alpha")
	require.Contains(t, prompt, "b.pdf (truncated)")
	require.Contains(t, prompt, "beta")
}

func TestResolveGroupRequiresIDWhenMultipleGroupsAdvertiseModel(t *testing.T) {
	lister := &stubGroupLister{groups: []Group{
		{ID: 1, Name: "sub", Platform: PlatformOpenAI, SubscriptionType: SubscriptionTypeSubscription, RateMultiplier: 1, AllowImageGeneration: true},
		{ID: 2, Name: "payg", Platform: PlatformOpenAI, RateMultiplier: 1.5, AllowImageGeneration: true},
	}}
	listing := &stubAdvertiser{byGroup: map[int64][]AdvertisedModel{
		1: {{Name: "gpt-4o"}},
		2: {{Name: "gpt-4o"}},
	}}
	svc := &PlaygroundService{groups: lister, listing: listing}

	_, err := svc.resolveGroup(context.Background(), 7, "gpt-4o", SupportedModelKindText, nil)
	require.Error(t, err)
	var playgroundErr *ErrPlayground
	require.ErrorAs(t, err, &playgroundErr)
	require.Equal(t, "GROUP_REQUIRED", playgroundErr.Code)

	groupID := int64(2)
	got, err := svc.resolveGroup(context.Background(), 7, "gpt-4o", SupportedModelKindText, &groupID)
	require.NoError(t, err)
	require.Equal(t, int64(2), got.ID)
}

func TestResolveGroupAutoSelectsWhenOnlyOneGroupAdvertisesModel(t *testing.T) {
	lister := &stubGroupLister{groups: []Group{
		{ID: 1, Name: "sub", Platform: PlatformOpenAI},
		{ID: 2, Name: "other", Platform: PlatformAnthropic},
	}}
	listing := &stubAdvertiser{byGroup: map[int64][]AdvertisedModel{
		1: {{Name: "gpt-4o"}},
		2: {{Name: "claude-sonnet-4"}},
	}}
	svc := &PlaygroundService{groups: lister, listing: listing}

	got, err := svc.resolveGroup(context.Background(), 7, "gpt-4o", SupportedModelKindText, nil)
	require.NoError(t, err)
	require.Equal(t, int64(1), got.ID)
}

func TestIsAllowedPlaygroundImageURLRejectsPrivateAndNonHTTPS(t *testing.T) {
	require.False(t, IsAllowedPlaygroundImageURL("http://127.0.0.1/a.png"))
	require.False(t, IsAllowedPlaygroundImageURL("https://127.0.0.1/a.png"))
	require.False(t, IsAllowedPlaygroundImageURL("https://169.254.169.254/latest/meta-data/"))
	require.False(t, IsAllowedPlaygroundImageURL("http://example.com/a.png"))
	require.False(t, IsAllowedPlaygroundImageURL("https://localhost/a.png"))
	require.False(t, IsAllowedPlaygroundImageURL("https://10.0.0.8/a.png"))
	require.False(t, IsAllowedPlaygroundImageURL("https://192.168.1.1/a.png"))
	require.True(t, IsAllowedPlaygroundImageURL("https://cdn.example.com/a.png"))
}

func TestFetchPlaygroundImageURLRejectsLoopbackAndMetadataWithoutContact(t *testing.T) {
	hits := 0
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		hits++
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader([]byte("ok"))), Request: req}, nil
	})}

	_, _, err := fetchPlaygroundImageURL(context.Background(), "https://127.0.0.1/a.png", client)
	require.Error(t, err)
	_, _, err = fetchPlaygroundImageURL(context.Background(), "https://169.254.169.254/latest/meta-data/", client)
	require.Error(t, err)
	require.Zero(t, hits)
}

func TestFetchPlaygroundImageURLRejectsRedirectToPrivateHost(t *testing.T) {
	hits := 0
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		hits++
		return &http.Response{
			StatusCode: http.StatusFound,
			Header:     http.Header{"Location": []string{"https://169.254.169.254/latest/meta-data/"}},
			Body:       io.NopCloser(bytes.NewReader(nil)),
			Request:    req,
		}, nil
	})}
	client.CheckRedirect = playgroundImageCheckRedirect

	_, _, err := fetchPlaygroundImageURL(context.Background(), "https://images.example.test/a.png", client)
	require.Error(t, err)
	require.Contains(t, err.Error(), "redirect url is not allowed")
	require.Equal(t, 1, hits)
}

func TestExtractPlaygroundDocumentsMarkdown(t *testing.T) {
	docs, err := ExtractPlaygroundDocuments([]PlaygroundDocument{
		{Filename: "notes.md", Text: "# Hello"},
	})
	require.NoError(t, err)
	require.Equal(t, "notes.md", docs[0].Filename)
	require.Equal(t, "# Hello", docs[0].Text)
}

func TestPlaygroundSubscriptionHasRemaining(t *testing.T) {
	require.True(t, playgroundSubscriptionHasRemaining(PlaygroundSubscription{}))
	require.False(t, playgroundSubscriptionHasRemaining(PlaygroundSubscription{DailyLimitUSD: 10, DailyUsedUSD: 10}))
	require.True(t, playgroundSubscriptionHasRemaining(PlaygroundSubscription{DailyLimitUSD: 10, DailyUsedUSD: 9}))
}
