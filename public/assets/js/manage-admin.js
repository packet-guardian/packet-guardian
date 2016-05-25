// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
$.onReady(function() {
    var blacklistSelect = $('[name=blacklist-sel]');
    if (blacklistSelect.length !== 0) {
        if (blacklistSelect.data("blacklisted") === "true") {
            $('[name=black-user-option]').text("Remove Username");
        }
    }

    $("[name=blacklist-sel]").change(function(e) {
        var self = $(e.target);
        switch (self.value()) {
            case "username":
                var isBl = (self.data("blacklisted") === "true");
                blacklistUsername(isBl);
                break;
            case "black-all":
                blacklistDevices([], true);
                break;
            case "unblack-all":
                blacklistDevices([], false);
                break;
            case "black-sel":
                blacklistSelectedDevices(true);
                break;
            case "unblack-sel":
                blacklistSelectedDevices(false);
                break;
        }
        self.value("");
    });

    function getUsername(encode) {
        var u = $('[name=username]').value();
        if (encode) {
            return encodeURIComponent(u);
        }
        return u;
    }

    function blacklistUsername(isBlacklisted) {
        var method = "POST";
        if (isBlacklisted) {
            method = "DELETE";
        }

        $.ajax({
            method: method,
            url: '/api/blacklist/user/'+getUsername(),
            success: function() {
                location.reload();
            },
            error: function() {
                c.FlashMessage("Error blacklisting user");
            }
        });
    }

    function getCheckedDevices() {
        var checked = $('.device-select:checked');
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
        devices = devices.join(',');
        if (add) {
            addDevicesToBlacklist(devices);
        } else {
            removeDevicesFromBlacklist(devices);
        }
    }

    function addDevicesToBlacklist(devices) {
        $.post('/api/blacklist/device/'+getUsername(), {"mac": devices}, function() {
            location.reload();
        }, function() {
            c.FlashMessage("Error blacklisting devices");
        });
    }

    function removeDevicesFromBlacklist(devices) {
        $.ajax({
            method: 'DELETE',
            url: '/api/blacklist/device/'+getUsername(),
            params: {"mac": devices},
            success: function() {
                location.reload();
            },
            error: function() {
                c.FlashMessage("Error removing devices from blacklist");
            }
        });
    }

    $('[name=reassign-selected-btn]').click(function() {
        var devices = getCheckedDevices();
        var username = prompt("Reassign to user:");
        if (devices.length === 0 || !username) {
            return;
        }
        $.post("/api/device/_reassign", {"username": username, "macs": devices.join(',')}, function() {
            location.reload();
        }, function() {
            c.FlashMessage("Error blacklisting devices");
        });
    });
});
