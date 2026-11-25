package service

import "strings"

// Supported-model catalog kinds. A model may belong to more than one kind when
// its ID independently matches multiple generation families.
const (
	SupportedModelKindText  = "text"
	SupportedModelKindImage = "image"
	SupportedModelKindVideo = "video"
)

// Supported-model vendor slugs. These are model vendors, not gateway platforms
// such as Antigravity or OpenCodeGo (those remain in the platforms field).
const (
	SupportedModelCompanyOpenAI    = "openai"
	SupportedModelCompanyAnthropic = "anthropic"
	SupportedModelCompanyGoogle    = "google"
	SupportedModelCompanyXAI       = "xai"
	SupportedModelCompanyDeepSeek  = "deepseek"
	SupportedModelCompanyMoonshot  = "moonshot"
	SupportedModelCompanyZhipu     = "zhipu"
	SupportedModelCompanyMiniMax   = "minimax"
	SupportedModelCompanyAlibaba   = "alibaba"
	SupportedModelCompanyMeta      = "meta"
	SupportedModelCompanyMistral   = "mistral"
	SupportedModelCompanyByteDance = "bytedance"
	SupportedModelCompanyTencent   = "tencent"
	SupportedModelCompanyBaidu     = "baidu"
	SupportedModelCompanyBaichuan  = "baichuan"
	SupportedModelCompanyOther     = "other"
)

// gateway platforms must never be surfaced as a model company.
var supportedModelGatewayPlatforms = map[string]struct{}{
	PlatformAntigravity: {},
	PlatformOpenCodeGo:  {},
	PlatformTypeSafe:    {},
	PlatformCommandCode: {},
	PlatformCline:       {},
	PlatformComposite:   {},
}

// classifySupportedModel infers user-facing kinds and company for a request-side
// model ID. It reuses billing/scheduling family detectors and DetectModelPlatform,
// and never calls upstream or the model catalog sync.
func classifySupportedModel(name string) (kinds []string, company string) {
	return classifySupportedModelKinds(name), classifySupportedModelCompany(name)
}

// classifySupportedModelKinds assigns generation kinds from the model ID only.
//
// Rules:
//   - Image if the ID matches known image-generation families already used for
//     billing/routing: gpt-image-*, dall-e-*, grok-imagine / grok-imagine-edit /
//     grok-imagine-image*, Gemini *-image / *-image-*, imagen*.
//   - Video if the ID matches known video-generation families: grok-imagine-video*
//     / grok-video* (same口径 as video_model_prices and isGrokVideoGenerationModel),
//     sora*, veo*.
//   - Text if the ID matches neither image nor video. TTS, STT, embeddings and
//     moderation therefore appear under text, because this page only offers
//     text / image / video.
//   - A model is tagged with every kind its ID independently matches. Image and
//     video families are not also tagged text. Cross-type entries are possible
//     when one ID matches both an image family and a video family.
func classifySupportedModelKinds(name string) []string {
	bare := supportedModelBareID(name)
	if bare == "" {
		return []string{SupportedModelKindText}
	}

	return supportedModelKindsFromFlags(
		isSupportedImageGenerationModel(bare),
		isSupportedVideoGenerationModel(bare),
	)
}

// supportedModelKindsFromFlags is the cross-type assembly rule: a model receives
// every kind it independently matches. Image and video families are not also
// tagged text. When neither matches, the residual kind is text.
func supportedModelKindsFromFlags(image, video bool) []string {
	if !image && !video {
		return []string{SupportedModelKindText}
	}
	kinds := make([]string, 0, 2)
	if image {
		kinds = append(kinds, SupportedModelKindImage)
	}
	if video {
		kinds = append(kinds, SupportedModelKindVideo)
	}
	return kinds
}

func isSupportedImageGenerationModel(bare string) bool {
	if IsGPTImageGenerationModel(bare) || isGrokImageGenerationModel(bare) || isGeminiCompatibleImageModel(bare) {
		return true
	}
	if bare == "dall-e" || strings.HasPrefix(bare, "dall-e-") || bare == "gpt-image" {
		return true
	}
	// LiteLLM / Vertex image-generation families referenced in billing comments.
	if strings.HasPrefix(bare, "imagen") || strings.HasPrefix(bare, "imagegeneration") {
		return true
	}
	return false
}

func isSupportedVideoGenerationModel(bare string) bool {
	if isGrokVideoGenerationModel(bare) || CanonicalGrokImagineVideoPriceFamily(bare) != "" {
		return true
	}
	// OpenAI Sora and Google Veo are the video families named alongside
	// grok-imagine in the existing scheduling/billing discussion.
	if strings.HasPrefix(bare, "sora") || strings.HasPrefix(bare, "veo") {
		return true
	}
	return false
}

// classifySupportedModelCompany maps a model ID to a vendor slug.
//
// It first reuses DetectModelPlatform (the same prefix table used to route
// composite groups), then remaps provider platforms to vendors (Gemini→google,
// Grok→xai). Gateway platforms (Antigravity, OpenCodeGo, TypeSafe, Command Code,
// Cline) are not companies and fall through to additional ID-family rules or
// "other". Additional families come from prefixes the repo already treats as
// foreign vendor families (qwen, llama, mistral, doubao, hunyuan, ernie,
// baichuan, gemma) plus sora / imagen / veo, which DetectModelPlatform does
// not yet cover. Ambiguous or unknown IDs become "other".
func classifySupportedModelCompany(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return SupportedModelCompanyOther
	}
	if platform, ok := DetectModelPlatform(trimmed); ok {
		if company, ok := companyFromProviderPlatform(platform); ok {
			return company
		}
	}
	if company, ok := companyFromBareModelID(supportedModelBareID(trimmed)); ok {
		return company
	}
	return SupportedModelCompanyOther
}

func companyFromProviderPlatform(platform string) (string, bool) {
	if _, isGateway := supportedModelGatewayPlatforms[platform]; isGateway {
		return "", false
	}
	switch platform {
	case PlatformOpenAI:
		return SupportedModelCompanyOpenAI, true
	case PlatformAnthropic:
		return SupportedModelCompanyAnthropic, true
	case PlatformGemini:
		return SupportedModelCompanyGoogle, true
	case PlatformGrok:
		return SupportedModelCompanyXAI, true
	case PlatformDeepseek:
		return SupportedModelCompanyDeepSeek, true
	case PlatformKimi:
		return SupportedModelCompanyMoonshot, true
	case PlatformZhipu:
		return SupportedModelCompanyZhipu, true
	case PlatformMiniMax:
		return SupportedModelCompanyMiniMax, true
	default:
		return "", false
	}
}

func companyFromBareModelID(bare string) (string, bool) {
	if bare == "" {
		return "", false
	}
	switch {
	case strings.HasPrefix(bare, "sora"):
		return SupportedModelCompanyOpenAI, true
	case strings.HasPrefix(bare, "imagen"),
		strings.HasPrefix(bare, "imagegeneration"),
		strings.HasPrefix(bare, "veo"),
		strings.HasPrefix(bare, "gemma-"):
		return SupportedModelCompanyGoogle, true
	case strings.HasPrefix(bare, "qwen-"),
		strings.HasPrefix(bare, "qwen2-"),
		strings.HasPrefix(bare, "qwen3-"),
		strings.HasPrefix(bare, "qwen4-"),
		strings.HasPrefix(bare, "qwq-"):
		return SupportedModelCompanyAlibaba, true
	case strings.HasPrefix(bare, "llama-"),
		strings.HasPrefix(bare, "llama2-"),
		strings.HasPrefix(bare, "llama3-"),
		strings.HasPrefix(bare, "meta-llama"):
		return SupportedModelCompanyMeta, true
	case strings.HasPrefix(bare, "mistral-"),
		strings.HasPrefix(bare, "mixtral-"):
		return SupportedModelCompanyMistral, true
	case strings.HasPrefix(bare, "doubao-"):
		return SupportedModelCompanyByteDance, true
	case strings.HasPrefix(bare, "hunyuan-"):
		return SupportedModelCompanyTencent, true
	case strings.HasPrefix(bare, "ernie-"):
		return SupportedModelCompanyBaidu, true
	case strings.HasPrefix(bare, "baichuan-"):
		return SupportedModelCompanyBaichuan, true
	default:
		return "", false
	}
}

func supportedModelBareID(name string) string {
	normalized := strings.ToLower(strings.TrimSpace(name))
	normalized = strings.TrimPrefix(normalized, "models/")
	if slash := strings.LastIndexByte(normalized, '/'); slash >= 0 {
		normalized = strings.TrimSpace(normalized[slash+1:])
		normalized = strings.TrimPrefix(normalized, "models/")
	}
	return normalized
}
