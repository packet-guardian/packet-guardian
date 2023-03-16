import $ from "@/jlib2";
import api from "@/pg-api";
import flashMessage from "@/flash";

function login() {
    const data = {
        username: $("[name=username]").value(),
        password: $("[name=password]").value(),
    };

    if (data.username === "" || data.password === "") {
        return;
    }

    $("#login-btn").prop("disabled", "true");
    $("#login-btn").text("Logging in...");

    api.login(
        data,
        () => (location.href = "/"),
        (req: any) => {
            $("#login-btn").text("Login");
            $("#login-btn").prop("disabled", "false");
            if (req.status === 401) {
                flashMessage("Incorrect username or password");
            } else {
                flashMessage("Unknown error");
            }
        }
    );
}

function checkKeyAndLogin(e: Event) {
    if ((e as KeyboardEvent).keyCode === 13) {
        login();
    }
}

$("#login-btn").click(login);
$("[name=username]").keyup(checkKeyAndLogin);
$("[name=password]").keyup(checkKeyAndLogin);
