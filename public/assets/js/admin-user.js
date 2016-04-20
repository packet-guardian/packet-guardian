/*jslint browser:true */
/*globals j*/
j.OnReady(function () {
    'use strict';

    var devExpirationTypes = {
        "never": 0,
        "global": 1,
        "specific": 2,
        "duration": 3,
        "daily": 4
    };

    // Device limit select box init
    function checkLimit() {
        var limit = j.$("[name=device-limit]");
        var specialLimits = j.$("[name=special-limit]");
        if (limit.value === "-1") {
            specialLimits.value = "global";
            limit.value = "";
            limit.disabled = true;
        } else if (limit.value === "0") {
            specialLimits.value = "unlimited";
            limit.value = "";
            limit.disabled = true;
        } else {
            specialLimits.value = "specific";
        }
    }
    checkLimit();

    // Expiration textboxes init
    function checkExpires() {
        var limit = j.$("[name=device-expiration]");
        var devExpSel = j.$("[name=dev-exp-sel]");
        var expires = devExpSel.getAttribute("data-expires");
        if (expires === "1") {
            devExpSel.value = "global";
            limit.value = "";
            limit.disabled = true;
        } else if (expires === "0") {
            devExpSel.value = "never";
            limit.value = "";
            limit.disabled = true;
        } else if (expires === "3") {
            devExpSel.value = "duration";
        } else if (expires === "4") {
            devExpSel.value = "daily";
            // Remove "Daily at " text
            limit.value = limit.value.slice(10);
        } else {
            devExpSel.value = "specific";
        }

        var valAfter = j.$("[name=valid-after]");
        var valBefore = j.$("[name=valid-before]");
        var valBefSel = j.$("[name=val-bef-sel]");
        var forever = valBefSel.getAttribute("data-forever");
        if (forever === "true") {
            valBefSel.value = "forever";
            valBefore.value = "";
            valBefore.disabled = true;
            valAfter.value = "";
            valAfter.disabled = true;
        } else {
            valBefSel.value = "specific";
        }
    }
    checkExpires();

    // Select boxes change events
    j.Change("[name=special-limit]", function() {
        j.$("[name=device-limit]").value = "";
        j.$("[name=device-limit]").disabled = (this.value !== "specific");
    });

    j.Change('[name=dev-exp-sel]', function() {
        if (this.value === "specific" || this.value === "daily" || this.value === "duration") {
            j.$("[name=device-expiration]").disabled = false;
        } else {
            j.$("[name=device-expiration]").disabled = true;
        }

        if (this.value === "specific") {
            setTextboxToToday('[name=device-expiration]');
        } else {
            j.$("[name=device-expiration]").value = "";
        }
    });

    j.Change('[name=val-bef-sel]', function() {
        j.$('[name=valid-before]').disabled = (this.value === "forever");
        j.$("[name=valid-after]").disabled = (this.value === "forever");

        if (this.value === "specific") {
            setTextboxToToday('[name=valid-before]');
            setTextboxToToday('[name=valid-after]');
        } else {
            j.$("[name=valid-before]").value = "";
            j.$("[name=valid-after]").value = "";
        }
    });

    j.Click('[name=delete-btn]', function() {
        var username = j.$('[name=username]').value;
        j.Delete('/api/user', {"username": username}, function(resp) {
            resp = JSON.parse(resp);
            if (resp.Code !== 0) {
                c.FlashMessage(resp.Message);
                return;
            }
            location.href = "/admin/users";
        });
    });

    // Form submittion
    j.Submit("#user-form", function(e) {
        var formData = {
            "username": j.$('[name=username]').value,
            "password": j.$('[name=password]').value,
            "device_limit": "",
            "expiration_type": "",
            "device_expiration": j.$('[name=device-expiration]').value,
            "valid_start": 0,
            "valid_end": 0
        };

        if (j.$('[name=clear-pass]').checked) {
            formData.password = -1;
        }

        var devLimit = j.$('[name=special-limit]').value;
        if (devLimit === "global") {
            formData.device_limit = -1;
        } else if (devLimit === "unlimited") {
            formData.device_limit = 0;
        } else {
            formData.device_limit = j.$('[name=device-limit]').value;
        }

        var devExpSel = j.$("[name=dev-exp-sel]").value;
        formData.expiration_type = devExpirationTypes[devExpSel];

        if (j.$("[name=val-bef-sel]").value !== "forever") {
            formData.valid_start = j.$('[name=valid-after]').value;
            formData.valid_end = j.$('[name=valid-before]').value;
        }

        j.Post('/api/user', formData, function(resp) {
            resp = JSON.parse(resp);
            if (resp.Code !== 0) {
                c.FlashMessage(resp.Message);
                return;
            }

            c.FlashMessage(resp.Message, 'success');
            j.$('[name=password]').value = "";
            j.$('[name=clear-pass]').checked = false;
        });

        e.preventDefault();
    });

    // Utility functions
    function setTextboxToToday(el) {
        var date = new Date();
        var dateStr = date.getFullYear() + '-' +
            ('0' + (date.getMonth()+1)).slice(-2) + '-' +
            ('0' + date.getDate()).slice(-2);

        var timeStr = ('0' + date.getHours()).slice(-2) + ':' +
            ('0' + (date.getMinutes())).slice(-2) + ':' +
            ('0' + (date.getSeconds())).slice(-2);
        j.$(el).value = dateStr+" "+timeStr;
    }
});
