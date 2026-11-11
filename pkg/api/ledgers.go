package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

// LedgersApi represents shared ledger api
type LedgersApi struct {
	ApiUsingConfig
	ledgers *services.LedgerService
}

// Initialize a shared ledger api singleton instance
var (
	Ledgers = &LedgersApi{
		ApiUsingConfig: ApiUsingConfig{
			container: settings.Container,
		},
		ledgers: services.Ledgers,
	}
)

// LedgerListHandler returns all ledgers which the current login user can access
func (a *LedgersApi) LedgerListHandler(c *core.WebContext) (any, *errs.Error) {
	uid := c.GetCurrentLoginUid()

	ledgers, err := a.ledgers.ListLedgers(c, uid)

	if err != nil {
		log.Errorf(c, "[ledgers.LedgerListHandler] failed to list ledgers for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	personalLedgerName := ""

	user, err := services.Users.GetUserById(c, uid)

	if err == nil {
		personalLedgerName = user.Nickname
	}

	for i := 0; i < len(ledgers); i++ {
		if ledgers[i].IsPersonal {
			ledgers[i].Name = personalLedgerName
		}
	}

	return ledgers, nil
}

// LedgerCreateHandler creates a shared ledger for the current login user
func (a *LedgersApi) LedgerCreateHandler(c *core.WebContext) (any, *errs.Error) {
	var req models.LedgerCreateRequest
	err := c.ShouldBindJSON(&req)

	if err != nil {
		log.Warnf(c, "[ledgers.LedgerCreateHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentLoginUid()

	info, err := a.ledgers.CreateLedger(c, uid, req.Name)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[ledgers.LedgerCreateHandler] failed to create ledger for user \"uid:%d\", because %s", uid, err.Error())
		}

		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return info, nil
}

// LedgerJoinHandler joins a shared ledger by invite code
func (a *LedgersApi) LedgerJoinHandler(c *core.WebContext) (any, *errs.Error) {
	var req models.LedgerJoinRequest
	err := c.ShouldBindJSON(&req)

	if err != nil {
		log.Warnf(c, "[ledgers.LedgerJoinHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentLoginUid()

	info, err := a.ledgers.JoinLedger(c, uid, req.InviteCode)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[ledgers.LedgerJoinHandler] failed to join ledger for user \"uid:%d\", because %s", uid, err.Error())
		}

		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return info, nil
}

// LedgerRenameHandler updates the name of a shared ledger
func (a *LedgersApi) LedgerRenameHandler(c *core.WebContext) (any, *errs.Error) {
	var req models.LedgerRenameRequest
	err := c.ShouldBindJSON(&req)

	if err != nil {
		log.Warnf(c, "[ledgers.LedgerRenameHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentLoginUid()

	if err = a.ledgers.RenameLedger(c, uid, req.LedgerId, req.Name); err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[ledgers.LedgerRenameHandler] failed to rename ledger \"%d\", because %s", req.LedgerId, err.Error())
		}

		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return true, nil
}

// LedgerInviteCodeRegenerateHandler regenerates the invite code of a shared ledger
func (a *LedgersApi) LedgerInviteCodeRegenerateHandler(c *core.WebContext) (any, *errs.Error) {
	var req models.LedgerIdRequest
	err := c.ShouldBindJSON(&req)

	if err != nil {
		log.Warnf(c, "[ledgers.LedgerInviteCodeRegenerateHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentLoginUid()

	inviteCode, err := a.ledgers.RegenerateInviteCode(c, uid, req.LedgerId)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[ledgers.LedgerInviteCodeRegenerateHandler] failed to regenerate invite code, because %s", err.Error())
		}

		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return inviteCode, nil
}

// LedgerMemberListHandler returns all members of a shared ledger
func (a *LedgersApi) LedgerMemberListHandler(c *core.WebContext) (any, *errs.Error) {
	var req models.LedgerIdRequest
	err := c.ShouldBindQuery(&req)

	if err != nil {
		log.Warnf(c, "[ledgers.LedgerMemberListHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentLoginUid()

	members, err := a.ledgers.GetMembers(c, uid, req.LedgerId)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[ledgers.LedgerMemberListHandler] failed to list members of ledger \"%d\", because %s", req.LedgerId, err.Error())
		}

		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return members, nil
}

// LedgerMemberRoleUpdateHandler updates the role of a ledger member
func (a *LedgersApi) LedgerMemberRoleUpdateHandler(c *core.WebContext) (any, *errs.Error) {
	var req models.LedgerMemberRoleUpdateRequest
	err := c.ShouldBindJSON(&req)

	if err != nil {
		log.Warnf(c, "[ledgers.LedgerMemberRoleUpdateHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentLoginUid()

	if err = a.ledgers.UpdateMemberRole(c, uid, req.LedgerId, req.Uid, req.Role); err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[ledgers.LedgerMemberRoleUpdateHandler] failed to update member role, because %s", err.Error())
		}

		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return true, nil
}

// LedgerMemberRemoveHandler removes a member from a shared ledger
func (a *LedgersApi) LedgerMemberRemoveHandler(c *core.WebContext) (any, *errs.Error) {
	var req models.LedgerMemberRemoveRequest
	err := c.ShouldBindJSON(&req)

	if err != nil {
		log.Warnf(c, "[ledgers.LedgerMemberRemoveHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentLoginUid()

	if err = a.ledgers.RemoveMember(c, uid, req.LedgerId, req.Uid); err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[ledgers.LedgerMemberRemoveHandler] failed to remove member, because %s", err.Error())
		}

		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return true, nil
}

// LedgerLeaveHandler removes the current login user from a shared ledger
func (a *LedgersApi) LedgerLeaveHandler(c *core.WebContext) (any, *errs.Error) {
	var req models.LedgerIdRequest
	err := c.ShouldBindJSON(&req)

	if err != nil {
		log.Warnf(c, "[ledgers.LedgerLeaveHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentLoginUid()

	if err = a.ledgers.LeaveLedger(c, uid, req.LedgerId); err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[ledgers.LedgerLeaveHandler] failed to leave ledger \"%d\", because %s", req.LedgerId, err.Error())
		}

		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return true, nil
}

// LedgerDeleteHandler deletes a shared ledger owned by the current login user
func (a *LedgersApi) LedgerDeleteHandler(c *core.WebContext) (any, *errs.Error) {
	var req models.LedgerIdRequest
	err := c.ShouldBindJSON(&req)

	if err != nil {
		log.Warnf(c, "[ledgers.LedgerDeleteHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentLoginUid()

	if err = a.ledgers.DeleteLedger(c, uid, req.LedgerId); err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[ledgers.LedgerDeleteHandler] failed to delete ledger \"%d\", because %s", req.LedgerId, err.Error())
		}

		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return true, nil
}
