package core

// LedgerRole represents the role of the current user in the current ledger
type LedgerRole byte

// Ledger roles
const (
	LEDGER_ROLE_OWNER  LedgerRole = 1
	LEDGER_ROLE_EDITOR LedgerRole = 2
	LEDGER_ROLE_VIEWER LedgerRole = 3
)

// context keys used by the ledger (shared bookkeeping book) data scope
const (
	webContextCurrentLoginUidFieldKey   = "webContextCurrentLoginUid"
	webContextCurrentLedgerIdFieldKey   = "webContextCurrentLedgerId"
	webContextCurrentLedgerRoleFieldKey = "webContextCurrentLedgerRole"
)

// SetCurrentLoginUid sets the uid of the login identity (before the data scope is switched to a ledger)
func (c *WebContext) SetCurrentLoginUid(uid int64) {
	c.Set(webContextCurrentLoginUidFieldKey, uid)
}

// GetCurrentLoginUid returns the uid of the login identity, never the data owner of a shared ledger
func (c *WebContext) GetCurrentLoginUid() int64 {
	value, exists := c.Get(webContextCurrentLoginUidFieldKey)

	if !exists || value == nil {
		return c.GetCurrentUid()
	}

	uid, ok := value.(int64)

	if !ok {
		return c.GetCurrentUid()
	}

	return uid
}

// SetCurrentLedgerId sets the current ledger id (0 means the personal ledger)
func (c *WebContext) SetCurrentLedgerId(ledgerId int64) {
	c.Set(webContextCurrentLedgerIdFieldKey, ledgerId)
}

// GetCurrentLedgerId returns the current ledger id (0 means the personal ledger)
func (c *WebContext) GetCurrentLedgerId() int64 {
	value, exists := c.Get(webContextCurrentLedgerIdFieldKey)

	if !exists || value == nil {
		return 0
	}

	ledgerId, ok := value.(int64)

	if !ok {
		return 0
	}

	return ledgerId
}

// SetCurrentLedgerRole sets the role of the current user in the current ledger
func (c *WebContext) SetCurrentLedgerRole(role LedgerRole) {
	c.Set(webContextCurrentLedgerRoleFieldKey, role)
}

// GetCurrentLedgerRole returns the role of the current user in the current ledger
func (c *WebContext) GetCurrentLedgerRole() LedgerRole {
	value, exists := c.Get(webContextCurrentLedgerRoleFieldKey)

	if !exists || value == nil {
		return LEDGER_ROLE_OWNER
	}

	role, ok := value.(LedgerRole)

	if !ok {
		return LEDGER_ROLE_OWNER
	}

	return role
}
