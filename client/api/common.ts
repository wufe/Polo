import { AxiosError, AxiosResponse } from "axios";

export enum APIRequestResult {
    FAILED    = "failed",
    SUCCEEDED = "succeeded",
}

export enum APIRequestFailReason {
    UNKNOWN      = "unknown",
    NOT_FOUND    = "not-found",
    SERVER_ERROR = "server-error",
}

export type APIPayload<T = void> = {
    result : APIRequestResult.FAILED;
    reason?: APIRequestFailReason;
} | {
    result: APIRequestResult.SUCCEEDED;
    payload: T;
}

export type APIResponseObject<T = unknown> = 
    APIResponseObejctSucceeded<T> | APIReponseObjectFailed;

export type APIResponseObejctSucceeded<T = unknown> = {
    message: string;
    result: T;
};
export type APIReponseObjectFailed = {
    message: string;
    reason: string;
};

export function isAxiosError(e: Error): e is AxiosError {
    return 'isAxiosError' in e;
}

export async function buildRequest<T>(
    request: () => Promise<AxiosResponse<APIResponseObject<T>>>
): Promise<APIPayload<T>> {
    try {
        const response = await request();
        return {
            result: APIRequestResult.SUCCEEDED,
            payload: (response.data as APIResponseObejctSucceeded<T>).result
        };
    } catch (e) {
        let reason: APIRequestFailReason = APIRequestFailReason.UNKNOWN;
        if (isAxiosError(e)) {
            if (e.response?.status == 404) {
                reason = APIRequestFailReason.NOT_FOUND;
            }
        }
        return {
            result: APIRequestResult.FAILED,
            reason
        } as APIPayload<T>;
    }
}