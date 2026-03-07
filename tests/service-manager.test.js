'use strict';

// Mock child_process so we don't actually run system commands during tests
jest.mock('child_process', () => ({
  execSync: jest.fn()
}));

const { execSync } = require('child_process');
const serviceManager = require('../src/service-manager');

const mockServices = [
  { name: 'Sonarr',   type: 'windows-service', id: 'Sonarr',   enabled: true },
  { name: 'Radarr',   type: 'windows-service', id: 'Radarr',   enabled: true },
  { name: 'Disabled', type: 'windows-service', id: 'Disabled', enabled: false }
];

beforeEach(() => {
  execSync.mockClear();
  execSync.mockReturnValue('');
});

describe('serviceManager.stopAll', () => {
  test('returns results for each enabled service', () => {
    const results = serviceManager.stopAll(mockServices);
    // Disabled service should be excluded
    expect(results.length).toBe(2);
    expect(results[0].name).toBe('Sonarr');
    expect(results[1].name).toBe('Radarr');
  });

  test('marks results as success when execSync does not throw', () => {
    const results = serviceManager.stopAll(mockServices);
    results.forEach(r => expect(r.success).toBe(true));
  });

  test('marks result as failed when execSync throws', () => {
    execSync.mockImplementation(() => { throw new Error('service not found'); });
    const results = serviceManager.stopAll(mockServices);
    results.forEach(r => {
      expect(r.success).toBe(false);
      expect(r.message).toMatch(/service not found/);
    });
  });

  test('skips disabled services', () => {
    serviceManager.stopAll(mockServices);
    // execSync should only be called for enabled services (2 calls)
    expect(execSync).toHaveBeenCalledTimes(2);
  });
});

describe('serviceManager.startAll', () => {
  test('returns results for each enabled service', () => {
    const results = serviceManager.startAll(mockServices);
    expect(results.length).toBe(2);
  });

  test('marks results as success when execSync succeeds', () => {
    const results = serviceManager.startAll(mockServices);
    results.forEach(r => expect(r.success).toBe(true));
  });
});

describe('serviceManager.stopService / startService', () => {
  test('stopService returns success on success', () => {
    const r = serviceManager.stopService(mockServices[0]);
    expect(r.success).toBe(true);
  });

  test('stopService returns failure on error', () => {
    execSync.mockImplementation(() => { throw new Error('access denied'); });
    const r = serviceManager.stopService(mockServices[0]);
    expect(r.success).toBe(false);
    expect(r.message).toMatch(/access denied/);
  });

  test('startService returns success on success', () => {
    const r = serviceManager.startService(mockServices[0]);
    expect(r.success).toBe(true);
  });
});
