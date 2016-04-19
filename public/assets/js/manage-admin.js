j.OnReady(function() {
    var blackUserBtn = j.$('[name=black-user-btn]');
    if (blackUserBtn !== null) {
        if (blackUserBtn.getAttribute("data-blacklisted") === "true") {
            blackUserBtn.innerHTML = "Unblacklist Username";
        }
    }

    // Admin buttons
    j.Click('[name=black-user-btn]', function() {
        var username = encodeURIComponent(j.$('[name=username]').value);
        var isBl = (this.getAttribute("data-blacklisted") === "true");
        blacklistUser(username, !isBl);
    });
    j.Click('[name=black-all-btn]', function() {
        var username = encodeURIComponent(j.$('[name=username]').value);
        j.Post('/api/blacklist/device', {"username": username}, function(resp) {
            resp = JSON.parse(resp);
            if (resp.Code === 0) {
                blacklistUser(username, true);
                return;
            }
            c.FlashMessage("Error blacklisting devices and user");
        });
    });
    j.Click('[name=unblack-all-btn]', function() {
        var username = encodeURIComponent(j.$('[name=username]').value);
        j.Delete('/api/blacklist/device', {"username": username}, function(resp) {
            resp = JSON.parse(resp);
            if (resp.Code === 0) {
                blacklistUser(username, false);
                return;
            }
            c.FlashMessage("Error removing devices and user from blacklist");
        });
    });

    function blacklistUser(username, blacklist) {
        var method = j.Delete;

        if (blacklist) {
            method = j.Post;
        }

        method('/api/blacklist/user', {"username": username}, function(resp) {
            resp = JSON.parse(resp);
            if (resp.Code === 0) {
                location.reload();
                return;
            }
            c.FlashMessage("Error blacklisting user");
        });
    }

    j.Click('[name=black-sel-btn]', function() {
        blacklistSelectedDevices(true);
    });
    j.Click('[name=unblack-sel-btn]', function() {
        blacklistSelectedDevices(false);
    });

    function blacklistSelectedDevices(addTo) {
        var username = encodeURIComponent(j.$('[name=username]').value);
        var checked = j.$('.device-select:checked', true);
        var devicesToRemove = [];
        for (var i = 0; i < checked.length; i++) {
            devicesToRemove.push(checked[i].value);
        }
        var method = j.Delete;
        if (addTo) {
            method = j.Post;
        }

        method('/api/blacklist/device', {"mac": devicesToRemove, "username": username}, function(resp) {
            resp = JSON.parse(resp);
            if (resp.Code === 0) {
                location.reload();
                return;
            }
            c.FlashMessage("Error removing devices from blacklist");
        });
    }
});
