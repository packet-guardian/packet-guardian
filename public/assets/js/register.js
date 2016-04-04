/*jslint browser:true */
/*globals j*/
j.OnReady(function () {
    'use strict';
    var regBtnEnabled = true;
    var register = function () {
        if (!regBtnEnabled) { return; }
        var data = {
            username: j.$('[name=username]').value,
            password: j.$('[name=password]').value
        };

        if (data.username === '' || data.password === '') {
            return;
        }

        j.Post('/register', data, function (resp, req) {
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
            j.$('.success-message').style.display = 'block';
        }, function (req) {
            c.FlashMessage("Server error, please call the IT help desk");
            return;
        });
    };

    j.Click('#register-btn', register);
});
