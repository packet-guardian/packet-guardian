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

    // Event callbacks
    // Delete buttons
    function deleteAllDevices() {
        var username = $('[name=username]').value();
        $.ajax({
            method: "DELETE",
            url: "/api/device/"+username,
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
            url: '/api/device/'+username,
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
