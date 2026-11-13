import { ref } from 'vue';
import { defineStore } from 'pinia';

export const useDesktopPageStore = defineStore('desktopPages', () => {
    const showAddTransactionDialogInTransactionList = ref<boolean>(false);
    const showAiTransactionDialogInTransactionList = ref<boolean>(false);

    function setShowAddTransactionDialogInTransactionList(): void {
        showAddTransactionDialogInTransactionList.value = true;
    }

    function resetShowAddTransactionDialogInTransactionList(): void {
        showAddTransactionDialogInTransactionList.value = false;
    }

    function setShowAiTransactionDialogInTransactionList(): void {
        showAiTransactionDialogInTransactionList.value = true;
    }

    function resetShowAiTransactionDialogInTransactionList(): void {
        showAiTransactionDialogInTransactionList.value = false;
    }

    return {
        // states
        showAddTransactionDialogInTransactionList,
        showAiTransactionDialogInTransactionList,
        // functions
        setShowAddTransactionDialogInTransactionList,
        resetShowAddTransactionDialogInTransactionList,
        setShowAiTransactionDialogInTransactionList,
        resetShowAiTransactionDialogInTransactionList
    }
});
