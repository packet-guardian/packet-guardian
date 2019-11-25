import $ from "jlib2";

let flashTimeout = 0;

function flashMessage(text, type) {
  const flash = $("#flashDiv");
  const flashClass = type === "success" ? "flashSuccess" : "flashFailure";

  // Post is a callback which is called after the message has faded out
  function clear(post) {
    flash.fadeOut(500, function() {
      flash.removeClass("flashSuccess");
      flash.removeClass("flashFailure");
      $("#flashText").html("");
      flashTimeout = 0;
      if (post) {
        post();
      }
    });
  }

  function show() {
    $("#flashText").html(text);
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

function checkAndFlash(element) {
  const flashElm = $(element);
  if (flashElm.length === 0) {
    return;
  }

  var flashMsg = flashElm.html();
  if (flashMsg !== "") {
    flashMessage(flashMsg);
  }
}

const defaultFlashElem = "#flashText";

checkAndFlash(defaultFlashElem);

export { flashMessage, checkAndFlash };
export default flashMessage;
