// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
$.onReady(function() {
    "use strict";
    // Event handlers
    $("[name=blacklist-sel]").change(function(e) {
        var self = $(e.target);
        var cmodal = new jsConfirm();
        switch (self.value()) {
            case "username":
                var isBl = (self.data("blacklisted") === "true");
                if (isBl) {
                    cmodal.show("Remove username from blacklist?",
                        function() { blacklistUsername(true); });
                } else {
                    cmodal.show("Add username to blacklist?",
                        function() { blacklistUsername(false); });
                }
                break;
            case "black-all":
                cmodal.show("Add all user's devices to blacklist?", addDevicesToBlacklist);
                break;
            case "unblack-all":
                cmodal.show("Remove all user's devices from blacklist?", removeDevicesFromBlacklist);
                break;
            case "black-sel":
                cmodal.show("Add selected user's devices to blacklist?",
                    function() { blacklistSelectedDevices(true); });
                break;
            case "unblack-sel":
                cmodal.show("Remove selected user's devices from blacklist?",
                    function() { blacklistSelectedDevices(false); });
                break;
        }
        self.value("");
    });

    $('[name=reassign-selected-btn]').click(function() {
        var pmodal = new jsPrompt();
        pmodal.show("New owner's username:", reassignSelectedDevices);
    });

    // Event callbacks
    var blacklistSelect = $('[name=blacklist-sel]');
    if (blacklistSelect.length !== 0) {
        if (blacklistSelect.data("blacklisted") === "true") {
            $('[name=black-user-option]').text("Remove Username");
        }
    }

    function getUsername(encode) {
        var u = $('[name=username]').value();
        if (encode) {
            return encodeURIComponent(u);
        }
        return u;
    }

    function blacklistUsername(isBlacklisted) {
        var success = function() { location.reload(); };
        var error = function() { c.FlashMessage("Error blacklisting user"); };

        if (isBlacklisted) {
            API.unblacklistUser(getUsername(), success, error);
        } else {
            API.blacklistUser(getUsername(), success, error);
        }
    }

    function getCheckedDevices() {
        var checked = $('.device-checkbox:checked');
        var devices = [];
        for (var i = 0; i < checked.length; i++) {
            devices.push(checked[i].value);
        }
        return devices;
    }

    function blacklistSelectedDevices(add) {
        var devicesToRemove = getCheckedDevices();
        if (devicesToRemove.length === 0) {
            return;
        }
        blacklistDevices(devicesToRemove, add);
    }

    function blacklistDevices(devices, add) {
        if (add) {
            addDevicesToBlacklist(devices);
        } else {
            removeDevicesFromBlacklist(devices);
        }
    }

    function addDevicesToBlacklist(devices) {
        if (devices) {
            API.blacklistDevices(devices, reloadPage, errorAdding);
        } else {
            API.blacklistAllDevices(getUsername(), reloadPage, errorAdding);
        }
    }

    function removeDevicesFromBlacklist(devices) {
        if (devices) {
            API.unblacklistDevices(devices, reloadPage, errorRemoving);
        } else {
            API.unblacklistAllDevices(getUsername(), reloadPage, errorRemoving);
        }
    }

    function reassignSelectedDevices(username) {
        var devices = getCheckedDevices();
        if (devices.length === 0 || !username) {
            return;
        }

        API.reassignDevices(username, devices, function() {
            location.reload();
        }, function() {
            c.FlashMessage("Error reassigning devices");
        })
    }

    function reloadPage() {
        location.reload();
    }

    function errorAdding() {
        c.FlashMessage("Error blacklisting devices");
    }

    function errorRemoving() {
        c.FlashMessage("Error removing devices from blacklist");
    }
});
