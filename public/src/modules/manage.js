import $ from 'jLib';
import api from 'pg-api';
import flashMessage from 'flash';
import { jsPrompt, jsConfirm } from 'modals';

function initManage() {
    // Event handlers
    $('[name=add-device-btn]').click(function(e) {
        var isAdmin = $(e.target).data("admin");
        var user = $('[name=username]').value();
        if (isAdmin !== null) {
            location.href = `/register?manual=1&username=${user}`;
        } else {
            location.href = '/register?manual=1';
        }
        return;
    });

    // Delete buttons
    $('[name=del-all-btn]').click(function() {
        var cmodal = new jsConfirm();
        cmodal.show('Are you sure you want to delete all devices?', deleteAllDevices);
    });

    $('[name=del-selected-btn]').click(function() {
        var cmodal = new jsConfirm();
        cmodal.show('Are you sure you want to delete selected devices?', deleteSelectedDevices);
    });

    $('[name=dev-sel-all]').click(function(e) {
        $('.device-checkbox').prop('checked', $(e.target).prop('checked'));
    });

    $('.device-checkbox').click(function(e) {
        $('[name=dev-sel-all]').prop('checked', false);
    });

    $('.edit-dev-desc').click(function(e) {
        e.stopPropagation();
        var id = $(e.target).data('device');
        var pmodal = new jsPrompt();
        pmodal.show('Device Description:', $(`#device-${id}-desc`).text(), function(desc) {
            editDeviceDescription(id, desc);
        });
    });
}

// Event callbacks
function editDeviceDescription(id, desc) {
    var mac = $(`#device-${id}-mac`).text();
    api.saveDeviceDescription(mac, desc, function(resp, req) {
        $(`#device-${id}-desc`).text(desc);
        flashMessage('Device description saved', 'success');
    }, function(req) {
        var resp = JSON.parse(req.responseText);
        switch (req.status) {
            case 500:
                flashMessage(`Internal Server Error - ${resp.Message}`);
                break;
            default:
                flashMessage(resp.Message);
                break;
        }
    }
    );
}

// Delete buttons
function deleteAllDevices() {
    api.deleteDevices($('[name=username]').value(), [], function() {
        location.reload();
    }, function() {
        flashMessage('Error deleting devices');
    })
}

function deleteSelectedDevices() {
    var checked = $('.device-checkbox:checked');
    if (checked.length === 0) {
        return;
    }

    var devicesToRemove = [];
    for (var i = 0; i < checked.length; i++) {
        devicesToRemove.push(checked[i].value);
    }

    api.deleteDevices($('[name=username]').value(), devicesToRemove,
        () => location.reload(),
        () => flashMessage('Error deleting devices')
    );
}

initManage();
