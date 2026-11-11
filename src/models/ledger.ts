export enum LedgerMemberRole {
    Owner = 1,
    Editor = 2,
    Viewer = 3
}

export interface LedgerInfoResponse {
    readonly id: string;
    readonly name: string;
    readonly isPersonal: boolean;
    readonly role: LedgerMemberRole;
    readonly memberCount: number;
    readonly inviteCode?: string;
    readonly ownerUid?: string;
    readonly ownerName?: string;
}

export interface LedgerMemberInfoResponse {
    readonly uid: string;
    readonly username: string;
    readonly nickname: string;
    readonly role: LedgerMemberRole;
    readonly joinedAt: number;
    readonly isCurrent?: boolean;
}

export interface LedgerCreateRequest {
    readonly name: string;
}

export interface LedgerJoinRequest {
    readonly inviteCode: string;
}

export interface LedgerRenameRequest {
    readonly ledgerId: string;
    readonly name: string;
}

export interface LedgerInviteCodeRegenerateRequest {
    readonly ledgerId: string;
}

export interface LedgerMemberRoleUpdateRequest {
    readonly ledgerId: string;
    readonly uid: string;
    readonly role: LedgerMemberRole;
}

export interface LedgerMemberRemoveRequest {
    readonly ledgerId: string;
    readonly uid: string;
}

export interface LedgerDeleteRequest {
    readonly ledgerId: string;
}

export interface LedgerLeaveRequest {
    readonly ledgerId: string;
}

export const ALL_LEDGER_MEMBER_ROLES: LedgerMemberRole[] = [
    LedgerMemberRole.Owner,
    LedgerMemberRole.Editor,
    LedgerMemberRole.Viewer
];
