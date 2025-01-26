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

// Handle browser action clicks
browser.browserAction.onClicked.addListener(async (tab) => {
  await clipWebsite(tab.url);
});

// Handle page action clicks
browser.pageAction.onClicked.addListener(async (tab) => {
  await clipWebsite(tab.url);
});

async function clipWebsite(url) {
  if (!endpoint) {
    browser.notifications.create({
      type: 'basic',
      title: 'Website Clipper',
      message: 'Please configure the endpoint in the extension settings first.'
    });
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

    // Show success notification
    browser.notifications.create({
      type: 'basic',
      title: 'Website Clipper',
      message: 'Successfully clipped website!'
    });
  } catch (error) {
    console.error('Error clipping website:', error);
    browser.notifications.create({
      type: 'basic',
      title: 'Website Clipper',
      message: 'Error clipping website: ' + error.message
    });
  }
}
