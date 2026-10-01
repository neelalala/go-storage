// State management and event hub for Go Storage Web

class State {
  constructor() {
    this.username = localStorage.getItem('gs_username') || '';
    this.buckets = [];
    this.activeBucket = localStorage.getItem('gs_active_bucket') || '';
    this.currentPrefix = '';
    this.objects = [];
    this.commonPrefixes = [];
    this.searchQuery = '';
    this.viewMode = localStorage.getItem('gs_view_mode') || 'list'; // 'list' | 'grid'
    this.selectedItem = null;
    this.inspectorOpen = false;
    this.isLoading = false;
    this.isGatewayOnline = null; // null = checking, true = online, false = offline
    this.listeners = new Map();
  }

  on(event, callback) {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, []);
    }
    this.listeners.get(event).push(callback);
    return () => this.off(event, callback);
  }

  off(event, callback) {
    const list = this.listeners.get(event);
    if (!list) return;
    this.listeners.set(
      event,
      list.filter((cb) => cb !== callback)
    );
  }

  emit(event, payload) {
    const list = this.listeners.get(event);
    if (list) {
      list.forEach((cb) => {
        try {
          cb(payload);
        } catch (err) {
          console.error(`Error in listener for event ${event}:`, err);
        }
      });
    }
  }

  setUsername(user) {
    this.username = (user || '').trim();
    localStorage.setItem('gs_username', this.username);
    this.activeBucket = '';
    localStorage.removeItem('gs_active_bucket');
    this.emit('user:changed', this.username);
  }

  setBuckets(buckets) {
    this.buckets = buckets || [];
    // If the active bucket no longer exists in loaded buckets, reset it
    if (this.activeBucket && !this.buckets.some((b) => b.name === this.activeBucket)) {
      this.setActiveBucket('');
    }
    this.emit('buckets:changed', this.buckets);
  }

  setActiveBucket(bucketName) {
    this.activeBucket = bucketName || '';
    this.currentPrefix = '';
    this.selectedItem = null;
    this.inspectorOpen = false;
    if (this.activeBucket) {
      localStorage.setItem('gs_active_bucket', this.activeBucket);
    } else {
      localStorage.removeItem('gs_active_bucket');
    }
    this.emit('bucket:selected', this.activeBucket);
  }

  setCurrentPrefix(prefix) {
    // Normalize prefix: if non-empty, ensure it ends with '/'
    let p = prefix ? prefix.trim().replace(/^\/+/, '') : '';
    if (p && !p.endsWith('/')) {
      p += '/';
    }
    this.currentPrefix = p;
    this.emit('prefix:changed', this.currentPrefix);
  }

  navigateUp() {
    if (!this.currentPrefix) return;
    const parts = this.currentPrefix.split('/').filter(Boolean);
    parts.pop();
    const newPrefix = parts.length > 0 ? parts.join('/') + '/' : '';
    this.setCurrentPrefix(newPrefix);
  }

  setContents(objects, commonPrefixes) {
    this.objects = objects;
    this.commonPrefixes = commonPrefixes;
    this.emit('contents:changed', { objects, commonPrefixes });
  }

  setSearchQuery(q) {
    this.searchQuery = q.trim().toLowerCase();
    this.emit('search:changed', this.searchQuery);
  }

  setViewMode(mode) {
    this.viewMode = mode === 'grid' ? 'grid' : 'list';
    localStorage.setItem('gs_view_mode', this.viewMode);
    this.emit('viewmode:changed', this.viewMode);
  }

  setSelectedItem(item) {
    this.selectedItem = item;
    this.inspectorOpen = Boolean(item);
    this.emit('selection:changed', item);
  }

  setLoading(loading) {
    this.isLoading = loading;
    this.emit('loading:changed', loading);
  }

  setGatewayStatus(isOnline) {
    this.isGatewayOnline = isOnline;
    this.emit('gateway:status', isOnline);
  }

  getFilteredItems() {
    let folders = this.commonPrefixes;
    let files = this.objects;

    if (this.searchQuery) {
      folders = folders.filter((p) => {
        let name = p.replace(this.currentPrefix, '').replace(/\/$/, '');
        try {
          name = decodeURIComponent(name);
        } catch { }
        return name.toLowerCase().includes(this.searchQuery) || p.toLowerCase().includes(this.searchQuery);
      });
      files = files.filter((f) => {
        let name = f.key.replace(this.currentPrefix, '');
        try {
          name = decodeURIComponent(name);
        } catch { }
        return name.toLowerCase().includes(this.searchQuery) || f.key.toLowerCase().includes(this.searchQuery);
      });
    }

    return { folders, files };
  }
}

export const state = new State();
