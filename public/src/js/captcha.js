import { playCaptchaAudio, reloadCaptcha } from '../modules/captcha';
import { $ } from '../modules/jLib';
import { checkAndFlashDefault } from '../modules/flash';

checkAndFlashDefault();

$('#reload-captcha-btn').click(reloadCaptcha);
$('#play-captcha-btn').click(playCaptchaAudio);
