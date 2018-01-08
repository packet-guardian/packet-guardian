function setTextboxToToday(el) {
    'use strict';
    var date = new Date();
    var dateStr = date.getFullYear() + '-' +
        ('0' + (date.getMonth() + 1)).slice(-2) + '-' +
        ('0' + date.getDate()).slice(-2);

    var timeStr = ('0' + date.getHours()).slice(-2) + ':' +
        ('0' + (date.getMinutes())).slice(-2);
    document.querySelector(el).value = `${dateStr} ${timeStr}`;
}

export { setTextboxToToday };
