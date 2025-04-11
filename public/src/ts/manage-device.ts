import $ from "@/jlib2";
import api from "@/pg-api";
import flashMessage from "@/flash";
import { ModalPrompt, ModalConfirm } from "@/modals";
import { setTextboxToToday } from "@/utils";

let oldExpiration = "";

$("#delete-btn").click(() => {
    const cmodal = new ModalConfirm();
    cmodal.show("Are you sure you want to delete this device?", () =>
        api.deleteDevices(getUsername(), [getMacAddress()], reloadPage, () =>
            flashMessage("Error deleting device")
        )
    );
});

$("#unflag-dev-btn").click(() => {
    const cmodal = new ModalConfirm();
    cmodal.show("Are you sure you want to unflag this device?", () =>
        api.flagDevice(getMacAddress(), false, reloadPage, () =>
            flashMessage("Error unflagging device")
        )
    );
});

$("#flag-dev-btn").click(() => {
    const cmodal = new ModalConfirm();
    cmodal.show("Are you sure you want to flag this device?", () =>
        api.flagDevice(getMacAddress(), true, reloadPage, () =>
            flashMessage("Error flagging device")
        )
    );
});

$("#unblacklist-btn").click(() => {
    const cmodal = new ModalConfirm();
    cmodal.show(
        "Are you sure you want to remove this device from the block list?",
        () =>
            api.unblacklistDevices([getMacAddress()], reloadPage, () =>
                flashMessage("Error removing device from block list")
            )
    );
});

$("#blacklist-btn").click(() => {
    const cmodal = new ModalConfirm();
    cmodal.show("Are you sure you want to block this device?", () =>
        api.blacklistDevices([getMacAddress()], reloadPage, () =>
            flashMessage("Error blocking device")
        )
    );
});

$("#reassign-btn").click(() => {
    const pmodal = new ModalPrompt();
    pmodal.show("New owner's username:", (newUser) =>
        api.reassignDevices(newUser, [getMacAddress()], reloadPage, (req) =>
            flashMessage(JSON.parse(req.responseText).Message)
        )
    );
});

$("#register-btn").click(() => {
    const pmodal = new ModalPrompt();
    pmodal.show("Username:", (newUser) =>
        api.registerDevice(
            {
                username: newUser,
                "mac-address": getMacAddress(),
                description: "",
                platform: "",
            },
            reloadPage,
            (req) => flashMessage(JSON.parse(req.responseText).Message)
        )
    );
});

function getMacAddress() {
    return $("#mac-address").text();
}

function getUsername() {
    return $("#username").value();
}

function getDescription() {
    return $("#device-desc").text();
}

$("#edit-dev-desc").click((e) => {
    e.stopPropagation();
    const pmodal = new ModalPrompt();
    pmodal.show("Device Description:", getDescription(), editDeviceDescription);
});

$("#edit-dev-expiration").click((e) => {
    e.stopPropagation();
    oldExpiration = $("#device-expiration").text();
    $("#device-expiration").text("");
    if (oldExpiration === "Never") {
        $("#dev-exp-sel").value("never");
    } else if (oldExpiration === "Rolling") {
        $("#dev-exp-sel").value("rolling");
    } else {
        $("#dev-exp-sel").value("specific");
        $("#dev-exp-val").value(oldExpiration);
        $("#dev-exp-val").style("display", "inline");
    }
    $("#edit-dev-expiration").style("display", "none");
    $("#edit-expire-controls").style("display", "inline");
    $("#confirmation-icons").style("display", "inline");
});

$("#dev-exp-sel").change((e) => {
    if ($(e.target).value() !== "specific") {
        $("#dev-exp-val").style("display", "none");
    } else {
        setTextboxToToday("#dev-exp-val");
        $("#dev-exp-val").style("display", "inline");
    }
});

$("#dev-expiration-cancel").click((e) => {
    e.stopPropagation();
    clearExpirationControls(oldExpiration);
});

$("#dev-expiration-ok").click((e) => {
    e.stopPropagation();
    editDeviceExpiration($("#dev-exp-sel").value(), $("#dev-exp-val").value());
});

$("#edit-notes").click((e) => {
    e.stopPropagation();
    $("#notes-textarea").value($("#notes-text").text());

    $("#edit-notes").style("display", "none");
    $("#notes-text").style("display", "none");

    $("#notes-text-edit").style("display", "inline");
    $("#notes-confirmation-icons").style("display", "inline");
});

$("#notes-edit-cancel").click((e) => {
    e.stopPropagation();
    clearNotesEdit();
});

$("#notes-edit-ok").click((e) => {
    e.stopPropagation();
    editDeviceNotes($("#notes-textarea").value());
});

// Event callbacks
function clearExpirationControls(value: string) {
    $("#edit-expire-controls").style("display", "none");
    $("#confirmation-icons").style("display", "none");
    $("#device-expiration").text(value);
    $("#dev-exp-val").value("");
    $("#edit-dev-expiration").style("display", "inline");
}

function clearNotesEdit() {
    $("#edit-notes").style("display", "inline");
    $("#notes-text").style("display", "inline");

    $("#notes-text-edit").style("display", "none");
    $("#notes-confirmation-icons").style("display", "none");
}

function editDeviceDescription(desc: string) {
    api.saveDeviceDescription(
        getMacAddress(),
        desc,
        () => {
            $("#device-desc").text(desc);
            flashMessage("Device description saved", "success");
        },
        apiResponseCheck
    );
}

function editDeviceExpiration(type: string, value: string) {
    api.saveDeviceExpiration(
        getMacAddress(),
        type,
        value,
        (resp, req) => {
            clearExpirationControls(resp.Data.newExpiration);
            flashMessage("Device expiration saved", "success");
        },
        apiResponseCheck
    );
}

function editDeviceNotes(notes: string) {
    api.saveDeviceNotes(
        getMacAddress(),
        notes,
        () => {
            $("#notes-text").text(notes);
            clearNotesEdit();
            flashMessage("Device notes saved", "success");
        },
        apiResponseCheck
    );
}

function apiResponseCheck(req: XMLHttpRequest) {
    const resp = JSON.parse(req.responseText);
    switch (req.status) {
        case 500:
            flashMessage(`Internal Server Error - ${resp.Message}`);
            break;
        default:
            flashMessage(resp.Message);
            break;
    }
}

const reloadPage = () => location.reload();
