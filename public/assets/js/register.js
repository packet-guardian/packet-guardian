/*jslint browser:true */
/*globals j*/
j.OnReady(function () {
    'use strict';
    var register = function () {
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
                // Redirect to success page
            }

            c.FlashMessage("Incorrect username or password");
            window.scrollTo(0, 0);

        }, function (req) {
            console.log("an error was encountered");
        });
    };

    j.Click('#register-btn', function () {
        register();
    });
});
