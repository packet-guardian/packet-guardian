import $ from 'jLib';
import api from 'pg-api';
import flashMessage from 'flash';
import { jsPrompt, jsConfirm } from 'modals';

var oldExpiration = '';

$('#delete-btn').click(function() {
    var cmodal = new jsConfirm();
    cmodal.show('Are you sure you want to delete this device?', function() {
        var username = getUsername();

        api.deleteDevices(username, [getMacAddress()], function() {
            location.reload();
        }, function() {
            flashMessage('Error deleting device');
        })
    });
});

$('#unflag-dev-btn').click(function() {
    var cmodal = new jsConfirm();
    cmodal.show('Are you sure you want to unflag this device?', function() {
        api.flagDevice(getMacAddress(), false, function() {
            location.reload();
        }, function() {
            flashMessage('Error unflagging device');
        })
    });
});

$('#flag-dev-btn').click(function() {
    var cmodal = new jsConfirm();
    cmodal.show('Are you sure you want to flag this device?', function() {
        api.flagDevice(getMacAddress(), true, function() {
            location.reload();
        }, function() {
            flashMessage('Error flagging device');
        })
    });
});

$('#unblacklist-btn').click(function() {
    var cmodal = new jsConfirm();
    cmodal.show('Are you sure you want to remove this device from the blacklist?', function() {
        api.unblacklistDevices([getMacAddress()], function() {
            location.reload();
        }, function() {
            flashMessage('Error removing device from blacklist');
        });
    });
});

$('#blacklist-btn').click(function() {
    var cmodal = new jsConfirm();
    cmodal.show('Are you sure you want to blacklist this device?', function() {
        api.blacklistDevices([getMacAddress()], function() {
            location.reload();
        }, function() {
            flashMessage('Error blacklisting device');
        });
    });
});

$('#reassign-btn').click(function() {
    var pmodal = new jsPrompt();
    pmodal.show("New owner's username:", function(newUser) {
        api.reassignDevices(newUser, [getMacAddress()], function() {
            location.reload();
        }, function(req) {
            data = JSON.parse(req.responseText);
            flashMessage(data.Message);
        });
    });
});

function getMacAddress() {
    return $('#mac-address').text();
}

function getUsername() {
    return $('#username').value();
}

function getDescription() {
    return $('#device-desc').text();
}

function setDescription(desc) {
    $('#device-desc').text(desc);
}

$('#edit-dev-desc').click(function(e) {
    e.stopPropagation();
    var pmodal = new jsPrompt();
    pmodal.show('Device Description:', getDescription(), editDeviceDescription);
});

$('#edit-dev-expiration').click(function(e) {
    e.stopPropagation();
    oldExpiration = $('#device-expiration').text();
    $('#device-expiration').text('');
    if (oldExpiration === 'Never') {
        $('#dev-exp-sel').value('never');
    } else if (oldExpiration === 'Rolling') {
        $('#dev-exp-sel').value('rolling');
    } else {
        $('#dev-exp-sel').value('specific');
        $('#dev-exp-val').value(oldExpiration);
        $('#dev-exp-val').style('display', 'inline');
    }
    $('#edit-dev-expiration').style('display', 'none');
    $('#edit-expire-controls').style('display', 'inline');
    $('#confirmation-icons').style('display', 'inline');
});

$('#dev-exp-sel').change(function(e) {
    if ($(e.target).value() !== 'specific') {
        $('#dev-exp-val').style('display', 'none');
    } else {
        c.setTextboxToToday('#dev-exp-val');
        $('#dev-exp-val').style('display', 'inline');
    }
});

$('#dev-expiration-cancel').click(function(e) {
    e.stopPropagation();
    clearExpirationControls(oldExpiration);
});

$('#dev-expiration-ok').click(function(e) {
    e.stopPropagation();
    editDeviceExpiration($('#dev-exp-sel').value(), $('#dev-exp-val').value());
});

// Event callbacks
function clearExpirationControls(value) {
    $('#edit-expire-controls').style('display', 'none');
    $('#confirmation-icons').style('display', 'none');
    $('#device-expiration').text(value);
    $('#dev-exp-val').value('');
    $('#edit-dev-expiration').style('display', 'inline');
}

function editDeviceDescription(desc) {
    api.saveDeviceDescription(getMacAddress(), desc, function() {
        $('#device-desc').text(desc);
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
    });
}

function editDeviceExpiration(type, value) {
    api.saveDeviceExpiration(getMacAddress(), type, value, function(resp, req) {
        resp = JSON.parse(resp);
        clearExpirationControls(resp.Data.newExpiration);
        flashMessage('Device expiration saved', 'success');
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
    });
}
