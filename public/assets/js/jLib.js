/*
 * The jLib library is a collection of commonly used functions to add quick functionality to a project.
 * This library is exremely light-weight and only adds what is necessary for convience.
 *
 * Licensed under the MIT license. Text available online: https://opensource.org/licenses/MIT
**/
/*jslint browser:true */
var j = {
    $: function (id, all) {
        'use strict';
        // If id is already a node, return it
        if (typeof id !== 'string') {
            return id;
        }

        all = all !== undefined ? all : false;

        if (all) {
            return document.querySelectorAll(id);
        }
        return document.querySelector(id);
    },

    // Document on ready
    OnReady: function (fn) {
        'use strict';
        if (document.readyState !== 'loading') {
            fn(); // DOM is already ready
        } else if (document.addEventListener) {
            // IE 9+, modern
            document.addEventListener('DOMContentLoaded', fn);
        } else {
            // IE 8
            document.attachEvent('onreadystatechange', function () {
                if (document.readyState !== 'loading') {
                    fn();
                }
            });
        }
    },

    // Event handling

    On: function (el, eventName, handler) {
        'use strict';
        el = j.$(el);
        if (el.addEventListener) {
            el.addEventListener(eventName, handler);
        } else {
            el.attachEvent('on' + eventName, function () {
                handler.call(el);
            });
        }
    },

    Off: function (el, eventName, handler) {
        'use strict';
        el = j.$(el);

        if (el.removeEventListener) {
            el.removeEventListener(eventName, handler);
        } else {
            el.detachEvent('on' + eventName, handler);
        }
    },

    Click: function (el, handler) {
        'use strict';
        j.On(el, 'click', handler);
    },

    Submit: function (el, handler) {
        'use strict';
        j.On(el, 'submit', handler);
    },

    // Class manipulation

    AddClass: function (el, className) {
        'use strict';
        el = j.$(el);
        if (el.classList) {
            el.classList.add(className);
        } else {
            el.className += ' ' + className;
        }
    },

    HasClass: function (el, className) {
        'use strict';
        el = j.$(el);
        if (el.classList) {
            el.classList.contains(className);
        } else {
            new RegExp('(^| )' + className + '( |$)', 'gi').test(el.className);
        }
    },

    RemoveClass: function (el, className) {
        'use strict';
        el = j.$(el);
        if (el.classList) {
            el.classList.remove(className);
        } else {
            el.className = el.className.replace(new RegExp('(^|\\b)' + className.split(' ').join('|') + '(\\b|$)', 'gi'), ' ');
        }
    },

    ToggleClass: function (el, className) {
        'use strict';
        el = j.$(el);
        if (el.classList) {
            el.classList.toggle(className);
        } else {
            if (j.HasClass(el, className)) {
                j.RemoveClass(el, className);
            } else {
                j.AddClass(el, className);
            }
        }
    },

    // AJAX functions
    Get: function (url, data, successFn, errorFn) {
        'use strict';
        j.Ajax('GET', url, data, successFn, errorFn, null);
    },

    Post: function (url, data, successFn, errorFn) {
        'use strict';
        var middleware = function (req) {
            req.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded; charset=UTF-8');
            return req;
        };
        j.Ajax('POST', url, data, successFn, errorFn, middleware);
    },

    Ajax: function (method, url, data, successFn, errorFn, middleware) {
        'use strict';
        method = method.toUpperCase();
        successFn = (successFn !== undefined && successFn !== null) ? successFn : function () { return; };
        errorFn = (errorFn !== undefined && errorFn !== null) ? errorFn : function () { return; };
        middleware = (middleware !== undefined && middleware !== null) ? middleware : function (r) { return r; };
        var dataStr = j.formatAJAXData(data),
            request = new XMLHttpRequest();

        if (dataStr === null) {
            console.log('Invalid data to j.Post.');
            return;
        }

        if (method === 'GET') {
            url += '?' + dataStr;
            dataStr = '';
        }

        request.open(method.toUpperCase(), url, true);
        request = middleware(request);

        request.onreadystatechange = function () {
            if (this.readyState === 4) {
                if (this.status >= 200 && this.status < 400) {
                    successFn(this.responseText, this);
                } else {
                    errorFn(this);
                }
            }
        };

        request.send(dataStr);
    },

    formatAJAXData: function (data) {
        'use strict';
        var dataStr = '',
            dataParts = [],
            key = '';
        if (typeof data === 'object') {
            for (key in data) {
                if (data.hasOwnProperty(key)) {
                    dataParts.push(encodeURIComponent(key) + '=' + encodeURIComponent(data[key]));
                }
            }
            dataStr = dataParts.join('&');
        } else if (typeof data === 'string') {
            dataStr = data;
        } else {
            return null;
        }
        return dataStr;
    },
};
