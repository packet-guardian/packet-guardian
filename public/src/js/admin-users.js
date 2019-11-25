import $ from "jlib2";
import "flash";

// Form submittion
$("#new-user-form").submit(function(e) {
  e.preventDefault();
  const username = $("[name=username]").value();
  if (username !== "") {
    location.href = `/admin/users/${username}`;
  }
});
