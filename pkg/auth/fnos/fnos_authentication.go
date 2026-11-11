package fnos

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/httpclient"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

// FnOSUserInfo represents the user information returned by fnOS authorization server
type FnOSUserInfo struct {
	Uid          int64
	UserName     string
	NickName     string
	IsAdmin      bool
	LanguageCode string
}

// FnOSAuthentication represents a fnOS OAuth 2.0 (PKCE) client
type FnOSAuthentication struct {
	config     *settings.Config
	httpClient *http.Client
}

// NewFnOSAuthentication returns a new fnOS authentication instance
func NewFnOSAuthentication(config *settings.Config) *FnOSAuthentication {
	return &FnOSAuthentication{
		config:     config,
		httpClient: httpclient.NewHttpClient(config.FnOSRequestTimeout, "system", config.FnOSSkipTLSVerify, core.GetOutgoingUserAgent(), config.EnableDebugLog, []uint16{}),
	}
}

// GenerateCodeVerifier returns a new PKCE code verifier
func GenerateCodeVerifier() (string, error) {
	buf := make([]byte, 48)

	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// GenerateCodeChallenge returns the S256 code challenge of the given verifier
func GenerateCodeChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// GenerateRandomState returns a new random state string
func GenerateRandomState() (string, error) {
	buf := make([]byte, 16)

	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", buf), nil
}

// GetAuthorizationUrl returns the fnOS authorization url. fnOS does not accept a redirect_uri, the user signs in
// on the fnOS page and copies the one-time authorization code back to ezBookkeeping.
func (a *FnOSAuthentication) GetAuthorizationUrl(state string, codeChallenge string, deviceId string) string {
	query := url.Values{}
	query.Set("client_id", a.config.FnOSClientId)
	query.Set("scope", a.config.FnOSScope)
	query.Set("device_id", deviceId)
	query.Set("device_name", "ezBookkeeping")
	query.Set("device_model", "NAS")
	query.Set("code_challenge", codeChallenge)
	query.Set("code_challenge_method", "S256")
	query.Set("state", state)

	return a.config.FnOSBaseUrl + "/oauthapi/authorize?" + query.Encode()
}

// ExchangeCode exchanges the one-time authorization code for an access token and user information
func (a *FnOSAuthentication) ExchangeCode(code string, codeVerifier string, deviceId string) (*FnOSUserInfo, error) {
	form := url.Values{}
	form.Set("client_id", a.config.FnOSClientId)
	form.Set("code", code)
	form.Set("code_verifier", codeVerifier)
	form.Set("device_id", deviceId)
	form.Set("grant_type", "authorization_code")

	if a.config.FnOSClientSecret != "" {
		form.Set("client_secret", a.config.FnOSClientSecret)
	}

	req, err := http.NewRequest(http.MethodPost, a.config.FnOSBaseUrl+"/oauthapi/third-part/token", strings.NewReader(form.Encode()))

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := a.httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))

	if err != nil {
		return nil, err
	}

	var tokenResp struct {
		AccessToken          string `json:"access_token"`
		AccessTokenAlt       string `json:"accessToken"`
		Uid                  int64  `json:"uid"`
		UserId               int64  `json:"user_id"`
		UserIdAlt            int64  `json:"userId"`
		Username             string `json:"username"`
		UserName             string `json:"user_name"`
		UserNameAlt          string `json:"userName"`
		Nickname             string `json:"nickname"`
		NicknameAlt          string `json:"nickName"`
		Admin                bool   `json:"admin"`
		AuthorizationPending bool   `json:"authorization_pending"`
		Error                string `json:"error"`
		Message              string `json:"message"`
		Msg                  string `json:"msg"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}

	accessToken := tokenResp.AccessToken

	if accessToken == "" {
		accessToken = tokenResp.AccessTokenAlt
	}

	if accessToken == "" {
		message := tokenResp.Error

		if tokenResp.Message != "" {
			message = tokenResp.Message
		} else if tokenResp.Msg != "" {
			message = tokenResp.Msg
		}

		if message == "" {
			message = fmt.Sprintf("http status %d", resp.StatusCode)
		}

		return nil, fmt.Errorf("fnos token exchange failed: %s", message)
	}

	uid := tokenResp.Uid

	if uid <= 0 {
		uid = tokenResp.UserId
	}

	if uid <= 0 {
		uid = tokenResp.UserIdAlt
	}

	userName := tokenResp.Username

	if userName == "" {
		userName = tokenResp.UserName
	}

	if userName == "" {
		userName = tokenResp.UserNameAlt
	}

	if uid <= 0 || userName == "" {
		return nil, errs.ErrFnOSUserInfoIsEmpty
	}

	nickName := tokenResp.Nickname

	if nickName == "" {
		nickName = tokenResp.NicknameAlt
	}

	if nickName == "" {
		nickName = userName
	}

	return &FnOSUserInfo{
		Uid:      uid,
		UserName: userName,
		NickName: nickName,
		IsAdmin:  tokenResp.Admin,
	}, nil
}
