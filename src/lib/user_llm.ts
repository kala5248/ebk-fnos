import type { UserLlmConfigResponse } from '@/models/user_llm_config.ts';

const USER_LLM_CONFIG_CACHE_KEY = 'ezbookkeeping.userLlmConfig';

let cachedUserLlmConfig: UserLlmConfigResponse | null | undefined = undefined;

export function getCachedUserLlmConfig(): UserLlmConfigResponse | null {
    if (cachedUserLlmConfig !== undefined) {
        return cachedUserLlmConfig;
    }

    try {
        const value = localStorage.getItem(USER_LLM_CONFIG_CACHE_KEY);

        if (!value) {
            cachedUserLlmConfig = null;
            return null;
        }

        cachedUserLlmConfig = JSON.parse(value) as UserLlmConfigResponse;
        return cachedUserLlmConfig;
    } catch {
        cachedUserLlmConfig = null;
        return null;
    }
}

export function setCachedUserLlmConfig(config: UserLlmConfigResponse | null): void {
    cachedUserLlmConfig = config;

    try {
        if (!config) {
            localStorage.removeItem(USER_LLM_CONFIG_CACHE_KEY);
        } else {
            localStorage.setItem(USER_LLM_CONFIG_CACHE_KEY, JSON.stringify(config));
        }
    } catch {
        // ignore local cache failures
    }
}

export function clearCachedUserLlmConfig(): void {
    setCachedUserLlmConfig(null);
}

export function isUserLlmTextRecognitionEnabled(): boolean {
    const config = getCachedUserLlmConfig();
    return !!config && config.enableTextRecognition && (config.hasApiKey || config.provider === 'ollama' || config.provider === 'lmstudio');
}

export function isUserLlmImageRecognitionEnabled(): boolean {
    const config = getCachedUserLlmConfig();
    return !!config && config.enableImageRecognition && (config.hasApiKey || config.provider === 'ollama' || config.provider === 'lmstudio');
}
