import $ from "@/jlib2";
import api, { RegisterDeviceInput } from "@/pg-api";
import flashMessage from "@/flash";

function register() {
    disableRegBtn();
    const data = {
        username: "",
        "mac-address": "",
        description: $("[name=dev-desc]").value(),
        platform: ""
    };

    // It's not guaranteed that all fields will be shown
    // The username box will always be shown, sometimes disabled
    const username = $("[name=username]");
    if (username.length !== 0) {
        data.username = username.value();
    }
    if (data.username === "") {
        enableRegBtn();
        return;
    } // Required

    // The password box will only show if the user isn't logged in
    let password = "";
    const passwordElem = $("[name=password]");
    if (passwordElem.length !== 0) {
        password = passwordElem.value();
        if (password === "") {
            enableRegBtn();
            return;
        }
    }

    // The mac-address field will only show for a manual registration
    const mac = $("[name=mac-address]");
    if (mac.length !== 0) {
        data["mac-address"] = mac.value();
        if (data["mac-address"] === "") {
            enableRegBtn();
            return;
        }
    }

    // The platform field will only show for a manual registration
    const platform = $("[name=platform]");
    if (platform.length !== 0) {
        data.platform = platform.value();
        if (data.platform === "") {
            enableRegBtn();
            return;
        }
    }

    if (password !== "") {
        // Need to login first
        api.login(
            { username: data.username, password },
            () => registerDevice(data, true),
            req => {
                window.scrollTo(0, 0);
                enableRegBtn();
                if (req.status === 401) {
                    flashMessage("Incorrect username or password");
                } else {
                    flashMessage("Unknown error");
                }
            }
        );
    } else {
        registerDevice(data, false);
    }
}

function disableRegBtn() {
    $("#register-btn").prop("disabled", true);
    $("#register-btn").text("Registering...");
}

function enableRegBtn() {
    $("#register-btn").text("Register");
    $("#register-btn").prop("disabled", false);
}

function registerDevice(data: RegisterDeviceInput, logout: boolean) {
    api.registerDevice(
        data,
        resp => {
            window.scrollTo(0, 0);
            flashMessage("Registration successful", "success");
            $(".register-box").hide();

            if (logout) {
                // If the user had to login to register, let's log them out.
                // It may be a bit confusing if they go back and forget they
                // had to enter a password.
                api.logout();
            }

            if (data["mac-address"] === "") {
                $("#suc-msg-auto").show();
                return;
            }

            location.href = resp.Data.Location;
        },
        req => {
            window.scrollTo(0, 0);
            enableRegBtn();
            const resp = JSON.parse(req.responseText);
            switch (req.status) {
                case 500:
                    flashMessage(`Internal Server Error - ${resp.Message}`);
                    break;
                default:
                    flashMessage(resp.Message);
                    break;
            }
            if (logout) {
                // If the user had to login to register, let's log them out.
                // It may be a bit confusing if they go back and forget they
                // had to enter a password.
                api.logout();
            }
        }
    );
}

$("#register-btn").click(register);
