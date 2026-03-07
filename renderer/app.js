'use strict';

/* global warden */

// ---------- State ----------
let currentState = { gamingMode: false, services: [], hotlinks: [] };

// ---------- DOM helpers ----------
const $ = id => document.getElementById(id);

// ---------- UI update ----------
function applyState(state) {
  currentState = state;
  const gaming = state.gamingMode;

  // Body class
  document.body.classList.toggle('gaming', gaming);

  // Header badge
  $('gamingBadgeIcon').textContent = gaming ? '🎮' : '🖥️';
  $('gamingBadgeText').textContent = gaming ? 'Gaming Mode ON' : 'Gaming Mode OFF';

  // Toggle switch
  $('gamingToggle').checked = gaming;
  $('toggleLabel').textContent = gaming ? 'ON' : 'OFF';

  // Shield emoji in header
  $('headerShield').textContent = gaming ? '⚔️' : '🛡️';

  // Description
  $('gamingDesc').textContent = gaming
    ? 'Gaming mode is active. Heavy background processes have been stopped.'
    : 'Disable heavy background processes to boost gaming performance.';

  // Services list
  renderServiceList(state.services, gaming);

  // Hotlinks
  renderHotlinks(state.hotlinks);
}

function renderServiceList(services, gamingOn) {
  const list = $('serviceList');
  if (!services || services.length === 0) {
    list.innerHTML = '<li class="loading">No services configured.</li>';
    return;
  }
  list.innerHTML = services.map(svc => {
    const dot = svc.enabled
      ? (gamingOn ? 'stopped' : 'active')
      : '';
    const statusLabel = !svc.enabled
      ? 'disabled in config'
      : (gamingOn ? 'stopped' : 'running');
    return `
      <li class="service-item">
        <span class="service-dot ${dot}" title="${statusLabel}"></span>
        <span class="service-name">${escHtml(svc.name)}</span>
        <span class="service-type">${escHtml(svc.type)}</span>
      </li>`;
  }).join('');
}

function renderHotlinks(hotlinks) {
  const grid = $('hotlinks');
  if (!hotlinks || hotlinks.length === 0) {
    grid.innerHTML = '<div class="loading">No links configured.</div>';
    return;
  }
  grid.innerHTML = hotlinks.map(link => `
    <div class="hotlink-btn" data-url="${escAttr(link.url)}" role="button" tabindex="0"
         aria-label="Open ${escAttr(link.name)}">
      <span class="hotlink-icon">${link.icon || '🔗'}</span>
      <span class="hotlink-name">${escHtml(link.name)}</span>
    </div>
  `).join('');

  grid.querySelectorAll('.hotlink-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      const url = btn.dataset.url;
      if (url) warden.openExternal(url);
    });
    btn.addEventListener('keydown', e => {
      if (e.key === 'Enter' || e.key === ' ') {
        e.preventDefault();
        btn.click();
      }
    });
  });
}

function showServiceResults(results) {
  const container = $('serviceResults');
  if (!results || results.length === 0) {
    container.classList.add('hidden');
    return;
  }
  container.classList.remove('hidden');
  container.innerHTML = results.map(r => `
    <div class="result-item ${r.success ? 'ok' : 'fail'}">
      ${r.success ? '✔' : '✖'} ${escHtml(r.name)}: ${escHtml(r.message || '')}
    </div>
  `).join('');
  // Auto-hide after 5 s
  setTimeout(() => container.classList.add('hidden'), 5000);
}

// ---------- Security helpers ----------
function escHtml(str) {
  return String(str)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}

function escAttr(str) {
  return escHtml(str);
}

// ---------- Event wiring ----------
$('gamingToggle').addEventListener('change', async () => {
  $('gamingToggle').disabled = true;
  try {
    let res;
    if ($('gamingToggle').checked) {
      res = await warden.enableGamingMode();
    } else {
      res = await warden.disableGamingMode();
    }
    applyState({ ...currentState, gamingMode: res.gamingMode });
    if (res.results) showServiceResults(res.results);
  } catch (err) {
    console.error('Failed to toggle gaming mode:', err);
    $('gamingToggle').checked = !$('gamingToggle').checked;
  } finally {
    $('gamingToggle').disabled = false;
  }
});

// Listen for state changes pushed from the main process (e.g. tray left-click)
warden.onGamingModeChanged(data => {
  applyState({ ...currentState, gamingMode: data.gamingMode });
});

// ---------- Bootstrap ----------
(async () => {
  try {
    const state = await warden.getState();
    applyState(state);
  } catch (err) {
    console.error('Failed to load state:', err);
  }
})();
