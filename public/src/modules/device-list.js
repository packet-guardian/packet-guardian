import $ from "jLib";

let stayOpen = false;

function deviceListInit() {
  let reallyOpen = false;
  $(".device-header").click(function(e) {
    if (!reallyOpen) {
      return;
    }
    let self = $(e.target);
    if (self.hasClass("device-checkbox")) {
      return;
    }
    if (self[0].tagName === "A") {
      return;
    }

    while (!self.hasClass("device-header")) {
      self = $(self[0].parentNode);
    }
    const bodyNum = self.data("deviceId");
    expandBody(bodyNum);
  });

  function expandBody(bodyNum) {
    const thebody = $(`#device-body-${bodyNum}`);
    // Get the max-height value before setting it back to 0
    const mh = thebody.style("max-height");
    if (!stayOpen) {
      // Close all
      $(".device-body").style("max-height", "0px");
    }

    const newMaxHeight = mh === "1000px" ? "0px" : "1000px";
    thebody.style("max-height", newMaxHeight);
  }

  const preOpenID = location.hash.substring(1);
  if (preOpenID) {
    expandBody(preOpenID);
  }

  $(".device-header").on("mousedown", () => {
    reallyOpen = true;
  });
  $(".device-header").on("mousemove", () => {
    reallyOpen = false;
  });
}

function keepDevicesOpen(stay) {
  stayOpen = stay;
}

deviceListInit();

export { keepDevicesOpen };
