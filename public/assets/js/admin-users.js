// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
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
