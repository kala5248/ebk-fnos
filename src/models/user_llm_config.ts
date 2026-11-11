export interface UserLlmConfigResponse {
    readonly provider: string;
    readonly baseUrl: string;
    readonly apiKeyMasked: string;
    readonly textModelId: string;
    readonly imageModelId: string;
    readonly enableTextRecognition: boolean;
    readonly enableImageRecognition: boolean;
    readonly hasApiKey: boolean;
}

export interface UserLlmConfigUpdateRequest {
    readonly provider: string;
    readonly baseUrl: string;
    readonly apiKey: string;
    readonly textModelId: string;
    readonly imageModelId: string;
    readonly enableTextRecognition: boolean;
    readonly enableImageRecognition: boolean;
}

export interface UserLlmConfigTestResponse {
    readonly success: boolean;
    readonly content: string;
}

export interface LlmProviderOption {
    readonly value: string;
    readonly name: string;
    readonly defaultBaseUrl: string;
    readonly supportImage: boolean;
}

export const ALL_LLM_PROVIDERS: LlmProviderOption[] = [
    { value: 'openai', name: 'OpenAI', defaultBaseUrl: 'https://api.openai.com/v1', supportImage: true },
    { value: 'anthropic', name: 'Anthropic', defaultBaseUrl: 'https://api.anthropic.com', supportImage: true },
    { value: 'googleai', name: 'Google AI', defaultBaseUrl: 'https://generativelanguage.googleapis.com', supportImage: true },
    { value: 'ollama', name: 'Ollama', defaultBaseUrl: 'http://127.0.0.1:11434', supportImage: false },
    { value: 'lmstudio', name: 'LM Studio', defaultBaseUrl: 'http://127.0.0.1:1234', supportImage: false }
];
