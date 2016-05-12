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
            "mac-address": "",
            "description": $('[name=dev-desc]').value()
        };

        // It's not guaranteed that all fields will be shown
        // The username box will always be shown, sometimes disabled
        var username = $('[name=username]');
        if (username.length !== 0) {
            data.username = username.value();
        }
        if (data.username === "") { return; } // Required

        // The password box will only show if the user isn't logged in
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

        // The platform field will only show for a manual registration
        var platform = $('[name=platform]');
        if (platform.length !== 0) {
            data.platform = platform.value();
            if (data.platform === "") { return; }
        }

        if (data.password) { // Need to login first
            $.post('/login', {"username": data.username, "password": data.password}, function(resp, req) {
                delete(data.password);
                registerDevice(data);
            }, function(req) {
                window.scrollTo(0, 0);
                regBtnEnabled = true;
                if (req.status === 401) {
                    c.FlashMessage("Incorrect username or password");
                } else {
                    c.FlashMessage("Unknown error");
                }
            });
        } else {
            registerDevice(data);
        }
    };

    function registerDevice(data) {
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
            switch(req.status) {
                case 400:
                case 409:
                    var resp = JSON.parse(req.responseText);
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
    }

    $('#register-btn').click(register);
});
