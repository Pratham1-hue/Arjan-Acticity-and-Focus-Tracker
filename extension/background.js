// StudyTracker Bridge — background service worker (Manifest V3).
//
// Keeps one long-lived Native Messaging port open to nmhost.exe
// (com.studytracker), so Chrome launches the bridge process once
// instead of once per tab event. Messages are one JSON object each:
//
//   {"type":"tab","url":"...","title":"...","incognito":false}
//
// Events that trigger a send:
//   - switching to another tab          (tabs.onActivated)
//   - a page finished loading           (tabs.onUpdated, status=complete)
//   - the browser window regains focus  (windows.onFocusChanged)
//   - browser start / extension load    (onStartup / onInstalled)
//
// When the daemon is down, connectNative fails; we retry quietly
// and re-sync the active tab as soon as it comes back.

const HOST = "com.studytracker";
const RECONNECT_DELAY_MS = 2000;

let port = null;
let reconnectTimer = null;

function connect() {
  if (port) return;
  try {
    port = chrome.runtime.connectNative(HOST);
    port.onDisconnect.addListener(() => {
      // Daemon stopped, or the host manifest is missing/misconfigured.
      port = null;
      clearTimeout(reconnectTimer);
      reconnectTimer = setTimeout(connect, RECONNECT_DELAY_MS);
    });
  } catch (e) {
    // Host not registered yet (or this extension ID is not in
    // allowed_origins). Retry; do not spam the user with errors.
    port = null;
    clearTimeout(reconnectTimer);
    reconnectTimer = setTimeout(connect, RECONNECT_DELAY_MS);
  }
}

function send(tab) {
  if (!port) connect();
  if (!port) return;
  try {
    port.postMessage({
      type: "tab",
      url: tab.url || "",
      title: tab.title || "",
      incognito: tab.incognito || false, // daemon records these as "Incognito"
    });
  } catch (e) {
    port = null; // port died mid-send; the next event reconnects
  }
}

// Send the active tab of a window (no-op if that window isn't a browser).
function currentTab(windowId) {
  chrome.tabs.query({ active: true, windowId: windowId }, (tabs) => {
    if (chrome.runtime.lastError) return;
    if (tabs && tabs.length > 0) send(tabs[0]);
  });
}

chrome.runtime.onInstalled.addListener(() => {
  connect();
  currentTab();
});

chrome.runtime.onStartup.addListener(() => {
  connect();
  currentTab();
});

chrome.tabs.onActivated.addListener((activeInfo) => {
  currentTab(activeInfo.windowId);
});

// Only when the page has finished loading, so we report the final URL
// after redirects rather than every intermediate hop. Background tabs
// are ignored: only the visible one matters.
chrome.tabs.onUpdated.addListener((tabId, changeInfo, tab) => {
  if (changeInfo.status === "complete" && tab.active) send(tab);
});

// The browser regained focus: re-sync, in case tabs changed while the
// user was in another app (the daemon keys sessions by URL, and its
// own window hook has no idea what happened inside Chrome).
chrome.windows.onFocusChanged.addListener((windowId) => {
  if (windowId !== chrome.windows.WINDOW_ID_NONE) currentTab(windowId);
});

connect();
