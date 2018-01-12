import $ from 'jLib';
import api from 'pg-api';
import flashMessage from 'flash';
import { setTextboxToToday } from 'utils';

const devExpirationTypes = {
    never: 0,
    global: 1,
    specific: 2,
    duration: 3,
    daily: 4,
    rolling: 5
};

// Device limit select box init
function checkLimit() {
    const limit = $('[name=device-limit]');
    const specialLimits = $('[name=special-limit]');
    if (limit.value() === '-1') {
        specialLimits.value('global');
        limit.value('');
        limit.prop('disabled', true);
    } else if (limit.value() === '0') {
        specialLimits.value('unlimited');
        limit.value('');
        limit.prop('disabled', true);
    } else {
        specialLimits.value('specific');
    }
}
checkLimit();

// Expiration textboxes init
function checkExpires() {
    const limit = $('[name=device-expiration]');
    const devExpSel = $('[name=dev-exp-sel]');
    const expires = devExpSel.data('expires');
    switch (expires) {
        case '0':
            devExpSel.value('never');
            limit.value('');
            limit.prop('disabled', true);
            break;
        case '1':
            devExpSel.value('global');
            limit.value('');
            limit.prop('disabled', true);
            break;
        case '3':
            devExpSel.value('duration');
            break;
        case '4':
            devExpSel.value('daily');
            // Remove "Daily at " text
            limit.value(limit.value().slice(9));
            break;
        case '5':
            devExpSel.value('rolling');
            limit.value('');
            limit.prop('disabled', true);
            break;
        default:
            devExpSel.value('specific');
    }

    const valAfter = $('[name=valid-after]');
    const valBefore = $('[name=valid-before]');
    const valBefSel = $('[name=val-bef-sel]');
    const forever = valBefSel.data('forever');
    if (forever === 'true') {
        valBefSel.value('forever');
        valBefore.value('');
        valBefore.prop('disabled', true);
        valAfter.value('');
        valAfter.prop('disabled', true);
    } else {
        valBefSel.value('specific');
    }
}
checkExpires();

// Select boxes change events
$('[name=special-limit]').change(function(e) {
    const devLimit = $('[name=device-limit]');
    devLimit.value('');
    devLimit.prop('disabled', ($(e.target).value() !== 'specific'));
});

$('[name=dev-exp-sel]').change(function(e) {
    const self = $(e.target);
    // Enable/disable appropiate textboxes
    if (self.value() === 'specific' ||
        self.value() === 'daily' ||
        self.value() === 'duration') {
        $('[name=device-expiration]').prop('disabled', false);
    } else {
        $('[name=device-expiration]').prop('disabled', true);
    }

    // Zero field by default
    $('[name=device-expiration]').value('');

    // Fill in textbox and tooltip as needed
    switch (self.value()) {
        case 'specific':
            setTextboxToToday('[name=device-expiration]');
            setExpirationToolTop('(YYYY-MM-DD HH:mm)');
            break;
        case 'duration':
            setExpirationToolTop('(5h30m = 5 hours and 30 minutes)');
            break;
        case 'daily':
            setExpirationToolTop('(HH:mm)');
            break;
        default:
            setExpirationToolTop('');
    }
});

$('[name=val-bef-sel]').change(function(e) {
    const self = $(e.target);
    $('[name=valid-before]').prop('disabled', (self.value() === 'forever'));
    $('[name=valid-after]').prop('disabled', (self.value() === 'forever'));

    if (self.value() === 'specific') {
        setTextboxToToday('[name=valid-before]');
        setTextboxToToday('[name=valid-after]');
        setUserExpirationToolTip('(YYYY-MM-DD HH:mm)');
    } else {
        $('[name=valid-before]').value('');
        $('[name=valid-after]').value('');
        setUserExpirationToolTip('');
    }
});

$('[name=delete-btn]').click(function() {
    api.deleteUser($('[name=username]').value(), function(resp, req) {
        if (req.status > 204) {
            resp = JSON.parse(resp);
            flashMessage(resp.Message);
            return;
        }
        location.href = '/admin/users';
    });
});

// Form submittion
$('#user-form').submit(function(e) {
    e.preventDefault();
    const formData = {
        'username': $('[name=username]').value(),
        'password': $('[name=password]').value(),
        'device_limit': '',
        'expiration_type': '',
        'device_expiration': $('[name=device-expiration]').value(),
        'valid_start': 0,
        'valid_end': 0,
        'can_manage': $('[name=can-manage]').prop('checked') ? 1 : 0,
        'can_autoreg': $('[name=can-autoreg]').prop('checked') ? 1 : 0,
        "allow_status_api": $('[name=user-api-status]').prop('checked') ? 1 : 0,
        "ui_group": $('[name=user-ui-group]').value(),
        "api_group": $('[name=user-api-group]').value()
    };

    if ($('[name=clear-pass]').prop('checked')) {
        formData.password = -1;
    }

    var devLimit = $('[name=special-limit]').value();
    if (devLimit === 'global') {
        formData.device_limit = -1;
    } else if (devLimit === 'unlimited') {
        formData.device_limit = 0;
    } else {
        formData.device_limit = $('[name=device-limit]').value();
    }

    var devExpSel = $('[name=dev-exp-sel]').value();
    formData.expiration_type = devExpirationTypes[devExpSel];

    if ($('[name=val-bef-sel]').value() !== 'forever') {
        formData.valid_start = $('[name=valid-after]').value();
        formData.valid_end = $('[name=valid-before]').value();
    }

    api.saveUser(formData, function(resp, req) {
        window.scroll(0, 0);
        if (req.status > 204) {
            resp = JSON.parse(resp);
            flashMessage(resp.Message);
            return;
        }

        flashMessage('User saved', 'success');
        $('[name=password]').value('');
        $('[name=clear-pass]').prop('checked', false);
        if (formData.password === -1 || formData.password === '') {
            $('#has-password').text('No');
        } else {
            $('#has-password').text('Yes');
        }
        $('#submit-btn').text('Save');
    }, function(req) {
        var resp = JSON.parse(req.responseText);
        flashMessage(resp.Message);
    });
});

// Utility functions
function setExpirationToolTop(tip) {
    $('#dev-exp-tooltip').text(tip);
}

function setUserExpirationToolTip(tip) {
    $('#user-exp-tooltip').text(tip);
}
