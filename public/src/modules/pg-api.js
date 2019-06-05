// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
import $ from 'jLib';

// API is a collection of methods used to interact with the API
// in Packet Guardian. They're centralized here for easy maintenance.
class API {
  // Authentication functions

  // data = {"username": "", "password": ""}
  login(data, ok, error) {
    $.post('/login', data, ok, error);
  }

  logout(ok, error) {
    $.get('/logout?noredirect', {}, ok, error);
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
  saveUser(data, ok, error) {
    $.post('/api/user', data, ok, error);
  }

  deleteUser(username, ok, error) {
    username = encodeURIComponent(username);
    $.ajax({
      method: 'DELETE',
      url: '/api/user',
      params: { username: username },
      success: ok,
      error: error
    });
  }

  blacklistUser(username, ok, error) {
    username = encodeURIComponent(username);
    $.post(`/api/blacklist/user/${username}`, {}, ok, error);
  }

  unblacklistUser(username, ok, error) {
    username = encodeURIComponent(username);
    $.ajax({
      method: 'DELETE',
      url: `/api/blacklist/user/${username}`,
      success: ok,
      error: error
    });
  }

  // Device functions
  saveDeviceDescription(mac, desc, ok, error) {
    mac = encodeURIComponent(mac);
    $.post(`/api/device/mac/${mac}/description`, { description: desc }, ok, error);
  }

  saveDeviceExpiration(mac, type, val, ok, error) {
    mac = encodeURIComponent(mac);
    const data = {
      type: type,
      value: val
    };
    $.post(`/api/device/mac/${mac}/expiration`, data, ok, error);
  }

  // flagged is bool
  flagDevice(mac, flagged, ok, error) {
    mac = encodeURIComponent(mac);
    $.post(`/api/device/mac/${mac}/flag`, { flagged }, ok, error);
  }

  // macs is an array of MAC addresses
  deleteDevices(username, macs, ok, error) {
    username = encodeURIComponent(username);
    $.ajax({
      method: 'DELETE',
      url: `/api/device/user/${username}`,
      params: { mac: macs.join(',') },
      success: ok,
      error: error
    });
  }

  // macs is an array of MAC addresses
  blacklistDevices(macs, ok, error) {
    $.post('/api/blacklist/device', { mac: macs.join(',') }, ok, error);
  }

  blacklistAllDevices(username, ok, error) {
    username = encodeURIComponent(username);
    $.post('/api/blacklist/device', { username: username }, ok, error);
  }

  // macs is an array of MAC addresses
  unblacklistDevices(macs, ok, error) {
    $.ajax({
      method: 'DELETE',
      url: '/api/blacklist/device',
      params: { mac: macs.join(',') },
      success: ok,
      error: error
    });
  }

  unblacklistAllDevices(username, ok, error) {
    username = encodeURIComponent(username);
    $.ajax({
      method: 'DELETE',
      url: '/api/blacklist/device',
      params: { username: username },
      success: ok,
      error: error
    });
  }

  // macs is an array of MAC addresses
  reassignDevices(username, macs, ok, error) {
    $.post('/api/device/reassign', { username: username, macs: macs.join(',') }, ok, error);
  }

  // data = {
  //     "username": "",
  //     "mac-address": "",
  //     "description": "",
  //     "password",  <- Optional
  //     "platform"   <- Optional
  // }
  registerDevice(data, ok, error) {
    if (!('username' in data) || !('mac-address' in data) || !('description' in data)) {
      console.error('Invalid data object');
      return;
    }
    $.post('/api/device', data, ok, error);
  }
}

export default new API();
