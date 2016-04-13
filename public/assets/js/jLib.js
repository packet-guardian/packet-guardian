/*
 * The jLib library is a collection of commonly used functions to add quick functionality to a project.
 * This library is extremely light-weight and only adds what is necessary for convience.
 *
 * Licensed under the MIT license. Text available online: https://opensource.org/licenses/MIT
**/
/* exported j */
/* jshint -W083 */
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

    Noop: function() {
        return;
    },

    // Event handling

    On: function (el, eventName, handler) {
        'use strict';
        var els = j.$(el, true),
            i = 0;
        for (i = 0; i < els.length; i++) {
            if (els[i].addEventListener) {
                els[i].addEventListener(eventName, handler);
            } else {
                els[i].attachEvent('on' + eventName, function () {
                    handler.call(el[i]);
                });
            }
        }
    },

    Off: function (el, eventName, handler) {
        'use strict';
        var els = j.$(el, true),
            i = 0;
        for (i = 0; i < els.length; i++) {
            if (els[i].removeEventListener) {
                els[i].removeEventListener(eventName, handler);
            } else {
                els[i].detachEvent('on' + eventName, handler);
            }
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

    Keyup: function (el, handler) {
        'use strict';
        j.On(el, 'keyup', handler);
    },

    Keydown: function (el, handler) {
        'use strict';
        j.On(el, 'keydown', handler);
    },

    Keypress: function (el, handler) {
        'use strict';
        j.On(el, 'keypress', handler);
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
        };
        j.Ajax('POST', url, data, successFn, errorFn, middleware);
    },

    Delete: function (url, data, successFn, errorFn) {
        'use strict';
        j.Ajax('DELETE', url, data, successFn, errorFn, null);
    },

    Ajax: function (method, url, data, successFn, errorFn, middleware) {
        'use strict';
        method = method.toUpperCase();
        successFn = (successFn !== undefined && successFn !== null) ? successFn : j.Noop;
        errorFn = (errorFn !== undefined && errorFn !== null) ? errorFn : j.Noop;
        middleware = (middleware !== undefined && middleware !== null) ? middleware : j.Noop;
        var dataStr = j.formatAJAXData(data),
            request = new XMLHttpRequest();

        if (dataStr === null) {
            console.log('Invalid data to j.Post.');
            return;
        }

        if (method === 'GET' || method === 'DELETE') {
            url += '?' + dataStr;
            dataStr = '';
        }

        request.open(method.toUpperCase(), url, true);
        middleware(request);

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

    // Effects

    FadeIn: function (el, speed, post) {
        'use strict';
        j.fadeGeneric(el, speed, post, "in");
    },

    FadeOut: function (el, speed, post) {
        'use strict';
        j.fadeGeneric(el, speed, post, "out");
    },

    fadeGeneric: function (el, speed, post, inOut) {
        'use strict';
        if (inOut !== "in" && inOut !== "out") {
            console.error("Fade type must be either 'in' or 'out'");
            return;
        }
        el = j.$(el);
        post = (post !== undefined && post !== null) ? post : j.Noop;
        var opacity = (inOut === "in") ? 0 : 1;

        el.style.opacity = 0;
        el.style.filter = '';

        var last = +new Date();
        var tick = function() {
            if (inOut === "in") {
                opacity += (new Date() - last) / speed;
            } else {
                opacity -= (new Date() - last) / speed;
            }
            el.style.opacity = opacity;
            el.style.filter = 'alpha(opacity=' + (100 * opacity)|0 + ')';

            last = +new Date();

            if ((inOut === "out" && opacity > 0) || (inOut === "in" && opacity < 1)) {
                if (window.requestAnimationFrame) {
                    requestAnimationFrame(tick);
                } else {
                    setTimeout(tick, 16);
                }
            } else {
                post();
            }
        };

        tick();
    },

    Show: function (el) {
        'use strict';
        j.$(el).style.display = '';
    },

    Hide: function (el) {
        'use strict';
        j.$(el).style.display = 'none';
    },
};
