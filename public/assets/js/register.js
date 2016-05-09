/*jslint browser:true */
/*globals $*/
$.onReady(function () {
    'use strict';
    var regBtnEnabled = true;
    var register = function () {
        if (!regBtnEnabled) { return; }
        regBtnEnabled = false;
        var data = {
            "username": "",
            "password": "",
            "mac-address": ""
        };

        // It's not guaranteed that all fields will be shown
        // The username and password boxes will only show if the user isn't logged in
        var username = $('[name=username]');
        if (username.length !== 0) {
            data.username = username.value();
            if (data.username === "") { return; }
        }

        var password = $('[name=password]');
        if (password.length !== 0) {
            data.password = password.value();
            if (data.password === "") { return; }
        }

        // The mac-address field will only show for a manual registration
        var mac = $('[name=mac-address]');
        if (mac.length !== 0) {
            data["mac-address"] = mac.value();
            if (data["mac-address"] === "") { return; }
        }

        var platform = $('[name=platform]');
        if (platform.length !== 0) {
            data.platform = platform.value();
            if (data.platform === "") { return; }
        }

        $.post('/api/device', data, function(resp, req) {
            resp = JSON.parse(resp);
            window.scrollTo(0, 0);
            c.FlashMessage("Registration successful", 'success');
            $('.register-box').hide();

            if (data["mac-address"] === "") {
                $('#suc-msg-auto').show();
                return;
            }

            $('#suc-msg-manual').show();
            setTimeout(function() { location.href = resp.Data.Location; }, 3000);
        }, function (req) {
            regBtnEnabled = true;
            window.scrollTo(0, 0);
            var resp = JSON.parse(req.responseText);
            switch(req.status) {
                case 400:
                case 409:
                    c.FlashMessage(resp.Message);
                    break;
                case 401:
                    c.FlashMessage("Login failed");
                    break;
                case 403:
                    c.FlashMessage("Invalid permissions");
                    break;
                default:
                    c.FlashMessage("Internal Server Error");
            }
        });
    };

    $('#register-btn').click(register);
});
