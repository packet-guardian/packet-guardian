function setSrcQuery(e: HTMLAudioElement | HTMLImageElement, q: string) {
    let src = e.src;
    const p = src.indexOf("?");
    if (p >= 0) {
        src = src.substr(0, p);
    }
    e.src = `${src}?${q}`;
}

function playCaptchaAudio() {
    const e = document.getElementById("captchaAudio") as HTMLAudioElement;
    setSrcQuery(e, "lang=en");
    e.style.display = "block";
    e.autoplay = true;
    return false;
}

function reloadCaptcha() {
    setSrcQuery(
        document.getElementById("captchaImage") as HTMLImageElement,
        "reload=" + new Date().getTime()
    );
    setSrcQuery(
        document.getElementById("captchaAudio") as HTMLAudioElement,
        new Date().getTime().toString()
    );
    return false;
}

export { playCaptchaAudio, reloadCaptcha };
