<template>
    <v-row>
        <v-col cols="12">
            <v-card class="setting-items" :title="tt('AI Accounting Model')">
                <v-card-text class="pa-0 text-body-medium">
                    <div class="px-4 pt-2 pb-4 text-body-medium text-medium-emphasis">
                        {{ tt('Configure your own large language model for AI text and receipt image recognition. The configuration is stored per user and encrypted on the server.') }}
                    </div>

                    <v-form>
                        <div class="setting-item">
                            <span>{{ tt('Model Provider') }}</span>
                            <v-spacer/>
                            <v-select
                                class="ms-4"
                                density="compact"
                                item-title="name"
                                item-value="value"
                                persistent-placeholder
                                max-width="400px"
                                :placeholder="tt('Model Provider')"
                                :items="allProviders"
                                :disabled="loading || saving"
                                v-model="provider"
                                @update:model-value="onProviderChanged"
                            />
                        </div>

                        <div class="setting-item">
                            <span>{{ tt('API Base URL') }}</span>
                            <v-spacer/>
                            <v-text-field
                                class="ms-4"
                                density="compact"
                                persistent-placeholder
                                max-width="400px"
                                :placeholder="defaultBaseUrl || tt('API Base URL')"
                                :disabled="loading || saving"
                                v-model.trim="baseUrl"
                            />
                        </div>

                        <div class="setting-item">
                            <span>{{ tt('API Key') }}</span>
                            <v-spacer/>
                            <v-text-field
                                class="ms-4"
                                density="compact"
                                persistent-placeholder
                                type="password"
                                autocomplete="off"
                                max-width="400px"
                                :placeholder="hasApiKey ? tt('Saved, leave blank to keep it unchanged') : tt('API Key')"
                                :disabled="loading || saving || !providerSupportApiKey"
                                v-model.trim="apiKey"
                            />
                        </div>

                        <div class="setting-item">
                            <span>{{ tt('Text Recognition Model') }}</span>
                            <v-spacer/>
                            <v-text-field
                                class="ms-4"
                                density="compact"
                                persistent-placeholder
                                max-width="400px"
                                :placeholder="tt('Example: gpt-4o-mini')"
                                :disabled="loading || saving"
                                v-model.trim="textModelId"
                            />
                        </div>

                        <div class="setting-item">
                            <span>{{ tt('Receipt Image Recognition Model') }}</span>
                            <v-spacer/>
                            <v-text-field
                                class="ms-4"
                                density="compact"
                                persistent-placeholder
                                max-width="400px"
                                :placeholder="tt('Example: gpt-4o, defaults to the text model')"
                                :disabled="loading || saving"
                                v-model.trim="imageModelId"
                            />
                        </div>

                        <div class="setting-item">
                            <span>{{ tt('Enable AI Text Accounting') }}</span>
                            <v-spacer/>
                            <v-switch
                                class="ms-4"
                                density="compact"
                                color="primary"
                                :disabled="loading || saving"
                                v-model="enableTextRecognition"
                            />
                        </div>

                        <div class="setting-item">
                            <span>{{ tt('Enable AI Receipt Image Accounting') }}</span>
                            <v-spacer/>
                            <v-switch
                                class="ms-4"
                                density="compact"
                                color="primary"
                                :disabled="loading || saving || !providerSupportImage"
                                v-model="enableImageRecognition"
                            />
                        </div>
                    </v-form>
                </v-card-text>

                <v-card-actions>
                    <v-spacer/>
                    <v-btn variant="text" color="default" :disabled="loading || saving || testing"
                           @click="clearConfig">{{ tt('Remove Configuration') }}</v-btn>
                    <v-btn variant="outlined" color="default" :loading="testing" :disabled="loading || saving || !hasApiKey && !providerSupportApiKey"
                           @click="testConfig">
                        {{ tt('Test Connection') }}
                        <template #loader>
                            <v-progress-circular indeterminate size="20"/>
                        </template>
                    </v-btn>
                    <v-btn color="primary" :loading="saving" :disabled="loading || testing"
                           @click="saveConfig">
                        {{ tt('Save') }}
                        <template #loader>
                            <v-progress-circular indeterminate size="20"/>
                        </template>
                    </v-btn>
                </v-card-actions>
            </v-card>
        </v-col>
    </v-row>

    <snack-bar ref="snackbar" />
</template>

<script setup lang="ts">
import SnackBar from '@/components/desktop/SnackBar.vue';

import { ref, computed, onMounted, useTemplateRef } from 'vue';

import { useI18n } from '@/locales/helpers.ts';

import { ALL_LLM_PROVIDERS } from '@/models/user_llm_config.ts';
import type { UserLlmConfigUpdateRequest } from '@/models/user_llm_config.ts';

import services from '@/lib/services.ts';
import { setCachedUserLlmConfig } from '@/lib/user_llm.ts';

type SnackBarType = InstanceType<typeof SnackBar>;

const { tt } = useI18n();

const snackbar = useTemplateRef<SnackBarType>('snackbar');

const allProviders = computed(() => ALL_LLM_PROVIDERS.map(item => ({ ...item, name: tt(item.name) })));

const provider = ref<string>('openai');
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

const providerSupportApiKey = computed<boolean>(() => provider.value !== 'ollama' && provider.value !== 'lmstudio');
const providerSupportImage = computed<boolean>(() => {
    const item = ALL_LLM_PROVIDERS.find(p => p.value === provider.value);
    return item ? item.supportImage : false;
});
const defaultBaseUrl = computed<string>(() => {
    const item = ALL_LLM_PROVIDERS.find(p => p.value === provider.value);
    return item ? item.defaultBaseUrl : '';
});

function onProviderChanged(): void {
    if (!baseUrl.value) {
        baseUrl.value = defaultBaseUrl.value;
    }
}

function loadConfig(): void {
    loading.value = true;

    services.getUserLlmConfig().then(response => {
        loading.value = false;

        const result = response.data ? response.data.result : null;

        if (!result || typeof result !== 'object') {
            return;
        }

        provider.value = result.provider || 'openai';
        baseUrl.value = result.baseUrl || '';
        textModelId.value = result.textModelId || '';
        imageModelId.value = result.imageModelId || '';
        enableTextRecognition.value = !!result.enableTextRecognition;
        enableImageRecognition.value = !!result.enableImageRecognition;
        hasApiKey.value = !!result.hasApiKey;
    }).catch(error => {
        loading.value = false;
        snackbar.value?.showError(error);
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
        snackbar.value?.showMessage(tt('The text recognition model cannot be empty'));
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

        snackbar.value?.showMessage(tt('Your AI model configuration has been saved'));
    }).catch(error => {
        saving.value = false;
        snackbar.value?.showError(error);
    });
}

function testConfig(): void {
    testing.value = true;

    services.testUserLlmConfig().then(response => {
        testing.value = false;

        if (response.data && response.data.success) {
            snackbar.value?.showMessage(tt('The model connection test succeeded'));
        } else {
            snackbar.value?.showMessage(tt('The model connection test failed'));
        }
    }).catch(error => {
        testing.value = false;
        snackbar.value?.showError(error);
    });
}

function clearConfig(): void {
    saving.value = true;

    services.deleteUserLlmConfig().then(() => {
        saving.value = false;
        apiKey.value = '';
        hasApiKey.value = false;
        setCachedUserLlmConfig(null);
        snackbar.value?.showMessage(tt('Your AI model configuration has been removed'));
    }).catch(error => {
        saving.value = false;
        snackbar.value?.showError(error);
    });
}

onMounted(() => {
    loadConfig();
});
</script>
