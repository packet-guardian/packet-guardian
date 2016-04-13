/* exported c */
var c = {
    flashTimeout: 0,

    FlashMessage: function (text, type) {
        var flash = j.$('#flashDiv'),
            flashClass = (type === 'success') ? 'flashSuccess' : 'flashFailure';

        // Post is a callback which is called after the message has faded out
        var clear = function(post) {
            post = (post !== undefined && post !== null) ? post : j.Noop;
            j.FadeOut(flash, 500, function() {
                j.RemoveClass(flash, 'flashSuccess');
                j.RemoveClass(flash, 'flashFailure');
                j.$('#flashText').innerHTML = "";
                c.flashTimeout = 0;
                post();
            });
        };

        var show = function() {
            j.$('#flashText').innerHTML = text;
            j.AddClass(flash, flashClass);
            j.FadeIn(flash, 500);
            c.flashTimeout = setTimeout(clear, 10000);
        };

        if (c.flashTimeout) {
            clearTimeout(c.flashTimeout);
            clear(show);
            return;
        }
        show(j.Noop);
    },
};

j.OnReady(function() {
    var flashMsg = j.$('#flashText').innerHTML;
    if (flashMsg !== '') {
        c.FlashMessage(flashMsg);
    }
});
