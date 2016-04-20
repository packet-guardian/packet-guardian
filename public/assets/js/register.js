/*jslint browser:true */
/*globals j*/
j.OnReady(function () {
    'use strict';
    var regBtnEnabled = true;
    var register = function () {
        if (!regBtnEnabled) { return; }
        var data = {
            "username": "",
            "password": "",
            "mac-address": ""
        };

        // It's not guaranteed that all fields will be shown
        // The username and password boxes will only show if the user isn't logged in
        var username = j.$('[name=username]');
        if (username !== null) {
            data.username = username.value;
            if (data.username === "") { return; }
        }

        var password = j.$('[name=password]');
        if (password !== null) {
            data.password = password.value;
            if (data.password === "") { return; }
        }

        // The mac-address field will only show for a manual registration
        var mac = j.$('[name=mac-address]');
        if (mac !== null) {
            data["mac-address"] = mac.value;
            if (data["mac-address"] === "") { return; }
        }

        var platform = j.$('[name=platform]');
        if (platform !== null) {
            data.platform = platform.value;
            if (data.platform === "") { return; }
        }

        j.Post('/api/device/register', data, function (resp, req) {
            if (resp === '') {
                c.FlashMessage("Server error, please call the IT help desk");
                return;
            }
            resp = JSON.parse(resp);
            window.scrollTo(0, 0);
            if (resp.Code !== 0) {
                c.FlashMessage(resp.Message);
                return;
            }

            c.FlashMessage(resp.Message, 'success');
            regBtnEnabled = false;
            j.Hide('.register-box');

            if (data["mac-address"] === "") {
                j.$('#suc-msg-auto').style.display = 'block';
                return;
            }

            j.$('#suc-msg-manual').style.display = 'block';
            setTimeout(function() { location.href = resp.Data.Location; }, 3000);
        }, function (req) {
            c.FlashMessage("Server error, please call the IT help desk");
            return;
        });
    };

    j.Click('#register-btn', register);
});
