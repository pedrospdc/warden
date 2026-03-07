'use strict';

const {
  app,
  Tray,
  Menu,
  BrowserWindow,
  ipcMain,
  shell,
  nativeImage
} = require('electron');

const path = require('path');
const fs = require('fs');

const configModule = require('./src/config');
const gamingMode = require('./src/gaming-mode');

// Prevent multiple instances
if (!app.requestSingleInstanceLock()) {
  app.quit();
}

let tray = null;
let portalWindow = null;
let appConfig = null;

// ---------- Icon helpers ----------

function loadIcon(filename) {
  const iconPath = path.join(__dirname, 'assets', filename);
  if (fs.existsSync(iconPath)) {
    const img = nativeImage.createFromPath(iconPath);
    if (!img.isEmpty()) return img;
  }
  // Fallback: programmatically create a simple 16×16 coloured square
  return createFallbackIcon(filename.includes('gaming') ? '#e02828' : '#1e64dc');
}

function createFallbackIcon(color) {
  // Build a tiny 16×16 PNG from scratch using raw bytes
  const size = 16;
  const [r, g, b] = hexToRgb(color);
  const raw = Buffer.alloc(size * (1 + size * 3));
  for (let y = 0; y < size; y++) {
    raw[y * (1 + size * 3)] = 0; // filter byte
    for (let x = 0; x < size; x++) {
      const offset = y * (1 + size * 3) + 1 + x * 3;
      raw[offset] = r;
      raw[offset + 1] = g;
      raw[offset + 2] = b;
    }
  }
  const png = buildPng(raw, size, size);
  return nativeImage.createFromBuffer(png);
}

function hexToRgb(hex) {
  const m = hex.replace('#', '').match(/.{2}/g);
  return m.map(h => parseInt(h, 16));
}

function buildPng(rawPixels, width, height) {
  const zlib = require('zlib');

  const makeChunk = (type, data) => {
    const buf = Buffer.alloc(12 + data.length);
    buf.writeUInt32BE(data.length, 0);
    buf.write(type, 4, 'ascii');
    data.copy(buf, 8);
    const crc = crc32(Buffer.concat([Buffer.from(type, 'ascii'), data]));
    buf.writeUInt32BE(crc, 8 + data.length);
    return buf;
  };

  const sig = Buffer.from('\x89PNG\r\n\x1a\n', 'binary');

  const ihdrData = Buffer.alloc(13);
  ihdrData.writeUInt32BE(width, 0);
  ihdrData.writeUInt32BE(height, 4);
  ihdrData[8] = 8;  // bit depth
  ihdrData[9] = 2;  // color type: RGB
  const ihdr = makeChunk('IHDR', ihdrData);

  const compressed = zlib.deflateSync(rawPixels);
  const idat = makeChunk('IDAT', compressed);
  const iend = makeChunk('IEND', Buffer.alloc(0));

  return Buffer.concat([sig, ihdr, idat, iend]);
}

function crc32(buf) {
  const table = crc32.table || (crc32.table = (() => {
    const t = new Uint32Array(256);
    for (let i = 0; i < 256; i++) {
      let c = i;
      for (let j = 0; j < 8; j++) c = (c & 1) ? 0xedb88320 ^ (c >>> 1) : c >>> 1;
      t[i] = c;
    }
    return t;
  })());
  let crc = 0xffffffff;
  for (const byte of buf) crc = table[(crc ^ byte) & 0xff] ^ (crc >>> 8);
  return (crc ^ 0xffffffff) >>> 0;
}

// ---------- Tray ----------

function buildContextMenu() {
  const gaming = gamingMode.isEnabled();
  return Menu.buildFromTemplate([
    {
      label: gaming ? '🎮 Gaming Mode: ON' : '🖥️  Gaming Mode: OFF',
      enabled: false
    },
    { type: 'separator' },
    {
      label: gaming ? 'Disable Gaming Mode' : 'Enable Gaming Mode',
      click: () => {
        gamingMode.toggle();
        updateTray();
      }
    },
    { type: 'separator' },
    {
      label: 'Open Portal…',
      click: () => openPortal()
    },
    { type: 'separator' },
    {
      label: 'Quit Warden',
      click: () => app.quit()
    }
  ]);
}

function updateTray() {
  if (!tray) return;
  const gaming = gamingMode.isEnabled();
  tray.setImage(loadIcon(gaming ? 'icon-gaming.png' : 'icon-normal.png'));
  tray.setToolTip(gaming ? 'Warden – Gaming Mode ON 🎮' : 'Warden – Gaming Mode OFF');
  tray.setContextMenu(buildContextMenu());

  // Notify any open portal window
  if (portalWindow && !portalWindow.isDestroyed()) {
    portalWindow.webContents.send('gaming-mode-changed', { gamingMode: gaming });
  }
}

function createTray() {
  tray = new Tray(loadIcon('icon-normal.png'));
  tray.setToolTip('Warden – Gaming Mode OFF');
  tray.setContextMenu(buildContextMenu());

  // Left-click on the tray icon toggles gaming mode
  tray.on('click', () => {
    gamingMode.toggle();
    updateTray();
  });
}

// ---------- Portal window ----------

function openPortal() {
  if (portalWindow && !portalWindow.isDestroyed()) {
    portalWindow.focus();
    return;
  }

  portalWindow = new BrowserWindow({
    width: 900,
    height: 650,
    title: 'Warden – Portal',
    icon: loadIcon('icon-normal.png'),
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      contextIsolation: true,
      nodeIntegration: false
    }
  });

  portalWindow.loadFile(path.join(__dirname, 'renderer', 'index.html'));

  portalWindow.on('closed', () => {
    portalWindow = null;
  });
}

// ---------- IPC handlers ----------

ipcMain.handle('get-state', () => {
  return {
    gamingMode: gamingMode.isEnabled(),
    services: appConfig.services,
    hotlinks: appConfig.portal.hotlinks
  };
});

ipcMain.handle('toggle-gaming-mode', () => {
  const newState = gamingMode.toggle();
  updateTray();
  return { gamingMode: newState };
});

ipcMain.handle('enable-gaming-mode', () => {
  const results = gamingMode.enable();
  updateTray();
  return { gamingMode: true, results };
});

ipcMain.handle('disable-gaming-mode', () => {
  const results = gamingMode.disable();
  updateTray();
  return { gamingMode: false, results };
});

ipcMain.handle('open-external', (_event, url) => {
  // Only allow http/https URLs
  if (/^https?:\/\//.test(url)) {
    shell.openExternal(url);
  }
});

// ---------- App lifecycle ----------

app.whenReady().then(() => {
  // Hide from macOS dock; on Windows there is no dock equivalent
  if (app.dock) app.dock.hide();

  appConfig = configModule.load();

  gamingMode.init(appConfig, (enabled) => {
    updateTray();
  });

  createTray();
});

app.on('window-all-closed', () => {
  // Keep the app running in the tray even when all windows are closed
});

app.on('second-instance', () => {
  // If a second instance is started, open the portal instead
  openPortal();
});
