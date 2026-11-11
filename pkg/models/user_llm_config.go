package models

// UserLlmConfig represents the per-user large language model configuration stored in database
type UserLlmConfig struct {
	Uid                    int64  `xorm:"PK"`
	Provider               string `xorm:"VARCHAR(32) NOT NULL"`
	BaseUrl                string `xorm:"VARCHAR(255) NOT NULL"`
	ApiKey                 string `xorm:"VARCHAR(2048) NOT NULL"`
	TextModelId            string `xorm:"VARCHAR(128) NOT NULL"`
	ImageModelId           string `xorm:"VARCHAR(128)"`
	EnableTextRecognition  bool   `xorm:"NOT NULL"`
	EnableImageRecognition bool   `xorm:"NOT NULL"`
	UpdatedUnixTime        int64
}

// UserLlmConfigUpdateRequest represents all parameters of per-user large language model configuration update request
type UserLlmConfigUpdateRequest struct {
	Provider               string `json:"provider" binding:"required,max=32"`
	BaseUrl                string `json:"baseUrl" binding:"max=255"`
	ApiKey                 string `json:"apiKey" binding:"max=2048"`
	TextModelId            string `json:"textModelId" binding:"max=128"`
	ImageModelId           string `json:"imageModelId" binding:"max=128"`
	EnableTextRecognition  bool   `json:"enableTextRecognition"`
	EnableImageRecognition bool   `json:"enableImageRecognition"`
}

// UserLlmConfigResponse represents a view-object of per-user large language model configuration
type UserLlmConfigResponse struct {
	Provider               string `json:"provider"`
	BaseUrl                string `json:"baseUrl"`
	ApiKeyMasked           string `json:"apiKeyMasked"`
	TextModelId            string `json:"textModelId"`
	ImageModelId           string `json:"imageModelId"`
	EnableTextRecognition  bool   `json:"enableTextRecognition"`
	EnableImageRecognition bool   `json:"enableImageRecognition"`
	HasApiKey              bool   `json:"hasApiKey"`
}
