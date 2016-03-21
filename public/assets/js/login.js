/*jslint browser:true */
/*globals j*/
j.OnReady(function () {
    'use strict';
    j.Click('#login-btn', function () {
        var data = {};
        data.username = j.$('[name=username]').value;
        data.password = j.$('[name=password]').value;

        j.Post('/login', data, function (_, req) {
            if (req.status === 200) {
                window.location.href = "/manage";
            } else {
                console.log("malformed response");
            }
        }, function (req) {
            if (req.status === 403) {
                console.log("bad login");
            } else {
                console.log("an error was encountered");
            }
        });
    });
});
