import $ from 'jLib';
import api from 'pg-api';
import flashMessage from 'flash';

function login() {
  $('#login-btn').prop('disabled', true);

  var data = {
    username: $('[name=username]').value(),
    password: $('[name=password]').value()
  };

  if (data.username === '' || data.password === '') {
    return;
  }

  $('#login-btn').text('Logging in...');

  api.login(data, function() {
    location.href = '/';
  }, function(req) {
    $('#login-btn').text('Login >');
    $('#login-btn').prop('disabled', false);
    if (req.status === 401) {
      flashMessage('Incorrect username or password');
    } else {
      flashMessage('Unknown error');
    }
  });
}

function checkKeyAndLogin(e) {
  if (e.keyCode === 13) { login(); }
}

$('#login-btn').click(login);
$('[name=username]').keyup(checkKeyAndLogin);
$('[name=password]').keyup(checkKeyAndLogin);
