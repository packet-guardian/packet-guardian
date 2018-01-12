/*
 * The jLib library is a collection of commonly used functions to add quick functionality to a project.
 * This library is extremely light-weight and only adds what is necessary for convenience.
 *
 * Licensed under the MIT license. Text available online: https://opensource.org/licenses/MIT
**/
var jLib = function(params) {
    return new jLib.fn.init(params);
};

jLib.noop = function() {};

// Document ready
jLib.onReady = function (fn) {
    if (document.readyState !== 'loading') {
        fn(); // DOM is already ready
        return;
    }
    document.addEventListener('DOMContentLoaded', fn);
};

jLib.fn = jLib.prototype = {
    constructor: jLib,
    forEach: function(callback) {
        this.map(callback);
        return this;
    },
    map: function(callback) {
        var results = [], j, ln;
        ln = this.length;
        for (j = 0; j < ln; j++) {
            results.push(callback.call(this, this[j], j));
        }
        return results;
    },
    mapOne: function(callback) {
        var m = this.map(callback);
        return m.length > 1 ? m : m[0];
    },

    // Event handling
    on: (function () {
        if (document.addEventListener) {
            return function(evt, dgt, fn) {
                var nme;
                var delegate = function(ev) {
                    var tg = ev.target,
                        cs;

                    if (typeof dgt !== 'string') {
                        return dgt(ev, tg);
                    }

                    nme = dgt.substr(1);
                    if (dgt[0] === '.') {
                        cs = tg.className.split(' ');
                    } else if (dgt[0] === '#') {
                        cs = tg.id;
                    } else {
                        cs = tg.nodeName.toLowerCase();
                        nme = dgt.substr(0);
                    }

                    if (cs.indexOf(nme) === -1) {
                        ev.preventDefault();
                        ev.stopPropagation();
                        return;
                    }
                    return fn(ev, tg);
                };

                return this.forEach(function(el) {
                    if (el === null) {
                        return;
                    }
                    el.addEventListener(evt, delegate, false);
                });
            };
        }
        if (document.attachEvent) {
            return function (evt, fn) {
                return this.forEach(function(el) {
                    el.attachEvent('on' + evt, fn);
                });
            };
        }
        return function (evt, fn) {
            return this.forEach(function(el) {
                el['on' + evt] = fn;
            });
        };

    })(),

    off: (function () {
        if (document.removeEventListener) {
            return function (evt, fn) {
                return this.forEach(function(el) {
                    el.removeEventListener(evt, fn, false);
                });
            };
        }
        if (document.detachEvent) {
            return function (evt, fn) {
                return this.forEach(function(el) {
                    el.detachEvent('on' + evt, fn);
                });
            };
        }
        return function (evt) {
            return this.forEach(function(el) {
                el['on' + evt] = null;
            });
        };
    })(),

    // Event handling - Convenience functions
    click:    function (dgt, fn) { this.on('click', dgt, fn); },
    submit:   function (dgt, fn) { this.on('submit', dgt, fn); },
    change:   function (dgt, fn) { this.on('change', dgt, fn); },
    keyup:    function (dgt, fn) { this.on('keyup', dgt, fn); },
    keydown:  function (dgt, fn) { this.on('keydown', dgt, fn); },
    keypress: function (dgt, fn) { this.on('keypress', dgt, fn); },

    // Element property functions
    text: function(text) {
        if (text !== undefined) {
            return this.forEach(function(el) {
                el.textContent = text;
            });
        }
        return this.mapOne(function(el) {
            return el.textContent;
        });
    },
    html: function(html) {
        if (html !== undefined) {
            return this.forEach(function(el) {
                el.innerHTML = html;
            });
        }
        return this.mapOne(function(el) {
            return el.innerHTML;
        });
    },
    value: function(value) {
        if (value !== undefined) {
            return this.forEach(function(el) {
                el.value = value;
            });
        }
        return this.mapOne(function(el) {
            return el.value;
        });
    },
    data: function(dk, dv) {
        if (dv !== undefined) {
            return this.forEach(function(el) {
                if (el.dataset === undefined) {
                    el.setAttribute(`data-${dk}`, dv);
                    return;
                }
                el.dataset[dk] = dv;
            });
        }
        return this.mapOne(function(el) {
            if (el.dataset === undefined) {
                return el.getAttribute(`data-${dk}`);
            }
            return el.dataset[dk];
        });
    },
    attr: function(attr, val) {
        if (val !== undefined) {
            return this.forEach(function(el) {
                el.setAttribute(attr, val);
            });
        }
        return this.mapOne(function(el) {
            return el.getAttribute(attr);
        });
    },
    prop: function(prop, val) {
        if (val !== undefined) {
            return this.forEach(function(el) {
                el[prop] = val;
            });
        }
        return this.mapOne(function(el) {
            return el[prop];
        });
    },

    // Class manipulation
    addClass: function(className) {
        return this.forEach(function(el) {
            if (el.className.indexOf(className) !== -1) {
                return;
            }
            addClass(el, className);
        });
    },

    removeClass: function(className) {
        if (className === undefined) {
            return this.forEach(function(el) {
                el.removeAttribute('class');
            });
        }
        return this.forEach(function(el) {
            removeClass(el, className);
        });
    },

    hasClass: function(className) {
        return this.mapOne(function(el) {
            return hasClass(el, className);
        });
    },

    toggleClass: function(className) {
        return this.forEach(function(el) {
            if (el.classList) {
                el.classList.toggle(className);
            } else {
                if (hasClass(el, className)) {
                    removeClass(el, className);
                } else {
                    addClass(el, className);
                }
            }
        });
    },

    show: function() {
        return this.forEach(function(el) {
            el.style.display = 'block';
        });
    },

    hide: function() {
        return this.forEach(function(el) {
            el.style.display = 'none';
        });
    },

    style: function(s, v) {
        if (v !== undefined) {
            return this.forEach(function(el) {
                el.style[s] = v;
            });
        }
        return this.mapOne(function(el) {
            return el.style[s];
        });
    },

    fadeIn: function(speed, callback) {
        this.fadeGeneric(speed, callback, 'in');
    },

    fadeOut: function(speed, callback) {
        this.fadeGeneric(speed, callback, 'out');
    },

    fadeGeneric: function(speed, callback, inOut) {
        if (inOut !== 'in' && inOut !== 'out') {
            console.error("Fade type must be either 'in' or 'out'");
            return;
        }
        callback = (callback) ? callback : jLib.noop;
        var opacity = (inOut === 'in') ? 0 : 1,
            self = this;

        self.forEach(function(el) {
            el.style.opacity = opacity;
            el.style.filter = '';
        });

        var last = +new Date();
        var tick = function() {
            if (inOut === 'in') {
                opacity += (new Date() - last) / speed;
            } else {
                opacity -= (new Date() - last) / speed;
            }
            self.forEach(function(el) {
                el.style.opacity = opacity;
                el.style.filter = `alpha(opacity=${(100 * opacity)|0})`;
            });

            last = +new Date();

            if ((inOut === 'out' && opacity > 0) || (inOut === 'in' && opacity < 1)) {
                requestAnimationFrame(tick);
            } else {
                return callback();
            }
        };
        tick();
    }
};

var init = jLib.fn.init = function(s) {
    var els, chr, i, cl;
    if (!s) {
        return this;
    }
    if (typeof s === 'string') {
        chr = s.substr(1);
        if (s[0] === '#') {
            els = [document.getElementById(chr)];
        } else {
            els = document.querySelectorAll(s);
        }
    } else if (s.length && s.isArray) {
        els = s;
    } else {
        els = [s];
    }

    cl = els.length;

    for (i = 0; i < cl; i++) {
        this[i] = els[i];
    }
    this.length = cl;
    return this;
};

init.prototype = jLib.fn;

// ajaxSettings = {
//     method: "GET",
//     url: '',
//     params: '',
//     data: '',
//     contentType: "application/x-www-form-urlencoded; charset=UTF-8",
//     success: jLib().noop(),
//     error: jLib().noop(),
//     headers: {},
// }

jLib.new = function(tag, attrs) {
    var el = jLib([document.createElement(tag)]), key;
    if (attrs) {
        if (attrs.text) {
            el.text(attrs.text);
            delete attrs.text;
        }
        if (attrs.html) {
            el.html(attrs.html);
            delete attrs.html;
        }
        for (key in attrs) {
            if (attrs.hasOwnProperty(key)) {
                el.attr(key, attrs[key]);
            }
        }
    }
    return el;
};

jLib.ajax = function(options) {
    if (options === undefined) {
        return null;
    }
    options = checkAjaxSettings(options);

    var xhr = new XMLHttpRequest(), xhrURL = '';

    xhrURL = (options.params === '') ? options.url : options.url + '?' + options.params;
    xhr.open(options.method, xhrURL, true);
    xhr.setRequestHeader('Content-Type', options.contentType);
    for (var key in options.headers) {
        if (options.headers.hasOwnProperty(key)) {
            xhr.setRequestHeader(key, options.headers[key]);
        }
    }
    xhr.setRequestHeader('X-Requested-With', 'XMLHttpRequest');
    xhr.onreadystatechange = function() {
        if (this.readyState === 4) {
            if (this.status >= 200 && this.status < 400) {
                options.success(this.responseText, this);
            } else {
                options.error(this);
            }
        }
    };
    xhr.send(options.data);
};

jLib.params = function(data) {
    var dataStr = '',
        dataParts = [],
        key = '';
    if (typeof data === 'string') {
        return data;
    }
    if (typeof data === 'object') {
        for (key in data) {
            if (data.hasOwnProperty(key)) {
                dataParts.push(encodeURIComponent(key) + '=' + encodeURIComponent(data[key]));
            }
        }
        dataStr = dataParts.join('&');
    }
    return dataStr;
};

jLib.get = function(url, data, success, error) {
    jLib.ajax({
        method: 'GET',
        url: url,
        params: jLib.params(data),
        success: success,
        error: error
    });
};

jLib.post = function(url, data, success, error) {
    jLib.ajax({
        method: 'POST',
        url: url,
        data: jLib.params(data),
        success: success,
        error: error
    });
};

function checkAjaxSettings(options) {
    if (!options.method) {
        options.method = 'GET';
    }
    options.method.toUpperCase();
    if (!options.url) {
        options.url = '';
    }
    if (!options.data) {
        options.data = '';
    }
    if (typeof options.data !== 'string') {
        options.data = jLib.params(options.data);
    }
    if (!options.params) {
        options.params = '';
    }
    if (typeof options.params !== 'string') {
        options.params = jLib.params(options.params);
    }
    if (!options.contentType) {
        options.contentType = 'application/x-www-form-urlencoded; charset=UTF-8';
    }
    if (!options.success) {
        options.success = jLib.noop;
    }
    if (!options.error) {
        options.error = jLib.noop;
    }
    if (!options.headers) {
        options.headers = {};
    }
    return options;
}

// Wrappers for class manipulators to use classList of available
// or fallback to className
function removeClass(el, className) {
    if (el.classList) {
        el.classList.remove(className);
    } else {
        el.className = el.className.replace(new RegExp('(^|\\b)' + className.split(' ').join('|') + '(\\b|$)', 'gi'), ' ');
    }
}

function addClass(el, className) {
    if (el.classList) {
        el.classList.add(className);
    } else {
        el.className += ' ' + className;
    }
}

function hasClass(el, className) {
    if (el.classList) {
        return el.classList.contains(className);
    } else {
        return (el.className.indexOf(className) !== -1);
    }
}

export default jLib;
