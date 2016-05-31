function jsOverlay() {
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

function jsAlert(){
    var cb = function(){};
    var overlay = new jsOverlay();
    var container = null;

	this.show = function(dialog, callback){
        cb = (callback) ? callback : function(){};
		var winW = window.innerWidth;
	    var winH = window.innerHeight;
	    var dialogbox = document.getElementById('dialogbox');

        container = document.createElement("div");

        var header = document.createElement("div");
        header.innerHTML = "Alert";
        $(header).addClass("js-modal-header");
        container.appendChild(header);


        var body = document.createElement("div");
        body.innerHTML = dialog;
        $(body).addClass("js-modal-body");
        container.appendChild(body);

        var footer = document.createElement("div");
        $(footer).addClass("js-modal-footer");
        var okButton = document.createElement("button");
        okButton.innerHTML = "OK";
        $(okButton).click(this.ok.bind(this));
        footer.appendChild(okButton);
        container.appendChild(footer);

        $(container).addClass("js-modal");
        document.body.insertBefore(container, document.body.firstChild);
        $(container).style("left", (winW/2) - (500 * 0.5)+"px");
        $(container).style("top", (winH/2) - (250 * 0.5)+"px");

        overlay.show();
        $(container).show();
	};

	this.ok = function(){
        $(container).hide();
        overlay.hide();
        cb();
        document.body.removeChild(document.getElementsByClassName("js-modal")[0]);
	};
}

function jsConfirm(){
    var okCallback = function(){};
    var cancelCallback = function(){};
    var overlay = new jsOverlay();
    var container = null;

    this.show = function(dialog, okCallback, cnlCallback){
        okCallback = (okCallback) ? okCallback : function(){};
        cancelCallback = (cnlCallback) ? cnlCallback : function(){};
        var winW = window.innerWidth;
        var winH = window.innerHeight;
        var dialogbox = document.getElementById('dialogbox');

        container = document.createElement("div");

        var header = document.createElement("div");
        header.innerHTML = "Confirm";
        $(header).addClass("js-modal-header");
        container.appendChild(header);


        var body = document.createElement("div");
        body.innerHTML = dialog;
        $(body).addClass("js-modal-body");
        container.appendChild(body);

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
        container.appendChild(footer);

        $(container).addClass("js-modal");
        document.body.insertBefore(container, document.body.firstChild);
        $(container).style("left", (winW/2) - (500 * 0.5)+"px");
        $(container).style("top", (winH/2) - (250 * 0.5)+"px");

        this._overlay = new jsOverlay();
        overlay.show();
        $(container).show();
    };

    this.ok = function(){
        $(container).hide();
        overlay.hide();
        okCallback();
        document.body.removeChild(document.getElementsByClassName("js-modal")[0]);
    };

    this.cancel = function(){
        $(container).hide();
        overlay.hide();
        cancelCallback();
        document.body.removeChild(document.getElementsByClassName("js-modal")[0]);
    };
}

function jsPrompt(){
    var okCallback = function(){};
    var cancelCallback = function(){};
    var overlay = new jsOverlay();
    var container = null;

    this.show = function(dialog, okkCallback, cnlCallback){
        okCallback = (okkCallback) ? okkCallback : function(){};
        cancelCallback = (cnlCallback) ? cnlCallback : function(){};
        var winW = window.innerWidth;
        var winH = window.innerHeight;
        var dialogbox = document.getElementById('dialogbox');

        container = document.createElement("div");

        var header = document.createElement("div");
        header.innerHTML = "Prompt";
        $(header).addClass("js-modal-header");
        container.appendChild(header);


        var body = document.createElement("div");
        body.innerHTML = dialog;
        body.innerHTML += "<form><input type=\"text\" id=\"js-modal-prompt-input\" size=\"50\"></form>";
        $(body).addClass("js-modal-body");
        container.appendChild(body);

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
        container.appendChild(footer);

        $(container).addClass("js-modal");
        document.body.insertBefore(container, document.body.firstChild);
        $(container).style("left", (winW/2) - (500 * 0.5)+"px");
        $(container).style("top", (winH/2) - (250 * 0.5)+"px");

        overlay.show();
        $(container).show();
    };

    this.ok = function(){
        $(container).hide();
        overlay.hide();
        okCallback($("#js-modal-prompt-input").value());
        document.body.removeChild(document.getElementsByClassName("js-modal")[0]);
    };

    this.cancel = function(){
        $(container).hide();
        overlay.hide();
        cancelCallback();
        document.body.removeChild(document.getElementsByClassName("js-modal")[0]);
    };
}
