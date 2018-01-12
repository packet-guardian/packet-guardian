import $ from 'jLib';

let flashTimeout = 0;

function flashMessage(text, type) {
    'use strict';
    const flash = $('#flashDiv');
    const flashClass = (type === 'success') ? 'flashSuccess' : 'flashFailure';

    // Post is a callback which is called after the message has faded out
    function clear(post) {
        post = (post) ? post : $.noop;
        flash.fadeOut(500, function() {
            flash.removeClass('flashSuccess');
            flash.removeClass('flashFailure');
            $('#flashText').html('');
            flashTimeout = 0;
            post();
        });
    }

    function show() {
        $('#flashText').html(text);
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
    'use strict';
    var flashMsg = $(element).html();
    if (flashMsg !== '') {
        flashMessage(flashMsg);
    }
}

const defaultFlashElem = '#flashText';

checkAndFlash(defaultFlashElem);

export { flashMessage, checkAndFlash };
export default flashMessage;
