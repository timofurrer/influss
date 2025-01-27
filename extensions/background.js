let endpoint = '';
let username = '';
let password = '';

// Load stored settings
browser.storage.local.get(['endpoint', 'username', 'password']).then(items => {
  endpoint = items.endpoint || '';
  username = items.username || '';
  password = items.password || '';
});

// Listen for storage changes
browser.storage.onChanged.addListener((changes, area) => {
  if (area === 'local') {
    if (changes.endpoint) endpoint = changes.endpoint.newValue;
    if (changes.username) username = changes.username.newValue;
    if (changes.password) password = changes.password.newValue;
  }
});

// Handle action clicks
browser.action.onClicked.addListener(async (tab) => {
  await clipWebsite(tab.url);
});

async function clipWebsite(url) {
  if (!endpoint) {
    console.error('Endpoint not configured');
    return;
  }

  try {
    const response = await fetch(endpoint, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Basic ' + btoa(username + ':' + password)
      },
      body: JSON.stringify({ url: url })
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    // Android doesn't support notifications API, so we use console.log
    console.log('Successfully clipped website!');
  } catch (error) {
    console.error('Error clipping website:', error);
  }
}

// Listen for messages from popup
browser.runtime.onMessage.addListener(async (message) => {
  if (message.action === 'clipWebsite') {
    await clipWebsite(message.url);
  }
});
