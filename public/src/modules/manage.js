import $ from "jLib";
import api from "pg-api";
import flashMessage from "flash";
import { ModalPrompt, ModalConfirm } from "modals";

function initManage() {
  // Event handlers
  $("[name=add-device-btn]").click(function(e) {
    var isAdmin = $(e.target).data("admin");
    var user = $("[name=username]").value();
    if (isAdmin !== null) {
      location.href = `/register?manual=1&username=${user}`;
    } else {
      location.href = "/register?manual=1";
    }
  });

  // Delete buttons
  $("[name=del-selected-btn]").click(function() {
    var cmodal = new ModalConfirm();
    cmodal.show(
      "Are you sure you want to delete selected devices?",
      deleteSelectedDevices
    );
  });

  $(".device-checkbox-target").click(function(e) {
    $("#select-all-checkbox").prop("checked", false);
  });

  $(".edit-dev-desc").click(function(e) {
    e.stopPropagation();
    var id = $(e.target).data("device");
    var pmodal = new ModalPrompt();
    pmodal.show("Device Description:", $(`#device-${id}-desc`).text(), function(
      desc
    ) {
      console.dir(id);
      editDeviceDescription(id, desc);
    });
  });

  $("#select-all").click(function(e) {
    const state = !$("#select-all-checkbox").prop("checked");
    $(".device-checkbox").prop("checked", state);
  });
}

// Event callbacks
function editDeviceDescription(id, desc) {
  var mac = $(`#device-${id}-mac`)
    .text()
    .trim();
  api.saveDeviceDescription(
    mac,
    desc,
    function(resp, req) {
      $(`#device-${id}-desc`).text(desc);
      flashMessage("Device description saved", "success");
    },
    function(req) {
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
function deleteSelectedDevices() {
  var checked = $(".device-checkbox:checked");
  if (checked.length === 0) {
    return;
  }

  var devicesToRemove = [];
  for (var i = 0; i < checked.length; i++) {
    devicesToRemove.push(checked[i].value);
  }

  api.deleteDevices(
    $("[name=username]").value(),
    devicesToRemove,
    () => location.reload(),
    () => flashMessage("Error deleting devices")
  );
}

initManage();
