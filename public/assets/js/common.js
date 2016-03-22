/* exported c */
var c = {
    FlashMessage: function (text) {
        var flash = j.$('#flashDiv');
        j.$('#flashText').innerHTML = "Incorrect username or password";
        j.AddClass(flash, 'flashFailure');
        j.FadeIn(flash, 500);

        var clear = function() {
            j.FadeOut(flash, 500, function() {
                j.RemoveClass(flash, 'flashFailure');
                j.$('#flashText').innerHTML = "";
            });
        };
        setTimeout(clear, 10000);
    },
};
