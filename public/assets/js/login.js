j.OnReady(function () {
    'use strict';
    var login = function () {
        var data = {};
        data.username = j.$('[name=username]').value;
        data.password = j.$('[name=password]').value;

        if (data.username === '' || data.password === '') {
            return;
        }

        j.Post('/login', data, function (resp, req) {
            if (resp === '') {
                console.log("malformed response");
            }
            resp = JSON.parse(resp);
            if (resp.Code === 0) {
                console.log("Login successful");
                // Redirect to management page
            }

            c.FlashMessage("Incorrect username or password");

        }, function (req) {
            console.log("an error was encountered");
        });
    };

    j.Click('#login-btn', function () {
        login();
    });

    j.Keyup('[name=password]', function (e) {
        if (e.keyCode === 13) {
            login();
        }
    });
});
