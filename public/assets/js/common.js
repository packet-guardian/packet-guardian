/* exported c */
var c = {
    flashTimeout: 0,

    FlashMessage: function (text) {
        clearTimeout(c.flashTimeout);
        var flash = j.$('#flashDiv');
        j.$('#flashText').innerHTML = text;
        j.AddClass(flash, 'flashFailure');
        j.FadeIn(flash, 500);

        var clear = function() {
            j.FadeOut(flash, 500, function() {
                j.RemoveClass(flash, 'flashFailure');
                j.$('#flashText').innerHTML = "";
            });
        };
        c.flashTimeout = setTimeout(clear, 10000);
    },
};
