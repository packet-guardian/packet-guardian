/*jslint browser:true */
/*globals $*/
$.onReady(function () {
    'use strict';

    // Form submittion
    $("#new-user-form").submit(function(e) {
        e.preventDefault();
        var username = $("[name=username]").value();
        if (username !== "") {
            location.href = "/admin/users/"+username;
        }
    });
});
