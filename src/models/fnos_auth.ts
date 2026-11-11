export interface FnOSAuthUrlResponse {
    readonly url: string;
    readonly state: string;
}

export interface FnOSLoginRequest {
    readonly state: string;
    readonly code: string;
}
