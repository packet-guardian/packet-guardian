import $ from "jlib2";

/* eslint-disable max-statements */
class ModalOverlay {
  constructor() {
    this.o = $(document.createElement("div"));
    this.o.addClass("js-modal-overlay");
  }

  show() {
    document.body.insertBefore(this.o.elements[0], document.body.firstChild);
    this.o.show();
  }

  hide() {
    this.o.hide();
    document.body.removeChild(this.o.elements[0]);
  }
}

class ModalAlert {
  constructor() {
    this._okCallback = function() {};
    this._overlay = new ModalOverlay();
    this._container = null;
  }

  show(dialog, callback) {
    this._okCallback = callback || function() {};
    const winW = window.innerWidth;
    const winH = window.innerHeight;

    this._container = document.createElement("div");

    const header = document.createElement("div");
    header.innerHTML = "Alert";
    $(header).addClass("js-modal-header");
    $(header).addClass("grabbable");
    $(header).prop("id", "js-modal-header-id");
    this._container.appendChild(header);

    const body = document.createElement("div");
    body.innerHTML = dialog;
    $(body).addClass("js-modal-body");
    this._container.appendChild(body);

    const footer = document.createElement("div");
    $(footer).addClass("js-modal-footer");
    const okButton = document.createElement("button");
    okButton.innerHTML = "OK";
    $(okButton).click(this.ok.bind(this));
    footer.appendChild(okButton);
    this._container.appendChild(footer);

    const jContainer = $(this._container);
    jContainer.addClass("js-modal");
    jContainer.prop("id", "js-modal-container");
    document.body.insertBefore(this._container, document.body.firstChild);
    jContainer.style("left", winW / 2 - 500 * 0.5 + "px");
    jContainer.style("top", winH / 2 - 250 * 0.5 + "px");

    this._overlay = new ModalOverlay();
    this._overlay.show();
    jContainer.show();
    bindMouseMove("#js-modal-header-id", "#js-modal-container");
  }

  ok() {
    $(this._container).hide();
    this._overlay.hide();
    this._okCallback();
    document.body.removeChild(document.getElementsByClassName("js-modal")[0]);
  }
}

class ModalConfirm {
  constructor() {
    this._okCallback = function() {};
    this._cancelCallback = function() {};
    this._overlay = null;
    this._container = null;
  }

  show(dialog, okCallback, cnlCallback) {
    this._okCallback = okCallback || function() {};
    this._cancelCallback = cnlCallback || function() {};
    const winW = window.innerWidth;
    const winH = window.innerHeight;

    this._container = document.createElement("div");

    const header = document.createElement("div");
    header.innerHTML = "Confirm";
    $(header).addClass("js-modal-header");
    $(header).addClass("grabbable");
    $(header).prop("id", "js-modal-header-id");
    this._container.appendChild(header);

    const body = document.createElement("div");
    body.innerHTML = dialog;
    $(body).addClass("js-modal-body");
    this._container.appendChild(body);

    const footer = document.createElement("div");
    $(footer).addClass("js-modal-footer");

    const okButton = document.createElement("button");
    okButton.innerHTML = "OK";
    $(okButton).click(this.ok.bind(this));

    const cnlButton = document.createElement("button");
    cnlButton.innerHTML = "Cancel";
    $(cnlButton).click(this.cancel.bind(this));

    footer.appendChild(okButton);
    footer.appendChild(cnlButton);
    this._container.appendChild(footer);

    const jContainer = $(this._container);
    jContainer.addClass("js-modal");
    jContainer.prop("id", "js-modal-container");
    document.body.insertBefore(this._container, document.body.firstChild);
    jContainer.style("left", winW / 2 - 500 * 0.5 + "px");
    jContainer.style("top", winH / 2 - 250 * 0.5 + "px");

    this._overlay = new ModalOverlay();
    this._overlay.show();
    jContainer.show();
    bindMouseMove("#js-modal-header-id", "#js-modal-container");
  }

  ok() {
    $(this._container).hide();
    this._overlay.hide();
    this._okCallback();
    document.body.removeChild(document.getElementsByClassName("js-modal")[0]);
  }

  cancel() {
    $(this._container).hide();
    this._overlay.hide();
    this._cancelCallback();
    document.body.removeChild(document.getElementsByClassName("js-modal")[0]);
  }
}

class ModalPrompt {
  constructor() {
    this._okCallback = function() {};
    this._cancelCallback = function() {};
    this._overlay = new ModalOverlay();
    this._container = null;
  }

  show(dialog, value, okkCallback, cnlCallback) {
    if (typeof value === "function") {
      // Shift variables
      cnlCallback = okkCallback;
      okkCallback = value;
      value = "";
    }
    this._okCallback = okkCallback || function() {};
    this._cancelCallback = cnlCallback || function() {};
    const winW = window.innerWidth;
    const winH = window.innerHeight;

    this._container = document.createElement("div");

    const header = document.createElement("div");
    header.innerHTML = "Prompt";
    $(header).addClass("js-modal-header");
    $(header).addClass("grabbable");
    $(header).prop("id", "js-modal-header-id");
    this._container.appendChild(header);

    // Create dialog body
    const body = document.createElement("div");
    $(body).addClass("js-modal-body");
    body.innerHTML = dialog;
    // Create the main form
    const form = document.createElement("form");
    $(form).submit(this.enter.bind(this));
    // Create the input for user data
    const input = document.createElement("input");
    input.type = "text";
    input.id = "js-modal-prompt-input";
    input.size = 50;
    input.value = value;
    // Add input to form
    form.appendChild(input);
    // Add form to body
    body.appendChild(form);
    // Add body to container
    this._container.appendChild(body);

    const footer = document.createElement("div");
    $(footer).addClass("js-modal-footer");
    const okButton = document.createElement("button");
    okButton.innerHTML = "OK";
    $(okButton).click(this.ok.bind(this));
    const cnlButton = document.createElement("button");
    cnlButton.innerHTML = "Cancel";
    $(cnlButton).click(this.cancel.bind(this));
    footer.appendChild(okButton);
    footer.appendChild(cnlButton);
    this._container.appendChild(footer);

    const jContainer = $(this._container);
    jContainer.addClass("js-modal");
    jContainer.prop("id", "js-modal-container");
    document.body.insertBefore(this._container, document.body.firstChild);
    jContainer.style("left", winW / 2 - 500 * 0.5 + "px");
    jContainer.style("top", winH / 2 - 250 * 0.5 + "px");

    this._overlay = new ModalOverlay();
    this._overlay.show();
    jContainer.show();
    document.getElementById("js-modal-prompt-input").focus();
    bindMouseMove("#js-modal-header-id", "#js-modal-container");
  }

  // This is called when a user presses enter in the input box
  // It stops the form submittion and calls the ok function
  enter(e) {
    e.preventDefault();
    e.stopPropagation();
    this.ok();
    return false;
  }

  ok() {
    $(this._container).hide();
    this._overlay.hide();
    this._okCallback($("#js-modal-prompt-input").value());
    document.body.removeChild(document.getElementsByClassName("js-modal")[0]);
  }

  cancel() {
    $(this._container).hide();
    this._overlay.hide();
    this._cancelCallback();
    document.body.removeChild(document.getElementsByClassName("js-modal")[0]);
  }
}

function bindMouseMove(binderID, movediv) {
  const binder = $(binderID);
  binder.on("mousedown", function(e) {
    if (e.which !== 1) {
      return;
    }
    const self = $(e.target);
    const mover = $(movediv);
    self.style("position", "relative");

    document.onmousemove = function(e) {
      mover.style("left", e.pageX - 250 + "px");
      mover.style("top", e.pageY - 10 + "px");
    };
  });
  binder.on("mouseup", function() {
    document.onmousemove = null;
  });

  binder.on("dragstart", function() {
    return false;
  });
}

export { ModalOverlay, ModalAlert, ModalConfirm, ModalPrompt };
