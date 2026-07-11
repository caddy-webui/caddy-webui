const API_BASE = '';

const api = {
    token: localStorage.getItem('token'),

    setToken(token) {
        this.token = token;
        localStorage.setItem('token', token);
    },

    clearToken() {
        this.token = null;
        localStorage.removeItem('token');
    },

    async request(method, path, body, isFormData) {
        const headers = {};
        if (this.token) {
            headers['Authorization'] = 'Bearer ' + this.token;
        }
        if (!isFormData && body) {
            headers['Content-Type'] = 'application/json';
        }

        const options = { method, headers };
        if (body) {
            options.body = isFormData ? body : JSON.stringify(body);
        }

        const resp = await fetch(API_BASE + path, options);
        const data = await resp.json();

        if (resp.status === 401) {
            this.clearToken();
            if (window.location.hash !== '#/login') {
                window.location.hash = '#/login';
            }
            throw new Error('认证已过期，请重新登录');
        }

        return data;
    },

    get(path) { return this.request('GET', path); },
    post(path, body) { return this.request('POST', path, body); },
    put(path, body) { return this.request('PUT', path, body); },
    delete(path) { return this.request('DELETE', path); },
    upload(path, formData) { return this.request('POST', path, formData, true); },
    uploadPut(path, formData) { return this.request('PUT', path, formData, true); },

    async getStatus() { return this.get('/api/auth/status'); },
    async setup(data) { return this.post('/api/auth/setup', data); },
    async login(data) { return this.post('/api/auth/login', data); },
    async changePassword(data) { return this.put('/api/auth/password', data); },

    async getDashboard() { return this.get('/api/dashboard'); },
    async getSites(page, pageSize) {
        const p = new URLSearchParams();
        if (page) p.set('page', page);
        if (pageSize) p.set('page_size', pageSize);
        return this.get('/api/sites?' + p.toString());
    },
    async getSite(id) { return this.get('/api/sites/' + id); },
    async createSite(data) { return this.post('/api/sites', data); },
    async updateSite(id, data) { return this.put('/api/sites/' + id, data); },
    async deleteSite(id) { return this.delete('/api/sites/' + id); },
    async toggleSite(id, enabled) { return this.put('/api/sites/' + id + '/toggle', { enabled }); },

    async getCaddyStatus() { return this.get('/api/caddy/status'); },
    async caddyStart() { return this.post('/api/caddy/start'); },
    async caddyStop() { return this.post('/api/caddy/stop'); },
    async caddyRestart() { return this.post('/api/caddy/restart'); },
    async caddyReload() { return this.post('/api/caddy/reload'); },

    async getCertificates() { return this.get('/api/certificates'); },
    async renewCert(siteId) { return this.post('/api/certificates/' + siteId + '/renew'); },
    async uploadCert(siteId, certFile, keyFile) {
        const fd = new FormData();
        fd.append('cert_file', certFile);
        fd.append('key_file', keyFile);
        return this.upload('/api/certificates/' + siteId + '/upload', fd);
    },
    async updateCert(siteId, certFile, keyFile) {
        const fd = new FormData();
        if (certFile) fd.append('cert_file', certFile);
        if (keyFile) fd.append('key_file', keyFile);
        return this.uploadPut('/api/certificates/' + siteId + '/update', fd);
    },
    async setCertMode(siteId, mode) { return this.put('/api/certificates/' + siteId + '/mode', { cert_mode: mode }); },

    async getSettings() { return this.get('/api/settings'); },
    async updateSettings(data) { return this.put('/api/settings', data); },

    async getCaddyfile() { return this.get('/api/files/caddyfile'); },
    async saveCaddyfile(content) { return this.put('/api/files/caddyfile', { content }); },
    async uploadFile(siteDomain, file, path) {
        const fd = new FormData();
        fd.append('file', file);
        fd.append('site_domain', siteDomain);
        if (path) fd.append('path', path);
        return this.upload('/api/files/upload', fd);
    },
};
