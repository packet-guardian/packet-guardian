/*jslint browser:true */
/*globals j*/
j.OnReady(function () {
    'use strict';

    // Form submittion
    j.Submit("#new-user-form", function(e) {
        var username = j.$("[name=username]").value;
        if (username !== "") {
            location.href = "/admin/users/"+username;
        }

        e.preventDefault();
    });
});
