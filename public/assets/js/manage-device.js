// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

$.onReady(function() {
    $('#delete-btn').click(function() {
        var cmodal = new jsConfirm();
        cmodal.show("Are you sure you want to delete this device?", function() {
            var mac = getMacAddress();
            var username = getUsername();

            $.ajax({
                method: 'DELETE',
                url: '/api/device/'+username,
                params: {"mac": mac},
                success: function() {
                    location.href = '/admin/manage/user/'+username;
                },
                error: function() {
                    c.FlashMessage("Error deleting device");
                }
            });
        });
    });

    $('#unblacklist-btn').click(function() {
        var cmodal = new jsConfirm();
        cmodal.show("Are you sure you want to remove this device from the blacklist?", function() {
            var mac = getMacAddress();
            $.ajax({
                method: 'DELETE',
                url: '/api/blacklist/device/'+getUsername(),
                params: {"mac": mac},
                success: function() {
                    location.reload();
                },
                error: function() {
                    c.FlashMessage("Error removing device from blacklist");
                }
            });
        });
    });

    $('#blacklist-btn').click(function() {
        var cmodal = new jsConfirm();
        cmodal.show("Are you sure you want to blacklist this device?", function() {
            var mac = getMacAddress();
            $.post('/api/blacklist/device/'+getUsername(), {"mac": mac}, function() {
                location.reload();
            }, function() {
                c.FlashMessage("Error blacklisting device");
            });
        });
    });

    $('#reassign-btn').click(function() {
        var pmodal = new jsPrompt();
        pmodal.show("New owner's username:", function(newUser) {
            var mac = getMacAddress();
            $.post("/api/device/_reassign", {"username": newUser, "macs": mac}, function() {
                location.reload();
            }, function(req) {
                data = JSON.parse(req.responseText);
                c.FlashMessage(data.Message);
            });
        });
    });

    function getMacAddress() {
        return $('#mac-address').text();
    }

    function getUsername() {
        return $('#username').value();
    }
});
