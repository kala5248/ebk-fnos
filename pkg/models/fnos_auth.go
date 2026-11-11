package models

// FnOSAuthUrlRequest represents all parameters of requesting a fnOS authorization url
type FnOSAuthUrlRequest struct {
	ClientSessionId string `form:"clientSessionId"`
}

// FnOSAuthUrlResponse represents the response of fnOS authorization url
type FnOSAuthUrlResponse struct {
	Url   string `json:"url"`
	State string `json:"state"`
}

// FnOSLoginRequest represents all parameters of fnOS authorization code login request
type FnOSLoginRequest struct {
	State string `json:"state" binding:"required"`
	Code  string `json:"code" binding:"required,notBlank,max=2048"`
}
