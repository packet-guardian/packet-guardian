// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
$.onReady(function() {
    // Event handlers
    $('[name=add-device-btn]').click(function(e) {
        isAdmin = $(e.target).data("admin");
        user = $('[name=username]').value();
        if (isAdmin !== null) {
            location.href = "/register?manual=1&username="+user;
        } else {
            location.href = "/register?manual=1";
        }
        return;
    });

    // Delete buttons
    $('[name=del-all-btn]').click(function() {
        var cmodal = new jsConfirm();
        cmodal.show("Are you sure you want to delete all devices?", deleteAllDevices);
    });

    $('[name=del-selected-btn]').click(function() {
        var cmodal = new jsConfirm();
        cmodal.show("Are you sure you want to delete selected devices?", deleteSelectedDevices);
    });

    $('[name=dev-sel-all]').click(function(e) {
        $('.device-checkbox').prop("checked", $(e.target).prop("checked"));
    });

    $('.device-checkbox').click(function(e) {
        $('[name=dev-sel-all]').prop("checked", false);
    });

    $('.edit-dev-desc').click(function(e) {
        e.stopPropagation();
        id = $(e.target).data("device");
        var pmodal = new jsPrompt();
        pmodal.show("Device Description:", $('#device-'+id+'-desc').text(), function(desc) {
            editDeviceDescription(id, desc);
        });
    });

    // Event callbacks
    function editDeviceDescription(id, desc) {
        mac = $('#device-'+id+'-mac').text();
        $.post(
            '/api/device/mac/'+encodeURIComponent(mac)+'/description',
            {"description": desc}, function(resp, req) {
                $('#device-'+id+'-desc').text(desc);
                c.FlashMessage("Device description saved", 'success');
            }, function(req) {
                var resp = JSON.parse(req.responseText);
                switch(req.status) {
                    case 500:
                        c.FlashMessage("Internal Server Error - "+resp.Message);
                        break;
                    default:
                        c.FlashMessage(resp.Message);
                        break;
                    }
            }
        );
    }

    // Delete buttons
    function deleteAllDevices() {
        var username = $('[name=username]').value();
        $.ajax({
            method: "DELETE",
            url: "/api/device/user/"+username,
            success: function() {
                location.reload();
            },
            error: function() {
                c.FlashMessage("Error deleting devices");
            }
        });
    }

    function deleteSelectedDevices() {
        var checked = $('.device-checkbox:checked');
        if (checked.length === 0) {
            return;
        }

        var username = $('[name=username]').value();
        var devicesToRemove = [];
        for (var i = 0; i < checked.length; i++) {
            devicesToRemove.push(checked[i].value);
        }

        $.ajax({
            method: 'DELETE',
            url: '/api/device/user/'+username,
            params: {"mac": devicesToRemove.join(',')},
            success: function() {
                location.reload();
            },
            error: function() {
                c.FlashMessage("Error deleting devices");
            }
        });
    }
});
