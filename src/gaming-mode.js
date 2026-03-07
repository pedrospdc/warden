'use strict';

const serviceManager = require('./service-manager');
const config = require('./config');

let _config = null;
let _onChangeCallback = null;

function init(cfg, onChange) {
  _config = cfg;
  _onChangeCallback = onChange;
}

/**
 * Returns the current gaming mode state.
 * @returns {boolean}
 */
function isEnabled() {
  return _config ? _config.gamingMode : false;
}

/**
 * Enable gaming mode: stop all configured services.
 * @returns {Array} results of service stop operations
 */
function enable() {
  if (!_config) return [];
  const results = serviceManager.stopAll(_config.services);
  _config.gamingMode = true;
  config.save(_config);
  if (_onChangeCallback) _onChangeCallback(true, results);
  return results;
}

/**
 * Disable gaming mode: restart all configured services.
 * @returns {Array} results of service start operations
 */
function disable() {
  if (!_config) return [];
  const results = serviceManager.startAll(_config.services);
  _config.gamingMode = false;
  config.save(_config);
  if (_onChangeCallback) _onChangeCallback(false, results);
  return results;
}

/**
 * Toggle gaming mode on/off.
 * @returns {boolean} new state
 */
function toggle() {
  if (isEnabled()) {
    disable();
    return false;
  } else {
    enable();
    return true;
  }
}

module.exports = { init, isEnabled, enable, disable, toggle };
