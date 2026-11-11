<template>
    <v-select
        v-if="ledgers.length > 1"
        class="ledger-switcher"
        density="compact"
        variant="outlined"
        hide-details
        item-title="displayName"
        item-value="id"
        :items="selectableLedgers"
        :model-value="currentLedgerId"
        :menu-props="{ maxHeight: 400 }"
        @update:model-value="switchLedger"
    >
        <template #prepend-inner>
            <v-icon :icon="mdiAccountMultipleOutline" size="20"/>
        </template>
    </v-select>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';

import { useI18n } from '@/locales/helpers.ts';

import services from '@/lib/services.ts';
import { getCurrentLedgerId, setCurrentLedgerId } from '@/lib/ledger.ts';

import { mdiAccountMultipleOutline } from '@mdi/js';

interface LedgerSwitchItem {
    readonly id: number;
    readonly displayName: string;
}

const { tt } = useI18n();

const ledgers = ref<LedgerSwitchItem[]>([]);
const currentLedgerId = ref<number>(getCurrentLedgerId());

const selectableLedgers = computed<LedgerSwitchItem[]>(() => ledgers.value);

function loadLedgers(): void {
    services.getLedgers().then(response => {
        const result = (response.data && response.data.result) ? response.data.result : [];

        ledgers.value = result.map(item => ({
            id: parseInt(item.id, 10),
            displayName: item.isPersonal ? `${item.name} (${tt('Personal')})` : item.name
        }));

        const current = getCurrentLedgerId();
        const exists = ledgers.value.some(item => item.id === current);

        if (!exists) {
            currentLedgerId.value = 0;
            setCurrentLedgerId(0);
        }
    }).catch(() => {
        ledgers.value = [];
    });
}

function switchLedger(ledgerId: number): void {
    if (ledgerId === currentLedgerId.value) {
        return;
    }

    setCurrentLedgerId(ledgerId);
    window.location.reload();
}

onMounted(() => {
    loadLedgers();
});
</script>

<style scoped>
.ledger-switcher {
    max-width: 220px;
    min-width: 140px;
}
</style>
