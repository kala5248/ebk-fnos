package services

import (
	"crypto/rand"
	"strings"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"xorm.io/xorm"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
)

const ledgerInviteCodeCharset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
const ledgerInviteCodeLength = 8

// LedgerService represents a shared ledger service
type LedgerService struct {
	ServiceUsingDB
	ServiceUsingUuid
}

// Initialize a ledger service singleton instance
var (
	Ledgers = &LedgerService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
		ServiceUsingUuid: ServiceUsingUuid{
			container: uuid.Container,
		},
	}
)

// ListLedgers returns all ledgers which the given user can access (the personal ledger is always the first item)
func (s *LedgerService) ListLedgers(c core.Context, uid int64) ([]*models.LedgerInfo, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	ledgers := make([]*models.LedgerInfo, 0)
	ledgers = append(ledgers, &models.LedgerInfo{
		LedgerId:    0,
		Name:        "",
		IsPersonal:  true,
		Role:        models.LEDGER_MEMBER_ROLE_OWNER,
		MemberCount: 1,
	})

	sess := s.UserDB().NewSession(c)
	members := make([]*models.LedgerMember, 0, 20)
	err := sess.Where("uid=?", uid).Find(&members)

	if err != nil {
		return nil, err
	}

	for i := 0; i < len(members); i++ {
		ledger, err := s.GetLedgerById(c, members[i].LedgerId)

		if err != nil {
			if errs.IsCustomError(err) {
				continue
			}

			return nil, err
		}

		memberCount, err := s.GetLedgerMemberCount(c, ledger.LedgerId)

		if err != nil {
			return nil, err
		}

		info := &models.LedgerInfo{
			LedgerId:    ledger.LedgerId,
			Name:        ledger.Name,
			IsPersonal:  false,
			Role:        members[i].Role,
			MemberCount: memberCount,
			InviteCode:  ledger.InviteCode,
			OwnerUid:    ledger.OwnerUid,
		}

		if ledger.OwnerUid == uid {
			info.OwnerName, _ = s.getUserNickname(c, uid)
		} else {
			info.OwnerName, _ = s.getUserNickname(c, ledger.OwnerUid)
		}

		ledgers = append(ledgers, info)
	}

	return ledgers, nil
}

// GetLedgerById returns the ledger model by ledger id
func (s *LedgerService) GetLedgerById(c core.Context, ledgerId int64) (*models.Ledger, error) {
	if ledgerId <= 0 {
		return nil, errs.ErrLedgerNotFound
	}

	ledger := &models.Ledger{}
	has, err := s.UserDB().NewSession(c).ID(ledgerId).Where("deleted=?", false).Get(ledger)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrLedgerNotFound
	}

	return ledger, nil
}

// GetLedgerMemberCount returns the member count of the given ledger
func (s *LedgerService) GetLedgerMemberCount(c core.Context, ledgerId int64) (int, error) {
	count, err := s.UserDB().NewSession(c).Where("ledger_id=?", ledgerId).Count(new(models.LedgerMember))

	if err != nil {
		return 0, err
	}

	return int(count), nil
}

// ResolveDataScope resolves the data scope (which uid owns the data and which role the current user has) of the given ledger
func (s *LedgerService) ResolveDataScope(c core.Context, uid int64, ledgerId int64) (*models.LedgerDataScope, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if ledgerId <= 0 {
		return &models.LedgerDataScope{
			LedgerId: 0,
			DataUid:  uid,
			Role:     models.LEDGER_MEMBER_ROLE_OWNER,
		}, nil
	}

	ledger, err := s.GetLedgerById(c, ledgerId)

	if err != nil {
		return nil, err
	}

	member := &models.LedgerMember{}
	has, err := s.UserDB().NewSession(c).Where("ledger_id=? AND uid=?", ledgerId, uid).Get(member)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrLedgerNotMember
	}

	return &models.LedgerDataScope{
		LedgerId: ledger.LedgerId,
		DataUid:  ledger.DataUid,
		Role:     member.Role,
	}, nil
}

// CreateLedger creates a new shared ledger, and creates an independent shadow data owner for it
func (s *LedgerService) CreateLedger(c core.Context, uid int64, name string) (*models.LedgerInfo, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	name = strings.TrimSpace(name)

	if name == "" {
		return nil, errs.ErrLedgerNameIsEmpty
	}

	if len([]rune(name)) > 64 {
		return nil, errs.ErrLedgerNameIsTooLong
	}

	owner, err := Users.GetUserById(c, uid)

	if err != nil {
		return nil, err
	}

	ledgerId := s.GenerateUuid(uuid.UUID_TYPE_DEFAULT)

	if ledgerId <= 0 {
		return nil, errs.ErrSystemIsBusy
	}

	inviteCode, err := s.generateInviteCode(c)

	if err != nil {
		return nil, err
	}

	dataUser := &models.User{
		Username:              "ledger_" + utils.Int64ToString(ledgerId),
		Email:                 "ledger+" + utils.Int64ToString(ledgerId) + "@ezbookkeeping.local",
		Nickname:              name,
		Password:              "",
		Language:              owner.Language,
		DefaultCurrency:       owner.DefaultCurrency,
		FirstDayOfWeek:        owner.FirstDayOfWeek,
		FiscalYearStart:       owner.FiscalYearStart,
		CalendarDisplayType:   owner.CalendarDisplayType,
		DateDisplayType:       owner.DateDisplayType,
		LongDateFormat:        owner.LongDateFormat,
		ShortDateFormat:       owner.ShortDateFormat,
		LongTimeFormat:        owner.LongTimeFormat,
		ShortTimeFormat:       owner.ShortTimeFormat,
		FiscalYearFormat:      owner.FiscalYearFormat,
		CurrencyDisplayType:   owner.CurrencyDisplayType,
		NumeralSystem:         owner.NumeralSystem,
		DecimalSeparator:      owner.DecimalSeparator,
		DigitGroupingSymbol:   owner.DigitGroupingSymbol,
		DigitGrouping:         owner.DigitGrouping,
		CoordinateDisplayType: owner.CoordinateDisplayType,
		ExpenseAmountColor:    owner.ExpenseAmountColor,
		IncomeAmountColor:     owner.IncomeAmountColor,
		TransactionEditScope:  models.TRANSACTION_EDIT_SCOPE_ALL,
		FeatureRestriction:    0,
		EmailVerified:         true,
	}

	err = Users.CreateUser(c, dataUser, true)

	if err != nil {
		log.Errorf(c, "[ledgers.CreateLedger] failed to create shadow data owner for ledger \"%s\", because %s", name, err.Error())
		return nil, err
	}

	ledger := &models.Ledger{
		LedgerId:        ledgerId,
		Name:            name,
		OwnerUid:        uid,
		DataUid:         dataUser.Uid,
		InviteCode:      inviteCode,
		Deleted:         false,
		CreatedUnixTime: time.Now().Unix(),
		UpdatedUnixTime: time.Now().Unix(),
	}

	member := &models.LedgerMember{
		LedgerId:        ledger.LedgerId,
		Uid:             uid,
		Role:            models.LEDGER_MEMBER_ROLE_OWNER,
		CreatedUnixTime: time.Now().Unix(),
		UpdatedUnixTime: time.Now().Unix(),
	}

	err = s.UserDB().DoTransaction(c, func(sess *xorm.Session) error {
		_, err := sess.Insert(ledger)

		if err != nil {
			return err
		}

		_, err = sess.Insert(member)

		return err
	})

	if err != nil {
		log.Errorf(c, "[ledgers.CreateLedger] failed to create ledger \"%s\" for user \"uid:%d\", because %s", name, uid, err.Error())
		return nil, err
	}

	err = s.copyDefaultCategories(c, uid, dataUser.Uid)

	if err != nil {
		log.Errorf(c, "[ledgers.CreateLedger] failed to copy categories for ledger \"%s\", because %s", name, err.Error())
	}

	info := &models.LedgerInfo{
		LedgerId:    ledger.LedgerId,
		Name:        ledger.Name,
		IsPersonal:  false,
		Role:        models.LEDGER_MEMBER_ROLE_OWNER,
		MemberCount: 1,
		InviteCode:  ledger.InviteCode,
		OwnerUid:    ledger.OwnerUid,
		OwnerName:   owner.Nickname,
	}

	log.Infof(c, "[ledgers.CreateLedger] user \"uid:%d\" created ledger \"%s\" (ledger id: %d)", uid, name, ledger.LedgerId)

	return info, nil
}

// JoinLedger joins a shared ledger by invite code
func (s *LedgerService) JoinLedger(c core.Context, uid int64, inviteCode string) (*models.LedgerInfo, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	inviteCode = strings.ToUpper(strings.TrimSpace(inviteCode))

	if inviteCode == "" {
		return nil, errs.ErrLedgerInviteCodeIsEmpty
	}

	ledger := &models.Ledger{}
	has, err := s.UserDB().NewSession(c).Where("invite_code=? AND deleted=?", inviteCode, false).Get(ledger)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrLedgerInviteCodeIsInvalid
	}

	if ledger.OwnerUid == uid {
		return nil, errs.ErrLedgerAlreadyJoined
	}

	existing := &models.LedgerMember{}
	has, err = s.UserDB().NewSession(c).Where("ledger_id=? AND uid=?", ledger.LedgerId, uid).Get(existing)

	if err != nil {
		return nil, err
	} else if has {
		return nil, errs.ErrLedgerAlreadyJoined
	}

	member := &models.LedgerMember{
		LedgerId:        ledger.LedgerId,
		Uid:             uid,
		Role:            models.LEDGER_MEMBER_ROLE_EDITOR,
		CreatedUnixTime: time.Now().Unix(),
		UpdatedUnixTime: time.Now().Unix(),
	}

	err = s.UserDB().DoTransaction(c, func(sess *xorm.Session) error {
		_, err := sess.Insert(member)
		return err
	})

	if err != nil {
		return nil, err
	}

	memberCount, err := s.GetLedgerMemberCount(c, ledger.LedgerId)

	if err != nil {
		return nil, err
	}

	ownerName, _ := s.getUserNickname(c, ledger.OwnerUid)

	log.Infof(c, "[ledgers.JoinLedger] user \"uid:%d\" joined ledger \"%s\" (ledger id: %d)", uid, ledger.Name, ledger.LedgerId)

	return &models.LedgerInfo{
		LedgerId:    ledger.LedgerId,
		Name:        ledger.Name,
		Role:        models.LEDGER_MEMBER_ROLE_EDITOR,
		MemberCount: memberCount,
		InviteCode:  ledger.InviteCode,
		OwnerUid:    ledger.OwnerUid,
		OwnerName:   ownerName,
	}, nil
}

// GetMembers returns all members of the given ledger
func (s *LedgerService) GetMembers(c core.Context, uid int64, ledgerId int64) ([]*models.LedgerMemberInfo, error) {
	if _, err := s.ResolveDataScope(c, uid, ledgerId); err != nil {
		return nil, err
	}

	members := make([]*models.LedgerMember, 0, 20)
	err := s.UserDB().NewSession(c).Where("ledger_id=?", ledgerId).Find(&members)

	if err != nil {
		return nil, err
	}

	infos := make([]*models.LedgerMemberInfo, 0, len(members))

	for i := 0; i < len(members); i++ {
		user, err := Users.GetUserById(c, members[i].Uid)

		if err != nil {
			log.Warnf(c, "[ledgers.GetMembers] cannot get user \"uid:%d\" of ledger \"%d\", because %s", members[i].Uid, ledgerId, err.Error())
			continue
		}

		infos = append(infos, &models.LedgerMemberInfo{
			Uid:       members[i].Uid,
			Username:  user.Username,
			Nickname:  user.Nickname,
			Role:      members[i].Role,
			JoinedAt:  members[i].CreatedUnixTime,
			IsCurrent: members[i].Uid == uid,
		})
	}

	return infos, nil
}

// RenameLedger updates the name of the given ledger
func (s *LedgerService) RenameLedger(c core.Context, uid int64, ledgerId int64, name string) error {
	if err := s.requireLedgerOwner(c, uid, ledgerId); err != nil {
		return err
	}

	name = strings.TrimSpace(name)

	if name == "" {
		return errs.ErrLedgerNameIsEmpty
	}

	if len([]rune(name)) > 64 {
		return errs.ErrLedgerNameIsTooLong
	}

	_, err := s.UserDB().NewSession(c).Table(new(models.Ledger)).
		Where("ledger_id=? AND deleted=?", ledgerId, false).
		Cols("name", "updated_unix_time").
		Update(&models.Ledger{Name: name, UpdatedUnixTime: time.Now().Unix()})

	return err
}

// RegenerateInviteCode regenerates the invite code of the given ledger
func (s *LedgerService) RegenerateInviteCode(c core.Context, uid int64, ledgerId int64) (string, error) {
	if err := s.requireLedgerOwner(c, uid, ledgerId); err != nil {
		return "", err
	}

	inviteCode, err := s.generateInviteCode(c)

	if err != nil {
		return "", err
	}

	_, err = s.UserDB().NewSession(c).Table(new(models.Ledger)).
		Where("ledger_id=? AND deleted=?", ledgerId, false).
		Cols("invite_code", "updated_unix_time").
		Update(&models.Ledger{InviteCode: inviteCode, UpdatedUnixTime: time.Now().Unix()})

	if err != nil {
		return "", err
	}

	return inviteCode, nil
}

// UpdateMemberRole updates the role of a member in the given ledger
func (s *LedgerService) UpdateMemberRole(c core.Context, uid int64, ledgerId int64, memberUid int64, role models.LedgerMemberRole) error {
	if err := s.requireLedgerOwner(c, uid, ledgerId); err != nil {
		return err
	}

	if role != models.LEDGER_MEMBER_ROLE_EDITOR && role != models.LEDGER_MEMBER_ROLE_VIEWER {
		return errs.ErrLedgerMemberRoleIsInvalid
	}

	member := &models.LedgerMember{}
	has, err := s.UserDB().NewSession(c).Where("ledger_id=? AND uid=?", ledgerId, memberUid).Get(member)

	if err != nil {
		return err
	} else if !has {
		return errs.ErrLedgerNotMember
	}

	if member.Role == models.LEDGER_MEMBER_ROLE_OWNER {
		return errs.ErrLedgerOwnerRoleCannotBeChanged
	}

	_, err = s.UserDB().NewSession(c).Table(new(models.LedgerMember)).
		Where("ledger_id=? AND uid=?", ledgerId, memberUid).
		Cols("role", "updated_unix_time").
		Update(&models.LedgerMember{Role: role, UpdatedUnixTime: time.Now().Unix()})

	return err
}

// RemoveMember removes a member from the given ledger
func (s *LedgerService) RemoveMember(c core.Context, uid int64, ledgerId int64, memberUid int64) error {
	if err := s.requireLedgerOwner(c, uid, ledgerId); err != nil {
		return err
	}

	if memberUid == uid {
		return errs.ErrLedgerOwnerCannotBeRemoved
	}

	_, err := s.UserDB().NewSession(c).Where("ledger_id=? AND uid=?", ledgerId, memberUid).Delete(new(models.LedgerMember))

	return err
}

// LeaveLedger removes the current user from the given ledger
func (s *LedgerService) LeaveLedger(c core.Context, uid int64, ledgerId int64) error {
	scope, err := s.ResolveDataScope(c, uid, ledgerId)

	if err != nil {
		return err
	}

	if scope.Role == models.LEDGER_MEMBER_ROLE_OWNER {
		return errs.ErrLedgerOwnerCannotLeave
	}

	_, err = s.UserDB().NewSession(c).Where("ledger_id=? AND uid=?", ledgerId, uid).Delete(new(models.LedgerMember))

	return err
}

// DeleteLedger deletes the given ledger (only the owner can do this)
func (s *LedgerService) DeleteLedger(c core.Context, uid int64, ledgerId int64) error {
	if err := s.requireLedgerOwner(c, uid, ledgerId); err != nil {
		return err
	}

	_, err := s.UserDB().NewSession(c).Table(new(models.Ledger)).
		Where("ledger_id=? AND deleted=?", ledgerId, false).
		Cols("deleted", "deleted_unix_time", "updated_unix_time").
		Update(&models.Ledger{Deleted: true, DeletedUnixTime: time.Now().Unix(), UpdatedUnixTime: time.Now().Unix()})

	if err != nil {
		return err
	}

	_, err = s.UserDB().NewSession(c).Where("ledger_id=?", ledgerId).Delete(new(models.LedgerMember))

	return err
}

func (s *LedgerService) requireLedgerOwner(c core.Context, uid int64, ledgerId int64) error {
	scope, err := s.ResolveDataScope(c, uid, ledgerId)

	if err != nil {
		return err
	}

	if scope.Role != models.LEDGER_MEMBER_ROLE_OWNER {
		return errs.ErrLedgerNotOwner
	}

	return nil
}

func (s *LedgerService) generateInviteCode(c core.Context) (string, error) {
	for i := 0; i < 5; i++ {
		buf := make([]byte, ledgerInviteCodeLength)

		if _, err := rand.Read(buf); err != nil {
			return "", err
		}

		code := make([]byte, ledgerInviteCodeLength)

		for j := 0; j < ledgerInviteCodeLength; j++ {
			code[j] = ledgerInviteCodeCharset[int(buf[j])%len(ledgerInviteCodeCharset)]
		}

		existed, err := s.UserDB().NewSession(c).Where("invite_code=?", string(code)).Exist(new(models.Ledger))

		if err != nil {
			return "", err
		} else if !existed {
			return string(code), nil
		}
	}

	return "", errs.ErrSystemIsBusy
}

func (s *LedgerService) getUserNickname(c core.Context, uid int64) (string, error) {
	user, err := Users.GetUserById(c, uid)

	if err != nil {
		return "", err
	}

	return user.Nickname, nil
}

// copyDefaultCategories copies all categories of the owner to the shadow data owner, so that the shared ledger
// starts with the same category tree as the owner's personal ledger
func (s *LedgerService) copyDefaultCategories(c core.Context, fromUid int64, toUid int64) error {
	categories, err := TransactionCategories.GetAllCategoriesByUid(c, fromUid, 0, -1)

	if err != nil {
		return err
	}

	if len(categories) == 0 {
		return nil
	}

	subCategories := make(map[int64][]*models.TransactionCategory)
	createMap := make(map[*models.TransactionCategory][]*models.TransactionCategory)
	newPrimaryByOriginalId := make(map[int64]*models.TransactionCategory)

	for i := 0; i < len(categories); i++ {
		category := categories[i]

		if category.Hidden {
			continue
		}

		if category.ParentCategoryId != models.LevelOneTransactionCategoryParentId {
			subCategories[category.ParentCategoryId] = append(subCategories[category.ParentCategoryId], category)
			continue
		}

		primary := &models.TransactionCategory{
			Type:         category.Type,
			Name:         category.Name,
			DisplayOrder: category.DisplayOrder,
			Icon:         category.Icon,
			IconType:     category.IconType,
			Color:        category.Color,
			Comment:      category.Comment,
		}

		createMap[nil] = append(createMap[nil], primary)
		createMap[primary] = make([]*models.TransactionCategory, 0, 10)
		newPrimaryByOriginalId[category.CategoryId] = primary
	}

	for originalParentId, subs := range subCategories {
		primary, exists := newPrimaryByOriginalId[originalParentId]

		if !exists {
			continue
		}

		for i := 0; i < len(subs); i++ {
			createMap[primary] = append(createMap[primary], &models.TransactionCategory{
				Type:         subs[i].Type,
				Name:         subs[i].Name,
				DisplayOrder: subs[i].DisplayOrder,
				Icon:         subs[i].Icon,
				IconType:     subs[i].IconType,
				Color:        subs[i].Color,
				Comment:      subs[i].Comment,
			})
		}
	}

	_, err = TransactionCategories.CreateCategories(c, toUid, createMap)

	return err
}
