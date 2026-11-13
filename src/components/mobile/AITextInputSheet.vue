<template>
    <f7-sheet class="ai-text-input-sheet" swipe-to-close swipe-handler=".swipe-handler" style="height:auto"
              :opened="show" @sheet:closed="close">
        <f7-toolbar class="toolbar-with-swipe-handler">
            <div class="swipe-handler"></div>
            <div class="left">
                <f7-link icon-f7="xmark" :aria-label="tt('Cancel')" @click="close"></f7-link>
            </div>
            <div class="right"></div>
        </f7-toolbar>
        <f7-page-content class="margin-top">
            <div class="padding-horizontal padding-bottom">
                <p class="text-color-gray">{{ tt('Enter one sentence and AI will recognize it.') }}</p>
                <textarea class="ai-text-input-textarea" v-model="text"
                          :placeholder="tt('For example: Lunch yesterday cost 35 yuan')"></textarea>
                <p class="text-align-center"><small>{{ tt('Uploaded text and personal data will be sent to the large language model, please be aware of potential privacy risks.') }}</small></p>
                <f7-button fill large class="margin-top" :class="{ 'disabled': !text.trim() }" @click="confirm(false)">
                    {{ tt('Recognize and Review') }}
                </f7-button>
                <f7-button tonal large class="margin-top-half" :class="{ 'disabled': !text.trim() }" @click="confirm(true)">
                    {{ tt('Recognize and Save') }}
                </f7-button>
            </div>
        </f7-page-content>
    </f7-sheet>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';

import { useI18n } from '@/locales/helpers.ts';

const props = defineProps<{
    show: boolean;
}>();

const emit = defineEmits<{
    (e: 'update:show', value: boolean): void;
    (e: 'text:confirm', text: string, autoSave: boolean): void;
}>();

const { tt } = useI18n();
const text = ref<string>('');

watch(() => props.show, (value) => {
    if (value) {
        text.value = '';
    }
});

function close(): void {
    emit('update:show', false);
}

function confirm(autoSave: boolean): void {
    const value = text.value.trim();

    if (!value) {
        return;
    }

    emit('text:confirm', value, autoSave);
    close();
}
</script>

<style scoped>
.ai-text-input-textarea {
    width: 100%;
    min-height: 110px;
    padding: 12px;
    border: 1px solid rgba(128, 128, 128, 0.3);
    border-radius: 8px;
    background: var(--f7-page-bg-color);
    color: var(--f7-text-color);
    font-size: 15px;
    resize: none;
}
</style>
