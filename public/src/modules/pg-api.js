// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
import { ajax, get, post } from "reckit";

// API is a collection of methods used to interact with the API
// in Packet Guardian. They're centralized here for easy maintenance.
class API {
  // Authentication functions

  // data = {"username": "", "password": ""}
  login(data, success, error) {
    post("/login", data, success, error);
  }

  logout(success, error) {
    get("/logout?noredirect", {}, success, error);
  }

  // User functions

  // data = {
  //     "username": "",
  //     "password": "",
  //     "device_limit": "",
  //     "expiration_type": "",
  //     "device_expiration": "",
  //     "valid_start": "",
  //     "valid_end": "",
  //     "can_manage": "",
  //     "can_autoreg": ""
  // }
  saveUser(data, success, error) {
    post("/api/user", data, success, error);
  }

  deleteUser(username, success, error) {
    username = encodeURIComponent(username);
    ajax({
      method: "DELETE",
      url: "/api/user",
      params: { username: username },
      success,
      error
    });
  }

  blacklistUser(username, success, error) {
    username = encodeURIComponent(username);
    post(`/api/blacklist/user/${username}`, {}, success, error);
  }

  unblacklistUser(username, success, error) {
    username = encodeURIComponent(username);
    ajax({
      method: "DELETE",
      url: `/api/blacklist/user/${username}`,
      success,
      error
    });
  }

  // Device functions
  saveDeviceDescription(mac, desc, success, error) {
    mac = encodeURIComponent(mac);
    post(
      `/api/device/mac/${mac}/description`,
      { description: desc },
      success,
      error
    );
  }

  saveDeviceExpiration(mac, type, val, success, error) {
    mac = encodeURIComponent(mac);
    const data = {
      type: type,
      value: val
    };
    post(`/api/device/mac/${mac}/expiration`, data, success, error);
  }

  // flagged is bool
  flagDevice(mac, flagged, success, error) {
    mac = encodeURIComponent(mac);
    post(`/api/device/mac/${mac}/flag`, { flagged }, success, error);
  }

  // macs is an array of MAC addresses
  deleteDevices(username, macs, success, error) {
    username = encodeURIComponent(username);
    ajax({
      method: "DELETE",
      url: `/api/device/user/${username}`,
      params: { mac: macs.join(",") },
      success,
      error
    });
  }

  // macs is an array of MAC addresses
  blacklistDevices(macs, success, error) {
    post("/api/blacklist/device", { mac: macs.join(",") }, success, error);
  }

  blacklistAllDevices(username, success, error) {
    username = encodeURIComponent(username);
    post("/api/blacklist/device", { username: username }, success, error);
  }

  // macs is an array of MAC addresses
  unblacklistDevices(macs, success, error) {
    ajax({
      method: "DELETE",
      url: "/api/blacklist/device",
      params: { mac: macs.join(",") },
      success,
      error
    });
  }

  unblacklistAllDevices(username, success, error) {
    username = encodeURIComponent(username);
    ajax({
      method: "DELETE",
      url: "/api/blacklist/device",
      params: { username: username },
      success,
      error
    });
  }

  // macs is an array of MAC addresses
  reassignDevices(username, macs, success, error) {
    post(
      "/api/device/reassign",
      { username: username, macs: macs.join(",") },
      success,
      error
    );
  }

  // data = {
  //     "username": "",
  //     "mac-address": "",
  //     "description": "",
  //     "password",  <- Optional
  //     "platform"   <- Optional
  // }
  registerDevice(data, success, error) {
    if (
      !("username" in data) ||
      !("mac-address" in data) ||
      !("description" in data)
    ) {
      console.error("Invalid data object");
      return;
    }
    post("/api/device", data, success, error);
  }
}

export default new API();
