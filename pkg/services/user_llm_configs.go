package services

import (
	"strings"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"xorm.io/xorm"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
)

const defaultUserLlmRequestTimeout uint32 = 60000

// UserLlmConfigService represents a per-user large language model configuration service
type UserLlmConfigService struct {
	ServiceUsingDB
	ServiceUsingConfig
}

// Initialize a user large language model configuration service singleton instance
var (
	UserLlmConfigs = &UserLlmConfigService{
		ServiceUsingDB:     ServiceUsingDB{container: datastore.Container},
		ServiceUsingConfig: ServiceUsingConfig{container: settings.Container},
	}
)

// GetUserLlmConfig returns the per-user large language model configuration
func (s *UserLlmConfigService) GetUserLlmConfig(c core.Context, uid int64) (*models.UserLlmConfig, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	llmConfig := &models.UserLlmConfig{}
	has, err := s.UserDB().NewSession(c).ID(uid).Get(llmConfig)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, nil
	}

	return llmConfig, nil
}

// SaveUserLlmConfig saves the per-user large language model configuration
func (s *UserLlmConfigService) SaveUserLlmConfig(c core.Context, uid int64, req *models.UserLlmConfigUpdateRequest) (*models.UserLlmConfig, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	normalizedProvider, err := normalizeLlmProvider(req.Provider)

	if err != nil {
		return nil, err
	}

	textModelId := strings.TrimSpace(req.TextModelId)
	imageModelId := strings.TrimSpace(req.ImageModelId)

	if req.EnableTextRecognition && textModelId == "" {
		return nil, errs.ErrUserLlmTextModelIsEmpty
	}

	if req.EnableImageRecognition && imageModelId == "" {
		imageModelId = textModelId
	}

	existed, err := s.GetUserLlmConfig(c, uid)

	if err != nil {
		return nil, err
	}

	encryptedApiKey := ""

	if existed != nil {
		encryptedApiKey = existed.ApiKey
	}

	apiKey := strings.TrimSpace(req.ApiKey)

	if apiKey != "" {
		encryptedApiKey, err = utils.EncryptSecret(apiKey, s.CurrentConfig().SecretKey)

		if err != nil {
			return nil, err
		}
	}

	llmConfig := &models.UserLlmConfig{
		Uid:                    uid,
		Provider:               normalizedProvider,
		BaseUrl:                strings.TrimRight(strings.TrimSpace(req.BaseUrl), "/"),
		ApiKey:                 encryptedApiKey,
		TextModelId:            textModelId,
		ImageModelId:           imageModelId,
		EnableTextRecognition:  req.EnableTextRecognition,
		EnableImageRecognition: req.EnableImageRecognition,
		UpdatedUnixTime:        time.Now().Unix(),
	}

	err = s.UserDB().DoTransaction(c, func(sess *xorm.Session) error {
		if existed == nil {
			_, err := sess.Insert(llmConfig)
			return err
		}

		_, err := sess.ID(uid).Cols("provider", "base_url", "api_key", "text_model_id", "image_model_id", "enable_text_recognition", "enable_image_recognition", "updated_unix_time").Update(llmConfig)
		return err
	})

	if err != nil {
		return nil, err
	}

	return llmConfig, nil
}

// DeleteUserLlmConfig deletes the per-user large language model configuration
func (s *UserLlmConfigService) DeleteUserLlmConfig(c core.Context, uid int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	_, err := s.UserDB().NewSession(c).ID(uid).Delete(new(models.UserLlmConfig))

	return err
}

// ResolveLLMConfig returns the effective large language model configuration of the given user.
// The per-user configuration is preferred, and it returns nil when the user does not enable it.
func (s *UserLlmConfigService) ResolveLLMConfig(c core.Context, uid int64, imageRecognition bool) (*settings.LLMConfig, *models.UserLlmConfig, error) {
	llmConfig, err := s.GetUserLlmConfig(c, uid)

	if err != nil {
		return nil, nil, err
	}

	if llmConfig == nil {
		return nil, nil, nil
	}

	if imageRecognition && (!llmConfig.EnableImageRecognition || llmConfig.ImageModelId == "") {
		return nil, llmConfig, nil
	}

	if !imageRecognition && (!llmConfig.EnableTextRecognition || llmConfig.TextModelId == "") {
		return nil, llmConfig, nil
	}

	apiKey, err := utils.DecryptSecret(llmConfig.ApiKey, s.CurrentConfig().SecretKey)

	if err != nil {
		return nil, llmConfig, err
	}

	modelId := llmConfig.TextModelId

	if imageRecognition {
		modelId = llmConfig.ImageModelId
	}

	effectiveConfig := buildLLMConfig(llmConfig.Provider, llmConfig.BaseUrl, apiKey, modelId)

	return effectiveConfig, llmConfig, nil
}

// ToUserLlmConfigResponse converts the model to a view-object
func (s *UserLlmConfigService) ToUserLlmConfigResponse(llmConfig *models.UserLlmConfig) *models.UserLlmConfigResponse {
	if llmConfig == nil {
		return &models.UserLlmConfigResponse{}
	}

	apiKeyMasked := ""
	hasApiKey := llmConfig.ApiKey != ""

	if hasApiKey {
		apiKey, err := utils.DecryptSecret(llmConfig.ApiKey, s.CurrentConfig().SecretKey)

		if err == nil && len(apiKey) > 4 {
			apiKeyMasked = "*" + apiKey[len(apiKey)-4:]
		}
	}

	return &models.UserLlmConfigResponse{
		Provider:               llmConfig.Provider,
		BaseUrl:                llmConfig.BaseUrl,
		ApiKeyMasked:           apiKeyMasked,
		TextModelId:            llmConfig.TextModelId,
		ImageModelId:           llmConfig.ImageModelId,
		EnableTextRecognition:  llmConfig.EnableTextRecognition,
		EnableImageRecognition: llmConfig.EnableImageRecognition,
		HasApiKey:              hasApiKey,
	}
}

func normalizeLlmProvider(provider string) (string, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))

	switch provider {
	case settings.OpenAILLMProvider,
		settings.OpenAICompatibleLLMProvider,
		settings.OpenAIResponsesCompatibleLLMProvider,
		settings.AnthropicLLMProvider,
		settings.AnthropicCompatibleLLMProvider,
		settings.OpenRouterLLMProvider,
		settings.OllamaLLMProvider,
		settings.LMStudioLLMProvider,
		settings.GoogleAILLMProvider:
		return provider, nil
	}

	return "", errs.ErrInvalidLLMProvider
}

func buildLLMConfig(provider string, baseUrl string, apiKey string, modelId string) *settings.LLMConfig {
	llmConfig := &settings.LLMConfig{
		LLMProvider:                         provider,
		EnableThinking:                      settings.LLMThinkingDefault,
		LargeLanguageModelAPIProxy:          "system",
		LargeLanguageModelAPIRequestTimeout: defaultUserLlmRequestTimeout,
		LargeLanguageModelAPISkipTLSVerify:  false,
		AnthropicMaxTokens:                  4096,
		AnthropicCompatibleMaxTokens:        4096,
	}

	switch provider {
	case settings.OpenAILLMProvider:
		llmConfig.OpenAIAPIKey = apiKey
		llmConfig.OpenAIModelID = modelId
	case settings.OpenAICompatibleLLMProvider, settings.OpenAIResponsesCompatibleLLMProvider:
		llmConfig.OpenAICompatibleBaseURL = baseUrl
		llmConfig.OpenAICompatibleAPIKey = apiKey
		llmConfig.OpenAICompatibleModelID = modelId
	case settings.OpenRouterLLMProvider:
		llmConfig.OpenRouterAPIKey = apiKey
		llmConfig.OpenRouterModelID = modelId
	case settings.AnthropicLLMProvider:
		llmConfig.AnthropicAPIKey = apiKey
		llmConfig.AnthropicModelID = modelId
	case settings.AnthropicCompatibleLLMProvider:
		llmConfig.AnthropicCompatibleBaseURL = baseUrl
		llmConfig.AnthropicCompatibleAPIKey = apiKey
		llmConfig.AnthropicCompatibleModelID = modelId
		llmConfig.AnthropicCompatibleAPIVersion = "2023-06-01"
	case settings.OllamaLLMProvider:
		llmConfig.OllamaServerURL = baseUrl
		llmConfig.OllamaModelID = modelId
	case settings.LMStudioLLMProvider:
		llmConfig.LMStudioServerURL = baseUrl
		llmConfig.LMStudioToken = apiKey
		llmConfig.LMStudioModelID = modelId
	case settings.GoogleAILLMProvider:
		llmConfig.GoogleAIAPIKey = apiKey
		llmConfig.GoogleAIModelID = modelId
	}

	return llmConfig
}
