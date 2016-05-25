// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
/*jslint browser:true */
/*globals $*/
$.onReady(function () {
    'use strict';

    var devExpirationTypes = {
        "never": 0,
        "global": 1,
        "specific": 2,
        "duration": 3,
        "daily": 4,
        "rolling": 5
    };

    // Device limit select box init
    function checkLimit() {
        var limit = $("[name=device-limit]");
        var specialLimits = $("[name=special-limit]");
        if (limit.value() === "-1") {
            specialLimits.value("global");
            limit.value("");
            limit.prop("disabled", true);
        } else if (limit.value() === "0") {
            specialLimits.value("unlimited");
            limit.value("");
            limit.prop("disabled", true);
        } else {
            specialLimits.value("specific");
        }
    }
    checkLimit();

    // Expiration textboxes init
    function checkExpires() {
        var limit = $("[name=device-expiration]");
        var devExpSel = $("[name=dev-exp-sel]");
        var expires = devExpSel.data("expires");
        if (expires === "1") {
            devExpSel.value("global");
            limit.value("");
            limit.prop("disabled", true);
        } else if (expires === "0") {
            devExpSel.value("never");
            limit.value("");
            limit.prop("disabled", true);
        } else if (expires === "3") {
            devExpSel.value("duration");
        } else if (expires === "4") {
            devExpSel.value("daily");
            // Remove "Daily at " text
            limit.value(limit.value().slice(10));
        } else {
            devExpSel.value("specific");
        }

        var valAfter = $("[name=valid-after]");
        var valBefore = $("[name=valid-before]");
        var valBefSel = $("[name=val-bef-sel]");
        var forever = valBefSel.data("forever");
        if (forever === "true") {
            valBefSel.value("forever");
            valBefore.value("");
            valBefore.prop("disabled", true);
            valAfter.value("");
            valAfter.prop("disabled", true);
        } else {
            valBefSel.value("specific");
        }
    }
    checkExpires();

    // Select boxes change events
    $("[name=special-limit]").change(function(e) {
        var devLimit = $("[name=device-limit]");
        devLimit.value("");
        devLimit.prop("disabled", ($(e.target).value() !== "specific"));
    });

    $('[name=dev-exp-sel]').change(function(e) {
        var self = $(e.target);
        if (self.value() === "specific" ||
            self.value() === "daily" ||
            self.value() === "duration") {
            $("[name=device-expiration]").prop("disabled", false);
        } else {
            $("[name=device-expiration]").prop("disabled", true);
        }

        if (self.value() === "specific") {
            setTextboxToToday('[name=device-expiration]');
        } else {
            $("[name=device-expiration]").value("");
        }
    });

    $('[name=val-bef-sel]').change(function(e) {
        var self = $(e.target);
        $('[name=valid-before]').prop("disabled", (self.value() === "forever"));
        $("[name=valid-after]").prop("disabled", (self.value() === "forever"));

        if (self.value() === "specific") {
            setTextboxToToday('[name=valid-before]');
            setTextboxToToday('[name=valid-after]');
        } else {
            $("[name=valid-before]").value("");
            $("[name=valid-after]").value("");
        }
    });

    $('[name=delete-btn]').click(function() {
        var username = $('[name=username]').value();
        $.ajax({
            method: "DELETE",
            url: '/api/user',
            params: {"username": username},
            success: function(resp, req) {
                if (req.status > 204) {
                    resp = JSON.parse(resp);
                    c.FlashMessage(resp.Message);
                    return;
                }
                location.href = "/admin/users";
            },
        });
    });

    // Form submittion
    $("#user-form").submit(function(e) {
        e.preventDefault();
        var formData = {
            "username": $('[name=username]').value(),
            "password": $('[name=password]').value(),
            "device_limit": "",
            "expiration_type": "",
            "device_expiration": $('[name=device-expiration]').value(),
            "valid_start": 0,
            "valid_end": 0
        };

        if ($('[name=clear-pass]').prop("checked")) {
            formData.password = -1;
        }

        var devLimit = $('[name=special-limit]').value();
        if (devLimit === "global") {
            formData.device_limit = -1;
        } else if (devLimit === "unlimited") {
            formData.device_limit = 0;
        } else {
            formData.device_limit = $('[name=device-limit]').value();
        }

        var devExpSel = $("[name=dev-exp-sel]").value();
        formData.expiration_type = devExpirationTypes[devExpSel];

        if ($("[name=val-bef-sel]").value() !== "forever") {
            formData.valid_start = $('[name=valid-after]').value();
            formData.valid_end = $('[name=valid-before]').value();
        }

        $.post('/api/user', formData, function(resp, req) {
            if (req.status > 204) {
                resp = JSON.parse(resp);
                c.FlashMessage(resp.Message);
                return;
            }

            c.FlashMessage("User saved", 'success');
            $('[name=password]').value("");
            $('[name=clear-pass]').prop("checked", false);
        });
    });

    // Utility functions
    function setTextboxToToday(el) {
        var date = new Date();
        var dateStr = date.getFullYear() + '-' +
            ('0' + (date.getMonth()+1)).slice(-2) + '-' +
            ('0' + date.getDate()).slice(-2);

        var timeStr = ('0' + date.getHours()).slice(-2) + ':' +
            ('0' + (date.getMinutes())).slice(-2);
        $(el).value(dateStr+" "+timeStr);
    }
});
