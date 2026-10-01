import { api } from './js/api.js';
import { state } from './js/state.js';
import { icons } from './js/icons.js';
import {
  renderUser,
  renderGatewayStatus,
  renderBuckets,
  renderBreadcrumbs,
  renderExplorer,
  showToast,
  setupModalTriggers,
} from './js/ui.js';

// Initialize Application
async function init() {
  setupIcons();
  setupModalTriggers();
  setupEventListeners();

  renderUser();
  renderGatewayStatus();
  api.checkGatewayHealth();
  setInterval(() => api.checkGatewayHealth(), 15000);

  await loadBuckets();
}

function setupIcons() {
  // Inject static icons into HTML placeholders
  document.getElementById('icon-brand')?.insertAdjacentHTML('beforeend', icons.database);
  document.getElementById('icon-upload')?.insertAdjacentHTML('beforeend', icons.upload);
  document.getElementById('icon-new-folder')?.insertAdjacentHTML('beforeend', icons.folderPlus);
  document.getElementById('icon-refresh')?.insertAdjacentHTML('beforeend', icons.refresh);
  document.getElementById('icon-new-bucket')?.insertAdjacentHTML('beforeend', icons.plus);
  document.getElementById('icon-dropzone')?.insertAdjacentHTML('beforeend', icons.upload);
  document.getElementById('icon-search')?.insertAdjacentHTML('beforeend', icons.search);
  document.getElementById('icon-view-list')?.insertAdjacentHTML('beforeend', icons.list);
  document.getElementById('icon-view-grid')?.insertAdjacentHTML('beforeend', icons.grid);
  document.getElementById('icon-settings')?.insertAdjacentHTML('beforeend', icons.server);
  document.getElementById('icon-close-inspector')?.insertAdjacentHTML('beforeend', icons.close);

  updateViewModeButtons();
}

function updateViewModeButtons() {
  const btnList = document.getElementById('btn-view-list');
  const btnGrid = document.getElementById('btn-view-grid');
  if (btnList && btnGrid) {
    if (state.viewMode === 'grid') {
      btnGrid.classList.add('btn-secondary');
      btnGrid.classList.remove('btn-ghost');
      btnList.classList.add('btn-ghost');
      btnList.classList.remove('btn-secondary');
    } else {
      btnList.classList.add('btn-secondary');
      btnList.classList.remove('btn-ghost');
      btnGrid.classList.add('btn-ghost');
      btnGrid.classList.remove('btn-secondary');
    }
  }
}

async function loadBuckets() {
  try {
    state.setLoading(true);
    const { buckets } = await api.listBuckets();
    state.setBuckets(buckets);
  } catch (err) {
    console.error('Failed to load buckets:', err);
    showToast(`Could not load buckets: ${err.message}`, 'error');
  } finally {
    state.setLoading(false);
  }
}

async function loadObjects() {
  if (!state.activeBucket) {
    state.setContents([], []);
    return;
  }

  try {
    state.setLoading(true);
    const { contents, commonPrefixes } = await api.listObjects(
      state.activeBucket,
      state.currentPrefix,
      '/', // delimiter for filesystem simulation!
      1000
    );
    state.setContents(contents, commonPrefixes);
  } catch (err) {
    console.error('Failed to load objects:', err);
    showToast(`Could not load objects: ${err.message}`, 'error');
    state.setContents([], []);
  } finally {
    state.setLoading(false);
  }
}

function setupEventListeners() {
  // State change reactions
  state.on('buckets:changed', () => renderBuckets());
  state.on('buckets:reload', () => loadBuckets());
  state.on('bucket:selected', () => {
    renderBuckets();
    renderBreadcrumbs();
    loadObjects();
  });
  state.on('prefix:changed', () => {
    renderBreadcrumbs();
    loadObjects();
  });
  state.on('contents:changed', () => renderExplorer());
  state.on('contents:reload', () => loadObjects());
  state.on('search:changed', () => renderExplorer());
  state.on('viewmode:changed', () => {
    updateViewModeButtons();
    renderExplorer();
  });
  state.on('user:changed', () => {
    renderUser();
    loadBuckets();
  });
  state.on('gateway:status', () => renderGatewayStatus());

  // Global Search
  const searchInput = document.getElementById('search-input');
  searchInput?.addEventListener('input', (e) => {
    state.setSearchQuery(e.target.value);
  });

  // View Mode Toggles
  document.getElementById('btn-view-list')?.addEventListener('click', () => state.setViewMode('list'));
  document.getElementById('btn-view-grid')?.addEventListener('click', () => state.setViewMode('grid'));

  // Refresh
  document.getElementById('btn-refresh')?.addEventListener('click', () => {
    loadObjects();
    showToast('Refreshed contents', 'info', 1500);
  });

  // User switcher modal
  document.getElementById('btn-user-profile')?.addEventListener('click', () => {
    const input = document.getElementById('input-username');
    if (input) input.value = state.username;
    document.getElementById('modal-user')?.classList.add('open');
  });

  document.getElementById('btn-switch-user')?.addEventListener('click', () => {
    const username = document.getElementById('input-username')?.value.trim();
    if (!username) {
      showToast('Please enter a username', 'error');
      return;
    }
    api.setUsername(username);
    state.setUsername(username);
    document.getElementById('modal-user')?.classList.remove('open');
    showToast(`Active user: ${username}`, 'success');
  });

  document.getElementById('btn-register-user')?.addEventListener('click', async () => {
    const username = document.getElementById('input-username')?.value.trim();
    if (!username) {
      showToast('Please enter a username to register', 'error');
      return;
    }
    try {
      await api.createUser(username);
      api.setUsername(username);
      state.setUsername(username);
      document.getElementById('modal-user')?.classList.remove('open');
      showToast(`User "${username}" registered and switched!`, 'success');
    } catch (err) {
      showToast(`Registration failed: ${err.message}`, 'error');
    }
  });

  // Create Bucket
  document.getElementById('btn-create-bucket-modal')?.addEventListener('click', () => {
    const input = document.getElementById('input-new-bucket-name');
    if (input) input.value = '';
    document.getElementById('modal-new-bucket')?.classList.add('open');
    setTimeout(() => input?.focus(), 50);
  });

  document.getElementById('btn-submit-create-bucket')?.addEventListener('click', async () => {
    const bucketName = document.getElementById('input-new-bucket-name')?.value.trim();
    if (!bucketName) {
      showToast('Please enter a bucket name', 'error');
      return;
    }
    if (!/^[a-z0-9.-]{3,63}$/.test(bucketName)) {
      showToast('Bucket name must be 3-63 chars: lowercase, numbers, dots, dashes', 'error');
      return;
    }

    try {
      await api.createBucket(bucketName);
      document.getElementById('modal-new-bucket')?.classList.remove('open');
      showToast(`Bucket "${bucketName}" created!`, 'success');
      state.setActiveBucket(bucketName);
      await loadBuckets();
    } catch (err) {
      showToast(`Create bucket failed: ${err.message}`, 'error');
    }
  });

  // Create Folder
  document.getElementById('btn-new-folder')?.addEventListener('click', () => {
    if (!state.activeBucket) {
      showToast('Select a bucket first', 'error');
      return;
    }
    const input = document.getElementById('input-new-folder-name');
    if (input) input.value = '';
    document.getElementById('modal-new-folder')?.classList.add('open');
    setTimeout(() => input?.focus(), 50);
  });

  document.getElementById('btn-submit-create-folder')?.addEventListener('click', async () => {
    const folderName = document.getElementById('input-new-folder-name')?.value.trim().replace(/\/+$/, '');
    if (!folderName) {
      showToast('Please enter a folder name', 'error');
      return;
    }

    // In object storage, create a folder marker file
    const folderKey = `${state.currentPrefix}${folderName}/.keep`;
    try {
      const emptyBlob = new Blob([''], { type: 'application/x-directory' });
      await api.putObject(state.activeBucket, folderKey, emptyBlob, { description: 'folder-marker' });
      document.getElementById('modal-new-folder')?.classList.remove('open');
      showToast(`Folder "${folderName}" created!`, 'success');
      await loadObjects();
    } catch (err) {
      showToast(`Create folder failed: ${err.message}`, 'error');
    }
  });

  // Upload Modal & Trigger
  document.getElementById('btn-upload-file')?.addEventListener('click', () => {
    if (!state.activeBucket) {
      showToast('Select a bucket first', 'error');
      return;
    }
    openUploadModal();
  });

  // Dropzone click
  const dropzone = document.getElementById('file-dropzone');
  const fileInputHidden = document.getElementById('file-input-hidden');

  dropzone?.addEventListener('click', () => {
    if (!state.activeBucket) {
      showToast('Select a bucket first', 'error');
      return;
    }
    fileInputHidden?.click();
  });

  fileInputHidden?.addEventListener('change', (e) => {
    if (e.target.files && e.target.files.length > 0) {
      openUploadModal(e.target.files[0]);
    }
  });

  // Drag and drop handlers
  ['dragenter', 'dragover'].forEach((eventName) => {
    window.addEventListener(eventName, (e) => {
      e.preventDefault();
      e.stopPropagation();
      dropzone?.classList.add('drag-active');
    });
  });

  ['dragleave', 'drop'].forEach((eventName) => {
    window.addEventListener(eventName, (e) => {
      e.preventDefault();
      e.stopPropagation();
      dropzone?.classList.remove('drag-active');
    });
  });

  window.addEventListener('drop', (e) => {
    if (!state.activeBucket) return;
    const files = e.dataTransfer?.files;
    if (files && files.length > 0) {
      openUploadModal(files[0]);
    }
  });

  // Settings Modal
  document.getElementById('btn-settings')?.addEventListener('click', () => {
    const input = document.getElementById('input-api-url');
    if (input) input.value = api.baseUrl;
    document.getElementById('modal-settings')?.classList.add('open');
  });

  document.getElementById('btn-save-settings')?.addEventListener('click', () => {
    const val = document.getElementById('input-api-url')?.value.trim() || '';
    api.setBaseUrl(val);
    document.getElementById('modal-settings')?.classList.remove('open');
    showToast('API Endpoint settings saved', 'success');
    loadBuckets();
  });

  // Upload Submission
  document.getElementById('btn-submit-upload')?.addEventListener('click', handleUploadSubmit);

  // Add metadata row in upload modal
  document.getElementById('btn-add-meta-field')?.addEventListener('click', () => {
    const container = document.getElementById('upload-meta-fields');
    if (!container) return;
    const row = document.createElement('div');
    row.className = 'meta-input-row';
    row.innerHTML = `
      <input type="text" class="form-input meta-key" placeholder="Key (e.g. author)" />
      <input type="text" class="form-input meta-val" placeholder="Value (e.g. John)" />
      <button type="button" class="btn btn-ghost btn-icon-only btn-sm remove-meta-row" title="Remove">${icons.close}</button>
    `;
    container.appendChild(row);
    row.querySelector('.remove-meta-row')?.addEventListener('click', () => row.remove());
  });

  // Keyboard shortcut: ESC to close modals & drawers
  window.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
      document.querySelectorAll('.modal-overlay.open').forEach((m) => m.classList.remove('open'));
      document.getElementById('inspector-overlay')?.classList.remove('open');
    }
  });
}

let pendingUploadFile = null;

function openUploadModal(file = null) {
  pendingUploadFile = file;
  const modal = document.getElementById('modal-upload');
  const fileInfo = document.getElementById('upload-file-info');
  const keyInput = document.getElementById('input-upload-key');
  const pathPrefixSpan = document.getElementById('upload-path-prefix');
  const metaContainer = document.getElementById('upload-meta-fields');
  const progressBar = document.getElementById('upload-progress-bar');
  const submitBtn = document.getElementById('btn-submit-upload');

  if (progressBar) progressBar.style.display = 'none';
  if (submitBtn) submitBtn.disabled = false;
  if (metaContainer) metaContainer.innerHTML = '';

  if (pathPrefixSpan) {
    pathPrefixSpan.textContent = state.currentPrefix ? `${state.currentPrefix}` : '/';
  }

  if (file) {
    fileInfo.innerHTML = `
      <div style="font-weight: 500; color: var(--text-primary);">${file.name}</div>
      <div style="font-size: 0.75rem; color: var(--text-muted);">${(file.size / 1024).toFixed(1)} KB • ${file.type || 'unknown type'}</div>
    `;
    if (keyInput) keyInput.value = file.name;
  } else {
    fileInfo.innerHTML = `
      <input type="file" id="upload-modal-file-picker" class="form-input" style="width: 100%;" />
    `;
    if (keyInput) keyInput.value = '';
    const picker = document.getElementById('upload-modal-file-picker');
    picker?.addEventListener('change', (e) => {
      if (e.target.files && e.target.files[0]) {
        pendingUploadFile = e.target.files[0];
        if (keyInput && !keyInput.value) {
          keyInput.value = pendingUploadFile.name;
        }
      }
    });
  }

  modal?.classList.add('open');
}

async function handleUploadSubmit() {
  if (!pendingUploadFile) {
    showToast('Please select a file to upload', 'error');
    return;
  }

  const keyInput = document.getElementById('input-upload-key');
  const filename = keyInput?.value.trim() || pendingUploadFile.name;
  if (!filename) {
    showToast('Please specify a valid object key/name', 'error');
    return;
  }

  const fullKey = `${state.currentPrefix}${filename}`;

  // Gather user metadata
  const userMetadata = {};
  document.querySelectorAll('#upload-meta-fields .meta-input-row').forEach((row) => {
    const k = row.querySelector('.meta-key')?.value.trim();
    const v = row.querySelector('.meta-val')?.value.trim();
    if (k && v) {
      userMetadata[k] = v;
    }
  });

  const progressBar = document.getElementById('upload-progress-bar');
  const progressFill = document.getElementById('upload-progress-fill');
  const submitBtn = document.getElementById('btn-submit-upload');

  if (progressBar) progressBar.style.display = 'block';
  if (progressFill) progressFill.style.width = '50%';
  if (submitBtn) submitBtn.disabled = true;

  try {
    await api.putObject(state.activeBucket, fullKey, pendingUploadFile, userMetadata, pendingUploadFile.type);
    if (progressFill) progressFill.style.width = '100%';

    showToast(`Uploaded "${filename}" successfully!`, 'success');
    document.getElementById('modal-upload')?.classList.remove('open');
    pendingUploadFile = null;
    await loadObjects();
  } catch (err) {
    showToast(`Upload failed: ${err.message}`, 'error');
  } finally {
    if (submitBtn) submitBtn.disabled = false;
    if (progressBar) progressBar.style.display = 'none';
  }
}

// Start application
window.addEventListener('DOMContentLoaded', init);
