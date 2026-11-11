package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/llm"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

// UserLlmConfigsApi represents per-user large language model configuration api
type UserLlmConfigsApi struct {
	ApiUsingConfig
	userLlmConfigs *services.UserLlmConfigService
}

// Initialize a per-user large language model configuration api singleton instance
var (
	UserLlmConfigs = &UserLlmConfigsApi{
		ApiUsingConfig: ApiUsingConfig{
			container: settings.Container,
		},
		userLlmConfigs: services.UserLlmConfigs,
	}
)

// UserLlmConfigGetHandler returns the large language model configuration of the current login user
func (a *UserLlmConfigsApi) UserLlmConfigGetHandler(c *core.WebContext) (any, *errs.Error) {
	uid := c.GetCurrentLoginUid()

	llmConfig, err := a.userLlmConfigs.GetUserLlmConfig(c, uid)

	if err != nil {
		log.Errorf(c, "[user_llm_configs.UserLlmConfigGetHandler] failed to get llm config for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return a.userLlmConfigs.ToUserLlmConfigResponse(llmConfig), nil
}

// UserLlmConfigSaveHandler saves the large language model configuration of the current login user
func (a *UserLlmConfigsApi) UserLlmConfigSaveHandler(c *core.WebContext) (any, *errs.Error) {
	var req models.UserLlmConfigUpdateRequest
	err := c.ShouldBindJSON(&req)

	if err != nil {
		log.Warnf(c, "[user_llm_configs.UserLlmConfigSaveHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentLoginUid()

	llmConfig, err := a.userLlmConfigs.SaveUserLlmConfig(c, uid, &req)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[user_llm_configs.UserLlmConfigSaveHandler] failed to save llm config for user \"uid:%d\", because %s", uid, err.Error())
		}

		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[user_llm_configs.UserLlmConfigSaveHandler] user \"uid:%d\" saved llm config (provider: %s)", uid, llmConfig.Provider)

	return a.userLlmConfigs.ToUserLlmConfigResponse(llmConfig), nil
}

// UserLlmConfigDeleteHandler deletes the large language model configuration of the current login user
func (a *UserLlmConfigsApi) UserLlmConfigDeleteHandler(c *core.WebContext) (any, *errs.Error) {
	uid := c.GetCurrentLoginUid()

	err := a.userLlmConfigs.DeleteUserLlmConfig(c, uid)

	if err != nil {
		log.Errorf(c, "[user_llm_configs.UserLlmConfigDeleteHandler] failed to delete llm config for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return true, nil
}

// UserLlmConfigTestHandler tests the saved large language model configuration of the current login user
func (a *UserLlmConfigsApi) UserLlmConfigTestHandler(c *core.WebContext) (any, *errs.Error) {
	uid := c.GetCurrentLoginUid()

	llmConfig, userConfig, err := a.userLlmConfigs.ResolveLLMConfig(c, uid, false)

	if err != nil {
		log.Errorf(c, "[user_llm_configs.UserLlmConfigTestHandler] failed to resolve llm config for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrUserLlmConfigTestFailed)
	}

	if userConfig == nil || llmConfig == nil {
		return nil, errs.ErrLargeLanguageModelProviderNotEnabled
	}

	req := &data.LargeLanguageModelRequest{
		Stream:       false,
		SystemPrompt: "You are a connection test helper. Reply with a JSON object only.",
		UserPrompt:   []byte("Return {\"ok\":true} as JSON."),
	}

	resp, err := llm.Container.GetJsonResponseByTextRecognitionModelWithConfig(c, uid, llmConfig, req)

	if err != nil {
		log.Warnf(c, "[user_llm_configs.UserLlmConfigTestHandler] llm connection test failed for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrUserLlmConfigTestFailed)
	}

	if resp == nil || len(resp.Content) == 0 {
		return nil, errs.ErrUserLlmConfigTestFailed
	}

	return map[string]any{
		"success": true,
		"content": resp.Content,
	}, nil
}
