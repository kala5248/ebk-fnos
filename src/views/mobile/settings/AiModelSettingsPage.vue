<template>
    <f7-page>
        <f7-navbar>
            <f7-nav-left :class="{ 'disabled': saving || loading }" :back-link="tt('Back')"></f7-nav-left>
            <f7-nav-title :title="tt('AI Accounting Model')"></f7-nav-title>
            <f7-nav-right>
                <f7-link :class="{ 'disabled': saving || loading }" icon-f7="checkmark_alt" :aria-label="tt('Save')" @click="saveConfig"></f7-link>
            </f7-nav-right>
        </f7-navbar>

        <f7-block class="no-margin">
            <p class="text-color-gray">{{ tt('Configure your own large language model for AI text and receipt image recognition. The configuration is stored per user and encrypted on the server.') }}</p>
        </f7-block>

        <f7-list strong inset dividers class="margin-vertical-half">
            <f7-list-item group-title :sortable="false">
                <small>{{ tt('Model Provider') }}</small>
            </f7-list-item>

            <f7-list-item link="#" no-chevron
                          :title="tt('Model Provider')"
                          :after="selectedProvider ? selectedProvider.name : ''"
                          @click="showProviderPopup = true">
                <list-item-selection-popup value-type="item"
                                           key-field="value" value-field="value"
                                           title-field="name"
                                           :title="tt('Model Provider')"
                                           :items="allProviders"
                                           v-model:show="showProviderPopup"
                                           v-model="selectedProvider">
                </list-item-selection-popup>
            </f7-list-item>

            <f7-list-item>
                <template #title>
                    <div class="item-actual-title">{{ tt('API Base URL') }}</div>
                </template>
                <f7-input type="url" :placeholder="selectedProvider ? selectedProvider.defaultBaseUrl : ''"
                          autocapitalize="off" autocorrect="off" spellcheck="false"
                          :value="baseUrl" @input="baseUrl = $event.target.value"></f7-input>
            </f7-list-item>

            <f7-list-item>
                <template #title>
                    <div class="item-actual-title">{{ tt('API Key') }}</div>
                </template>
                <f7-input type="password" autocomplete="off"
                          :placeholder="hasApiKey ? tt('Saved, leave blank to keep it unchanged') : tt('API Key')"
                          :value="apiKey" @input="apiKey = $event.target.value"></f7-input>
            </f7-list-item>

            <f7-list-item>
                <template #title>
                    <div class="item-actual-title">{{ tt('Text Recognition Model') }}</div>
                </template>
                <f7-input type="text" autocapitalize="off" autocorrect="off" spellcheck="false"
                          :placeholder="tt('Example: gpt-4o-mini')"
                          :value="textModelId" @input="textModelId = $event.target.value"></f7-input>
            </f7-list-item>

            <f7-list-item>
                <template #title>
                    <div class="item-actual-title">{{ tt('Receipt Image Recognition Model') }}</div>
                </template>
                <f7-input type="text" autocapitalize="off" autocorrect="off" spellcheck="false"
                          :placeholder="tt('Example: gpt-4o, defaults to the text model')"
                          :value="imageModelId" @input="imageModelId = $event.target.value"></f7-input>
            </f7-list-item>
        </f7-list>

        <f7-list strong inset dividers class="margin-vertical-half">
            <f7-list-item group-title :sortable="false">
                <small>{{ tt('AI & Sharing') }}</small>
            </f7-list-item>

            <f7-list-item :title="tt('Enable AI Text Accounting')">
                <f7-toggle :checked="enableTextRecognition" :disabled="saving || loading"
                           @toggle:change="enableTextRecognition = $event"></f7-toggle>
            </f7-list-item>

            <f7-list-item :title="tt('Enable AI Receipt Image Accounting')">
                <f7-toggle :checked="enableImageRecognition" :disabled="saving || loading || !providerSupportImage"
                           @toggle:change="enableImageRecognition = $event"></f7-toggle>
            </f7-list-item>
        </f7-list>

        <f7-list strong inset dividers class="margin-vertical-half">
            <f7-list-button :class="{ 'disabled': saving || loading || testing }"
                            :title="tt('Test Connection')" @click="testConfig"></f7-list-button>
            <f7-list-button :class="{ 'disabled': saving || loading || testing || (!hasApiKey && providerSupportApiKey) }"
                            color="red" :title="tt('Remove Configuration')" @click="clearConfig"></f7-list-button>
        </f7-list>

        <f7-block-footer class="margin-vertical-half">
            <p class="text-color-gray">{{ tt('The configuration is stored per user and encrypted on the server.') }}</p>
        </f7-block-footer>
    </f7-page>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';

import { useI18n } from '@/locales/helpers.ts';
import { useI18nUIComponents } from '@/lib/ui/mobile.ts';

import { ALL_LLM_PROVIDERS } from '@/models/user_llm_config.ts';
import type { LlmProviderOption, UserLlmConfigUpdateRequest } from '@/models/user_llm_config.ts';

import services from '@/lib/services.ts';
import { setCachedUserLlmConfig } from '@/lib/user_llm.ts';

const { tt } = useI18n();
const { showConfirm, showToast } = useI18nUIComponents();

const allProviders = ALL_LLM_PROVIDERS;

const selectedProvider = ref<LlmProviderOption | null>(ALL_LLM_PROVIDERS[0]);
const baseUrl = ref<string>('');
const apiKey = ref<string>('');
const textModelId = ref<string>('');
const imageModelId = ref<string>('');
const enableTextRecognition = ref<boolean>(true);
const enableImageRecognition = ref<boolean>(false);
const hasApiKey = ref<boolean>(false);

const loading = ref<boolean>(true);
const saving = ref<boolean>(false);
const testing = ref<boolean>(false);

const showProviderPopup = ref<boolean>(false);

const provider = computed<string>(() => selectedProvider.value ? selectedProvider.value.value : 'openai');
const providerSupportApiKey = computed<boolean>(() => provider.value !== 'ollama' && provider.value !== 'lmstudio');
const providerSupportImage = computed<boolean>(() => selectedProvider.value ? selectedProvider.value.supportImage : false);

function loadConfig(): void {
    loading.value = true;

    services.getUserLlmConfig().then(response => {
        loading.value = false;

        const result = response.data ? response.data.result : null;

        if (!result || typeof result !== 'object') {
            return;
        }

        selectedProvider.value = ALL_LLM_PROVIDERS.find(item => item.value === result.provider) || ALL_LLM_PROVIDERS[0];
        baseUrl.value = result.baseUrl || '';
        textModelId.value = result.textModelId || '';
        imageModelId.value = result.imageModelId || '';
        enableTextRecognition.value = !!result.enableTextRecognition;
        enableImageRecognition.value = !!result.enableImageRecognition;
        hasApiKey.value = !!result.hasApiKey;
    }).catch(error => {
        loading.value = false;
        showToast(error.message || error);
    });
}

function buildRequest(): UserLlmConfigUpdateRequest {
    return {
        provider: provider.value,
        baseUrl: baseUrl.value,
        apiKey: apiKey.value,
        textModelId: textModelId.value,
        imageModelId: imageModelId.value,
        enableTextRecognition: enableTextRecognition.value,
        enableImageRecognition: enableImageRecognition.value
    };
}

function saveConfig(): void {
    if (enableTextRecognition.value && !textModelId.value) {
        showToast(tt('The text recognition model cannot be empty'));
        return;
    }

    saving.value = true;

    services.saveUserLlmConfig(buildRequest()).then(response => {
        saving.value = false;

        const result = response.data ? response.data.result : null;

        if (result && typeof result === 'object') {
            hasApiKey.value = !!result.hasApiKey;
            apiKey.value = '';
            setCachedUserLlmConfig(result);
        }

        showToast(tt('Your AI model configuration has been saved'));
    }).catch(error => {
        saving.value = false;
        showToast(error.message || error);
    });
}

function testConfig(): void {
    testing.value = true;

    services.testUserLlmConfig().then(response => {
        testing.value = false;

        if (response.data && response.data.success) {
            showToast(tt('The model connection test succeeded'));
        } else {
            showToast(tt('The model connection test failed'));
        }
    }).catch(error => {
        testing.value = false;
        showToast(error.message || error);
    });
}

function clearConfig(): void {
    showConfirm(tt('Your AI model configuration will be removed. Are you sure?'), () => {
        saving.value = true;

        services.deleteUserLlmConfig().then(() => {
            saving.value = false;
            apiKey.value = '';
            hasApiKey.value = false;
            setCachedUserLlmConfig(null);
            showToast(tt('Your AI model configuration has been removed'));
        }).catch(error => {
            saving.value = false;
            showToast(error.message || error);
        });
    });
}

loadConfig();
</script>
