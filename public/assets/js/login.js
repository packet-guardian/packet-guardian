$.onReady(function () {
    'use strict';
    var login = function() {
        var data = {};
        data.username = $('[name=username]').value();
        data.password = $('[name=password]').value();

        if (data.username === '' || data.password === '') {
            return;
        }

        $.post('/login', data, function(resp, req) {
            location.href = '/';
        }, function(req) {
            if (req.status === 401) {
                c.FlashMessage("Incorrect username or password");
            } else {
                c.FlashMessage("Unknown error");
            }
        });
    };

    $('#login-btn').click(function() {
        login();
    });

    $('[name=password]').keyup(function(e) {
        if (e.keyCode === 13) {
            login();
        }
    });
});
