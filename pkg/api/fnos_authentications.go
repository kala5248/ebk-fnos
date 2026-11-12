package api

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/auth/fnos"
	"github.com/mayswind/ezbookkeeping/pkg/avatars"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/locales"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
)

const fnOSPendingStateExpiredSeconds = 600

type fnOSPendingState struct {
	Verifier        string
	DeviceId        string
	LinkUid         int64
	CreatedUnixTime int64
}

// FnOSAuthenticationsApi represents fnOS authorization login api
type FnOSAuthenticationsApi struct {
	ApiUsingConfig
	ApiWithUserInfo
	users             *services.UserService
	tokens            *services.TokenService
	userExternalAuths *services.UserExternalAuthService
	twoFactors        *services.TwoFactorAuthorizationService
	authentication    *fnos.FnOSAuthentication
	states            map[string]*fnOSPendingState
	statesMutex       sync.Mutex
}

// Initialize a fnOS authorization login api singleton instance
var (
	FnOSAuthentications = &FnOSAuthenticationsApi{
		ApiUsingConfig: ApiUsingConfig{
			container: settings.Container,
		},
		ApiWithUserInfo: ApiWithUserInfo{
			ApiUsingConfig: ApiUsingConfig{
				container: settings.Container,
			},
			ApiUsingAvatarProvider: ApiUsingAvatarProvider{
				container: avatars.Container,
			},
		},
		users:             services.Users,
		tokens:            services.Tokens,
		userExternalAuths: services.UserExternalAuths,
		twoFactors:        services.TwoFactorAuthorizations,
		states:            make(map[string]*fnOSPendingState, 128),
	}
)

// FnOSAuthUrlHandler returns the fnOS authorization url for the current client
func (a *FnOSAuthenticationsApi) FnOSAuthUrlHandler(c *core.WebContext) (any, *errs.Error) {
	if !a.CurrentConfig().EnableFnOSAuth {
		return nil, errs.ErrFnOSAuthNotEnabled
	}

	a.ensureAuthentication()

	state, err := fnos.GenerateRandomState()

	if err != nil {
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	verifier, err := fnos.GenerateCodeVerifier()

	if err != nil {
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	deviceId, err := fnos.GenerateRandomState()

	if err != nil {
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	codeChallenge := fnos.GenerateCodeChallenge(verifier)

	a.statesMutex.Lock()
	a.cleanupExpiredStates()
	a.states[state] = &fnOSPendingState{
		Verifier:        verifier,
		DeviceId:        deviceId,
		CreatedUnixTime: time.Now().Unix(),
	}
	a.statesMutex.Unlock()

	return &models.FnOSAuthUrlResponse{
		Url:   a.authentication.GetAuthorizationUrl(state, codeChallenge, deviceId),
		State: state,
	}, nil
}

// FnOSAuthLinkUrlHandler returns the fnOS authorization url for linking the current login user
func (a *FnOSAuthenticationsApi) FnOSAuthLinkUrlHandler(c *core.WebContext) (any, *errs.Error) {
	if !a.CurrentConfig().EnableFnOSAuth {
		return nil, errs.ErrFnOSAuthNotEnabled
	}

	a.ensureAuthentication()

	uid := c.GetCurrentLoginUid()
	state, err := fnos.GenerateRandomState()

	if err != nil {
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	verifier, err := fnos.GenerateCodeVerifier()

	if err != nil {
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	deviceId, err := fnos.GenerateRandomState()

	if err != nil {
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	codeChallenge := fnos.GenerateCodeChallenge(verifier)

	a.statesMutex.Lock()
	a.cleanupExpiredStates()
	a.states[state] = &fnOSPendingState{
		Verifier:        verifier,
		DeviceId:        deviceId,
		LinkUid:         uid,
		CreatedUnixTime: time.Now().Unix(),
	}
	a.statesMutex.Unlock()

	return &models.FnOSAuthUrlResponse{
		Url:   a.authentication.GetAuthorizationUrl(state, codeChallenge, deviceId),
		State: state,
	}, nil
}

// FnOSLoginHandler exchanges the one-time authorization code and logs the fnOS user in
func (a *FnOSAuthenticationsApi) FnOSLoginHandler(c *core.WebContext) (any, *errs.Error) {
	if !a.CurrentConfig().EnableFnOSAuth {
		return nil, errs.ErrFnOSAuthNotEnabled
	}

	a.ensureAuthentication()

	var req models.FnOSLoginRequest
	err := c.ShouldBindJSON(&req)

	if err != nil {
		log.Warnf(c, "[fnos_authentications.FnOSLoginHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	code := strings.TrimSpace(req.Code)

	if code == "" {
		return nil, errs.ErrFnOSCodeIsEmpty
	}

	pending := a.popState(req.State)

	if pending == nil {
		return nil, errs.ErrFnOSStateIsInvalid
	}

	if pending.LinkUid != 0 {
		return nil, errs.ErrFnOSStateIsInvalid
	}

	userInfo, err := a.authentication.ExchangeCode(code, pending.Verifier, pending.DeviceId)

	if err != nil {
		log.Warnf(c, "[fnos_authentications.FnOSLoginHandler] failed to exchange code with fnos, because %s", err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	user, externalAuth, err := a.getOrCreateUser(c, userInfo)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Warnf(c, "[fnos_authentications.FnOSLoginHandler] failed to get or create user, because %s", err.Error())
		}

		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if user.Disabled {
		return nil, errs.ErrFnOSUserIsDisabled
	}

	if user.FeatureRestriction.Contains(core.USER_FEATURE_RESTRICTION_TYPE_OAUTH2_LOGIN) {
		return nil, errs.ErrNotPermittedToPerformThisAction
	}

	exists, err := a.twoFactors.ExistsTwoFactorSetting(c, user.Uid)

	if err != nil {
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if exists {
		return nil, errs.ErrFnOSUserRequireTwoFactor
	}

	_ = externalAuth

	token, claims, err := a.tokens.CreateToken(c, user)

	if err != nil {
		log.Errorf(c, "[fnos_authentications.FnOSLoginHandler] failed to create token for user \"uid:%d\", because %s", user.Uid, err.Error())
		return nil, errs.Or(err, errs.ErrTokenGenerating)
	}

	c.SetTextualToken(token)
	c.SetTokenClaims(claims)

	userBasicInfo := a.GetUserBasicInfo(user)

	go func() {
		_ = a.users.UpdateUserLastLoginTime(c, user.Uid)
	}()

	log.Infof(c, "[fnos_authentications.FnOSLoginHandler] fnos user \"%s\" logged in as \"uid:%d\"", userInfo.UserName, user.Uid)

	return &models.AuthResponse{
		Token:   token,
		Need2FA: false,
		User:    userBasicInfo,
	}, nil
}

// FnOSLinkHandler links a fnOS account to the current login user
func (a *FnOSAuthenticationsApi) FnOSLinkHandler(c *core.WebContext) (any, *errs.Error) {
	if !a.CurrentConfig().EnableFnOSAuth {
		return nil, errs.ErrFnOSAuthNotEnabled
	}

	a.ensureAuthentication()

	var req models.FnOSLoginRequest
	err := c.ShouldBindJSON(&req)

	if err != nil {
		log.Warnf(c, "[fnos_authentications.FnOSLinkHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentLoginUid()
	pending := a.popState(req.State)

	if pending == nil || pending.LinkUid != uid {
		return nil, errs.ErrFnOSStateIsInvalid
	}

	userInfo, err := a.authentication.ExchangeCode(strings.TrimSpace(req.Code), pending.Verifier, pending.DeviceId)

	if err != nil {
		log.Warnf(c, "[fnos_authentications.FnOSLinkHandler] failed to exchange code with fnos, because %s", err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	existing, err := a.userExternalAuths.GetUserExternalAuthByExternalEmail(c, buildFnOSExternalEmail(userInfo.Uid), core.USER_EXTERNAL_AUTH_TYPE_FNOS)

	if err != nil && !errors.Is(err, errs.ErrUserExternalAuthNotFound) {
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if existing != nil {
		if existing.Uid != uid {
			return nil, errs.ErrFnOSExternalAuthAlreadyBound
		}

		return nil, errs.ErrFnOSExternalAuthAlreadyLinked
	}

	userExternalAuth := &models.UserExternalAuth{
		Uid:              uid,
		ExternalAuthType: core.USER_EXTERNAL_AUTH_TYPE_FNOS,
		ExternalUsername: buildFnOSExternalUserName(userInfo),
		ExternalEmail:    buildFnOSExternalEmail(userInfo.Uid),
	}

	err = a.userExternalAuths.CreateUserExternalAuth(c, userExternalAuth)

	if err != nil {
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return true, nil
}

func (a *FnOSAuthenticationsApi) getOrCreateUser(c *core.WebContext, userInfo *fnos.FnOSUserInfo) (*models.User, *models.UserExternalAuth, error) {
	externalEmail := buildFnOSExternalEmail(userInfo.Uid)
	externalAuth, err := a.userExternalAuths.GetUserExternalAuthByExternalEmail(c, externalEmail, core.USER_EXTERNAL_AUTH_TYPE_FNOS)

	if err != nil && !errors.Is(err, errs.ErrUserExternalAuthNotFound) {
		return nil, nil, err
	}

	if externalAuth != nil {
		user, err := a.users.GetUserById(c, externalAuth.Uid)

		if err != nil {
			return nil, nil, err
		}

		return user, externalAuth, nil
	}

	if !a.CurrentConfig().EnableUserRegister || !a.CurrentConfig().FnOSAutoRegister {
		return nil, nil, errs.ErrFnOSAutoRegisterDisabled
	}

	userName := buildFnOSLocalUsername(c, userInfo.UserName)
	nickName := strings.TrimSpace(userInfo.NickName)

	if nickName == "" {
		nickName = userInfo.UserName
	}

	if !utils.IsValidNickName(nickName) {
		nickName = userName
	}

	languageCode := parseClientLanguageCode(c.GetClientLocale())

	user := &models.User{
		Username:             userName,
		Email:                fmt.Sprintf("fnos-%d@ezbookkeeping.local", userInfo.Uid),
		Nickname:             nickName,
		Language:             languageCode,
		DefaultCurrency:      "CNY",
		FirstDayOfWeek:       core.WEEKDAY_MONDAY,
		FiscalYearStart:      core.FISCAL_YEAR_START_DEFAULT,
		TransactionEditScope: models.TRANSACTION_EDIT_SCOPE_ALL,
		FeatureRestriction:   a.CurrentConfig().DefaultFeatureRestrictions,
		EmailVerified:        true,
	}

	err = a.users.CreateUser(c, user, true)

	if err != nil {
		return nil, nil, err
	}

	userExternalAuth := &models.UserExternalAuth{
		Uid:              user.Uid,
		ExternalAuthType: core.USER_EXTERNAL_AUTH_TYPE_FNOS,
		ExternalUsername: buildFnOSExternalUserName(userInfo),
		ExternalEmail:    externalEmail,
	}

	err = a.userExternalAuths.CreateUserExternalAuth(c, userExternalAuth)

	if err != nil {
		return nil, nil, err
	}

	log.Infof(c, "[fnos_authentications.getOrCreateUser] fnos user \"%s\" registered as \"uid:%d\"", userInfo.UserName, user.Uid)

	return user, userExternalAuth, nil
}

func (a *FnOSAuthenticationsApi) popState(state string) *fnOSPendingState {
	a.statesMutex.Lock()
	defer a.statesMutex.Unlock()

	pending, exists := a.states[state]

	if !exists {
		return nil
	}

	delete(a.states, state)

	if time.Now().Unix()-pending.CreatedUnixTime > fnOSPendingStateExpiredSeconds {
		return nil
	}

	return pending
}

func (a *FnOSAuthenticationsApi) cleanupExpiredStates() {
	now := time.Now().Unix()

	for state, pending := range a.states {
		if now-pending.CreatedUnixTime > fnOSPendingStateExpiredSeconds {
			delete(a.states, state)
		}
	}
}

func (a *FnOSAuthenticationsApi) ensureAuthentication() {
	if a.authentication != nil {
		return
	}

	a.authentication = fnos.NewFnOSAuthentication(a.CurrentConfig())
}

func buildFnOSExternalEmail(uid int64) string {
	return fmt.Sprintf("fnos-%d@ezbookkeeping.local", uid)
}

func buildFnOSExternalUserName(userInfo *fnos.FnOSUserInfo) string {
	userName := strings.TrimSpace(userInfo.UserName)

	if userName == "" {
		userName = fmt.Sprintf("fnos-%d", userInfo.Uid)
	}

	return userName
}

func buildFnOSLocalUsername(c *core.WebContext, fnosUserName string) string {
	base := sanitizeFnOSUsername(fnosUserName)

	if base == "" {
		base = "fnos_user"
	}

	if len(base) > 24 {
		base = base[:24]
	}

	candidate := "fnos_" + base

	for i := 0; i < 20; i++ {
		if i > 0 {
			candidate = fmt.Sprintf("fnos_%s%d", base, i)
		}

		exists, err := services.Users.ExistsUsername(c, candidate)

		if err != nil {
			break
		}

		if !exists {
			return candidate
		}
	}

	return "fnos_user_" + strings.ReplaceAll(fmt.Sprintf("%d", time.Now().UnixNano()), "-", "")
}

func parseClientLanguageCode(acceptLanguage string) string {
	if acceptLanguage == "" {
		return ""
	}

	first := strings.TrimSpace(strings.Split(acceptLanguage, ",")[0])
	first = strings.TrimSpace(strings.Split(first, ";")[0])

	if _, exists := locales.AllLanguages[first]; exists {
		return first
	}

	primary := strings.Split(first, "-")[0]

	for language := range locales.AllLanguages {
		if strings.HasPrefix(language, primary+"-") {
			return language
		}
	}

	if _, exists := locales.AllLanguages[primary]; exists {
		return primary
	}

	return ""
}

func sanitizeFnOSUsername(userName string) string {
	var builder strings.Builder

	for _, r := range strings.ToLower(strings.TrimSpace(userName)) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			builder.WriteRune(r)
		} else if r == '.' || r == ' ' {
			builder.WriteRune('_')
		}
	}

	return builder.String()
}
