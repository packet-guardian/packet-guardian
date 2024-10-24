import $ from "@/jlib2";
import api, { SaveUserInput } from "@/pg-api";
import flashMessage from "@/flash";
import { setTextboxToToday } from "@/utils";
import { ModalPrompt } from "@/modals";

const devExpirationTypes = {
    never: 0,
    global: 1,
    specific: 2,
    duration: 3,
    daily: 4,
    rolling: 5,
};

// Device limit select box init
function checkLimit() {
    const limit = $("[name=device-limit]");
    const specialLimits = $("[name=special-limit]");
    if (limit.value() === "-1") {
        specialLimits.value("global");
        limit.value("");
        limit.prop("disabled", true);
    } else if (limit.value() === "0") {
        specialLimits.value("unlimited");
        limit.value("");
        limit.prop("disabled", true);
    } else {
        specialLimits.value("specific");
    }
}
checkLimit();

// Expiration textboxes init
function checkExpires() {
    const limit = $("[name=device-expiration]");
    const devExpSel = $("[name=dev-exp-sel]");
    const expires = devExpSel.data("expires");
    switch (expires) {
        case "0":
            devExpSel.value("never");
            limit.value("");
            limit.prop("disabled", true);
            break;
        case "1":
            devExpSel.value("global");
            limit.value("");
            limit.prop("disabled", true);
            break;
        case "3":
            devExpSel.value("duration");
            break;
        case "4":
            devExpSel.value("daily");
            // Remove "Daily at " text
            limit.value(limit.value().slice(9));
            break;
        case "5":
            devExpSel.value("rolling");
            limit.value("");
            limit.prop("disabled", true);
            break;
        default:
            devExpSel.value("specific");
    }

    const valAfter = $("[name=valid-after]");
    const valBefore = $("[name=valid-before]");
    const valBefSel = $("[name=val-bef-sel]");
    const forever = valBefSel.data("forever");
    if (forever === "true") {
        valBefSel.value("forever");
        valBefore.value("");
        valBefore.prop("disabled", true);
        valAfter.value("");
        valAfter.prop("disabled", true);
    } else {
        valBefSel.value("specific");
    }
}
checkExpires();

// Select boxes change events
$("[name=special-limit]").change((e) => {
    const devLimit = $("[name=device-limit]");
    devLimit.value("");
    devLimit.prop("disabled", $(e.target).value() !== "specific");
});

$("[name=dev-exp-sel]").change((e) => {
    const self = $(e.target);
    // Enable/disable appropiate textboxes
    $("[name=device-expiration]").prop(
        "disabled",
        !(
            self.value() === "specific" ||
            self.value() === "daily" ||
            self.value() === "duration"
        )
    );

    // Zero field by default
    $("[name=device-expiration]").value("");

    // Fill in textbox and tooltip as needed
    switch (self.value()) {
        case "specific":
            setTextboxToToday("[name=device-expiration]");
            setExpirationToolTop("(YYYY-MM-DD HH:mm)");
            break;
        case "duration":
            setExpirationToolTop("(5h30m = 5 hours and 30 minutes)");
            break;
        case "daily":
            setExpirationToolTop("(HH:mm)");
            break;
        default:
            setExpirationToolTop("");
    }
});

$("[name=val-bef-sel]").change((e) => {
    const self = $(e.target);
    $("[name=valid-before]").prop("disabled", self.value() === "forever");
    $("[name=valid-after]").prop("disabled", self.value() === "forever");

    if (self.value() === "specific") {
        setTextboxToToday("[name=valid-before]");
        setTextboxToToday("[name=valid-after]");
        setUserExpirationToolTip("(YYYY-MM-DD HH:mm)");
    } else {
        $("[name=valid-before]").value("");
        $("[name=valid-after]").value("");
        setUserExpirationToolTip("");
    }
});

$("[name=delete-btn]").click(() => {
    api.deleteUser(
        $("[name=username]").value(),
        (resp, req) => {
            if (req.status > 204) {
                flashMessage(resp.Message);
                return;
            }
            location.href = "/admin/users";
        },
        () => flashMessage("Error deleting user")
    );
});

function getDelegatesList(): string {
    return Array.from(document.querySelectorAll("p[data-delegate]"))
        .map(
            (item) =>
                `${item.querySelector("label")?.innerText}:${
                    item.querySelector("select")?.value
                }`
        )
        .join(",");
}

// Form submittion
$("#user-form").submit((e) => {
    e.preventDefault();
    const formData: SaveUserInput = {
        username: $("[name=username]").value(),
        password: $("[name=password]").value(),
        device_limit: -1,
        expiration_type: devExpirationTypes.global,
        device_expiration: $("[name=device-expiration]").value(),
        valid_start: "0",
        valid_end: "0",
        can_manage: $("[name=can-manage]").prop("checked") ? 1 : 0,
        can_autoreg: $("[name=can-autoreg]").prop("checked") ? 1 : 0,
        allow_status_api: $("[name=user-api-status]").prop("checked") ? 1 : 0,
        ui_group: $("[name=user-ui-group]").value(),
        api_group: $("[name=user-api-group]").value(),
        delegates: getDelegatesList(),
        notes: $("[name=notes]").value()
    };

    if ($("[name=clear-pass]").prop("checked")) {
        formData.password = "-1";
    }

    const devLimit = $("[name=special-limit]").value();
    if (devLimit === "global") {
        formData.device_limit = -1;
    } else if (devLimit === "unlimited") {
        formData.device_limit = 0;
    } else {
        formData.device_limit = parseInt($("[name=device-limit]").value());
    }

    const devExpSel = $("[name=dev-exp-sel]").value();
    if (devExpSel in devExpirationTypes) {
        formData.expiration_type = (<any>devExpirationTypes)[<any>devExpSel];
    }

    if ($("[name=val-bef-sel]").value() !== "forever") {
        formData.valid_start = $("[name=valid-after]").value();
        formData.valid_end = $("[name=valid-before]").value();
    }

    api.saveUser(
        formData,
        (resp, req) => {
            window.scroll(0, 0);
            if (req.status > 204) {
                flashMessage(resp.Message);
                return;
            }

            flashMessage("User saved", "success");
            $("[name=password]").value("");
            $("[name=clear-pass]").prop("checked", false);
            if (formData.password === "" || formData.password === "-1") {
                $("#has-password").text("No");
            } else {
                $("#has-password").text("Yes");
            }
            $("#submit-btn").text("Save");
        },
        (req) => {
            const resp = JSON.parse(req.responseText);
            flashMessage(resp.Message);
        }
    );
});

// Utility functions
function setExpirationToolTop(tip: string) {
    $("#dev-exp-tooltip").text(tip);
}

function setUserExpirationToolTip(tip: string) {
    $("#user-exp-tooltip").text(tip);
}

$("#add-delegate-btn").click(() => {
    const pmodal = new ModalPrompt();
    pmodal.show("Delegate username:", (newUser) => {
        const newDelegate = document.createElement("p");
        newDelegate.dataset["delegate"] = newUser;

        const newDelegateLabel = document.createElement("label");
        newDelegateLabel.innerText = newUser;
        newDelegateLabel.setAttribute("for", `${newUser}-permissions`);
        newDelegate.append(newDelegateLabel);
        newDelegate.appendChild(document.createTextNode(" "));
        newDelegate.appendChild(
            makeSelect(`${newUser}-permissions`, [
                {
                    value: "RO",
                    text: "RO",
                    selected: false,
                },
                {
                    value: "RW",
                    text: "RW",
                    selected: false,
                },
            ])
        );

        const delegateDelete = createElementFromHTML(
            `<i class="fa fa-times delete-icon" aria-role="button" data-delegate="${newUser}"title="Delete delegate"></i>`
        );
        $(delegateDelete).click(removeDelegate);

        newDelegate.appendChild(document.createTextNode(" "));
        newDelegate.appendChild(delegateDelete);

        document
            .querySelector("[name='delegate-list']")
            ?.appendChild(newDelegate);
    });
});

interface SelectOption {
    value: string;
    selected: boolean;
    text: string;
}

function makeSelect(name: string, options: SelectOption[]): HTMLSelectElement {
    const select = document.createElement("select");
    select.setAttribute("name", name);

    options.forEach((option) => {
        const opt = document.createElement("option");
        opt.value = option.value;
        opt.text = option.text;
        opt.selected = option.selected;
        select.appendChild(opt);
    });

    return select;
}

function createElementFromHTML(htmlString: string): HTMLElement {
    var div = document.createElement("div");
    div.innerHTML = htmlString.trim();

    // Change this to div.childNodes to support multiple top-level nodes
    return div.firstChild as HTMLElement;
}

$("i[data-delegate]").click(removeDelegate);

function removeDelegate(e: Event) {
    const target = e.target as HTMLElement;
    const delegateName = target?.dataset["delegate"];
    document.querySelector(`p[data-delegate=${delegateName}]`)?.remove();
}
