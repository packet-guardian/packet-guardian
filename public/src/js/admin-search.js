import { deviceListInit, keepDevicesOpen } from '../modules/device-list';
import { checkAndFlashDefault } from '../modules/flash';

checkAndFlashDefault();
keepDevicesOpen(true);
deviceListInit();
