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
  id: "clip-website",
  title: "Read it later with Influss",
  contexts: ["page", "link"]
});

// Handle action clicks
browser.action.onClicked.addListener(async (tab) => {
  await clipWebsite(tab.url);
});

async function clipWebsite(url) {
  if (!endpoint) {
    return { success: false, message: 'Please configure the endpoint in settings first.' };
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
      throw new Error(`Failed to clip website (${response.status})`);
    }

    return { success: true, message: 'Successfully clipped website!' };
  } catch (error) {
    console.error('Error clipping website:', error);
    return { success: false, message: `Error: ${error.message}` };
  }
}

// Create a reusable notification system
async function showToastNotification(result) {
  // Inject CSS
  await browser.tabs.insertCSS({
    code: `
      .influss-notification {
        position: fixed;
        top: 16px;
        right: 16px;
        padding: 12px 16px;
        border-radius: 8px;
        font-family: system-ui, -apple-system, sans-serif;
        font-size: 14px;
        max-width: 300px;
        z-index: 2147483647;
        animation: influssSlideIn 0.3s ease-out, influssSlideOut 0.3s ease-in 2.7s;
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
      }

      .influss-notification.success {
        background-color: #4ade80;
        color: #052e16;
      }

      .influss-notification.error {
        background-color: #f87171;
        color: #450a0a;
      }

      @keyframes influssSlideIn {
        from { transform: translateX(100%); opacity: 0; }
        to { transform: translateX(0); opacity: 1; }
      }

      @keyframes influssSlideOut {
        from { transform: translateX(0); opacity: 1; }
        to { transform: translateX(100%); opacity: 0; }
      }
    `
  });

  // Inject and remove notification
  await browser.tabs.executeScript({
    code: `
      (function() {
        const notification = document.createElement('div');
        notification.className = 'influss-notification ' + '${result.success ? 'success' : 'error'}';
        notification.textContent = '${result.message.replace(/'/g, "\\'")}';
        document.body.appendChild(notification);

        // Remove the notification after animation
        setTimeout(() => {
          notification.remove();
        }, 3000);
      })();
    `
  });
}

// Handle context menu clicks
browser.contextMenus.onClicked.addListener(async (info, tab) => {
  if (info.menuItemId === "clip-website") {
    const urlToClip = info.linkUrl || tab.url;
    const result = await clipWebsite(urlToClip);
    await showToastNotification(result);
  }
});

// Listen for messages from popup
browser.runtime.onMessage.addListener(async (message) => {
  if (message.action === 'clipWebsite') {
    await clipWebsite(message.url);
  }
});

