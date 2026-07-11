(function() {
    const mainContent = document.getElementById('mainContent');

    document.getElementById('menuToggle').addEventListener('click', () => {
        document.getElementById('sidebar').classList.toggle('open');
    });

    document.getElementById('logoutBtn').addEventListener('click', () => {
        api.clearToken();
        router.navigate('#/login');
    });

    document.getElementById('loginSubmit').addEventListener('click', async () => {
        const username = document.getElementById('loginUsername').value;
        const password = document.getElementById('loginPassword').value;
        const errorEl = document.getElementById('loginError');

        if (!username || !password) {
            errorEl.textContent = '请输入用户名和密码';
            errorEl.style.display = 'block';
            return;
        }

        try {
            const resp = await api.login({ username, password });
            if (resp.code === 0) {
                api.setToken(resp.data.token);
                router.navigate('#/dashboard');
            } else {
                errorEl.textContent = resp.message;
                errorEl.style.display = 'block';
            }
        } catch (e) {
            errorEl.textContent = e.message;
            errorEl.style.display = 'block';
        }
    });

    document.getElementById('setupSubmit').addEventListener('click', async () => {
        const username = document.getElementById('setupUsername').value;
        const password = document.getElementById('setupPassword').value;
        const confirm = document.getElementById('setupPasswordConfirm').value;
        const errorEl = document.getElementById('setupError');

        if (!username || !password) {
            errorEl.textContent = '请填写用户名和密码';
            errorEl.style.display = 'block';
            return;
        }
        if (password !== confirm) {
            errorEl.textContent = '两次输入的密码不一致';
            errorEl.style.display = 'block';
            return;
        }

        try {
            const resp = await api.setup({ username, password });
            if (resp.code === 0) {
                alert('初始化成功，请登录');
                router.navigate('#/login');
            } else {
                errorEl.textContent = resp.message;
                errorEl.style.display = 'block';
            }
        } catch (e) {
            errorEl.textContent = e.message;
            errorEl.style.display = 'block';
        }
    });

    function loadPage(name) {
        if (router.dashboardTimer) {
            clearInterval(router.dashboardTimer);
            router.dashboardTimer = null;
        }
        const script = document.createElement('script');
        script.src = '/js/pages/' + name + '.js';
        script.onload = () => script.remove();
        document.head.appendChild(script);
    }

    router.register('/dashboard', () => loadPage('dashboard'));
    router.register('/sites', () => loadPage('sites'));
    router.register('/sites/new', () => loadPage('site-edit'));
    router.register('/sites/:id/edit', (params) => loadPage('site-edit'));
    router.register('/caddyfile', () => loadPage('caddyfile'));
    router.register('/files', () => loadPage('files'));
    router.register('/settings', () => loadPage('settings'));

    async function checkAuth() {
        try {
            const resp = await api.getStatus();
            if (resp.code === 0) {
                document.getElementById('headerUsername').textContent = resp.data.username || '';
                if (!resp.data.initialized) {
                    router.navigate('#/setup');
                    return;
                }
                if (api.token) {
                    // 仅在当前不在已认证页面时才导航
                    const currentPath = (window.location.hash || '#/login').substring(1);
                    if (currentPath === '/login' || currentPath === '/setup') {
                        router.navigate('#/dashboard');
                    }
                } else {
                    router.navigate('#/login');
                }
            }
        } catch (e) {
            router.navigate('#/login');
        }
    }

    router.init();
    checkAuth();
})();

function formatDate(dateStr) {
    if (!dateStr) return '-';
    return dateStr.substring(0, 10);
}

function statusBadge(status) {
    const map = {
        'running': ['运行中', 'badge-success'],
        'stopped': ['已停止', 'badge-danger'],
        'unknown': ['未知', 'badge-info'],
        'valid': ['有效', 'badge-success'],
        'none': ['未申请', 'badge-info'],
        'expiring': ['即将过期', 'badge-warning'],
        'expired': ['已过期', 'badge-danger'],
    };
    const [text, cls] = map[status] || [status, 'badge-info'];
    return `<span class="badge ${cls}">${text}</span>`;
}

function escapeHtml(str) {
    if (!str) return '';
    return str.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}
