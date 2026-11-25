package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClassifySupportedModelKinds(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		model string
		want  []string
	}{
		{name: "chat text", model: "gpt-4o", want: []string{SupportedModelKindText}},
		{name: "claude text", model: "claude-sonnet-4", want: []string{SupportedModelKindText}},
		{name: "gemini text is not image", model: "gemini-2.5-flash", want: []string{SupportedModelKindText}},
		{name: "tts residual text", model: "gpt-4o-mini-tts", want: []string{SupportedModelKindText}},
		{name: "embedding residual text", model: "text-embedding-3-large", want: []string{SupportedModelKindText}},

		{name: "gpt-image family", model: "gpt-image-2", want: []string{SupportedModelKindImage}},
		{name: "gpt-image with vendor prefix", model: "openai/gpt-image-1", want: []string{SupportedModelKindImage}},
		{name: "dall-e family", model: "dall-e-3", want: []string{SupportedModelKindImage}},
		{name: "grok imagine default image", model: "grok-imagine", want: []string{SupportedModelKindImage}},
		{name: "grok imagine edit", model: "grok-imagine-edit", want: []string{SupportedModelKindImage}},
		{name: "grok imagine image", model: "grok-imagine-image-quality", want: []string{SupportedModelKindImage}},
		{name: "gemini image suffix", model: "gemini-2.5-flash-image", want: []string{SupportedModelKindImage}},
		{name: "gemini image-generation", model: "gemini-2.0-flash-exp-image-generation", want: []string{SupportedModelKindImage}},
		{name: "imagen family", model: "imagen-3.0-generate-002", want: []string{SupportedModelKindImage}},
		{name: "imagen with models prefix", model: "models/imagen-4.0-generate-preview", want: []string{SupportedModelKindImage}},

		{name: "grok imagine video", model: "grok-imagine-video", want: []string{SupportedModelKindVideo}},
		{name: "grok imagine video 1.5", model: "grok-imagine-video-1.5", want: []string{SupportedModelKindVideo}},
		{name: "grok video alias", model: "grok-video-1.5", want: []string{SupportedModelKindVideo}},
		{name: "sora family", model: "sora-2", want: []string{SupportedModelKindVideo}},
		{name: "sora with vendor prefix", model: "openai/sora", want: []string{SupportedModelKindVideo}},
		{name: "veo family", model: "veo-3.0-generate-preview", want: []string{SupportedModelKindVideo}},
		{name: "veo2 compact", model: "veo2", want: []string{SupportedModelKindVideo}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.want, classifySupportedModelKinds(tc.model))
		})
	}
}

func TestSupportedModelKindsFromFlags_CrossTypeRule(t *testing.T) {
	t.Parallel()
	// No in-tree family currently matches both an image detector and a video
	// detector (gpt-image-*/dall-e-*/imagen*/grok-imagine-image* vs
	// sora*/veo*/grok-imagine-video*). If both independent flags fire, both
	// kinds are emitted and text is omitted.
	require.Equal(t, []string{SupportedModelKindText}, supportedModelKindsFromFlags(false, false))
	require.Equal(t, []string{SupportedModelKindImage}, supportedModelKindsFromFlags(true, false))
	require.Equal(t, []string{SupportedModelKindVideo}, supportedModelKindsFromFlags(false, true))
	require.Equal(t, []string{SupportedModelKindImage, SupportedModelKindVideo}, supportedModelKindsFromFlags(true, true))
	require.True(t, isSupportedImageGenerationModel("gpt-image-2"))
	require.False(t, isSupportedVideoGenerationModel("gpt-image-2"))
	require.True(t, isSupportedVideoGenerationModel("sora-2"))
	require.False(t, isSupportedImageGenerationModel("sora-2"))
}

func TestClassifySupportedModelCompany(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		model string
		want  string
	}{
		{name: "gpt", model: "gpt-4o", want: SupportedModelCompanyOpenAI},
		{name: "codex", model: "codex-mini", want: SupportedModelCompanyOpenAI},
		{name: "dall-e", model: "dall-e-3", want: SupportedModelCompanyOpenAI},
		{name: "sora", model: "sora-2", want: SupportedModelCompanyOpenAI},
		{name: "o-series", model: "o3-mini", want: SupportedModelCompanyOpenAI},
		{name: "claude", model: "claude-sonnet-4", want: SupportedModelCompanyAnthropic},
		{name: "anthropic prefix", model: "anthropic/claude-opus-4", want: SupportedModelCompanyAnthropic},
		{name: "gemini", model: "gemini-2.5-pro", want: SupportedModelCompanyGoogle},
		{name: "gemma", model: "gemma-3-27b", want: SupportedModelCompanyGoogle},
		{name: "imagen", model: "imagen-3.0-generate-002", want: SupportedModelCompanyGoogle},
		{name: "veo", model: "veo-3.1", want: SupportedModelCompanyGoogle},
		{name: "google prefix", model: "google/gemini-2.5-flash", want: SupportedModelCompanyGoogle},
		{name: "grok", model: "grok-4.5", want: SupportedModelCompanyXAI},
		{name: "grok imagine", model: "grok-imagine-video", want: SupportedModelCompanyXAI},
		{name: "xai prefix", model: "xai/grok-4", want: SupportedModelCompanyXAI},
		{name: "deepseek", model: "deepseek-v4-pro", want: SupportedModelCompanyDeepSeek},
		{name: "kimi", model: "kimi-k2", want: SupportedModelCompanyMoonshot},
		{name: "kimi bare id", model: "k3", want: SupportedModelCompanyMoonshot},
		{name: "glm", model: "glm-4.6", want: SupportedModelCompanyZhipu},
		{name: "minimax", model: "minimax-m2", want: SupportedModelCompanyMiniMax},
		{name: "qwen", model: "qwen3-235b-a22b", want: SupportedModelCompanyAlibaba},
		{name: "llama", model: "llama-3.1-405b", want: SupportedModelCompanyMeta},
		{name: "mistral", model: "mistral-large", want: SupportedModelCompanyMistral},
		{name: "doubao", model: "doubao-pro-32k", want: SupportedModelCompanyByteDance},
		{name: "hunyuan", model: "hunyuan-t1", want: SupportedModelCompanyTencent},
		{name: "ernie", model: "ernie-4.0", want: SupportedModelCompanyBaidu},
		{name: "baichuan", model: "baichuan-13b", want: SupportedModelCompanyBaichuan},

		// Gateway / ambiguous / unknown IDs stay in "other".
		{name: "jev gateway model", model: "jev-latest", want: SupportedModelCompanyOther},
		{name: "custom alias", model: "my-custom-relay-model", want: SupportedModelCompanyOther},
		{name: "empty", model: "  ", want: SupportedModelCompanyOther},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.want, classifySupportedModelCompany(tc.model))
		})
	}
}

func TestClassifySupportedModelDoesNotUseGatewayAsCompany(t *testing.T) {
	t.Parallel()

	// Even when DetectModelPlatform would return a gateway platform, company
	// stays "other" unless a vendor family is visible in the ID itself.
	require.Equal(t, SupportedModelCompanyAnthropic, classifySupportedModelCompany("claude-sonnet-4"))
	kinds, company := classifySupportedModel("claude-sonnet-4")
	require.Equal(t, []string{SupportedModelKindText}, kinds)
	require.Equal(t, SupportedModelCompanyAnthropic, company)
}
