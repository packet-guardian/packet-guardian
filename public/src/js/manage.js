import { deviceListInit } from '../modules/device-list';
import { initManage } from '../modules/manage';
import { checkAndFlashDefault } from '../modules/flash';

checkAndFlashDefault();
deviceListInit();
initManage();
