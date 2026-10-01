import { icons, getFileIcon } from './icons.js';
import { state } from './state.js';
import { api } from './api.js';

export function formatBytes(bytes, decimals = 2) {
  if (!+bytes) return '0 B';
  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`;
}

export function formatDate(dateString) {
  if (!dateString) return '—';
  try {
    const d = new Date(dateString);
    if (isNaN(d.getTime())) return dateString;
    return d.toLocaleString(undefined, {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  } catch {
    return dateString;
  }
}

export function showToast(message, type = 'info', duration = 3500) {
  const container = document.getElementById('toast-container');
  if (!container) return;

  const toast = document.createElement('div');
  toast.className = `toast toast-${type}`;

  let iconSvg = icons.info;
  if (type === 'success') iconSvg = icons.check;
  if (type === 'error') iconSvg = icons.close;

  toast.innerHTML = `
    <div style="flex-shrink:0;">${iconSvg}</div>
    <div style="flex:1;">${escapeHtml(message)}</div>
  `;

  container.appendChild(toast);

  setTimeout(() => {
    toast.style.opacity = '0';
    toast.style.transform = 'translateY(10px)';
    toast.style.transition = 'all 0.2s ease';
    setTimeout(() => toast.remove(), 200);
  }, duration);
}

export function escapeHtml(str = '') {
  return String(str)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;');
}

export function decodeSafe(str = '') {
  if (!str) return '';
  if (str.includes('%')) {
    try {
      return decodeURIComponent(str);
    } catch { }
  }
  return str;
}

export function renderUser() {
  const nameEl = document.getElementById('current-username');
  const avatarEl = document.getElementById('user-avatar');
  if (nameEl) nameEl.textContent = state.username || 'Sign In';
  if (avatarEl) {
    const initial = state.username ? state.username.charAt(0).toUpperCase() : '?';
    avatarEl.textContent = initial;
  }
}

export function renderGatewayStatus() {
  const dot = document.getElementById('gateway-status-dot');
  const text = document.getElementById('gateway-status-text');
  const sub = document.getElementById('gateway-status-sub');
  if (!dot || !text) return;

  const currentHost = api.baseUrl ? api.baseUrl.replace(/^https?:\/\//, '') : 'localhost:8080';

  if (state.isGatewayOnline === true) {
    dot.className = 'status-dot online';
    text.textContent = 'Gateway Online';
    if (sub) sub.textContent = currentHost;
  } else if (state.isGatewayOnline === false) {
    dot.className = 'status-dot offline';
    text.textContent = 'Gateway Offline';
    if (sub) sub.textContent = 'Connection failed';
  } else {
    dot.className = 'status-dot checking';
    text.textContent = 'Connecting...';
    if (sub) sub.textContent = currentHost;
  }
}

export function renderBuckets() {
  const listEl = document.getElementById('bucket-list');
  if (!listEl) return;

  if (!state.username) {
    listEl.innerHTML = `
      <div style="padding: 16px 8px; text-align: center; color: var(--text-muted); font-size: 0.8rem;">
        Sign in to view your buckets.
      </div>
    `;
    return;
  }

  if (state.buckets.length === 0) {
    listEl.innerHTML = `
      <div style="padding: 16px 8px; text-align: center; color: var(--text-muted); font-size: 0.8rem;">
        No buckets found.<br/>Create one to start storing objects.
      </div>
    `;
    return;
  }

  listEl.innerHTML = state.buckets
    .map((bucket) => {
      const isActive = bucket.name === state.activeBucket;
      return `
      <li class="bucket-item ${isActive ? 'active' : ''}" data-bucket="${escapeHtml(bucket.name)}">
        <div class="bucket-item-left">
          ${icons.bucket}
          <span class="bucket-name" title="${escapeHtml(bucket.name)}">${escapeHtml(bucket.name)}</span>
        </div>
        <div class="bucket-actions">
          <button class="btn btn-danger-ghost btn-icon-only btn-sm delete-bucket-btn" title="Delete bucket" data-bucket="${escapeHtml(bucket.name)}">
            ${icons.trash}
          </button>
        </div>
      </li>
    `;
    })
    .join('');

  // Attach delete handlers
  listEl.querySelectorAll('.delete-bucket-btn').forEach((btn) => {
    btn.addEventListener('click', (e) => {
      e.stopPropagation();
      const bucketName = btn.dataset.bucket;
      openDeleteConfirmModal(`Delete bucket "${bucketName}"? (Bucket must be empty)`, async () => {
        try {
          await api.deleteBucket(bucketName);
          showToast(`Bucket "${bucketName}" deleted`, 'success');
          state.emit('buckets:reload');
        } catch (err) {
          showToast(err.message, 'error');
        }
      });
    });
  });

  // Attach click handlers
  listEl.querySelectorAll('.bucket-item').forEach((item) => {
    item.addEventListener('click', () => {
      const bucket = item.dataset.bucket;
      state.setActiveBucket(bucket);
    });
  });
}

export function renderBreadcrumbs() {
  const container = document.getElementById('breadcrumbs');
  if (!container) return;

  if (!state.username) {
    container.innerHTML = `<span class="breadcrumb-item active">Please sign in</span>`;
    return;
  }

  if (!state.activeBucket) {
    container.innerHTML = `<span class="breadcrumb-item active">Select a bucket</span>`;
    return;
  }

  let html = `
    <span class="breadcrumb-item root-crumb ${!state.currentPrefix ? 'active' : ''}" data-path="">
      ${icons.database}
      <span>${escapeHtml(state.activeBucket)}</span>
    </span>
  `;

  if (state.currentPrefix) {
    const segments = state.currentPrefix.split('/').filter(Boolean);
    let accum = '';

    segments.forEach((seg, idx) => {
      accum += seg + '/';
      const isLast = idx === segments.length - 1;
      const displaySeg = decodeSafe(seg);
      html += `
        <span class="breadcrumb-sep">${icons.chevronRight}</span>
        <span class="breadcrumb-item ${isLast ? 'active' : ''}" data-path="${escapeHtml(accum)}">
          ${escapeHtml(displaySeg)}
        </span>
      `;
    });
  }

  container.innerHTML = html;

  container.querySelectorAll('.breadcrumb-item:not(.active)').forEach((crumb) => {
    crumb.addEventListener('click', () => {
      const path = crumb.dataset.path;
      state.setCurrentPrefix(path);
    });
  });
}

export function renderExplorer() {
  const container = document.getElementById('explorer-container');
  if (!container) return;

  if (!state.username) {
    container.innerHTML = `
      <div class="empty-state">
        <div class="empty-icon">${icons.user}</div>
        <div class="empty-title">Welcome to Go Storage</div>
        <div class="empty-sub">Sign in or register a new user to start storing and managing files.</div>
        <button class="btn btn-primary" id="btn-empty-user-prompt" style="margin-top: 14px;">Sign In / Register</button>
      </div>
    `;
    container.querySelector('#btn-empty-user-prompt')?.addEventListener('click', () => {
      document.getElementById('btn-user-profile')?.click();
    });
    return;
  }

  if (!state.activeBucket) {
    container.innerHTML = `
      <div class="empty-state">
        <div class="empty-icon">${icons.bucket}</div>
        <div class="empty-title">No bucket selected</div>
        <div class="empty-sub">Choose a bucket from the sidebar or create a new one to browse objects.</div>
      </div>
    `;
    return;
  }

  const { folders, files } = state.getFilteredItems();

  if (folders.length === 0 && files.length === 0) {
    container.innerHTML = `
      <div class="empty-state">
        <div class="empty-icon">${icons.folderOpen}</div>
        <div class="empty-title">This folder is empty</div>
        <div class="empty-sub">
          ${state.searchQuery ? 'No objects match your search.' : 'Upload files or create subfolders using the actions above.'}
        </div>
      </div>
    `;
    return;
  }

  if (state.viewMode === 'grid') {
    renderGridView(container, folders, files);
  } else {
    renderTableView(container, folders, files);
  }
}

function renderTableView(container, folders, files) {
  let html = `
    <table class="file-table">
      <thead>
        <tr>
          <th>Name</th>
          <th>Size</th>
          <th>Last Modified</th>
          <th>ETag / MD5</th>
          <th style="text-align: right;">Actions</th>
        </tr>
      </thead>
      <tbody>
  `;

  // Up folder if in a subfolder
  if (state.currentPrefix) {
    html += `
      <tr class="clickable up-row" id="row-navigate-up">
        <td colspan="5">
          <div class="file-name-cell">
            <div class="file-icon-box" style="background: rgba(255,255,255,0.05); color: var(--text-secondary);">
              ${icons.arrowUp}
            </div>
            <span>.. (Up one level)</span>
          </div>
        </td>
      </tr>
    `;
  }

  // Folders
  folders.forEach((folderPrefix) => {
    const rawFolderName = folderPrefix.replace(state.currentPrefix, '').replace(/\/$/, '');
    const folderName = decodeSafe(rawFolderName);
    html += `
      <tr class="clickable folder-row" data-prefix="${escapeHtml(folderPrefix)}">
        <td>
          <div class="file-name-cell">
            <div class="file-icon-box">
              ${icons.folder}
            </div>
            <span>${escapeHtml(folderName)}</span>
          </div>
        </td>
        <td>—</td>
        <td>—</td>
        <td><span style="font-family: var(--font-mono); font-size: 0.75rem; color: var(--text-dim);">Directory</span></td>
        <td>
          <div class="cell-actions">
            <button class="btn btn-secondary btn-sm open-folder-btn" data-prefix="${escapeHtml(folderPrefix)}">Open</button>
          </div>
        </td>
      </tr>
    `;
  });

  // Files
  files.forEach((file) => {
    const rawFileName = file.key.replace(state.currentPrefix, '');
    const fileName = decodeSafe(rawFileName);
    const iconSvg = getFileIcon(fileName);
    const shortEtag = file.etag ? file.etag.substring(0, 8) + '...' : '—';

    html += `
      <tr class="clickable file-row" data-key="${escapeHtml(file.key)}">
        <td>
          <div class="file-name-cell">
            <div class="file-icon-box">
              ${iconSvg}
            </div>
            <span title="${escapeHtml(file.key)}">${escapeHtml(fileName)}</span>
          </div>
        </td>
        <td style="font-family: var(--font-mono);">${formatBytes(file.size)}</td>
        <td style="font-size: 0.8rem;">${formatDate(file.lastModified)}</td>
        <td>
          <span style="font-family: var(--font-mono); font-size: 0.775rem; color: var(--text-muted);" title="${escapeHtml(file.etag)}">
            ${shortEtag}
          </span>
        </td>
        <td>
          <div class="cell-actions">
            <button class="btn btn-ghost btn-icon-only btn-sm inspect-btn" title="Inspect metadata" data-key="${escapeHtml(file.key)}">
              ${icons.info}
            </button>
            <button class="btn btn-ghost btn-icon-only btn-sm download-btn" title="Download" data-key="${escapeHtml(file.key)}" data-name="${escapeHtml(fileName)}">
              ${icons.download}
            </button>
            <button class="btn btn-danger-ghost btn-icon-only btn-sm delete-file-btn" title="Delete object" data-key="${escapeHtml(file.key)}">
              ${icons.trash}
            </button>
          </div>
        </td>
      </tr>
    `;
  });

  html += `</tbody></table>`;
  container.innerHTML = html;

  // Up navigation
  const upRow = container.querySelector('#row-navigate-up');
  if (upRow) {
    upRow.addEventListener('click', () => state.navigateUp());
  }

  // Folder click navigation
  container.querySelectorAll('.folder-row').forEach((row) => {
    row.addEventListener('click', () => {
      const prefix = row.dataset.prefix;
      state.setCurrentPrefix(prefix);
    });
  });

  // File click to inspect
  container.querySelectorAll('.file-row').forEach((row) => {
    row.addEventListener('click', (e) => {
      // Don't trigger if clicked on an action button
      if (e.target.closest('button')) return;
      const key = row.dataset.key;
      openInspector(key);
    });
  });

  // Actions
  container.querySelectorAll('.inspect-btn').forEach((btn) => {
    btn.addEventListener('click', (e) => {
      e.stopPropagation();
      openInspector(btn.dataset.key);
    });
  });

  container.querySelectorAll('.download-btn').forEach((btn) => {
    btn.addEventListener('click', (e) => {
      e.stopPropagation();
      triggerDownload(state.activeBucket, btn.dataset.key, btn.dataset.name);
    });
  });

  container.querySelectorAll('.delete-file-btn').forEach((btn) => {
    btn.addEventListener('click', (e) => {
      e.stopPropagation();
      const key = btn.dataset.key;
      openDeleteConfirmModal(`Delete object "${key}"?`, async () => {
        try {
          await api.deleteObject(state.activeBucket, key);
          showToast(`Deleted ${key}`, 'success');
          state.emit('contents:reload');
        } catch (err) {
          showToast(err.message, 'error');
        }
      });
    });
  });
}

function renderGridView(container, folders, files) {
  let html = `<div class="file-grid">`;

  // Up folder
  if (state.currentPrefix) {
    html += `
      <div class="grid-card folder-card" id="grid-navigate-up">
        <div class="grid-icon" style="background: rgba(255,255,255,0.06); color: var(--text-secondary);">
          ${icons.arrowUp}
        </div>
        <div class="grid-name">.. (Up)</div>
        <div class="grid-meta">Go to parent folder</div>
      </div>
    `;
  }

  folders.forEach((folderPrefix) => {
    const rawFolderName = folderPrefix.replace(state.currentPrefix, '').replace(/\/$/, '');
    const folderName = decodeSafe(rawFolderName);
    html += `
      <div class="grid-card folder-card" data-prefix="${escapeHtml(folderPrefix)}">
        <div class="grid-icon">${icons.folder}</div>
        <div class="grid-name" title="${escapeHtml(folderName)}">${escapeHtml(folderName)}</div>
        <div class="grid-meta">Folder</div>
      </div>
    `;
  });

  files.forEach((file) => {
    const rawFileName = file.key.replace(state.currentPrefix, '');
    const fileName = decodeSafe(rawFileName);
    const iconSvg = getFileIcon(fileName);
    html += `
      <div class="grid-card file-card" data-key="${escapeHtml(file.key)}">
        <div class="grid-icon">${iconSvg}</div>
        <div class="grid-name" title="${escapeHtml(file.key)}">${escapeHtml(fileName)}</div>
        <div class="grid-meta">${formatBytes(file.size)}</div>
        <div class="grid-card-actions">
          <button class="btn btn-ghost btn-icon-only btn-sm inspect-btn" title="Inspect" data-key="${escapeHtml(file.key)}">${icons.info}</button>
          <button class="btn btn-ghost btn-icon-only btn-sm download-btn" title="Download" data-key="${escapeHtml(file.key)}" data-name="${escapeHtml(fileName)}">${icons.download}</button>
          <button class="btn btn-danger-ghost btn-icon-only btn-sm delete-file-btn" title="Delete" data-key="${escapeHtml(file.key)}">${icons.trash}</button>
        </div>
      </div>
    `;
  });

  html += `</div>`;
  container.innerHTML = html;

  const upCard = container.querySelector('#grid-navigate-up');
  if (upCard) {
    upCard.addEventListener('click', () => state.navigateUp());
  }

  container.querySelectorAll('.folder-card:not(#grid-navigate-up)').forEach((card) => {
    card.addEventListener('click', () => {
      state.setCurrentPrefix(card.dataset.prefix);
    });
  });

  container.querySelectorAll('.file-card').forEach((card) => {
    card.addEventListener('click', (e) => {
      if (e.target.closest('button')) return;
      openInspector(card.dataset.key);
    });
  });

  container.querySelectorAll('.inspect-btn').forEach((btn) => {
    btn.addEventListener('click', (e) => {
      e.stopPropagation();
      openInspector(btn.dataset.key);
    });
  });

  container.querySelectorAll('.download-btn').forEach((btn) => {
    btn.addEventListener('click', (e) => {
      e.stopPropagation();
      triggerDownload(state.activeBucket, btn.dataset.key, btn.dataset.name);
    });
  });

  container.querySelectorAll('.delete-file-btn').forEach((btn) => {
    btn.addEventListener('click', (e) => {
      e.stopPropagation();
      const key = btn.dataset.key;
      openDeleteConfirmModal(`Delete object "${key}"?`, async () => {
        try {
          await api.deleteObject(state.activeBucket, key);
          showToast(`Deleted ${key}`, 'success');
          state.emit('contents:reload');
        } catch (err) {
          showToast(err.message, 'error');
        }
      });
    });
  });
}

// Download helper
export async function triggerDownload(bucket, key, filename) {
  try {
    const downloadName = decodeSafe(filename || key.split('/').pop() || 'download');
    showToast(`Downloading ${downloadName}...`, 'info', 2000);
    const { blob } = await api.getObjectBlob(bucket, key);
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = downloadName;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    setTimeout(() => URL.revokeObjectURL(url), 10000);
  } catch (err) {
    showToast(`Download failed: ${err.message}`, 'error');
  }
}

// Inspector / Metadata Drawer
export async function openInspector(key) {
  const overlay = document.getElementById('inspector-overlay');
  const title = document.getElementById('inspector-title');
  const content = document.getElementById('inspector-content');
  const actions = document.getElementById('inspector-actions');
  if (!overlay || !content) return;

  overlay.classList.add('open');
  const displayTitle = decodeSafe(key.split('/').pop() || key);
  title.textContent = displayTitle;
  content.innerHTML = `
    <div style="display:flex;align-items:center;justify-content:center;height:200px;">
      <div class="spinner"></div>
    </div>
  `;

  try {
    const meta = await api.headObject(state.activeBucket, key);
    state.setSelectedItem(meta);

    // Build preview element
    let previewHtml = '';
    const ext = key.split('.').pop()?.toLowerCase();
    const ct = meta.contentType.toLowerCase();

    if (ct.startsWith('image/') || ['png', 'jpg', 'jpeg', 'gif', 'svg', 'webp'].includes(ext)) {
      previewHtml = `
        <div class="preview-container">
          <div class="spinner preview-loader"></div>
          <img id="inspector-preview-img" class="preview-img" style="display:none;" />
        </div>
      `;
    } else if (ct.startsWith('text/') || ['json', 'js', 'ts', 'go', 'py', 'txt', 'md', 'html', 'css', 'yaml', 'yml'].includes(ext)) {
      previewHtml = `
        <div class="preview-container">
          <div class="spinner preview-loader"></div>
          <pre id="inspector-preview-text" class="preview-text" style="display:none;"></pre>
        </div>
      `;
    } else if (ct.startsWith('audio/')) {
      previewHtml = `
        <div class="preview-container">
          <audio id="inspector-preview-audio" controls style="width:100%;"></audio>
        </div>
      `;
    } else if (ct.startsWith('video/')) {
      previewHtml = `
        <div class="preview-container">
          <video id="inspector-preview-video" controls style="max-width:100%;max-height:220px;"></video>
        </div>
      `;
    }

    // Build user tags
    const userMetaEntries = Object.entries(meta.userMetadata || {});
    let userMetaHtml = '';
    if (userMetaEntries.length > 0) {
      userMetaHtml = `
        <div class="meta-group">
          <div class="meta-group-title">${icons.tag} User Metadata (X-Amz-Meta)</div>
          <div class="meta-tags">
            ${userMetaEntries
          .map(
            ([k, v]) => `
              <div class="meta-tag">
                <strong>${escapeHtml(k)}:</strong> ${escapeHtml(v)}
              </div>
            `
          )
          .join('')}
          </div>
        </div>
      `;
    }

    // Build system tags
    const sysMetaEntries = Object.entries(meta.systemMetadata || {});
    let sysMetaHtml = '';
    if (sysMetaEntries.length > 0) {
      sysMetaHtml = `
        <div class="meta-group">
          <div class="meta-group-title">System Headers</div>
          <div class="meta-tags">
            ${sysMetaEntries
          .map(
            ([k, v]) => `
              <div class="meta-tag">
                <strong>${escapeHtml(k)}:</strong> ${escapeHtml(v)}
              </div>
            `
          )
          .join('')}
          </div>
        </div>
      `;
    }

    content.innerHTML = `
      ${previewHtml}

      <div class="meta-group">
        <div class="meta-group-title">${icons.info} Object Properties</div>
        
        <div class="meta-row">
          <span class="meta-label">Full Key (Path)</span>
          <div class="meta-val">
            <span title="${escapeHtml(meta.key)}">${escapeHtml(decodeSafe(meta.key))}</span>
            <button class="btn btn-ghost btn-sm btn-icon-only copy-key-btn" title="Copy Key">${icons.copy}</button>
          </div>
        </div>

        <div class="meta-row">
          <span class="meta-label">Bucket</span>
          <div class="meta-val">${escapeHtml(meta.bucket)}</div>
        </div>

        <div class="meta-row">
          <span class="meta-label">Size</span>
          <div class="meta-val">${formatBytes(meta.size)} (${meta.size} bytes)</div>
        </div>

        <div class="meta-row">
          <span class="meta-label">Content-Type</span>
          <div class="meta-val">${escapeHtml(meta.contentType)}</div>
        </div>

        <div class="meta-row">
          <span class="meta-label">ETag (MD5 Hash)</span>
          <div class="meta-val">${escapeHtml(meta.etag || '—')}</div>
        </div>

        <div class="meta-row">
          <span class="meta-label">Storage Node ID</span>
          <div class="meta-val" style="font-size:0.775rem;">${escapeHtml(meta.storageNodeId || '—')}</div>
        </div>

        <div class="meta-row">
          <span class="meta-label">Owner ID</span>
          <div class="meta-val" style="font-size:0.775rem;">${escapeHtml(meta.ownerId || '—')}</div>
        </div>

        <div class="meta-row">
          <span class="meta-label">Created At</span>
          <div class="meta-val">${formatDate(meta.createdAt)}</div>
        </div>

        <div class="meta-row">
          <span class="meta-label">Updated At</span>
          <div class="meta-val">${formatDate(meta.updatedAt)}</div>
        </div>
      </div>

      ${userMetaHtml}
      ${sysMetaHtml}
    `;

    // Copy key button
    content.querySelector('.copy-key-btn')?.addEventListener('click', () => {
      navigator.clipboard.writeText(meta.key);
      showToast('Key copied to clipboard', 'info');
    });

    // Actions in drawer
    actions.innerHTML = `
      <button class="btn btn-primary" id="drawer-download-btn" style="flex:1;">
        ${icons.download} Download
      </button>
      <button class="btn btn-danger-ghost" id="drawer-delete-btn">
        ${icons.trash}
      </button>
    `;

    document.getElementById('drawer-download-btn')?.addEventListener('click', () => {
      triggerDownload(meta.bucket, meta.key, meta.key.split('/').pop());
    });

    document.getElementById('drawer-delete-btn')?.addEventListener('click', () => {
      openDeleteConfirmModal(`Delete object "${meta.key}"?`, async () => {
        try {
          await api.deleteObject(meta.bucket, meta.key);
          showToast(`Deleted ${meta.key}`, 'success');
          closeInspector();
          state.emit('contents:reload');
        } catch (err) {
          showToast(err.message, 'error');
        }
      });
    });

    // Lazy load preview media if applicable
    if (previewHtml) {
      loadMediaPreview(meta.bucket, meta.key, ct, ext);
    }
  } catch (err) {
    content.innerHTML = `
      <div style="color:var(--status-danger);padding:20px;text-align:center;">
        Failed to load metadata: ${escapeHtml(err.message)}
      </div>
    `;
  }
}

async function loadMediaPreview(bucket, key, contentType, ext) {
  try {
    const { blob } = await api.getObjectBlob(bucket, key);
    const blobUrl = URL.createObjectURL(blob);

    const imgEl = document.getElementById('inspector-preview-img');
    const textEl = document.getElementById('inspector-preview-text');
    const audioEl = document.getElementById('inspector-preview-audio');
    const videoEl = document.getElementById('inspector-preview-video');
    const loader = document.querySelector('.preview-loader');

    if (imgEl) {
      imgEl.src = blobUrl;
      imgEl.onload = () => {
        imgEl.style.display = 'block';
        if (loader) loader.style.display = 'none';
      };
    } else if (textEl) {
      const text = await blob.text();
      textEl.textContent = text.slice(0, 10000) + (text.length > 10000 ? '\n... (truncated)' : '');
      textEl.style.display = 'block';
      if (loader) loader.style.display = 'none';
    } else if (audioEl) {
      audioEl.src = blobUrl;
    } else if (videoEl) {
      videoEl.src = blobUrl;
    }
  } catch (err) {
    console.warn('Failed to load preview:', err);
    const loader = document.querySelector('.preview-loader');
    if (loader) loader.style.display = 'none';
  }
}

export function closeInspector() {
  const overlay = document.getElementById('inspector-overlay');
  if (overlay) overlay.classList.remove('open');
  state.setSelectedItem(null);
}

// Confirmation Modal
let confirmCallback = null;
export function openDeleteConfirmModal(promptText, onConfirm) {
  const modal = document.getElementById('modal-delete-confirm');
  const textEl = document.getElementById('delete-confirm-text');
  if (!modal) return;
  textEl.textContent = promptText;
  confirmCallback = onConfirm;
  modal.classList.add('open');
}

export function setupModalTriggers() {
  // Close buttons on all modals
  document.querySelectorAll('.modal-overlay').forEach((overlay) => {
    overlay.addEventListener('click', (e) => {
      if (e.target === overlay) {
        overlay.classList.remove('open');
      }
    });

    overlay.querySelectorAll('.modal-close-btn, .modal-cancel-btn').forEach((btn) => {
      btn.addEventListener('click', () => {
        overlay.classList.remove('open');
      });
    });
  });

  // Confirm delete modal action
  const confirmBtn = document.getElementById('btn-confirm-delete');
  if (confirmBtn) {
    confirmBtn.addEventListener('click', async () => {
      const modal = document.getElementById('modal-delete-confirm');
      modal?.classList.remove('open');
      if (confirmCallback) {
        await confirmCallback();
        confirmCallback = null;
      }
    });
  }

  // Inspector close
  document.getElementById('inspector-close')?.addEventListener('click', closeInspector);
  document.getElementById('inspector-overlay')?.addEventListener('click', (e) => {
    if (e.target === document.getElementById('inspector-overlay')) {
      closeInspector();
    }
  });
}
