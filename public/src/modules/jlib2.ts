import "@/array-from-polyfill";

export default function init(
    selector:
        | string
        | Array<HTMLElement>
        | NodeListOf<HTMLElement>
        | HTMLElement
        | EventTarget
        | null
) {
    if (selector === null) {
        return new jlib2("");
    }
    return new jlib2(selector);
}

export class jlib2 {
    elements: Array<HTMLElement>;

    get length() {
        return this.elements.length;
    }

    constructor(
        selector:
            | string
            | Array<HTMLElement>
            | NodeListOf<HTMLElement>
            | HTMLElement
            | EventTarget
    ) {
        if (typeof selector === "string") {
            if (selector === "") {
                this.elements = [];
                return;
            }

            // ID
            if (selector[0] === "#") {
                const id = selector.substr(1);
                const elem = document.getElementById(id);

                if (elem) {
                    this.elements = [elem];
                } else {
                    this.elements = [];
                }
                return;
            }

            // Class, tag, or other selector
            this.elements = Array.from(document.querySelectorAll(selector));
            return;
        }

        // NodeList or Array<Node>
        if (selector instanceof NodeList || selector instanceof Array) {
            this.elements = Array.from(selector);
            return;
        }

        // Single Node
        this.elements = [selector as HTMLElement];
    }

    // Array functions
    forEach(callback: (element: HTMLElement) => void): void {
        this.elements.forEach(callback);
    }

    map<T>(callback: (element: HTMLElement) => T): Array<T> {
        return this.elements.map(callback);
    }

    mapOne<T>(callback: (element: HTMLElement) => T): T | null {
        if (this.elements.length > 0) {
            return callback(this.elements[0]);
        }
        return null;
    }

    reduce<T>(
        callback: (
            previousValue: T,
            currentValue: HTMLElement,
            currentIndex: number,
            array: HTMLElement[]
        ) => T,
        initialValue: T
    ): T {
        return this.elements.reduce(callback, initialValue);
    }

    filter<T>(
        callback: (
            value: HTMLElement,
            index: number,
            array: HTMLElement[]
        ) => boolean
    ): HTMLElement[] {
        return this.elements.filter(callback);
    }

    // Event handlers
    on(eventName: string, fn: EventListenerOrEventListenerObject) {
        return this.forEach(el => {
            el.addEventListener(eventName, fn, false);
        });
    }

    off(eventName: string, fn: EventListenerOrEventListenerObject) {
        return this.forEach(el => {
            el.removeEventListener(eventName, fn, false);
        });
    }

    // Event handling - Convenience functions
    click(fn: EventListenerOrEventListenerObject) {
        this.on("click", fn);
    }
    submit(fn: EventListenerOrEventListenerObject) {
        this.on("submit", fn);
    }
    change(fn: EventListenerOrEventListenerObject) {
        this.on("change", fn);
    }
    keyup(fn: EventListenerOrEventListenerObject) {
        this.on("keyup", fn);
    }
    keydown(fn: EventListenerOrEventListenerObject) {
        this.on("keydown", fn);
    }
    keypress(fn: EventListenerOrEventListenerObject) {
        this.on("keypress", fn);
    }

    focus() {
        this.mapOne(el => el.focus());
    }

    // Element property functions
    text(text?: string): string {
        if (text !== undefined) {
            this.forEach(el => (el.textContent = text));
        }
        return this.mapOne(el => el.textContent) ?? "";
    }
    html(html?: string): string {
        if (html !== undefined) {
            this.forEach(el => (el.innerHTML = html));
        }
        return this.mapOne(el => el.innerHTML) ?? "";
    }
    value(value?: string): string {
        if (value !== undefined) {
            this.forEach(el => ((el as HTMLInputElement).value = value));
        }
        return this.mapOne(el => (el as HTMLInputElement).value) ?? "";
    }
    data(dk: string, dv?: string): string | null {
        if (dv !== undefined) {
            this.forEach(el => {
                if (el.dataset === undefined) {
                    el.setAttribute(`data-${dk}`, dv);
                    return;
                }
                el.dataset[dk] = dv;
            });
        }
        return (
            this.mapOne(el => {
                if (el.dataset === undefined) {
                    return el.getAttribute(`data-${dk}`);
                }
                return el.dataset[dk];
            }) ?? null
        );
    }
    attr(attr: string, val?: string): string {
        if (val !== undefined) {
            this.forEach(el => el.setAttribute(attr, val));
        }
        return this.mapOne(el => el.getAttribute(attr)) ?? "";
    }
    prop(prop: string, val?: any): any {
        if (val !== undefined) {
            this.forEach(el => ((el as any)[prop] = val));
        }
        return this.mapOne(el => (el as any)[prop]);
    }

    // Class manipulation
    addClass(className: string) {
        return this.forEach(el => {
            if (!hasClass(el, className)) {
                addClass(el, className);
            }
        });
    }

    removeClass(className?: string) {
        if (className === undefined) {
            return this.forEach(el => el.removeAttribute("class"));
        }
        return this.forEach(el => removeClass(el, className));
    }

    hasClass(className: string) {
        return this.mapOne(el => hasClass(el, className));
    }

    toggleClass(className: string) {
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

    style(s: string, v?: any) {
        if (v !== undefined) {
            return this.forEach(el => (el.style[<any>s] = v));
        }
        return this.mapOne(el => el.style[<any>s]);
    }

    fadeIn(speed: number, callback?: () => void) {
        this.fadeGeneric(speed, callback, "in");
    }

    fadeOut(speed: number, callback?: () => void) {
        this.fadeGeneric(speed, callback, "out");
    }

    fadeGeneric(speed: number, callback?: () => void, inOut?: "in" | "out") {
        if (inOut !== "in" && inOut !== "out") {
            console.error("Fade type must be either 'in' or 'out'");
            return;
        }

        let opacity = inOut === "in" ? 0 : 1;

        this.forEach(el => {
            el.style.opacity = opacity.toString();
            el.style.filter = "";
        });

        let last = new Date().getTime();
        const self = this;
        const tick = () => {
            if (inOut === "in") {
                opacity += (new Date().getTime() - last) / speed;
            } else {
                opacity -= (new Date().getTime() - last) / speed;
            }

            self.forEach(el => {
                el.style.opacity = opacity.toString();
                el.style.filter = `alpha(opacity=${(100 * opacity) | 0})`;
            });

            last = new Date().getTime();

            if (
                (inOut === "out" && opacity > 0) ||
                (inOut === "in" && opacity < 1)
            ) {
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

export function onReady(fn: () => void) {
    if (document.readyState !== "loading") {
        fn(); // DOM is already ready
        return;
    }
    document.addEventListener("DOMContentLoaded", fn);
}

// Wrappers for class manipulators to use classList of available
// or fallback to className
function removeClass(el: HTMLElement, className: string) {
    if (el.classList) {
        el.classList.remove(className);
    } else {
        el.className = el.className.replace(
            new RegExp(
                "(^|\\b)" + className.split(" ").join("|") + "(\\b|$)",
                "gi"
            ),
            " "
        );
    }
}

function addClass(el: HTMLElement, className: string) {
    if (el.classList) {
        el.classList.add(className);
    } else {
        el.className += " " + className;
    }
}

function hasClass(el: HTMLElement, className: string) {
    if (el.classList) {
        return el.classList.contains(className);
    }
    return el.className.indexOf(className) !== -1;
}

export function newTag(tag: string, attrs: any) {
    const el = new jlib2([document.createElement(tag)]);

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
