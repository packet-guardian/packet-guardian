export default function init(selector) {
  return new jlib2(selector);
}

export class jlib2 {
  elements = [];

  get length() {
    return this.elements.length;
  }

  constructor(selector) {
    if (!selector) {
      return;
    }

    let elements = [];
    if (typeof selector === "string") {
      const chr = selector.substr(1);

      if (selector[0] === "#") {
        const elem = document.getElementById(chr);

        if (elem) {
          elements.push(elem);
        }
      } else {
        elements = Array.from(document.querySelectorAll(selector));
      }
    } else if (selector.length && selector.isArray) {
      elements = selector;
    } else {
      elements = [selector];
    }

    this.elements = elements;
  }

  // Array functions
  forEach(callback) {
    this.elements.forEach(callback);
  }

  map(callback) {
    return this.elements.map(callback);
  }

  mapOne(callback) {
    var m = this.map(callback);
    return m.length > 1 ? m : m[0];
  }

  reduce(callback, initialValue) {
    return this.elements.reduce(callback, initialValue);
  }

  filter(callback) {
    return this.elements.filter(callback);
  }

  // Event handlers
  on(eventName, dgt, fn) {
    var delegate = function(ev) {
      const tg = ev.target;

      if (typeof dgt !== "string") {
        return dgt(ev, tg);
      }

      let nme = "";
      let cs = null;
      nme = dgt.substr(1);
      if (dgt[0] === ".") {
        cs = tg.className.split(" ");
      } else if (dgt[0] === "#") {
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

    return this.forEach(el => {
      if (el !== null) {
        el.addEventListener(eventName, delegate, false);
      }
    });
  }

  off(evt, fn) {
    return this.forEach(el => {
      if (el !== null) {
        el.removeEventListener(evt, fn, false);
      }
    });
  }

  // Event handling - Convenience functions
  click(dgt, fn) {
    this.on("click", dgt, fn);
  }
  submit(dgt, fn) {
    this.on("submit", dgt, fn);
  }
  change(dgt, fn) {
    this.on("change", dgt, fn);
  }
  keyup(dgt, fn) {
    this.on("keyup", dgt, fn);
  }
  keydown(dgt, fn) {
    this.on("keydown", dgt, fn);
  }
  keypress(dgt, fn) {
    this.on("keypress", dgt, fn);
  }

  // Element property functions
  text(text) {
    if (text !== undefined) {
      return this.forEach(el => (el.textContent = text));
    }
    return this.mapOne(el => el.textContent);
  }
  html(html) {
    if (html !== undefined) {
      return this.forEach(el => (el.innerHTML = html));
    }
    return this.mapOne(el => el.innerHTML);
  }
  value(value) {
    if (value !== undefined) {
      return this.forEach(el => (el.value = value));
    }
    return this.mapOne(el => el.value);
  }
  data(dk, dv) {
    if (dv !== undefined) {
      return this.forEach(el => {
        if (el.dataset === undefined) {
          el.setAttribute(`data-${dk}`, dv);
          return;
        }
        el.dataset[dk] = dv;
      });
    }
    return this.mapOne(el => {
      if (el.dataset === undefined) {
        return el.getAttribute(`data-${dk}`);
      }
      return el.dataset[dk];
    });
  }
  attr(attr, val) {
    if (val !== undefined) {
      return this.forEach(el => el.setAttribute(attr, val));
    }
    return this.mapOne(el => el.getAttribute(attr));
  }
  prop(prop, val) {
    if (val !== undefined) {
      return this.forEach(el => (el[prop] = val));
    }
    return this.mapOne(el => el[prop]);
  }

  // Class manipulation
  addClass(className) {
    return this.forEach(el => {
      if (!hasClass(el, className)) {
        addClass(el, className);
      }
    });
  }

  removeClass(className) {
    if (className === undefined) {
      return this.forEach(el => el.removeAttribute("class"));
    }
    return this.forEach(el => removeClass(el, className));
  }

  hasClass(className) {
    return this.mapOne(el => hasClass(el, className));
  }

  toggleClass(className) {
    return this.forEach(el => {
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
  }

  // Fading/Display
  show() {
    return this.forEach(el => (el.style.display = "block"));
  }

  hide() {
    return this.forEach(el => (el.style.display = "none"));
  }

  style(s, v) {
    if (v !== undefined) {
      return this.forEach(el => (el.style[s] = v));
    }
    return this.mapOne(el => el.style[s]);
  }

  fadeIn(speed, callback) {
    this.fadeGeneric(speed, callback, "in");
  }

  fadeOut(speed, callback) {
    this.fadeGeneric(speed, callback, "out");
  }

  fadeGeneric(speed, callback, inOut) {
    if (inOut !== "in" && inOut !== "out") {
      console.error("Fade type must be either 'in' or 'out'");
      return;
    }

    let opacity = inOut === "in" ? 0 : 1;

    this.forEach(el => {
      el.style.opacity = opacity;
      el.style.filter = "";
    });

    let last = +new Date();
    const self = this;
    const tick = () => {
      if (inOut === "in") {
        opacity += (new Date() - last) / speed;
      } else {
        opacity -= (new Date() - last) / speed;
      }

      self.forEach(el => {
        el.style.opacity = opacity;
        el.style.filter = `alpha(opacity=${(100 * opacity) | 0})`;
      });

      last = +new Date();

      if ((inOut === "out" && opacity > 0) || (inOut === "in" && opacity < 1)) {
        requestAnimationFrame(tick);
      } else {
        if (callback) {
          return callback();
        }
      }
    };
    tick();
  }
}

export function onReady(fn) {
  if (document.readyState !== "loading") {
    fn(); // DOM is already ready
    return;
  }
  document.addEventListener("DOMContentLoaded", fn);
}

// Wrappers for class manipulators to use classList of available
// or fallback to className
function removeClass(el, className) {
  if (el.classList) {
    el.classList.remove(className);
  } else {
    el.className = el.className.replace(
      new RegExp("(^|\\b)" + className.split(" ").join("|") + "(\\b|$)", "gi"),
      " "
    );
  }
}

function addClass(el, className) {
  if (el.classList) {
    el.classList.add(className);
  } else {
    el.className += " " + className;
  }
}

function hasClass(el, className) {
  if (el.classList) {
    return el.classList.contains(className);
  }
  return el.className.indexOf(className) !== -1;
}

export function newTag(tag, attrs) {
  const el = jlib2([document.createElement(tag)]);

  if (attrs) {
    if (attrs.text) {
      el.text(attrs.text);
      delete attrs.text;
    }
    if (attrs.html) {
      el.html(attrs.html);
      delete attrs.html;
    }
    for (const key in attrs) {
      if (attrs.hasOwnProperty(key)) {
        el.attr(key, attrs[key]);
      }
    }
  }

  return el;
}
