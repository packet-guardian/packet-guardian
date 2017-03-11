// This source file is part of the Packet Guardian project.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
/* eslint-disable max-statements */
function jsOverlay() {
    "use strict";
    var o = $(document.createElement("div"));
    o.addClass("js-modal-overlay");
    this.show = function() {
        document.body.insertBefore(o[0], document.body.firstChild);
        o.show();
    };

    this.hide = function() {
        o.hide();
        document.body.removeChild(o[0]);
    };
}

function jsAlert() {
    "use strict";
    this._okCallback = function() { };
    this._overlay = new jsOverlay();
    this._container = null;

    this.show = function(dialog, callback) {
        this._okCallback = (callback) ? callback : function() { };
        var winW = window.innerWidth;
        var winH = window.innerHeight;
        var dialogbox = document.getElementById('dialogbox');

        this._container = document.createElement("div");

        var header = document.createElement("div");
        header.innerHTML = "Alert";
        $(header).addClass("js-modal-header");
        $(header).prop("id", "js-modal-header-id");
        this._container.appendChild(header);


        var body = document.createElement("div");
        body.innerHTML = dialog;
        $(body).addClass("js-modal-body");
        this._container.appendChild(body);

        var footer = document.createElement("div");
        $(footer).addClass("js-modal-footer");
        var okButton = document.createElement("button");
        okButton.innerHTML = "OK";
        $(okButton).click(this.ok.bind(this));
        footer.appendChild(okButton);
        this._container.appendChild(footer);

        var jContainer = $(this._container);
        jContainer.addClass("js-modal");
        document.body.insertBefore(this._container, document.body.firstChild);
        jContainer.style("left", (winW / 2) - (500 * 0.5) + "px");
        jContainer.style("top", (winH / 2) - (250 * 0.5) + "px");

        this._overlay = new jsOverlay();
        this._overlay.show();
        jContainer.show();
    };

    this.ok = function() {
        $(this._container).hide();
        this._overlay.hide();
        this._okCallback();
        document.body.removeChild(document.getElementsByClassName("js-modal")[0]);
    };
}

function jsConfirm() {
    "use strict";
    this._okCallback = function() { };
    this._cancelCallback = function() { };
    this._overlay = null;
    this._container = null;

    this.show = function(dialog, okCallback, cnlCallback) {
        this._okCallback = (okCallback) ? okCallback : function() { };
        this._cancelCallback = (cnlCallback) ? cnlCallback : function() { };
        var winW = window.innerWidth;
        var winH = window.innerHeight;
        var dialogbox = document.getElementById('dialogbox');

        this._container = document.createElement("div");

        var header = document.createElement("div");
        header.innerHTML = "Confirm";
        $(header).addClass("js-modal-header");
        $(header).prop("id", "js-modal-header-id");
        this._container.appendChild(header);


        var body = document.createElement("div");
        body.innerHTML = dialog;
        $(body).addClass("js-modal-body");
        this._container.appendChild(body);

        var footer = document.createElement("div");
        $(footer).addClass("js-modal-footer");
        var okButton = document.createElement("button");
        okButton.innerHTML = "OK";
        $(okButton).click(this.ok.bind(this));
        var cnlButton = document.createElement("button");
        cnlButton.innerHTML = "Cancel";
        $(cnlButton).click(this.cancel.bind(this));
        footer.appendChild(okButton);
        footer.appendChild(cnlButton);
        this._container.appendChild(footer);

        var jContainer = $(this._container);
        jContainer.addClass("js-modal");
        document.body.insertBefore(this._container, document.body.firstChild);
        jContainer.style("left", (winW / 2) - (500 * 0.5) + "px");
        jContainer.style("top", (winH / 2) - (250 * 0.5) + "px");

        this._overlay = new jsOverlay();
        this._overlay.show();
        jContainer.show();
    };

    this.ok = function() {
        $(this._container).hide();
        this._overlay.hide();
        this._okCallback();
        document.body.removeChild(document.getElementsByClassName("js-modal")[0]);
    };

    this.cancel = function() {
        $(this._container).hide();
        this._overlay.hide();
        this._cancelCallback();
        document.body.removeChild(document.getElementsByClassName("js-modal")[0]);
    };
}

function jsPrompt() {
    "use strict";
    this._okCallback = function() { };
    this._cancelCallback = function() { };
    this._overlay = new jsOverlay();
    this._container = null;

    this.show = function(dialog, value, okkCallback, cnlCallback) {
        if (typeof value === "function") {
            // Shift variables
            cnlCallback = okkCallback;
            okkCallback = value;
            value = "";
        }
        this._okCallback = (okkCallback) ? okkCallback : function() { };
        this._cancelCallback = (cnlCallback) ? cnlCallback : function() { };
        var winW = window.innerWidth;
        var winH = window.innerHeight;
        var dialogbox = document.getElementById('dialogbox');

        this._container = document.createElement("div");

        var header = document.createElement("div");
        header.innerHTML = "Prompt";
        $(header).addClass("js-modal-header");
        $(header).addClass("grabbable");
        $(header).prop("id", "js-modal-header-id");
        this._container.appendChild(header);

        // Create dialog body
        var body = document.createElement("div");
        $(body).addClass("js-modal-body");
        body.innerHTML = dialog;
        // Create the main form
        var form = document.createElement("form");
        $(form).submit(this.enter.bind(this));
        // Create the input for user data
        var input = document.createElement("input");
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

        var footer = document.createElement("div");
        $(footer).addClass("js-modal-footer");
        var okButton = document.createElement("button");
        okButton.innerHTML = "OK";
        $(okButton).click(this.ok.bind(this));
        var cnlButton = document.createElement("button");
        cnlButton.innerHTML = "Cancel";
        $(cnlButton).click(this.cancel.bind(this));
        footer.appendChild(okButton);
        footer.appendChild(cnlButton);
        this._container.appendChild(footer);

        var jContainer = $(this._container);
        jContainer.addClass("js-modal");
        jContainer.prop("id", "js-modal-container");
        document.body.insertBefore(this._container, document.body.firstChild);
        jContainer.style("left", (winW / 2) - (500 * 0.5) + "px");
        jContainer.style("top", (winH / 2) - (250 * 0.5) + "px");

        this._overlay = new jsOverlay();
        this._overlay.show();
        jContainer.show();
        document.getElementById("js-modal-prompt-input").focus();
        bindMouseMove("#js-modal-header-id", "#js-modal-container");
    };

    // This is called when a user presses enter in the input box
    // It stops the form submittion and calls the ok function
    this.enter = function(e) {
        e.preventDefault();
        e.stopPropagation();
        this.ok();
        return false;
    };

    this.ok = function() {
        $(this._container).hide();
        this._overlay.hide();
        this._okCallback($("#js-modal-prompt-input").value());
        document.body.removeChild(document.getElementsByClassName("js-modal")[0]);
    };

    this.cancel = function() {
        $(this._container).hide();
        this._overlay.hide();
        this._cancelCallback();
        document.body.removeChild(document.getElementsByClassName("js-modal")[0]);
    };
}

function bindMouseMove(binderID, movediv) {
    "use strict";
    var binder = $(binderID);
    binder.on("mousedown", function(e) {
        if (e.which !== 1) { return; }
        var self = $(e.target);
        var mover = $(movediv);
        self.style("position", "relative");

        document.onmousemove = function(e) {
            mover.style("left", e.pageX - 250 + 'px');
            mover.style("top", e.pageY - 10 + 'px');
        };
    });
    binder.on("mouseup", function() {
        document.onmousemove = null;
    });

    binder.on("dragstart", function() { return false; });
}
