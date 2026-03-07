'use strict';

// Mock service-manager so we don't execute system commands
jest.mock('../src/service-manager', () => ({
  stopAll:  jest.fn(() => [{ name: 'Sonarr', success: true, message: 'Sonarr stopped.' }]),
  startAll: jest.fn(() => [{ name: 'Sonarr', success: true, message: 'Sonarr started.' }])
}));

// Mock config module
jest.mock('../src/config', () => ({
  load: jest.fn(() => ({
    gamingMode: false,
    services: [
      { name: 'Sonarr', type: 'windows-service', id: 'Sonarr', enabled: true }
    ],
    portal: { hotlinks: [] }
  })),
  save: jest.fn()
}));

const gamingMode = require('../src/gaming-mode');
const configModule = require('../src/config');
const serviceManager = require('../src/service-manager');

describe('gaming-mode module', () => {
  let cfg;

  beforeEach(() => {
    jest.resetModules();
    // Re-require to get a fresh module state
    const freshGamingMode = jest.requireActual('../src/gaming-mode');
    cfg = { gamingMode: false, services: [{ name: 'Sonarr', type: 'windows-service', id: 'Sonarr', enabled: true }], portal: { hotlinks: [] } };
    freshGamingMode.init(cfg, jest.fn());
    // Use the fresh module for this describe block
  });

  test('isEnabled() returns false initially', () => {
    const fresh = jest.requireActual('../src/gaming-mode');
    const cfg2 = { gamingMode: false, services: [], portal: { hotlinks: [] } };
    fresh.init(cfg2, jest.fn());
    expect(fresh.isEnabled()).toBe(false);
  });

  test('enable() sets gamingMode to true in config', () => {
    const fresh = jest.requireActual('../src/gaming-mode');
    const cfg2 = { gamingMode: false, services: [], portal: { hotlinks: [] } };
    fresh.init(cfg2, jest.fn());
    fresh.enable();
    expect(cfg2.gamingMode).toBe(true);
  });

  test('disable() sets gamingMode to false in config', () => {
    const fresh = jest.requireActual('../src/gaming-mode');
    const cfg2 = { gamingMode: true, services: [], portal: { hotlinks: [] } };
    fresh.init(cfg2, jest.fn());
    fresh.disable();
    expect(cfg2.gamingMode).toBe(false);
  });

  test('toggle() flips gaming mode from false to true', () => {
    const fresh = jest.requireActual('../src/gaming-mode');
    const cfg2 = { gamingMode: false, services: [], portal: { hotlinks: [] } };
    fresh.init(cfg2, jest.fn());
    const result = fresh.toggle();
    expect(result).toBe(true);
    expect(cfg2.gamingMode).toBe(true);
  });

  test('toggle() flips gaming mode from true to false', () => {
    const fresh = jest.requireActual('../src/gaming-mode');
    const cfg2 = { gamingMode: true, services: [], portal: { hotlinks: [] } };
    fresh.init(cfg2, jest.fn());
    const result = fresh.toggle();
    expect(result).toBe(false);
    expect(cfg2.gamingMode).toBe(false);
  });

  test('enable() calls the onChange callback', () => {
    const fresh = jest.requireActual('../src/gaming-mode');
    const onChange = jest.fn();
    const cfg2 = { gamingMode: false, services: [], portal: { hotlinks: [] } };
    fresh.init(cfg2, onChange);
    fresh.enable();
    expect(onChange).toHaveBeenCalledWith(true, expect.any(Array));
  });

  test('disable() calls the onChange callback', () => {
    const fresh = jest.requireActual('../src/gaming-mode');
    const onChange = jest.fn();
    const cfg2 = { gamingMode: true, services: [], portal: { hotlinks: [] } };
    fresh.init(cfg2, onChange);
    fresh.disable();
    expect(onChange).toHaveBeenCalledWith(false, expect.any(Array));
  });
});
