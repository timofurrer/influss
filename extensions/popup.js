document.getElementById('clip-button').addEventListener('click', async () => {
  const tabs = await browser.tabs.query({active: true, currentWindow: true});
  const currentUrl = tabs[0].url;

  // Send message to background script
  browser.runtime.sendMessage({
    action: 'clipWebsite',
    url: currentUrl
  });

  // Close popup
  setTimeout(() => {
    window.close();
  }, 3000);
});

document.getElementById('settings-link').addEventListener('click', (e) => {
  e.preventDefault();
  browser.runtime.openOptionsPage();
});
