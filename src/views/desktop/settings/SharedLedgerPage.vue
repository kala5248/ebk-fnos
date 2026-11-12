<template>
    <v-row>
        <v-col cols="12">
            <v-card class="setting-items" :title="tt('Shared Ledgers')">
                <v-card-text class="pa-0 text-body-medium">
                    <div class="px-4 pt-2 pb-2 text-body-medium text-medium-emphasis">
                        {{ tt('Create a shared ledger and invite family members by invite code. All transactions of a shared ledger are visible to its members.') }}
                        <br/>
                        {{ tt('The current ledger is used by every page of the application.') }}
                    </div>

                    <v-list lines="two" class="py-0">
                        <v-list-item v-for="ledger in ledgers" :key="ledger.id">
                            <template #title>
                                <div class="d-flex align-center flex-wrap gap-x-2">
                                    <v-icon :icon="ledger.isPersonal ? mdiAccountOutline : mdiAccountMultipleOutline"
                                            size="20"/>
                                    <span class="text-body-large">{{ ledger.name }}</span>
                                    <v-chip size="small" :color="ledger.isPersonal ? 'default' : 'primary'" variant="tonal">
                                        {{ getRoleName(ledger) }}
                                    </v-chip>
                                    <v-chip size="small" variant="tonal" v-if="!ledger.isPersonal">
                                        {{ tt('{count} members', { count: ledger.memberCount }) }}
                                    </v-chip>
                                    <v-chip size="small" color="success" variant="tonal" v-if="isCurrentLedger(ledger)">
                                        {{ tt('In Use') }}
                                    </v-chip>
                                </div>
                            </template>
                            <template #subtitle>
                                <div class="d-flex align-center flex-wrap gap-x-2 text-body-medium">
                                    <span v-if="ledger.inviteCode">{{ tt('Invite Code') }}: {{ ledger.inviteCode }}</span>
                                    <span v-else>{{ tt('Personal ledger, visible only to you') }}</span>
                                </div>
                            </template>
                            <template #append>
                                <div class="d-flex align-center flex-wrap justify-end">
                                    <v-btn variant="text" color="default" size="small" :disabled="isCurrentLedger(ledger)"
                                           @click="switchLedger(ledger)">{{ tt('Use This Ledger') }}</v-btn>
                                    <v-btn variant="text" color="default" size="small"
                                           @click="showMembers(ledger)">{{ tt('Members') }}</v-btn>
                                    <v-btn variant="text" color="default" size="small" v-if="!ledger.isPersonal"
                                           @click="showRenameDialogFor(ledger)">{{ tt('Rename') }}</v-btn>
                                    <v-btn variant="text" color="default" size="small" v-if="ledger.role === LedgerMemberRole.Owner && !ledger.isPersonal"
                                           @click="regenerateInviteCode(ledger)">{{ tt('Reset Invite Code') }}</v-btn>
                                    <v-btn variant="text" color="error" size="small" v-if="!ledger.isPersonal"
                                           @click="showRemoveDialogFor(ledger)">
                                        {{ ledger.role === LedgerMemberRole.Owner ? tt('Delete Ledger') : tt('Leave Ledger') }}
                                    </v-btn>
                                </div>
                            </template>
                        </v-list-item>
                    </v-list>
                </v-card-text>

                <v-card-actions>
                    <v-spacer/>
                    <v-btn variant="outlined" color="default" @click="showJoinDialog = true">
                        {{ tt('Join a Shared Ledger') }}
                    </v-btn>
                    <v-btn color="primary" @click="showCreateDialog = true">
                        {{ tt('Create a Shared Ledger') }}
                    </v-btn>
                </v-card-actions>
            </v-card>
        </v-col>
    </v-row>

    <v-dialog width="500" v-model="showCreateDialog">
        <v-card :title="tt('Create a Shared Ledger')">
            <v-card-text>
                <v-text-field density="compact" persistent-placeholder maxlength="64"
                              :label="tt('Ledger Name')" :placeholder="tt('Example: Family Ledger')"
                              v-model.trim="newLedgerName"/>
            </v-card-text>
            <v-card-actions>
                <v-spacer/>
                <v-btn variant="text" color="default" @click="showCreateDialog = false">{{ tt('Cancel') }}</v-btn>
                <v-btn color="primary" :loading="submitting" @click="createLedger">{{ tt('Create') }}</v-btn>
            </v-card-actions>
        </v-card>
    </v-dialog>

    <v-dialog width="500" v-model="showJoinDialog">
        <v-card :title="tt('Join a Shared Ledger')">
            <v-card-text>
                <v-text-field density="compact" persistent-placeholder maxlength="16" autocapitalize="characters"
                              :label="tt('Invite Code')" :placeholder="tt('The invite code shared by the ledger owner')"
                              v-model.trim="joinInviteCode"/>
            </v-card-text>
            <v-card-actions>
                <v-spacer/>
                <v-btn variant="text" color="default" @click="showJoinDialog = false">{{ tt('Cancel') }}</v-btn>
                <v-btn color="primary" :loading="submitting" @click="joinLedger">{{ tt('Join') }}</v-btn>
            </v-card-actions>
        </v-card>
    </v-dialog>

    <v-dialog width="600" v-model="showRenameDialog">
        <v-card :title="tt('Rename Ledger')">
            <v-card-text>
                <v-text-field density="compact" persistent-placeholder maxlength="64"
                              :label="tt('Ledger Name')" v-model.trim="renameLedgerName"/>
            </v-card-text>
            <v-card-actions>
                <v-spacer/>
                <v-btn variant="text" color="default" @click="showRenameDialog = false">{{ tt('Cancel') }}</v-btn>
                <v-btn color="primary" :loading="submitting" @click="renameLedger">{{ tt('Save') }}</v-btn>
            </v-card-actions>
        </v-card>
    </v-dialog>

    <v-dialog width="700" v-model="showMembersDialog">
        <v-card :title="tt('Ledger Members')">
            <v-card-text>
                <div class="d-flex align-center gap-x-2 mb-3" v-if="currentLedger && currentLedger.role === LedgerMemberRole.Owner">
                    <v-text-field density="compact" persistent-placeholder maxlength="32" hide-details
                                  :label="tt('Username')" v-model.trim="newMemberUsername"/>
                    <v-btn color="primary" :loading="submitting" @click="addMemberByUsername">{{ tt('Add') }}</v-btn>
                </div>
                <v-list lines="two" class="py-0">
                    <v-list-item v-for="member in members" :key="member.uid">
                        <template #title>
                            <div class="d-flex align-center gap-x-2">
                                <span>{{ member.nickname || member.username }}</span>
                                <v-chip size="small" variant="tonal" v-if="member.isCurrent">{{ tt('Me') }}</v-chip>
                            </div>
                        </template>
                        <template #subtitle>
                            <span>{{ member.username }}</span>
                        </template>
                        <template #append>
                            <div class="d-flex align-center">
                                <v-select v-if="currentLedger && currentLedger.role === LedgerMemberRole.Owner && member.role !== LedgerMemberRole.Owner"
                                          class="me-2" density="compact" style="max-width: 140px"
                                          item-title="name" item-value="role"
                                          :items="editableRoles"
                                          v-model="member.role"
                                          @update:model-value="(role: number) => changeRole(member, role)"/>
                                <v-chip v-else size="small" variant="tonal">{{ getMemberRoleName(member.role) }}</v-chip>
                                <v-btn v-if="currentLedger && currentLedger.role === LedgerMemberRole.Owner && member.role !== LedgerMemberRole.Owner"
                                       variant="text" color="error" size="small" :loading="submitting"
                                       @click="removeMember(member)">{{ tt('Remove') }}</v-btn>
                            </div>
                        </template>
                    </v-list-item>
                </v-list>
            </v-card-text>
            <v-card-actions>
                <v-spacer/>
                <v-btn variant="text" color="default" @click="showMembersDialog = false">{{ tt('Close') }}</v-btn>
            </v-card-actions>
        </v-card>
    </v-dialog>

    <v-dialog width="500" v-model="showRemoveDialog">
        <v-card :title="removeDialogTitle">
            <v-card-text>{{ removeDialogContent }}</v-card-text>
            <v-card-actions>
                <v-spacer/>
                <v-btn variant="text" color="default" @click="showRemoveDialog = false">{{ tt('Cancel') }}</v-btn>
                <v-btn color="error" :loading="submitting" @click="confirmRemove">{{ tt('Confirm') }}</v-btn>
            </v-card-actions>
        </v-card>
    </v-dialog>

    <snack-bar ref="snackbar" />
</template>

<script setup lang="ts">
import SnackBar from '@/components/desktop/SnackBar.vue';

import { ref, computed, onMounted, useTemplateRef } from 'vue';

import { useI18n } from '@/locales/helpers.ts';

import { LedgerMemberRole } from '@/models/ledger.ts';
import type { LedgerInfoResponse, LedgerMemberInfoResponse } from '@/models/ledger.ts';

import services from '@/lib/services.ts';
import { getCurrentLedgerId, setCurrentLedgerId } from '@/lib/ledger.ts';

import {
    mdiAccountOutline,
    mdiAccountMultipleOutline
} from '@mdi/js';

type SnackBarType = InstanceType<typeof SnackBar>;

const { tt } = useI18n();

const snackbar = useTemplateRef<SnackBarType>('snackbar');

const ledgers = ref<LedgerInfoResponse[]>([]);
const members = ref<LedgerMemberInfoResponse[]>([]);
const currentLedger = ref<LedgerInfoResponse | null>(null);

const loading = ref<boolean>(false);
const submitting = ref<boolean>(false);

const showCreateDialog = ref<boolean>(false);
const showJoinDialog = ref<boolean>(false);
const showRenameDialog = ref<boolean>(false);
const showMembersDialog = ref<boolean>(false);
const showRemoveDialog = ref<boolean>(false);

const newLedgerName = ref<string>('');
const joinInviteCode = ref<string>('');
const newMemberUsername = ref<string>('');
const renameLedgerName = ref<string>('');
const removeTarget = ref<LedgerInfoResponse | null>(null);

const editableRoles = computed(() => [
    { name: tt('Editor'), role: LedgerMemberRole.Editor },
    { name: tt('Viewer (read only)'), role: LedgerMemberRole.Viewer }
]);

const removeDialogTitle = computed<string>(() => {
    if (!removeTarget.value) {
        return '';
    }

    return removeTarget.value.role === LedgerMemberRole.Owner ? tt('Delete Ledger') : tt('Leave Ledger');
});

const removeDialogContent = computed<string>(() => {
    if (!removeTarget.value) {
        return '';
    }

    return removeTarget.value.role === LedgerMemberRole.Owner
        ? tt('All transaction data of this shared ledger will be deleted permanently. Are you sure?')
        : tt('You will no longer be able to access the transactions of this shared ledger. Are you sure?');
});

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
    return getCurrentLedgerId() === parseInt(ledger.id, 10);
}

function loadLedgers(): void {
    loading.value = true;

    services.getLedgers().then(response => {
        loading.value = false;
        ledgers.value = (response.data && response.data.result) ? response.data.result : [];
    }).catch(error => {
        loading.value = false;
        snackbar.value?.showError(error);
    });
}

function switchLedger(ledger: LedgerInfoResponse): void {
    const ledgerId = parseInt(ledger.id, 10);
    setCurrentLedgerId(ledgerId);
    window.location.reload();
}

function createLedger(): void {
    if (!newLedgerName.value) {
        snackbar.value?.showMessage(tt('The ledger name cannot be empty'));
        return;
    }

    submitting.value = true;

    services.createLedger({ name: newLedgerName.value }).then(response => {
        submitting.value = false;
        showCreateDialog.value = false;
        newLedgerName.value = '';

        const result = response.data ? response.data.result : null;

        if (result) {
            snackbar.value?.showMessage(tt('The shared ledger has been created'));
        }

        loadLedgers();
    }).catch(error => {
        submitting.value = false;
        snackbar.value?.showError(error);
    });
}

function joinLedger(): void {
    if (!joinInviteCode.value) {
        snackbar.value?.showMessage(tt('The invite code cannot be empty'));
        return;
    }

    submitting.value = true;

    services.joinLedger({ inviteCode: joinInviteCode.value.toUpperCase() }).then(() => {
        submitting.value = false;
        showJoinDialog.value = false;
        joinInviteCode.value = '';
        snackbar.value?.showMessage(tt('You have joined the shared ledger'));
        loadLedgers();
    }).catch(error => {
        submitting.value = false;
        snackbar.value?.showError(error);
    });
}

function showRenameDialogFor(ledger: LedgerInfoResponse): void {
    currentLedger.value = ledger;
    renameLedgerName.value = ledger.name;
    showRenameDialog.value = true;
}

function renameLedger(): void {
    if (!currentLedger.value || !renameLedgerName.value) {
        return;
    }

    submitting.value = true;

    services.renameLedger({ ledgerId: currentLedger.value.id, name: renameLedgerName.value }).then(() => {
        submitting.value = false;
        showRenameDialog.value = false;
        snackbar.value?.showMessage(tt('The ledger has been renamed'));
        loadLedgers();
    }).catch(error => {
        submitting.value = false;
        snackbar.value?.showError(error);
    });
}

function regenerateInviteCode(ledger: LedgerInfoResponse): void {
    submitting.value = true;

    services.regenerateLedgerInviteCode({ ledgerId: ledger.id }).then(response => {
        submitting.value = false;
        snackbar.value?.showMessage(tt('The invite code has been reset'));
        loadLedgers();
    }).catch(error => {
        submitting.value = false;
        snackbar.value?.showError(error);
    });
}

function showMembers(ledger: LedgerInfoResponse): void {
    currentLedger.value = ledger;
    showMembersDialog.value = true;
    members.value = [];

    refreshMembers(ledger);
}

function addMemberByUsername(): void {
    if (!currentLedger.value) {
        return;
    }

    const username = newMemberUsername.value;

    if (!username) {
        snackbar.value?.showMessage(tt('The username cannot be empty'));
        return;
    }

    submitting.value = true;

    services.addLedgerMember({
        ledgerId: currentLedger.value.id,
        username: username,
        role: LedgerMemberRole.Editor
    }).then(response => {
        submitting.value = false;
        newMemberUsername.value = '';

        const result = response.data ? response.data.result : null;

        if (result) {
            currentLedger.value = result;
        }

        snackbar.value?.showMessage(tt('The member has been added'));
        loadLedgers();

        if (currentLedger.value) {
            refreshMembers(currentLedger.value);
        }
    }).catch(error => {
        submitting.value = false;
        snackbar.value?.showError(error);
    });
}

function refreshMembers(ledger: LedgerInfoResponse): void {
    services.getLedgerMembers({ ledgerId: ledger.id }).then(response => {
        members.value = (response.data && response.data.result) ? response.data.result : [];
    }).catch(error => {
        snackbar.value?.showError(error);
    });
}

function changeRole(member: LedgerMemberInfoResponse, role: number): void {
    if (!currentLedger.value) {
        return;
    }

    submitting.value = true;

    services.updateLedgerMemberRole({
        ledgerId: currentLedger.value.id,
        uid: member.uid,
        role: role as LedgerMemberRole
    }).then(() => {
        submitting.value = false;
        snackbar.value?.showMessage(tt('The member role has been updated'));
    }).catch(error => {
        submitting.value = false;
        snackbar.value?.showError(error);
    });
}

function removeMember(member: LedgerMemberInfoResponse): void {
    if (!currentLedger.value) {
        return;
    }

    submitting.value = true;

    services.removeLedgerMember({ ledgerId: currentLedger.value.id, uid: member.uid }).then(() => {
        submitting.value = false;
        showMembers(currentLedger.value as LedgerInfoResponse);
    }).catch(error => {
        submitting.value = false;
        snackbar.value?.showError(error);
    });
}

function showRemoveDialogFor(ledger: LedgerInfoResponse): void {
    removeTarget.value = ledger;
    showRemoveDialog.value = true;
}

function confirmRemove(): void {
    if (!removeTarget.value) {
        return;
    }

    submitting.value = true;

    const target = removeTarget.value;

    const promise = target.role === LedgerMemberRole.Owner
        ? services.deleteLedger({ ledgerId: target.id })
        : services.leaveLedger({ ledgerId: target.id });

    promise.then(() => {
        submitting.value = false;
        showRemoveDialog.value = false;

        if (getCurrentLedgerId() === parseInt(target.id, 10)) {
            setCurrentLedgerId(0);
        }

        snackbar.value?.showMessage(tt('The operation succeeded'));
        loadLedgers();
    }).catch(error => {
        submitting.value = false;
        snackbar.value?.showError(error);
    });
}

onMounted(() => {
    loadLedgers();
});
</script>
