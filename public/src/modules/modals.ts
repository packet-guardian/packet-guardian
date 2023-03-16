import $ from "@/jlib2";

class ModalOverlay {
    private o: any;

    constructor() {
        this.o = $(document.createElement("div"));
        this.o.addClass("js-modal-overlay");
    }

    show() {
        document.body.insertBefore(
            this.o.elements[0],
            document.body.firstChild
        );
        this.o.show();
    }

    hide() {
        this.o.hide();
        document.body.removeChild(this.o.elements[0]);
    }
}

class ModalAlert {
    private _okCallback: (() => void) | null = null;
    private _overlay: ModalOverlay = new ModalOverlay();
    private _container: HTMLDivElement | null = null;

    show(dialog: string, callback?: () => void) {
        if (callback) {
            this._okCallback = callback;
        }
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
        if (this._okCallback) {
            this._okCallback();
        }
        document.body.removeChild(
            document.getElementsByClassName("js-modal")[0]
        );
    }
}

class ModalConfirm {
    private _okCallback: (() => void) | null = null;
    private _cancelCallback: (() => void) | null = null;
    private _overlay: ModalOverlay = new ModalOverlay();
    private _container: HTMLDivElement | null = null;

    show(dialog: string, okCallback?: () => void, cnlCallback?: () => void) {
        if (okCallback) {
            this._okCallback = okCallback;
        }
        if (cnlCallback) {
            this._cancelCallback = cnlCallback;
        }
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
        if (this._okCallback) {
            this._okCallback();
        }
        document.body.removeChild(
            document.getElementsByClassName("js-modal")[0]
        );
    }

    cancel() {
        $(this._container).hide();
        this._overlay.hide();
        if (this._cancelCallback) {
            this._cancelCallback();
        }
        document.body.removeChild(
            document.getElementsByClassName("js-modal")[0]
        );
    }
}

class ModalPrompt {
    private _okCallback: ((input: string) => void) | null = null;
    private _cancelCallback: (() => void) | null = null;
    private _overlay: ModalOverlay = new ModalOverlay();
    private _container: HTMLDivElement | null = null;

    show(
        dialog: string,
        value: string | ((input: string) => void),
        okkCallback?: (input: string) => void | (() => void),
        cnlCallback?: () => void
    ) {
        if (typeof value === "function") {
            // Shift variables
            cnlCallback = okkCallback as () => void;
            okkCallback = value;
            value = "";
        }
        if (okkCallback) {
            this._okCallback = okkCallback;
        }
        if (okkCallback) {
            this._okCallback = okkCallback;
        }
        this._okCallback = okkCallback || function () {};
        this._cancelCallback = cnlCallback || function () {};
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
        $("#js-modal-prompt-input").focus();
        bindMouseMove("#js-modal-header-id", "#js-modal-container");
    }

    // This is called when a user presses enter in the input box
    // It stops the form submittion and calls the ok function
    enter(e: Event) {
        e.preventDefault();
        e.stopPropagation();
        this.ok();
        return false;
    }

    ok() {
        $(this._container).hide();
        this._overlay.hide();
        if (this._okCallback) {
            this._okCallback($("#js-modal-prompt-input").value() ?? "");
        }
        document.body.removeChild(
            document.getElementsByClassName("js-modal")[0]
        );
    }

    cancel() {
        $(this._container).hide();
        this._overlay.hide();
        if (this._cancelCallback) {
            this._cancelCallback();
        }
        document.body.removeChild(
            document.getElementsByClassName("js-modal")[0]
        );
    }
}

function bindMouseMove(binderID: string, movediv: string) {
    const binder = $(binderID);
    binder.on("mousedown", function (e) {
        if ((<MouseEvent>e).which !== 1) {
            return;
        }
        const self = $(<HTMLElement>e.target);
        const mover = $(movediv);
        self.style("position", "relative");

        document.onmousemove = function (e) {
            mover.style("left", e.pageX - 250 + "px");
            mover.style("top", e.pageY - 10 + "px");
        };
    });
    binder.on("mouseup", () => (document.onmousemove = null));

    binder.on("dragstart", () => false);
}

export { ModalOverlay, ModalAlert, ModalConfirm, ModalPrompt };
