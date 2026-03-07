'use strict';

const fs = require('fs');
const path = require('path');

const CONFIG_FILE = path.join(__dirname, '..', 'config.json');

const DEFAULT_CONFIG = {
  gamingMode: false,
  services: [
    { name: 'Sonarr', type: 'windows-service', id: 'Sonarr', enabled: true },
    { name: 'Radarr', type: 'windows-service', id: 'Radarr', enabled: true },
    { name: 'Lidarr', type: 'windows-service', id: 'Lidarr', enabled: true },
    { name: 'Readarr', type: 'windows-service', id: 'Readarr', enabled: true },
    { name: 'Prowlarr', type: 'windows-service', id: 'Prowlarr', enabled: true },
    { name: 'Jellyfin', type: 'windows-service', id: 'JellyfinServer', enabled: true },
    { name: 'qBittorrent', type: 'process', id: 'qbittorrent', enabled: false }
  ],
  portal: {
    hotlinks: [
      { name: 'Sonarr', url: 'http://localhost:8989', icon: '📺' },
      { name: 'Radarr', url: 'http://localhost:7878', icon: '🎬' },
      { name: 'Lidarr', url: 'http://localhost:8686', icon: '🎵' },
      { name: 'Readarr', url: 'http://localhost:8787', icon: '📚' },
      { name: 'Prowlarr', url: 'http://localhost:9696', icon: '🔍' },
      { name: 'Jellyfin', url: 'http://localhost:8096', icon: '🎞️' }
    ]
  }
};

function load() {
  try {
    if (fs.existsSync(CONFIG_FILE)) {
      const raw = fs.readFileSync(CONFIG_FILE, 'utf8');
      return Object.assign({}, DEFAULT_CONFIG, JSON.parse(raw));
    }
  } catch (err) {
    console.error('Failed to load config, using defaults:', err.message);
  }
  return Object.assign({}, DEFAULT_CONFIG);
}

function save(config) {
  try {
    fs.writeFileSync(CONFIG_FILE, JSON.stringify(config, null, 2), 'utf8');
  } catch (err) {
    console.error('Failed to save config:', err.message);
  }
}

module.exports = { load, save, DEFAULT_CONFIG };
