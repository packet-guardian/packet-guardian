import $ from "jLib";
import api from "pg-api";
import flashMessage from "flash";
import "manage";
import { ModalPrompt, ModalConfirm } from "modals";

// Event handlers
$("[name=blacklist-sel]").change(function(e) {
  var self = $(e.target);
  var cmodal = new ModalConfirm();
  switch (self.value()) {
    case "username":
      var isBl = self.data("blacklisted") === "true";
      if (isBl) {
        cmodal.show("Remove username from blacklist?", function() {
          blacklistUsername(true);
        });
      } else {
        cmodal.show("Add username to blacklist?", function() {
          blacklistUsername(false);
        });
      }
      break;
    case "black-all":
      cmodal.show(
        "Add all user's devices to blacklist?",
        addDevicesToBlacklist
      );
      break;
    case "unblack-all":
      cmodal.show(
        "Remove all user's devices from blacklist?",
        removeDevicesFromBlacklist
      );
      break;
    case "black-sel":
      cmodal.show("Add selected user's devices to blacklist?", function() {
        blacklistSelectedDevices(true);
      });
      break;
    case "unblack-sel":
      cmodal.show("Remove selected user's devices from blacklist?", function() {
        blacklistSelectedDevices(false);
      });
      break;
  }
  self.value("");
});

$("[name=reassign-selected-btn]").click(function() {
  var pmodal = new ModalPrompt();
  pmodal.show("New owner's username:", reassignSelectedDevices);
});

// Event callbacks
var blacklistSelect = $("[name=blacklist-sel]");
if (blacklistSelect.length !== 0) {
  if (blacklistSelect.data("blacklisted") === "true") {
    $("[name=black-user-option]").text("Remove User");
  }
}

function getUsername(encode) {
  var u = $("[name=username]").value();
  if (encode) {
    return encodeURIComponent(u);
  }
  return u;
}

function blacklistUsername(isBlacklisted) {
  var success = function() {
    location.reload();
  };
  var error = function() {
    flashMessage("Error blacklisting user");
  };

  if (isBlacklisted) {
    api.unblacklistUser(getUsername(), success, error);
  } else {
    api.blacklistUser(getUsername(), success, error);
  }
}

function getCheckedDevices() {
  var checked = $(".device-checkbox:checked");
  var devices = [];
  for (var i = 0; i < checked.length; i++) {
    devices.push(checked[i].value);
  }
  return devices;
}

function blacklistSelectedDevices(add) {
  var devicesToRemove = getCheckedDevices();
  if (devicesToRemove.length === 0) {
    return;
  }
  blacklistDevices(devicesToRemove, add);
}

function blacklistDevices(devices, add) {
  if (add) {
    addDevicesToBlacklist(devices);
  } else {
    removeDevicesFromBlacklist(devices);
  }
}

function addDevicesToBlacklist(devices) {
  if (devices) {
    api.blacklistDevices(devices, reloadPage, errorAdding);
  } else {
    api.blacklistAllDevices(getUsername(), reloadPage, errorAdding);
  }
}

function removeDevicesFromBlacklist(devices) {
  if (devices) {
    api.unblacklistDevices(devices, reloadPage, errorRemoving);
  } else {
    api.unblacklistAllDevices(getUsername(), reloadPage, errorRemoving);
  }
}

function reassignSelectedDevices(username) {
  var devices = getCheckedDevices();
  if (devices.length === 0 || !username) {
    return;
  }

  api.reassignDevices(username, devices, reloadPage, () =>
    flashMessage("Error reassigning devices")
  );
}

function reloadPage() {
  location.reload();
}

function errorAdding() {
  flashMessage("Error blacklisting devices");
}

function errorRemoving() {
  flashMessage("Error removing devices from blacklist");
}
