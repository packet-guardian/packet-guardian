import $ from "@/jlib2";
import "@/flash";
import { playCaptchaAudio, reloadCaptcha } from "@/captcha";

$("#reload-captcha-btn").click(reloadCaptcha);
$("#play-captcha-btn").click(playCaptchaAudio);
