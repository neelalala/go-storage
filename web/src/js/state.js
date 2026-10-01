// State management and event hub for Go Storage Web

class State {
  constructor() {
    this.username = localStorage.getItem('gs_username') || 'new-user';
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
    this.emit('user:changed', this.username);
  }

  setBuckets(buckets) {
    this.buckets = buckets;
    if (this.buckets.length > 0 && (!this.activeBucket || !this.buckets.some((b) => b.name === this.activeBucket))) {
      this.setActiveBucket(this.buckets[0].name);
    } else if (this.buckets.length === 0) {
      this.setActiveBucket('');
    }
    this.emit('buckets:changed', this.buckets);
  }

  setActiveBucket(bucketName) {
    this.activeBucket = bucketName;
    this.currentPrefix = '';
    this.selectedItem = null;
    this.inspectorOpen = false;
    localStorage.setItem('gs_active_bucket', bucketName);
    this.emit('bucket:selected', bucketName);
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

  getFilteredItems() {
    let folders = this.commonPrefixes;
    let files = this.objects;

    if (this.searchQuery) {
      folders = folders.filter((p) => {
        const name = p.replace(this.currentPrefix, '').replace(/\/$/, '');
        return name.toLowerCase().includes(this.searchQuery);
      });
      files = files.filter((f) => {
        const name = f.key.replace(this.currentPrefix, '');
        return name.toLowerCase().includes(this.searchQuery);
      });
    }

    return { folders, files };
  }
}

export const state = new State();
