// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
$.onReady(function() {
    var stayOpen = false;
    var reallyOpen = false;
    $(".device-header").click(function(e) {
        if (!reallyOpen) { return; }
        var self = $(e.target);
        if (self.hasClass("device-checkbox")) { return; }

        while (!self.hasClass("device-header")) {
            self = $(self[0].parentNode);
        }
        var bodyNum = self.data("deviceId");
        expandBody(bodyNum);
    });

    function expandBody(bodyNum) {
        var thebody = $("#device-body-"+bodyNum);
        // Get the max-height value before setting it back to 0
        var mh = thebody.style("max-height");
        if (!stayOpen) {
            $(".device-body").style("max-height", "0");
        }
        if (mh !== "200px") {
            thebody.style("max-height", "200px");
        }
    }

    var preOpenID = location.hash.substring(1);
    if (preOpenID) {
        expandBody(preOpenID);
    }

    $(".device-header").on("mousedown", function() {
        reallyOpen = true;
    });

    $(".device-header").on("mousemove", function() {
        reallyOpen = false;
    });

    window.keepDevicesOpen = function(stay) {
        stayOpen = stay;
    };
});