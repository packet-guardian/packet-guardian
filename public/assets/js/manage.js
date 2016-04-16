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
        j.Post('/device/delete', {"username": username}, function(resp) {
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
        j.Post('/device/delete', {"username": username, "mac": devicesToRemove}, function(resp) {
            resp = JSON.parse(resp);
            if (resp.Code === 0) {
                location.reload();
                return;
            }
            c.FlashMessage("Error deleteing devices");
        });
    });

    // Admin buttons
    // j.Click('[name=black-user-btn]', function() {
    //     var username = encodeURIComponent(j.$('[name=username]').value);
    //     var isBl = (this.getAttribute("data-blacklisted") === "true");
    //     var method = j.Post;
    //
    //     if (isBl) {
    //         method = j.Delete;
    //     }
    //
    //     method('/admin/blacklist/'+username, {}, function(resp) {
    //         resp = JSON.parse(resp);
    //         if (resp.Code === 0) {
    //             location.reload();
    //             return;
    //         }
    //         c.FlashMessage("Error blacklisting user");
    //     });
    // });
    // j.Click('[name=black-all-btn]', function() {
    //     var username = encodeURIComponent(j.$('[name=username]').value);
    //     j.Post('/admin/blacklist/'+username+'/all', {}, function(resp) {
    //         resp = JSON.parse(resp);
    //         if (resp.Code === 0) {
    //             location.reload();
    //             return;
    //         }
    //         c.FlashMessage("Error blacklisting user");
    //     });
    // });
    // j.Click('[name=unblack-all-btn]', function() {
    //     var username = encodeURIComponent(j.$('[name=username]').value);
    //     j.Delete('/admin/blacklist/'+username+'/all', {}, function(resp) {
    //         resp = JSON.parse(resp);
    //         if (resp.Code === 0) {
    //             location.reload();
    //             return;
    //         }
    //         c.FlashMessage("Error removing user from blacklist");
    //     });
    // });
    // j.Click('[name=black-sel-btn]', function() {
    //     var checked = j.$('.device-select:checked', true);
    //     var username = j.$('[name=username]').value;
    //     var devicesToRemove = [];
    //     for (var i = 0; i < checked.length; i++) {
    //         devicesToRemove.push(checked[i].value);
    //     }
    //     j.Post('/admin/blacklist', {"devices": devicesToRemove}, function(resp) {
    //         resp = JSON.parse(resp);
    //         if (resp.Code === 0) {
    //             location.reload();
    //             return;
    //         }
    //         c.FlashMessage("Error deleteing devices");
    //     });
    // });
    // j.Click('[name=unblack-sel-btn]', function() {
    //     var checked = j.$('.device-select:checked', true);
    //     var devicesToRemove = [];
    //     for (var i = 0; i < checked.length; i++) {
    //         devicesToRemove.push(checked[i].value);
    //     }
    //     j.Delete('/admin/blacklist', {"devices": devicesToRemove}, function(resp) {
    //         resp = JSON.parse(resp);
    //         if (resp.Code === 0) {
    //             location.reload();
    //             return;
    //         }
    //         c.FlashMessage("Error deleteing devices");
    //     });
    // });
    //
    // var blackUserBtn = j.$('[name=black-user-btn]');
    // if (blackUserBtn !== null) {
    //     if (blackUserBtn.getAttribute("data-blacklisted") === "true") {
    //         blackUserBtn.innerHTML = "Unblacklist Username";
    //     }
    // }
});
