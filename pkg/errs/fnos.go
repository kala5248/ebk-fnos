package errs

import "net/http"

// Error codes related to fnOS authorization login
var (
	ErrInvalidFnOSConfig             = NewNormalError(NormalSubcategoryFnOSAuth, 0, http.StatusBadRequest, "invalid fnos oauth configuration")
	ErrFnOSAuthNotEnabled            = NewNormalError(NormalSubcategoryFnOSAuth, 1, http.StatusBadRequest, "fnos authorization login is not enabled")
	ErrFnOSStateIsInvalid            = NewNormalError(NormalSubcategoryFnOSAuth, 2, http.StatusBadRequest, "fnos authorization state is invalid or expired")
	ErrFnOSCodeIsEmpty               = NewNormalError(NormalSubcategoryFnOSAuth, 3, http.StatusBadRequest, "fnos authorization code is empty")
	ErrFnOSUserInfoIsEmpty           = NewNormalError(NormalSubcategoryFnOSAuth, 4, http.StatusBadGateway, "fnos user information is empty")
	ErrFnOSAutoRegisterDisabled      = NewNormalError(NormalSubcategoryFnOSAuth, 5, http.StatusForbidden, "fnos user is not registered and auto registration is disabled")
	ErrFnOSUserRequireTwoFactor      = NewNormalError(NormalSubcategoryFnOSAuth, 6, http.StatusForbidden, "the fnos user has enabled two-factor authorization, please login with password instead")
	ErrFnOSExternalAuthAlreadyBound  = NewNormalError(NormalSubcategoryFnOSAuth, 7, http.StatusBadRequest, "the fnos account has been bound to another user")
	ErrFnOSExternalAuthAlreadyLinked = NewNormalError(NormalSubcategoryFnOSAuth, 8, http.StatusBadRequest, "current user has already linked a fnos account")
	ErrFnOSUserIsDisabled            = NewNormalError(NormalSubcategoryFnOSAuth, 9, http.StatusForbidden, "the fnos user has been disabled")
	ErrFnOSGatewayIdentityMissing    = NewNormalError(NormalSubcategoryFnOSAuth, 10, http.StatusUnauthorized, "fnos gateway identity header is missing")
)
