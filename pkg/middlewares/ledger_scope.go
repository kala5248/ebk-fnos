package middlewares

import (
	"strconv"
	"strings"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/services"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
)

// LedgerIdHeaderName is the request header which carries the selected shared ledger id
const LedgerIdHeaderName = "X-Ledger-Id"

// identityOnlyPathPrefixes contains those api paths whose handlers always operate on the login identity,
// instead of the data owner of the currently selected shared ledger
var identityOnlyPathPrefixes = []string{
	"/api/v1/tokens/",
	"/api/v1/users/",
	"/api/v1/ledgers/",
	"/api/v1/llm/user_config/",
	"/api/v1/fnos/",
}

// LedgerDataScope resolves the selected shared ledger and switches the data scope of the current request.
// This middleware must be registered after the token authorization middleware.
func LedgerDataScope(config *settings.Config) core.MiddlewareHandlerFunc {
	return func(c *core.WebContext) {
		claims := c.GetTokenClaims()

		if claims == nil {
			c.Next()
			return
		}

		loginUid := claims.Uid
		c.SetCurrentLoginUid(loginUid)

		if isIdentityOnlyPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		ledgerIdText := strings.TrimSpace(c.GetHeader(LedgerIdHeaderName))

		if ledgerIdText == "" {
			ledgerIdText = strings.TrimSpace(c.Query("ledgerId"))
		}

		var ledgerId int64 = 0

		if ledgerIdText != "" && ledgerIdText != "0" {
			parsedLedgerId, err := strconv.ParseInt(ledgerIdText, 10, 64)

			if err != nil || parsedLedgerId < 0 {
				log.Warnf(c, "[ledger_scope.LedgerDataScope] ledger id \"%s\" in request is invalid", ledgerIdText)
				utils.PrintJsonErrorResult(c, errs.ErrLedgerNotFound)
				return
			}

			ledgerId = parsedLedgerId
		}

		scope, err := services.Ledgers.ResolveDataScope(c, loginUid, ledgerId)

		if err != nil {
			log.Warnf(c, "[ledger_scope.LedgerDataScope] cannot resolve ledger \"%d\" for user \"uid:%d\", because %s", ledgerId, loginUid, err.Error())
			utils.PrintJsonErrorResult(c, errs.Or(err, errs.ErrLedgerNotFound))
			return
		}

		if scope.Role.IsReadOnly() && !isReadOnlyHttpMethod(c.Request.Method) {
			log.Warnf(c, "[ledger_scope.LedgerDataScope] user \"uid:%d\" has no write permission in ledger \"%d\"", loginUid, ledgerId)
			utils.PrintJsonErrorResult(c, errs.ErrLedgerReadOnly)
			return
		}

		claims.Uid = scope.DataUid

		c.SetCurrentLedgerId(scope.LedgerId)
		c.SetCurrentLedgerRole(core.LedgerRole(scope.Role))
		c.Next()
	}
}

func isIdentityOnlyPath(path string) bool {
	for i := 0; i < len(identityOnlyPathPrefixes); i++ {
		if strings.HasPrefix(path, identityOnlyPathPrefixes[i]) {
			return true
		}
	}

	return false
}

func isReadOnlyHttpMethod(method string) bool {
	return method == "GET" || method == "HEAD" || method == "OPTIONS"
}
