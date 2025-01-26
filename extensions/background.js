// Configuration
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

// Create context menu item
browser.contextMenus.create({
  id: "influss-clip-website",
  title: "Read it later with Influss",
  contexts: ["page"]
});

// Handle context menu clicks
browser.contextMenus.onClicked.addListener((info, tab) => {
  if (info.menuItemId === "influss-clip-website") {
    clipWebsite(tab.url);
  }
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
  } catch (error) {
    console.error('Error clipping website:', error);
  }
}
