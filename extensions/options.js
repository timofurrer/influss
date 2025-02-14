document.getElementById('options-form').addEventListener('submit', (e) => {
  e.preventDefault();

  const endpoint = document.getElementById('endpoint').value;
  const username = document.getElementById('username').value;
  const password = document.getElementById('password').value;

  browser.storage.sync.set({
    endpoint: endpoint,
    username: username,
    password: password
  }).then(() => {
    const saveMessage = document.getElementById('save-message');
    saveMessage.style.display = 'block';

    // Hide the message after 3 seconds
    setTimeout(() => {
      saveMessage.style.display = 'none';
    }, 3000);
  });
});
});

// Load saved options
browser.storage.sync.get(['endpoint', 'username', 'password']).then((result) => {
  if (result.endpoint) {
    document.getElementById('endpoint').value = result.endpoint;
  }
  if (result.username) {
    document.getElementById('username').value = result.username;
  }
  if (result.password) {
    document.getElementById('password').value = result.password;
  }
});
