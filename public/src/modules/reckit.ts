export const enum HTTPMethod {
    Get = "GET",
    Post = "POST",
    Put = "PUT",
    Delete = "DELETE",
    Options = "OPTIONS",
    Head = "HEAD",
}

export interface SuccessCallback {
    (resp: any, req: XMLHttpRequest): void;
}

export interface ErrorCallback {
    (req: XMLHttpRequest): void;
}

export interface HeadersMap {
    [index: string]: string;
}

export interface ParamMap {
    [index: string]: string | number | boolean;
}

export interface AjaxOptions {
    method: HTTPMethod;
    url: string;
    params?: ParamMap;
    data?: string | ParamMap;
    contentType?: string;
    headers?: HeadersMap;
    success?: SuccessCallback;
    error?: ErrorCallback;
}

export const ajax = (options: AjaxOptions) => {
    const xhr = new XMLHttpRequest();

    const params = options.params ? formatParams(options.params) : "";
    const xhrURL = params === "" ? options.url : options.url + "?" + params;

    xhr.open(options.method, xhrURL, true);
    xhr.setRequestHeader(
        "Content-Type",
        options.contentType ??
            "application/x-www-form-urlencoded; charset=UTF-8"
    );

    xhr.setRequestHeader("X-Requested-With", "XMLHttpRequest");

    if (options.headers) {
        for (let key in options.headers) {
            if (options.headers.hasOwnProperty(key)) {
                xhr.setRequestHeader(key, options.headers[key]);
            }
        }
    }

    xhr.onreadystatechange = function () {
        if (this.readyState === 4) {
            if (this.status >= 200 && this.status < 400) {
                options.success?.(this.responseText, this);
            } else {
                options.error?.(this);
            }
        }
    };

    xhr.send(options.data ? formatParams(options.data) : "");
};

export const formatParams = (data: string | ParamMap): string => {
    if (typeof data === "string") {
        return data;
    }

    let dataParts = [];

    for (const key in data) {
        dataParts.push(
            encodeURIComponent(key) + "=" + encodeURIComponent(data[key])
        );
    }

    return dataParts.join("&");
};

export const get = (
    url: string,
    params: ParamMap,
    success?: SuccessCallback,
    error?: ErrorCallback
) => {
    ajax({
        method: HTTPMethod.Get,
        url: url,
        params,
        success,
        error,
    });
};

export const post = (
    url: string,
    data: ParamMap,
    success?: SuccessCallback,
    error?: ErrorCallback
) => {
    ajax({
        method: HTTPMethod.Post,
        url: url,
        data,
        success,
        error,
    });
};
