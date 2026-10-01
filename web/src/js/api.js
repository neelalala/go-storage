// API Client for Go Storage

export class StorageApi {
  constructor() {
    this.baseUrl = localStorage.getItem('gs_api_url') || '';
    this.username = localStorage.getItem('gs_username') || 'new-user';
  }

  setBaseUrl(url) {
    this.baseUrl = url.replace(/\/+$/, '');
    localStorage.setItem('gs_api_url', this.baseUrl);
  }

  setUsername(username) {
    this.username = (username || '').trim();
    localStorage.setItem('gs_username', this.username);
  }

  getHeaders(extra = {}) {
    const headers = { ...extra };
    if (this.username) {
      headers['Authorization'] = `Username ${this.username}`;
    }
    return headers;
  }

  async request(path, options = {}) {
    const url = `${this.baseUrl}${path.startsWith('/') ? path : '/' + path}`;
    const opts = {
      ...options,
      headers: this.getHeaders(options.headers || {}),
    };

    const response = await fetch(url, opts);

    if (!response.ok) {
      let errorMessage = `HTTP ${response.status}: ${response.statusText}`;
      try {
        const data = await response.json();
        if (data.error && data.error.message) {
          errorMessage = data.error.message;
        }
      } catch {
        const text = await response.text().catch(() => '');
        if (text) errorMessage = text;
      }
      const err = new Error(errorMessage);
      err.status = response.status;
      throw err;
    }

    return response;
  }

  // User management
  async createUser(username) {
    const resp = await this.request(`/users/${encodeURIComponent(username)}`, {
      method: 'PUT',
    });
    return resp;
  }

  // Buckets
  async listBuckets(limit = 100, offset = 0) {
    const resp = await this.request(`/storage/?limit=${limit}&offset=${offset}`, {
      method: 'GET',
    });
    const data = await resp.json();
    const result = data?.listAllMyBucketsResult || {};
    return {
      owner: result.owner || { id: '', displayName: this.username },
      buckets: result.buckets?.bucket || [],
    };
  }

  async createBucket(bucketName) {
    const name = bucketName.trim();
    return await this.request(`/storage/${encodeURIComponent(name)}`, {
      method: 'PUT',
    });
  }

  async headBucket(bucketName) {
    const resp = await this.request(`/storage/${encodeURIComponent(bucketName)}`, {
      method: 'HEAD',
    });
    return {
      name: resp.headers.get('x-go-bucket-name') || bucketName,
      ownerId: resp.headers.get('x-go-owner-id') || '',
      creationDate: resp.headers.get('x-go-creation-date') || '',
    };
  }

  async deleteBucket(bucketName) {
    return await this.request(`/storage/${encodeURIComponent(bucketName)}`, {
      method: 'DELETE',
    });
  }

  // Objects & Filesystem
  async listObjects(bucketName, prefix = '', delimiter = '/', limit = 1000, offset = 0) {
    const params = new URLSearchParams({
      limit: String(limit),
      offset: String(offset),
    });
    if (prefix) params.set('prefix', prefix);
    if (delimiter) params.set('delimiter', delimiter);

    const resp = await this.request(`/storage/${encodeURIComponent(bucketName)}?${params.toString()}`, {
      method: 'GET',
    });
    const data = await resp.json();
    const result = data?.listBucketResult || {};

    return {
      name: result.name || bucketName,
      prefix: result.prefix || '',
      delimiter: result.delimiter || '',
      keyCount: result.keyCount || 0,
      isTruncated: Boolean(result.isTruncated),
      contents: (result.contents || []).map((item) => ({
        key: item.key,
        lastModified: item.lastModified,
        etag: (item.ETag || '').replace(/"/g, ''),
        size: Number(item.size) || 0,
        storageClass: item.storageClass || 'STANDARD',
      })),
      commonPrefixes: (result.commonPrefixes || []).map((cp) => cp.prefix),
    };
  }

  async putObject(bucketName, key, fileOrBlob, userMetadata = {}, contentType = '') {
    const ct = contentType || (fileOrBlob instanceof File && fileOrBlob.type ? fileOrBlob.type : 'application/octet-stream');
    const headers = {
      'Content-Type': ct,
    };

    // Add user metadata
    if (userMetadata && typeof userMetadata === 'object') {
      for (const [k, v] of Object.entries(userMetadata)) {
        if (!v) continue;
        const normKey = k.toLowerCase().startsWith('x-amz-meta-')
          ? k.toLowerCase()
          : `x-amz-meta-${k.toLowerCase().replace(/[^a-z0-9_-]/g, '-')}`;
        headers[normKey] = String(v);
      }
    }

    // Key can contain slashes, encode each path segment
    const encodedKey = key
      .split('/')
      .map((segment) => encodeURIComponent(segment))
      .join('/');

    return await this.request(`/storage/${encodeURIComponent(bucketName)}/${encodedKey}`, {
      method: 'PUT',
      headers,
      body: fileOrBlob,
    });
  }

  async headObject(bucketName, key) {
    const encodedKey = key
      .split('/')
      .map((segment) => encodeURIComponent(segment))
      .join('/');

    const resp = await this.request(`/storage/${encodeURIComponent(bucketName)}/${encodedKey}`, {
      method: 'HEAD',
    });

    const userMetadata = {};
    const systemMetadata = {};

    resp.headers.forEach((val, name) => {
      const lower = name.toLowerCase();
      if (lower.startsWith('x-amz-meta-')) {
        const metaKey = lower.replace('x-amz-meta-', '');
        userMetadata[metaKey] = val;
      } else if (['content-disposition', 'cache-control', 'content-encoding', 'content-language'].includes(lower)) {
        systemMetadata[lower] = val;
      }
    });

    return {
      key,
      bucket: bucketName,
      contentType: resp.headers.get('content-type') || 'application/octet-stream',
      size: Number(resp.headers.get('content-length')) || 0,
      etag: (resp.headers.get('etag') || '').replace(/"/g, ''),
      ownerId: resp.headers.get('x-go-owner-id') || '',
      storageNodeId: resp.headers.get('x-go-storage-node-id') || '',
      createdAt: resp.headers.get('x-go-created-at') || '',
      updatedAt: resp.headers.get('x-go-updated-at') || '',
      userMetadata,
      systemMetadata,
    };
  }

  async getObjectBlob(bucketName, key) {
    const encodedKey = key
      .split('/')
      .map((segment) => encodeURIComponent(segment))
      .join('/');

    const resp = await this.request(`/storage/${encodeURIComponent(bucketName)}/${encodedKey}`, {
      method: 'GET',
    });

    const blob = await resp.blob();
    const contentType = resp.headers.get('content-type') || blob.type || 'application/octet-stream';
    const etag = (resp.headers.get('etag') || '').replace(/"/g, '');

    return {
      blob,
      contentType,
      etag,
    };
  }

  async deleteObject(bucketName, key) {
    const encodedKey = key
      .split('/')
      .map((segment) => encodeURIComponent(segment))
      .join('/');

    return await this.request(`/storage/${encodeURIComponent(bucketName)}/${encodedKey}`, {
      method: 'DELETE',
    });
  }
}

export const api = new StorageApi();
