'use strict';

const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('warden', {
  /**
   * Get current application state (gaming mode, services, hotlinks)
   * @returns {Promise<object>}
   */
  getState: () => ipcRenderer.invoke('get-state'),

  /**
   * Toggle gaming mode on/off
   * @returns {Promise<{ gamingMode: boolean, results: Array }>}
   */
  toggleGamingMode: () => ipcRenderer.invoke('toggle-gaming-mode'),

  /**
   * Enable gaming mode
   * @returns {Promise<{ gamingMode: boolean, results: Array }>}
   */
  enableGamingMode: () => ipcRenderer.invoke('enable-gaming-mode'),

  /**
   * Disable gaming mode
   * @returns {Promise<{ gamingMode: boolean, results: Array }>}
   */
  disableGamingMode: () => ipcRenderer.invoke('disable-gaming-mode'),

  /**
   * Open an external URL in the default browser
   * @param {string} url
   */
  openExternal: (url) => ipcRenderer.invoke('open-external', url),

  /**
   * Listen for gaming mode state changes pushed from main process
   * @param {function} callback
   */
  onGamingModeChanged: (callback) => {
    ipcRenderer.on('gaming-mode-changed', (_event, data) => callback(data));
  }
});
