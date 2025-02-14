// Handle button clicks (important for mobile)
browser.action.onClicked.addListener((tab) => {
  saveForLater(tab.url);
});

browser.runtime.onInstalled.addListener(() => {
  // Create context menu item
  browser.contextMenus.create({
    id: "read-it-later-influss",
    title: "Read it later (influss)",
    contexts: ["page", "link"]
  });
});

// Handle context menu clicks
browser.contextMenus.onClicked.addListener((info, tab) => {
  if (info.menuItemId === "read-it-later-influss") {
    saveForLater(info.linkUrl || tab.url);
  }
});

// Handle messages from the page action
browser.runtime.onMessage.addListener((message) => {
  if (message.action === "saveCurrentPage") {
    browser.tabs.query({active: true, currentWindow: true})
      .then(tabs => {
        if (tabs[0]) {
          saveForLater(tabs[0].url);
        }
      });
  }
});

async function saveForLater(url) {
  try {
    // Get the endpoint and auth details from storage
    const { endpoint, username, password } = await browser.storage.sync.get(['endpoint', 'username', 'password']);

    if (!endpoint) {
      console.error('Endpoint not configured');
      return;
    }

    const authHeader = 'Basic ' + btoa(`${username}:${password}`);

    const response = await fetch(endpoint, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': authHeader
      },
      body: JSON.stringify({url: url})
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
  } catch (error) {
    console.error('Error saving page:', error);
  }
}
