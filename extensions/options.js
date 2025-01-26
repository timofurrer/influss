document.getElementById('options-form').addEventListener('submit', (e) => {
  e.preventDefault();

  browser.storage.local.set({
    endpoint: document.getElementById('endpoint').value,
    username: document.getElementById('username').value,
    password: document.getElementById('password').value
  });
});

// Load stored settings
browser.storage.local.get(['endpoint', 'username', 'password']).then(items => {
  if (items.endpoint) document.getElementById('endpoint').value = items.endpoint;
  if (items.username) document.getElementById('username').value = items.username;
  if (items.password) document.getElementById('password').value = items.password;
});
