import $ from "jlib2";
import api from "pg-api";
import flashMessage from "flash";
import "manage";
import { ModalPrompt, ModalConfirm } from "modals";

// Event handlers
$("[name=blacklist-sel]").change(e => {
  const self = $(e.target);
  const cmodal = new ModalConfirm();

  switch (self.value()) {
    case "username":
      if (self.data("blacklisted") === "true") {
        cmodal.show("Remove username from blacklist?", removeUsernameBlacklist);
      } else {
        cmodal.show("Add username to blacklist?", addUsernameBlacklist);
      }
      break;
    case "black-sel":
      cmodal.show("Add selected user's devices to blacklist?", addToBlacklist);
      break;
    case "unblack-sel":
      cmodal.show(
        "Remove selected user's devices from blacklist?",
        removeFromBlacklist
      );
      break;
  }

  self.value("");
});

$("[name=reassign-selected-btn]").click(() =>
  new ModalPrompt().show("New owner's username:", reassignSelectedDevices)
);

// Event callbacks
const blacklistSelect = $("[name=blacklist-sel]");
if (blacklistSelect.length !== 0) {
  if (blacklistSelect.data("blacklisted") === "true") {
    $("[name=black-user-option]").text("Remove User");
  }
}

const getUsername = () => $("[name=username]").value();

const removeUsernameBlacklist = () =>
  api.unblacklistUser(getUsername(), reloadPage, () =>
    flashMessage("Error removing user from blacklist")
  );

const addUsernameBlacklist = () =>
  api.blacklistUser(getUsername(), reloadPage, () =>
    flashMessage("Error adding user to blacklist")
  );

const getCheckedDevices = () =>
  $(".device-checkbox:checked").map(elem => elem.value);

function addToBlacklist() {
  const devicesToRemove = getCheckedDevices();
  if (devicesToRemove.length > 0) {
    api.blacklistDevices(devicesToRemove, reloadPage, errorAdding);
  }
}

function removeFromBlacklist() {
  const devicesToRemove = getCheckedDevices();
  if (devicesToRemove.length > 0) {
    api.unblacklistDevices(devicesToRemove, reloadPage, errorRemoving);
  }
}

function reassignSelectedDevices(username) {
  const devices = getCheckedDevices();
  if (devices.length > 0 && username) {
    api.reassignDevices(username, devices, reloadPage, () =>
      flashMessage("Error reassigning devices")
    );
  }
}

const reloadPage = () => location.reload();

const errorAdding = () => flashMessage("Error blacklisting devices");

const errorRemoving = () =>
  flashMessage("Error removing devices from blacklist");
