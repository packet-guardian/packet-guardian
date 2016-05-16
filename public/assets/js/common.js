/* exported c */
var c = {
    flashTimeout: 0,

    FlashMessage: function (text, type) {
        var flash = $('#flashDiv'),
            flashClass = (type === 'success') ? 'flashSuccess' : 'flashFailure';

        // Post is a callback which is called after the message has faded out
        var clear = function(post) {
            post = (post) ? post : $.noop;
            flash.fadeOut(500, function() {
                flash.removeClass('flashSuccess');
                flash.removeClass('flashFailure');
                $('#flashText').html("");
                c.flashTimeout = 0;
                post();
            });
        };

        var show = function() {
            $('#flashText').html(text);
            flash.addClass(flashClass);
            flash.fadeIn(500);
            c.flashTimeout = setTimeout(clear, 10000);
        };

        if (c.flashTimeout) {
            clearTimeout(c.flashTimeout);
            clear(show);
            return;
        }
        show();
    },

    BindSelectAll: function(target, sClass) {
        target = $(target);
        target.click(function() {
            $(sClass).prop("checked", target.prop("checked"));
        });
    },
};

$.onReady(function() {
    var flashMsg = $('#flashText').html();
    if (flashMsg !== '') {
        c.FlashMessage(flashMsg);
    }
});
