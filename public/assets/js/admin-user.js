/*jslint browser:true */
/*globals j*/
j.OnReady(function () {
    'use strict';
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
        } else {
            devExpSel.value = "specific";
        }

        var valBefore = j.$("[name=valid-before]");
        var valBefSel = j.$("[name=val-bef-sel]");
        var forever = valBefSel.getAttribute("data-forever");
        if (forever === "true") {
            valBefSel.value = "forever";
            valBefore.value = "";
            valBefore.disabled = true;
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
        j.$('[name=device-expiration]').disabled = (this.value !== "specific");

        if (this.value === "specific") {
            setTextboxToToday('[name=device-expiration]');
        } else {
            j.$("[name=device-expiration]").value = "";
        }
    });

    j.Change('[name=val-bef-sel]', function() {
        j.$('[name=valid-before]').disabled = (this.value !== "specific");

        if (this.value === "specific") {
            setTextboxToToday('[name=valid-before]');
        } else {
            j.$("[name=valid-before]").value = "";
        }
    });

    j.Click('[name=delete-btn]', function() {
        var username = j.$('[name=username]').value;
        j.Delete('/admin/users/'+username, {}, function(resp) {
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
            "clear-pass": j.$('[name=clear-pass]').checked,
            "special-limit": j.$('[name=special-limit]').value,
            "device-limit": j.$('[name=device-limit]').value,
            "device-expiration": j.$('[name=device-expiration]').value,
            "valid-after": j.$('[name=valid-after]').value,
            "valid-before": j.$('[name=valid-before]').value
        };

        var devExpSel = j.$("[name=dev-exp-sel]").value;
        if (devExpSel === "global") {
            formData["device-expiration"] = 1;
        } else if (devExpSel === "never") {
            formData["device-expiration"] = 0;
        }

        var valBefSel = j.$("[name=val-bef-sel]").value;
        if (valBefSel === "forever") {
            formData["valid-before"] = 0;
        }

        j.Post('/admin/users', formData, function(resp) {
            resp = JSON.parse(resp);
            if (resp.Code !== 0) {
                c.FlashMessage(resp.Message);
                return;
            }

            c.FlashMessage(resp.Message, 'success');
            j.$('[name=password]').value = "";
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
