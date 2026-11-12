package models

// LedgerMemberRole represents the role of a member in a shared ledger
type LedgerMemberRole byte

// Ledger member roles
const (
	LEDGER_MEMBER_ROLE_OWNER  LedgerMemberRole = 1
	LEDGER_MEMBER_ROLE_EDITOR LedgerMemberRole = 2
	LEDGER_MEMBER_ROLE_VIEWER LedgerMemberRole = 3
)

// IsReadOnly returns whether the role is not allowed to modify any data
func (r LedgerMemberRole) IsReadOnly() bool {
	return r == LEDGER_MEMBER_ROLE_VIEWER
}

// Ledger represents a shared ledger data stored in database
type Ledger struct {
	LedgerId        int64  `xorm:"PK"`
	Name            string `xorm:"VARCHAR(64) NOT NULL"`
	OwnerUid        int64  `xorm:"INDEX(IDX_ledger_owner_uid_deleted) NOT NULL"`
	DataUid         int64  `xorm:"INDEX(IDX_ledger_data_uid) NOT NULL"`
	InviteCode      string `xorm:"VARCHAR(16) UNIQUE NOT NULL"`
	Deleted         bool   `xorm:"INDEX(IDX_ledger_owner_uid_deleted) NOT NULL"`
	CreatedUnixTime int64
	UpdatedUnixTime int64
	DeletedUnixTime int64
}

// LedgerMember represents a member of a shared ledger stored in database
type LedgerMember struct {
	LedgerId        int64            `xorm:"PK"`
	Uid             int64            `xorm:"PK"`
	Role            LedgerMemberRole `xorm:"TINYINT NOT NULL"`
	CreatedUnixTime int64
	UpdatedUnixTime int64
}

// LedgerInfo represents a view-object of a ledger visible to the current user
type LedgerInfo struct {
	LedgerId    int64            `json:"id,string"`
	Name        string           `json:"name"`
	IsPersonal  bool             `json:"isPersonal"`
	Role        LedgerMemberRole `json:"role"`
	MemberCount int              `json:"memberCount"`
	InviteCode  string           `json:"inviteCode,omitempty"`
	OwnerUid    int64            `json:"ownerUid,string,omitempty"`
	OwnerName   string           `json:"ownerName,omitempty"`
}

// LedgerDataScope represents the data scope resolved from a ledger selection
type LedgerDataScope struct {
	LedgerId int64
	DataUid  int64
	Role     LedgerMemberRole
}

// LedgerCreateRequest represents all parameters of ledger creation request
type LedgerCreateRequest struct {
	Name string `json:"name" binding:"required,notBlank,max=64"`
}

// LedgerAddMemberRequest represents all parameters of adding a member to a ledger by username
type LedgerAddMemberRequest struct {
	LedgerId int64            `json:"ledgerId,string" binding:"min=0"`
	Username string           `json:"username" binding:"required,notBlank,max=32"`
	Role     LedgerMemberRole `json:"role"`
}

// LedgerJoinRequest represents all parameters of joining a shared ledger request
type LedgerJoinRequest struct {
	InviteCode string `json:"inviteCode" binding:"required,notBlank,max=16"`
}

// LedgerIdRequest represents all parameters of a request which only needs the ledger id
type LedgerIdRequest struct {
	LedgerId int64 `json:"ledgerId,string" form:"ledgerId,string" binding:"required,min=1"`
}

// LedgerRenameRequest represents all parameters of ledger renaming request
type LedgerRenameRequest struct {
	LedgerId int64  `json:"ledgerId,string" binding:"required,min=1"`
	Name     string `json:"name" binding:"required,notBlank,max=64"`
}

// LedgerMemberInfo represents a view-object of a ledger member
type LedgerMemberInfo struct {
	Uid       int64            `json:"uid,string"`
	Username  string           `json:"username"`
	Nickname  string           `json:"nickname"`
	Role      LedgerMemberRole `json:"role"`
	JoinedAt  int64            `json:"joinedAt"`
	IsCurrent bool             `json:"isCurrent,omitempty"`
}

// LedgerMemberRoleUpdateRequest represents all parameters of updating a ledger member role request
type LedgerMemberRoleUpdateRequest struct {
	LedgerId int64            `json:"ledgerId,string" binding:"required,min=1"`
	Uid      int64            `json:"uid,string" binding:"required,min=1"`
	Role     LedgerMemberRole `json:"role" binding:"required,min=2,max=3"`
}

// LedgerMemberRemoveRequest represents all parameters of removing a ledger member request
type LedgerMemberRemoveRequest struct {
	LedgerId int64 `json:"ledgerId,string" binding:"required,min=1"`
	Uid      int64 `json:"uid,string" binding:"required,min=1"`
}
