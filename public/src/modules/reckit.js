// options = {
//     method: "GET",
//     url: '',
//     params: '',
//     data: '',
//     contentType: "application/x-www-form-urlencoded; charset=UTF-8",
//     success: (() => {}),
//     error: (() => {}),
//     headers: {},
// }
export const ajax = options => {
  if (options === undefined) {
    return null;
  }
  options = checkAjaxSettings(options);

  const xhr = new XMLHttpRequest();
  const xhrURL =
    options.params === "" ? options.url : options.url + "?" + options.params;
  xhr.open(options.method, xhrURL, true);
  xhr.setRequestHeader("Content-Type", options.contentType);

  for (let key in options.headers) {
    if (options.headers.hasOwnProperty(key)) {
      xhr.setRequestHeader(key, options.headers[key]);
    }
  }

  xhr.setRequestHeader("X-Requested-With", "XMLHttpRequest");

  xhr.onreadystatechange = function() {
    if (this.readyState === 4) {
      if (this.status >= 200 && this.status < 400) {
        options.success(this.responseText, this);
      } else {
        options.error(this);
      }
    }
  };

  xhr.send(options.data);
};

export const params = data => {
  if (typeof data === "string") {
    return data;
  }

  let dataStr = "";

  if (typeof data === "object") {
    let dataParts = [];

    for (const key in data) {
      if (data.hasOwnProperty(key)) {
        dataParts.push(
          encodeURIComponent(key) + "=" + encodeURIComponent(data[key])
        );
      }
    }

    dataStr = dataParts.join("&");
  }

  return dataStr;
};

export const get = (url, data, success, error) => {
  ajax({
    method: "GET",
    url: url,
    params: params(data),
    success,
    error
  });
};

export const post = (url, data, success, error) => {
  ajax({
    method: "POST",
    url: url,
    data: params(data),
    success,
    error
  });
};

function checkAjaxSettings(options) {
  if (!options.method) {
    options.method = "GET";
  }
  options.method.toUpperCase();
  if (!options.url) {
    options.url = "";
  }
  if (!options.data) {
    options.data = "";
  }
  if (typeof options.data !== "string") {
    options.data = params(options.data);
  }
  if (!options.params) {
    options.params = "";
  }
  if (typeof options.params !== "string") {
    options.params = params(options.params);
  }
  if (!options.contentType) {
    options.contentType = "application/x-www-form-urlencoded; charset=UTF-8";
  }
  if (!options.success) {
    options.success = () => {};
  }
  if (!options.error) {
    options.error = () => {};
  }
  if (!options.headers) {
    options.headers = {};
  }
  return options;
}
