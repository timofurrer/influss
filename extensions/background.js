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
  try {
    // Get the active tab
    const tabs = await browser.tabs.query({ active: true, currentWindow: true });
    const activeTab = tabs[0];

    // Inject the notification CSS
    await browser.scripting.insertCSS({
      target: { tabId: activeTab.id },
      css: `
        .influss-toast {
          position: fixed;
          top: 20px;
          right: 20px;
          background: ${result.success ? '#10B981' : '#EF4444'};
          color: white;
          padding: 12px 20px;
          border-radius: 8px;
          font-family: system-ui, -apple-system, sans-serif;
          font-size: 14px;
          box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
          z-index: 2147483647;
          max-width: 300px;
          opacity: 0;
          transform: translateY(-20px);
          transition: all 0.3s ease-in-out;
        }
        .influss-toast.show {
          opacity: 1;
          transform: translateY(0);
        }
      `
    });

    // Inject the notification script
    await browser.scripting.executeScript({
      target: { tabId: activeTab.id },
      func: (message) => {
        // Remove any existing toast
        const existingToast = document.querySelector('.influss-toast');
        if (existingToast) {
          existingToast.remove();
        }

        // Create new toast
        const toast = document.createElement('div');
        toast.className = 'influss-toast';
        toast.textContent = message;
        document.body.appendChild(toast);

        // Trigger animation
        setTimeout(() => toast.classList.add('show'), 10);

        // Remove after delay
        setTimeout(() => {
          toast.style.opacity = '0';
          toast.style.transform = 'translateY(-20px)';
          setTimeout(() => toast.remove(), 300);
        }, 3000);
      },
      args: [result.message]
    });
  } catch (error) {
    console.error('Error showing notification:', error);
  }
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
