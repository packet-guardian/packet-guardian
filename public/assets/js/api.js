// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
/*jslint browser:true */
/*globals $*/

(function(w) {
    // API is a collection of methods used to interact with the API
    // in Packet Guardian. They're centralized here for easy maintenance.
    function API() { }

    // Authentication functions

    // data = {"username": "", "password": ""}
    API.prototype.login = function(data, ok, error) {
        $.post('/login', data, ok, error);
    }

    API.prototype.logout = function(ok, error) {
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
    API.prototype.saveUser = function(data, ok, error) {
        $.post('/api/user', data, ok, error);
    }

    API.prototype.deleteUser = function(username, ok, error) {
        $.ajax({
            method: "DELETE",
            url: '/api/user',
            params: { "username": username },
            success: ok,
            error: error
        });
    }

    API.prototype.blacklistUser = function(username, ok, error) {
        $.post('/api/blacklist/user/' + username, {}, ok, error);
    }

    API.prototype.unblacklistUser = function(username, ok, error) {
        $.ajax({
            method: 'DELETE',
            url: '/api/blacklist/user/' + username,
            success: ok,
            error: error
        });
    }

    // Device functions
    API.prototype.saveDeviceDescription = function(mac, desc, ok, error) {
        mac = encodeURIComponent(mac);
        $.post('/api/device/mac/' + mac + '/description', { "description": desc }, ok, error);
    }

    API.prototype.saveDeviceExpiration = function(mac, type, val, ok, error) {
        mac = encodeURIComponent(mac);
        var data = {
            "type": type,
            "value": val
        };
        $.post('/api/device/mac/' + mac + '/expiration', data, ok, error);
    }

    // macs is an array of MAC addresses
    API.prototype.deleteDevices = function(username, macs, ok, error) {
        $.ajax({
            method: "DELETE",
            url: "/api/device/user/" + username,
            params: { "mac": macs.join(',') },
            success: ok,
            error: error
        });
    }

    // macs is an array of MAC addresses
    API.prototype.blacklistDevices = function(username, macs, ok, error) {
        $.post('/api/blacklist/device/' + username, { "mac": macs.join(',') }, ok, error);
    }

    // macs is an array of MAC addresses
    API.prototype.unblacklistDevices = function(username, macs, ok, error) {
        $.ajax({
            method: 'DELETE',
            url: '/api/blacklist/device/' + username,
            params: { "mac": macs.join(',') },
            success: ok,
            error: error
        });
    }

    // macs is an array of MAC addresses
    API.prototype.reassignDevices = function(username, macs, ok, error) {
        $.post("/api/device/reassign", { "username": username, "macs": macs.join(',') }, ok, error);
    }

    // data = {
    //     "username": "",
    //     "mac-address": "",
    //     "description": "",
    //     "password",  <- Optional
    //     "platform"   <- Optional
    // }
    API.prototype.registerDevice = function(data, ok, error) {
        $.post('/api/device', data, ok, error);
    }

    w.API = new API();
})(window);
