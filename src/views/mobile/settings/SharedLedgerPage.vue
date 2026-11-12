<template>
    <f7-page ptr @ptr:refresh="loadLedgers">
        <f7-navbar>
            <f7-nav-left :back-link="tt('Back')"></f7-nav-left>
            <f7-nav-title :title="tt('Shared Ledgers')"></f7-nav-title>
            <f7-nav-right>
                <f7-link icon-f7="plus" :aria-label="tt('Create a Shared Ledger')" @click="createLedger"></f7-link>
            </f7-nav-right>
        </f7-navbar>

        <f7-list strong inset dividers class="margin-vertical-half">
            <f7-list-item group-title :sortable="false">
                <small>{{ tt('Shared Ledgers') }}</small>
            </f7-list-item>
            <f7-list-item v-for="ledger in ledgers" :key="ledger.id"
                          link="#" no-chevron
                          :title="ledger.name"
                          :after="getRoleName(ledger)"
                          @click="showLedgerActions(ledger)">
                <template #footer>
                    <span v-if="isCurrentLedger(ledger)" class="text-color-primary">{{ tt('In Use') }}</span>
                    <span v-else-if="ledger.inviteCode">{{ tt('Invite Code') }}: {{ ledger.inviteCode }}</span>
                    <span v-else>{{ tt('Personal ledger, visible only to you') }}</span>
                </template>
            </f7-list-item>
        </f7-list>

        <f7-list strong inset dividers class="margin-vertical-half">
            <f7-list-button :title="tt('Create a Shared Ledger')" @click="createLedger"></f7-list-button>
            <f7-list-button :title="tt('Join a Shared Ledger')" @click="joinLedger"></f7-list-button>
        </f7-list>

        <f7-block class="no-margin">
            <p class="text-color-gray">{{ tt('Create a shared ledger and invite family members by invite code. All transactions of a shared ledger are visible to its members.') }}</p>
            <p class="text-color-gray">{{ tt('The current ledger is used by every page of the application.') }}</p>
        </f7-block>

        <f7-actions close-by-outside-click close-on-escape :opened="showLedgerActionSheet" @actions:closed="showLedgerActionSheet = false">
            <f7-actions-group v-if="actionLedger">
                <f7-actions-label>{{ actionLedger.name }}</f7-actions-label>
                <f7-actions-button @click="switchLedger(actionLedger)" v-if="!isCurrentLedger(actionLedger)">{{ tt('Use This Ledger') }}</f7-actions-button>
                <f7-actions-button @click="showMembers(actionLedger)">{{ tt('Members') }}</f7-actions-button>
                <f7-actions-button @click="renameLedger(actionLedger)" v-if="!actionLedger.isPersonal">{{ tt('Rename') }}</f7-actions-button>
                <f7-actions-button @click="regenerateInviteCode(actionLedger)" v-if="actionLedger.role === LedgerMemberRole.Owner && !actionLedger.isPersonal">{{ tt('Reset Invite Code') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group v-if="actionLedger && !actionLedger.isPersonal">
                <f7-actions-button color="red" @click="removeLedger(actionLedger)">
                    {{ actionLedger.role === LedgerMemberRole.Owner ? tt('Delete Ledger') : tt('Leave Ledger') }}
                </f7-actions-button>
            </f7-actions-group>
            <f7-actions-group>
                <f7-actions-button bold close>{{ tt('Cancel') }}</f7-actions-button>
            </f7-actions-group>
        </f7-actions>

        <f7-actions close-by-outside-click close-on-escape :opened="showMembersActionSheet" @actions:closed="showMembersActionSheet = false">
            <f7-actions-group v-for="member in members" :key="member.uid">
                <f7-actions-label>
                    {{ member.nickname || member.username }} · {{ getMemberRoleName(member.role) }}{{ member.isCurrent ? ' · ' + tt('Me') : '' }}
                </f7-actions-label>
                <f7-actions-button @click="changeRole(member, LedgerMemberRole.Editor)"
                                   v-if="isOwner && member.role !== LedgerMemberRole.Owner && member.role !== LedgerMemberRole.Editor">{{ tt('Set as Editor') }}</f7-actions-button>
                <f7-actions-button @click="changeRole(member, LedgerMemberRole.Viewer)"
                                   v-if="isOwner && member.role !== LedgerMemberRole.Owner && member.role !== LedgerMemberRole.Viewer">{{ tt('Set as Viewer (read only)') }}</f7-actions-button>
                <f7-actions-button color="red" @click="removeMember(member)"
                                   v-if="isOwner && member.role !== LedgerMemberRole.Owner">{{ tt('Remove') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group v-if="isOwner">
                <f7-actions-button @click="promptAddMember">{{ tt('Add Member by Username') }}</f7-actions-button>
            </f7-actions-group>
            <f7-actions-group>
                <f7-actions-button bold close>{{ tt('Close') }}</f7-actions-button>
            </f7-actions-group>
        </f7-actions>
    </f7-page>
</template>

<script setup lang="ts">
import { ref } from 'vue';

import { useI18n } from '@/locales/helpers.ts';
import { useI18nUIComponents } from '@/lib/ui/mobile.ts';

import { LedgerMemberRole } from '@/models/ledger.ts';
import type { LedgerInfoResponse, LedgerMemberInfoResponse } from '@/models/ledger.ts';

import services from '@/lib/services.ts';
import { getCurrentLedgerId, setCurrentLedgerId } from '@/lib/ledger.ts';

const { tt } = useI18n();
const { showPrompt, showConfirm, showToast } = useI18nUIComponents();

const ledgers = ref<LedgerInfoResponse[]>([]);
const members = ref<LedgerMemberInfoResponse[]>([]);
const actionLedger = ref<LedgerInfoResponse | null>(null);
const membersLedger = ref<LedgerInfoResponse | null>(null);
const isOwner = ref<boolean>(false);

const showLedgerActionSheet = ref<boolean>(false);
const showMembersActionSheet = ref<boolean>(false);

function getRoleName(ledger: LedgerInfoResponse): string {
    if (ledger.isPersonal) {
        return tt('Personal');
    }

    return getMemberRoleName(ledger.role);
}

function getMemberRoleName(role: LedgerMemberRole): string {
    switch (role) {
        case LedgerMemberRole.Owner:
            return tt('Owner');
        case LedgerMemberRole.Editor:
            return tt('Editor');
        case LedgerMemberRole.Viewer:
            return tt('Viewer (read only)');
        default:
            return '';
    }
}

function isCurrentLedger(ledger: LedgerInfoResponse): boolean {
    return getCurrentLedgerId() === ledger.id;
}

function loadLedgers(): void {
    services.getLedgers().then(response => {
        ledgers.value = (response.data && response.data.result) ? response.data.result : [];

        const current = getCurrentLedgerId();

        if (current > 0 && !ledgers.value.some(item => parseInt(item.id, 10) === current)) {
            setCurrentLedgerId('');
        }
    }).catch(error => {
        showToast(error.message || error);
    });
}

function showLedgerActions(ledger: LedgerInfoResponse): void {
    actionLedger.value = ledger;
    showLedgerActionSheet.value = true;
}

function switchLedger(ledger: LedgerInfoResponse | null): void {
    if (!ledger || isCurrentLedger(ledger)) {
        return;
    }

    setCurrentLedgerId(ledger.id);
    showLedgerActionSheet.value = false;
    window.location.reload();
}

function createLedger(): void {
    showPrompt(tt('Ledger Name'), '', (value: string) => {
        if (!value) {
            return;
        }

        services.createLedger({ name: value }).then(() => {
            showToast(tt('The shared ledger has been created'));
            loadLedgers();
        }).catch(error => {
            showToast(error.message || error);
        });
    });
}

function joinLedger(): void {
    showPrompt(tt('Invite Code'), '', (value: string) => {
        if (!value) {
            return;
        }

        services.joinLedger({ inviteCode: value.toUpperCase() }).then(() => {
            showToast(tt('You have joined the shared ledger'));
            loadLedgers();
        }).catch(error => {
            showToast(error.message || error);
        });
    });
}

function renameLedger(ledger: LedgerInfoResponse | null): void {
    if (!ledger) {
        return;
    }

    showLedgerActionSheet.value = false;

    showPrompt(tt('Ledger Name'), ledger.name, (value: string) => {
        if (!value) {
            return;
        }

        services.renameLedger({ ledgerId: ledger.id, name: value }).then(() => {
            showToast(tt('The ledger has been renamed'));
            loadLedgers();
        }).catch(error => {
            showToast(error.message || error);
        });
    });
}

function regenerateInviteCode(ledger: LedgerInfoResponse | null): void {
    if (!ledger) {
        return;
    }

    showLedgerActionSheet.value = false;

    showConfirm(tt('The invite code will be reset and the old one will no longer be usable. Are you sure?'), () => {
        services.regenerateLedgerInviteCode({ ledgerId: ledger.id }).then(() => {
            showToast(tt('The invite code has been reset'));
            loadLedgers();
        }).catch(error => {
            showToast(error.message || error);
        });
    });
}

function removeLedger(ledger: LedgerInfoResponse | null): void {
    if (!ledger) {
        return;
    }

    showLedgerActionSheet.value = false;

    const isOwnerUser = ledger.role === LedgerMemberRole.Owner;
    const message = isOwnerUser
        ? tt('All transaction data of this shared ledger will be deleted permanently. Are you sure?')
        : tt('You will no longer be able to access the transactions of this shared ledger. Are you sure?');

    showConfirm(message, () => {
        const promise = isOwnerUser
            ? services.deleteLedger({ ledgerId: ledger.id })
            : services.leaveLedger({ ledgerId: ledger.id });

        promise.then(() => {
            if (isCurrentLedger(ledger)) {
                setCurrentLedgerId('');
            }

            showToast(tt('The operation succeeded'));
            loadLedgers();
        }).catch(error => {
            showToast(error.message || error);
        });
    });
}

function showMembers(ledger: LedgerInfoResponse | null): void {
    if (!ledger) {
        return;
    }

    showLedgerActionSheet.value = false;
    membersLedger.value = ledger;
    isOwner.value = ledger.role === LedgerMemberRole.Owner;
    members.value = [];

    services.getLedgerMembers({ ledgerId: ledger.id }).then(response => {
        members.value = (response.data && response.data.result) ? response.data.result : [];
        showMembersActionSheet.value = true;
    }).catch(error => {
        showToast(error.message || error);
    });
}

function promptAddMember(): void {
    if (!membersLedger.value) {
        return;
    }

    showMembersActionSheet.value = false;

    showPrompt(tt('Enter the username to add'), '', (value: string) => {
        const username = (value || '').trim();

        if (!username) {
            return;
        }

        services.addLedgerMember({
            ledgerId: membersLedger.value ? membersLedger.value.id : '0',
            username: username,
            role: LedgerMemberRole.Editor
        }).then(response => {
            const result = response.data ? response.data.result : null;

            if (result && membersLedger.value) {
                membersLedger.value = result;
            }

            showToast(tt('The member has been added'));
            loadLedgers();

            if (membersLedger.value) {
                services.getLedgerMembers({ ledgerId: membersLedger.value.id }).then(res => {
                    members.value = (res.data && res.data.result) ? res.data.result : [];
                    showMembersActionSheet.value = true;
                });
            }
        }).catch(error => {
            showToast(error.message || error);
        });
    });
}

function changeRole(member: LedgerMemberInfoResponse, role: LedgerMemberRole): void {
    if (!membersLedger.value) {
        return;
    }

    showMembersActionSheet.value = false;

    services.updateLedgerMemberRole({
        ledgerId: membersLedger.value.id,
        uid: member.uid,
        role: role
    }).then(() => {
        showToast(tt('The member role has been updated'));
        showMembers(membersLedger.value);
    }).catch(error => {
        showToast(error.message || error);
    });
}

function removeMember(member: LedgerMemberInfoResponse): void {
    if (!membersLedger.value) {
        return;
    }

    showMembersActionSheet.value = false;

    showConfirm(tt('Remove this member from the shared ledger?'), () => {
        services.removeLedgerMember({ ledgerId: membersLedger.value?.id || '', uid: member.uid }).then(() => {
            showToast(tt('The operation succeeded'));
            showMembers(membersLedger.value);
        }).catch(error => {
            showToast(error.message || error);
        });
    });
}

loadLedgers();
</script>
