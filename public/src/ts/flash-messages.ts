import $ from "@/jlib2";
import flashMessage from "@/flash";

function checkAndFlash(element: string) {
    const flashElm = $(element);
    if (flashElm.length === 0) {
        return;
    }

    const flashMsg = flashElm.html() ?? "";
    const flashType = flashElm.data("flashtype") ?? "";
    if (flashMsg !== "") {
        flashMessage(flashMsg, flashType);
    }
}

const defaultFlashElem = "#flash-text";

checkAndFlash(defaultFlashElem);
