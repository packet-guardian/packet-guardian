j.OnReady(function() {
    var blacklistSelect = j.$('[name=blacklist-sel]');
    if (blacklistSelect !== null) {
        if (blacklistSelect.getAttribute("data-blacklisted") === "true") {
            j.$('[name=black-user-option]').text = "Remove Username";
        }
    }

    j.Change("[name=blacklist-sel]", function() {
        switch (this.value) {
            case "username":
                var isBl = (this.getAttribute("data-blacklisted") === "true");
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
        this.value = "";
    });

    function getUsername(encode) {
        var u = j.$('[name=username]').value;
        if (encode) {
            return encodeURIComponent(u);
        }
        return u;
    }

    function blacklistUsername(isBlacklisted) {
        var method = j.Post;

        if (isBlacklisted) {
            method = j.Delete;
        }

        method('/api/blacklist/user', {"username": getUsername(true)}, function(resp) {
            resp = JSON.parse(resp);
            if (resp.Code === 0) {
                location.reload();
                return;
            }
            c.FlashMessage("Error blacklisting user");
        });
    }

    function blacklistSelectedDevices(add) {
        var checked = j.$('.device-select:checked', true);
        var devicesToRemove = [];
        for (var i = 0; i < checked.length; i++) {
            devicesToRemove.push(checked[i].value);
        }
        if (devicesToRemove.length === 0) {
            return;
        }
        blacklistDevices(devicesToRemove, add);
    }

    function blacklistDevices(devices, add) {
        var method = j.Delete;
        if (add) {
            method = j.Post;
        }

        method('/api/blacklist/device', {"mac": devices, "username": getUsername(true)}, function(resp) {
            resp = JSON.parse(resp);
            if (resp.Code === 0) {
                location.reload();
                return;
            }
            c.FlashMessage("Error removing devices from blacklist");
        });
    }
});
