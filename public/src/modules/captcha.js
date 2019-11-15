function setSrcQuery(e, q) {
  let src = e.src;
  const p = src.indexOf("?");
  if (p >= 0) {
    src = src.substr(0, p);
  }
  e.src = `${src}?${q}`;
}

function playCaptchaAudio() {
  const e = document.getElementById("captchaAudio");
  setSrcQuery(e, "lang=en");
  e.style.display = "block";
  e.autoplay = "true";
  return false;
}

function reloadCaptcha() {
  setSrcQuery(
    document.getElementById("captchaImage"),
    "reload=" + new Date().getTime()
  );
  setSrcQuery(document.getElementById("captchaAudio"), new Date().getTime());
  return false;
}

export { playCaptchaAudio, reloadCaptcha };
