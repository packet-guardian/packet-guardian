j.OnReady(function() {
    j.Click('[name=logout-btn]', function() {
        location.href = "/logout";
        return;
    });
    j.Click('[name=admin-btn]', function() {
        location.href = "/admin";
        return;
    });
    j.Click('[name=add-device-btn]', function() {
        isAdmin = this.getAttribute("data-admin");
        user = j.$('[name=username]').value;
        if (isAdmin !== null) {
            location.href = "/register?manual=1&username="+user;
        } else {
            location.href = "/register?manual=1";
        }
        return;
    });

    // Delete buttons
    j.Click('[name=del-all-btn]', function() {
        var username = j.$('[name=username]').value;
        j.Delete('/api/device/delete', {"username": username}, function(resp) {
            resp = JSON.parse(resp);
            if (resp.Code === 0) {
                location.reload();
                return;
            }
            c.FlashMessage("Error deleteing devices");
        });
    });
    j.Click('[name=del-selected-btn]', function() {
        var checked = j.$('.device-select:checked', true);
        var username = j.$('[name=username]').value;
        var devicesToRemove = [];
        for (var i = 0; i < checked.length; i++) {
            devicesToRemove.push(checked[i].value);
        }
        if (devicesToRemove.length === 0) {
            return;
        }

        j.Delete('/api/device/delete', {"username": username, "mac": devicesToRemove}, function(resp) {
            resp = JSON.parse(resp);
            if (resp.Code === 0) {
                location.reload();
                return;
            }
            c.FlashMessage("Error deleteing devices");
        });
    });
});
