package errs

import "net/http"

// Error codes related to per-user large language model configuration
var (
	ErrUserLlmTextModelIsEmpty = NewNormalError(NormalSubcategoryUserLlmConfig, 0, http.StatusBadRequest, "text recognition model id is empty")
	ErrUserLlmConfigNotFound   = NewNormalError(NormalSubcategoryUserLlmConfig, 1, http.StatusBadRequest, "user large language model configuration not found")
	ErrUserLlmConfigTestFailed = NewNormalError(NormalSubcategoryUserLlmConfig, 2, http.StatusBadGateway, "user large language model configuration test failed")
)
