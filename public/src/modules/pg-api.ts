import {
    ajax,
    get,
    post,
    SuccessCallback,
    ErrorCallback,
    HTTPMethod
} from "@/reckit";

export interface LoginInput {
    [index: string]: string;
    username: string;
    password: string;
}

export interface SaveUserInput {
    [index: string]: string | number;
    username: string;
    password: string;
    device_limit: number;
    expiration_type: number;
    device_expiration: string;
    valid_start: string;
    valid_end: string;
    can_manage: number;
    can_autoreg: number;
}

export interface RegisterDeviceInput {
    username: string;
    "mac-address": string;
    description: string;
    platform?: string;
}

// API is a collection of methods used to interact with the API
// in Packet Guardian. They're centralized here for easy maintenance.
class API {
    // Authentication functions

    login(
        data: LoginInput,
        success?: APISuccessCallback<EmptyResp>,
        error?: ErrorCallback
    ) {
        post("/login", data, apiRespWrapper(success), error);
    }

    logout(success?: APISuccessCallback<EmptyResp>, error?: ErrorCallback) {
        get("/logout?noredirect", {}, apiRespWrapper(success), error);
    }

    // User functions
    saveUser(
        data: SaveUserInput,
        success?: APISuccessCallback<EmptyResp>,
        error?: ErrorCallback
    ) {
        post("/api/user", data, apiRespWrapper(success), error);
    }

    deleteUser(
        username: string,
        success?: APISuccessCallback<EmptyResp>,
        error?: ErrorCallback
    ) {
        ajax({
            method: HTTPMethod.Delete,
            url: "/api/user",
            params: { username },
            success: apiRespWrapper(success),
            error
        });
    }

    blacklistUser(
        username: string,
        success?: APISuccessCallback<EmptyResp>,
        error?: ErrorCallback
    ) {
        username = encodeURIComponent(username);
        post(
            `/api/blacklist/user/${username}`,
            {},
            apiRespWrapper(success),
            error
        );
    }

    unblacklistUser(
        username: string,
        success?: APISuccessCallback<EmptyResp>,
        error?: ErrorCallback
    ) {
        username = encodeURIComponent(username);
        ajax({
            method: HTTPMethod.Delete,
            url: `/api/blacklist/user/${username}`,
            success: apiRespWrapper(success),
            error
        });
    }

    // Device functions
    saveDeviceDescription(
        mac: string,
        description: string,
        success?: APISuccessCallback<EmptyResp>,
        error?: ErrorCallback
    ) {
        mac = encodeURIComponent(mac);
        post(
            `/api/device/mac/${mac}/description`,
            { description },
            apiRespWrapper(success),
            error
        );
    }

    saveDeviceExpiration(
        mac: string,
        type: string,
        value: string,
        success?: APISuccessCallback<DeviceExpirationResp>,
        error?: ErrorCallback
    ) {
        mac = encodeURIComponent(mac);
        post(
            `/api/device/mac/${mac}/expiration`,
            {
                type,
                value
            },
            apiRespWrapper(success),
            error
        );
    }

    // flagged is bool
    flagDevice(
        mac: string,
        flagged: boolean,
        success?: APISuccessCallback<EmptyResp>,
        error?: ErrorCallback
    ) {
        mac = encodeURIComponent(mac);
        post(
            `/api/device/mac/${mac}/flag`,
            { flagged },
            apiRespWrapper(success),
            error
        );
    }

    // macs is an array of MAC addresses
    deleteDevices(
        username: string,
        macs: string[],
        success?: APISuccessCallback<EmptyResp>,
        error?: ErrorCallback
    ) {
        username = encodeURIComponent(username);
        ajax({
            method: HTTPMethod.Delete,
            url: `/api/device/user/${username}`,
            params: { mac: macs.join(",") },
            success: apiRespWrapper(success),
            error
        });
    }

    // macs is an array of MAC addresses
    blacklistDevices(
        macs: string[],
        success?: APISuccessCallback<EmptyResp>,
        error?: ErrorCallback
    ) {
        post(
            "/api/blacklist/device",
            { mac: macs.join(",") },
            apiRespWrapper(success),
            error
        );
    }

    blacklistAllDevices(
        username: string,
        success?: APISuccessCallback<EmptyResp>,
        error?: ErrorCallback
    ) {
        username = encodeURIComponent(username);
        post(
            "/api/blacklist/device",
            { username },
            apiRespWrapper(success),
            error
        );
    }

    // macs is an array of MAC addresses
    unblacklistDevices(
        macs: string[],
        success?: APISuccessCallback<EmptyResp>,
        error?: ErrorCallback
    ) {
        ajax({
            method: HTTPMethod.Delete,
            url: "/api/blacklist/device",
            params: { mac: macs.join(",") },
            success: apiRespWrapper(success),
            error
        });
    }

    unblacklistAllDevices(
        username: string,
        success?: APISuccessCallback<EmptyResp>,
        error?: ErrorCallback
    ) {
        username = encodeURIComponent(username);
        ajax({
            method: HTTPMethod.Delete,
            url: "/api/blacklist/device",
            params: { username },
            success: apiRespWrapper(success),
            error
        });
    }

    // macs is an array of MAC addresses
    reassignDevices(
        username: string,
        macs: string[],
        success?: APISuccessCallback<EmptyResp>,
        error?: ErrorCallback
    ) {
        post(
            "/api/device/reassign",
            { username, macs: macs.join(",") },
            apiRespWrapper(success),
            error
        );
    }

    registerDevice(
        data: RegisterDeviceInput,
        success?: APISuccessCallback<DeviceRegisterResp>,
        error?: ErrorCallback
    ) {
        post(
            "/api/device",
            {
                username: data.username,
                "mac-address": data["mac-address"],
                description: data.description,
                platform: data.platform ?? ""
            },
            apiRespWrapper(success),
            error
        );
    }
}

export default new API();

interface EmptyResp {
    Message: string;
    Data: {};
}

interface DeviceExpirationResp extends EmptyResp {
    Data: {
        newExpiration: string;
    };
}

interface DeviceRegisterResp extends EmptyResp {
    Data: {
        Location: string;
    };
}

export interface APISuccessCallback<T> {
    (resp: T, req: XMLHttpRequest): void;
}

function apiRespWrapper<T>(
    fn?: APISuccessCallback<T>
): SuccessCallback | undefined {
    if (fn) {
        return (resp: any, req: XMLHttpRequest) =>
            resp ? fn(JSON.parse(resp), req) : fn({} as any, req);
    }
}
