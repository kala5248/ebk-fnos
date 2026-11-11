package errs

import "net/http"

// Error codes related to shared ledger features
var (
	ErrLedgerNotFound                 = NewNormalError(NormalSubcategoryLedger, 0, http.StatusBadRequest, "ledger not found")
	ErrLedgerNotMember                = NewNormalError(NormalSubcategoryLedger, 1, http.StatusForbidden, "current user is not a member of this ledger")
	ErrLedgerNotOwner                 = NewNormalError(NormalSubcategoryLedger, 2, http.StatusForbidden, "current user is not the owner of this ledger")
	ErrLedgerNameIsEmpty              = NewNormalError(NormalSubcategoryLedger, 3, http.StatusBadRequest, "ledger name is empty")
	ErrLedgerNameIsTooLong            = NewNormalError(NormalSubcategoryLedger, 4, http.StatusBadRequest, "ledger name is too long")
	ErrLedgerInviteCodeIsEmpty        = NewNormalError(NormalSubcategoryLedger, 5, http.StatusBadRequest, "ledger invite code is empty")
	ErrLedgerInviteCodeIsInvalid      = NewNormalError(NormalSubcategoryLedger, 6, http.StatusBadRequest, "ledger invite code is invalid")
	ErrLedgerAlreadyJoined            = NewNormalError(NormalSubcategoryLedger, 7, http.StatusBadRequest, "user has already joined this ledger")
	ErrLedgerMemberRoleIsInvalid      = NewNormalError(NormalSubcategoryLedger, 8, http.StatusBadRequest, "ledger member role is invalid")
	ErrLedgerOwnerRoleCannotBeChanged = NewNormalError(NormalSubcategoryLedger, 9, http.StatusBadRequest, "the role of ledger owner cannot be changed")
	ErrLedgerOwnerCannotBeRemoved     = NewNormalError(NormalSubcategoryLedger, 10, http.StatusBadRequest, "the ledger owner cannot be removed")
	ErrLedgerOwnerCannotLeave         = NewNormalError(NormalSubcategoryLedger, 11, http.StatusBadRequest, "the ledger owner cannot leave the ledger, delete it instead")
	ErrLedgerReadOnly                 = NewNormalError(NormalSubcategoryLedger, 12, http.StatusForbidden, "current user has read-only permission in this ledger")
)
