import $ from "@/jlib2";

let flashTimeout = 0;

function flashMessage(text: string, type = "") {
    const flash = $("#flashDiv");
    const flashClass = type === "success" ? "flash-success" : "flash-failure";

    // Post is a callback which is called after the message has faded out
    function clear(post: () => void) {
        flash.fadeOut(500, function () {
            flash.removeClass("flash-success");
            flash.removeClass("flash-failure");
            $("#flash-text").html("");
            flashTimeout = 0;
            if (post) {
                post();
            }
        });
    }

    function show() {
        $("#flash-text").html(text);
        flash.addClass(flashClass);
        flash.fadeIn(500);
        flashTimeout = setTimeout(clear, 10000);
    }

    if (flashTimeout) {
        clearTimeout(flashTimeout);
        clear(show);
        return;
    }
    show();
}

export default flashMessage;
