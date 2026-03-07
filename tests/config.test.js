'use strict';

const path = require('path');
const fs = require('fs');
const os = require('os');

// Use a temporary config file for tests
const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'warden-test-'));
const tmpConfig = path.join(tmpDir, 'config.json');

// Redirect the config module to use the temp path
jest.mock('../src/config', () => {
  const actual = jest.requireActual('../src/config');
  return {
    ...actual,
    load: () => JSON.parse(JSON.stringify(actual.DEFAULT_CONFIG)),
    save: jest.fn()
  };
});

const configModule = require('../src/config');

describe('config module', () => {
  test('DEFAULT_CONFIG has gamingMode=false', () => {
    expect(configModule.DEFAULT_CONFIG.gamingMode).toBe(false);
  });

  test('DEFAULT_CONFIG has services array', () => {
    expect(Array.isArray(configModule.DEFAULT_CONFIG.services)).toBe(true);
    expect(configModule.DEFAULT_CONFIG.services.length).toBeGreaterThan(0);
  });

  test('DEFAULT_CONFIG services include *arr apps', () => {
    const names = configModule.DEFAULT_CONFIG.services.map(s => s.name);
    expect(names).toContain('Sonarr');
    expect(names).toContain('Radarr');
    expect(names).toContain('Lidarr');
    expect(names).toContain('Readarr');
    expect(names).toContain('Prowlarr');
  });

  test('DEFAULT_CONFIG services include Jellyfin', () => {
    const names = configModule.DEFAULT_CONFIG.services.map(s => s.name);
    expect(names).toContain('Jellyfin');
  });

  test('DEFAULT_CONFIG portal has hotlinks', () => {
    const { portal } = configModule.DEFAULT_CONFIG;
    expect(Array.isArray(portal.hotlinks)).toBe(true);
    expect(portal.hotlinks.length).toBeGreaterThan(0);
  });

  test('load() returns an object with gamingMode property', () => {
    const cfg = configModule.load();
    expect(typeof cfg.gamingMode).toBe('boolean');
  });
});

afterAll(() => {
  fs.rmSync(tmpDir, { recursive: true, force: true });
});
