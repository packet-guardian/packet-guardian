import $ from "@/jlib2";

// Form submittion
$("#new-user-form").submit((e) => {
    e.preventDefault();
    const username = $("[name=username]").value();
    if (username !== "") {
        location.href = `/admin/users/${username}`;
    }
});
