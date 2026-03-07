'use strict';

const { execSync } = require('child_process');

/**
 * Determines the current platform.
 * @returns {'win32'|'linux'|'darwin'|string}
 */
function getPlatform() {
  return process.platform;
}

/**
 * Stop a Windows service or Linux systemd unit.
 * @param {object} service - service descriptor from config
 * @returns {{ success: boolean, message: string }}
 */
function stopService(service) {
  try {
    if (getPlatform() === 'win32') {
      if (service.type === 'windows-service') {
        execSync(`net stop "${service.id}"`, { stdio: 'pipe' });
      } else if (service.type === 'process') {
        execSync(`taskkill /IM "${service.id}.exe" /F`, { stdio: 'pipe' });
      }
    } else {
      // Linux / macOS – use systemctl or kill
      if (service.type === 'windows-service') {
        execSync(`systemctl stop "${service.id}"`, { stdio: 'pipe' });
      } else if (service.type === 'process') {
        execSync(`pkill -f "${service.id}" || true`, { stdio: 'pipe' });
      }
    }
    return { success: true, message: `${service.name} stopped.` };
  } catch (err) {
    return { success: false, message: `Failed to stop ${service.name}: ${err.message}` };
  }
}

/**
 * Start a Windows service or Linux systemd unit.
 * @param {object} service - service descriptor from config
 * @returns {{ success: boolean, message: string }}
 */
function startService(service) {
  try {
    if (getPlatform() === 'win32') {
      if (service.type === 'windows-service') {
        execSync(`net start "${service.id}"`, { stdio: 'pipe' });
      } else if (service.type === 'process') {
        execSync(`start "" "${service.id}.exe"`, { stdio: 'pipe' });
      }
    } else {
      if (service.type === 'windows-service') {
        execSync(`systemctl start "${service.id}"`, { stdio: 'pipe' });
      } else if (service.type === 'process') {
        execSync(`${service.id} &`, { stdio: 'pipe', shell: true });
      }
    }
    return { success: true, message: `${service.name} started.` };
  } catch (err) {
    return { success: false, message: `Failed to start ${service.name}: ${err.message}` };
  }
}

/**
 * Stop all enabled services in the config list.
 * @param {Array} services
 * @returns {Array<{ name: string, success: boolean, message: string }>}
 */
function stopAll(services) {
  return services
    .filter(s => s.enabled)
    .map(s => ({ name: s.name, ...stopService(s) }));
}

/**
 * Start all enabled services in the config list.
 * @param {Array} services
 * @returns {Array<{ name: string, success: boolean, message: string }>}
 */
function startAll(services) {
  return services
    .filter(s => s.enabled)
    .map(s => ({ name: s.name, ...startService(s) }));
}

module.exports = { stopService, startService, stopAll, startAll };
