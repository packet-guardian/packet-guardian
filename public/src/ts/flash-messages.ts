import $ from "@/jlib2";
import flashMessage from "@/flash";

function checkAndFlash(element: string) {
    const flashElm = $(element);
    if (flashElm.length === 0) {
        return;
    }

    const flashMsg = flashElm.html() ?? "";
    if (flashMsg !== "") {
        flashMessage(flashMsg);
    }
}

const defaultFlashElem = "#flashText";

checkAndFlash(defaultFlashElem);
