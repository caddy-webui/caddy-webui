(async function() {
    const mc = document.getElementById('mainContent');

    const resp = await api.getSettings();
    const s = resp.code === 0 ? resp.data : {};

    mc.innerHTML = `
        <h2 class="page-title">全局设置</h2>

        <div class="card settings-section">
            <h3>服务设置</h3>
            <div class="form-group">
                <label>WebUI 端口（1024-65535）</label>
                <input type="number" id="settingsPort" value="${s.port || 8729}" min="1024" max="65535">
            </div>
            <div class="form-group">
                <label>日志级别</label>
                <select id="settingsLogLevel">
                    <option value="DEBUG" ${s.log_level === 'DEBUG' ? 'selected' : ''}>DEBUG</option>
                    <option value="INFO" ${s.log_level === 'INFO' ? 'selected' : ''}>INFO</option>
                    <option value="WARN" ${s.log_level === 'WARN' ? 'selected' : ''}>WARN</option>
                    <option value="ERROR" ${s.log_level === 'ERROR' ? 'selected' : ''}>ERROR</option>
                </select>
            </div>
            <button class="btn btn-primary" id="saveSettingsBtn">保存设置</button>
            <div id="settingsMsg" style="margin-top:8px;display:none;"></div>
        </div>

        <div class="card settings-section">
            <h3>修改密码</h3>
            <div class="form-group">
                <label>管理员用户名</label>
                <input type="text" value="${escapeHtml(s.username || '')}" disabled>
            </div>
            <div class="form-group">
                <label>旧密码</label>
                <input type="password" id="oldPassword">
            </div>
            <div class="form-group">
                <label>新密码（至少6位）</label>
                <input type="password" id="newPassword">
            </div>
            <button class="btn btn-primary" id="changePasswordBtn">修改密码</button>
            <div id="passwordMsg" style="margin-top:8px;display:none;"></div>
        </div>

        <div class="card settings-section">
            <h3>Caddy 服务控制</h3>
            <div style="display:flex;gap:8px;flex-wrap:wrap;">
                <button class="btn btn-success" id="caddyStartBtn">启动</button>
                <button class="btn btn-danger" id="caddyStopBtn">停止</button>
                <button class="btn btn-warning" id="caddyRestartBtn">重启</button>
                <button class="btn btn-primary" id="caddyReloadBtn">重载</button>
            </div>
            <div id="caddyMsg" style="margin-top:8px;display:none;"></div>
        </div>
    `;

    document.getElementById('saveSettingsBtn').addEventListener('click', async () => {
        const data = {
            port: parseInt(document.getElementById('settingsPort').value),
            log_level: document.getElementById('settingsLogLevel').value,
        };
        const resp = await api.updateSettings(data);
        const msgEl = document.getElementById('settingsMsg');
        if (resp.code === 0) {
            msgEl.className = 'success-msg';
            msgEl.textContent = '设置更新成功' + (resp.data.need_restart ? '。' + resp.data.restart_hint : '');
        } else {
            msgEl.className = 'error-msg';
            msgEl.textContent = resp.message;
        }
        msgEl.style.display = 'block';
    });

    document.getElementById('changePasswordBtn').addEventListener('click', async () => {
        const data = {
            old_password: document.getElementById('oldPassword').value,
            new_password: document.getElementById('newPassword').value,
        };
        const resp = await api.changePassword(data);
        const msgEl = document.getElementById('passwordMsg');
        if (resp.code === 0) {
            msgEl.className = 'success-msg';
            msgEl.textContent = '密码修改成功';
            document.getElementById('oldPassword').value = '';
            document.getElementById('newPassword').value = '';
        } else {
            msgEl.className = 'error-msg';
            msgEl.textContent = resp.message;
        }
        msgEl.style.display = 'block';
    });

    const caddyActions = [
        ['caddyStartBtn', 'caddyStop', '确定启动 Caddy？', api.caddyStart.bind(api)],
        ['caddyStopBtn', 'caddyStop', '确定停止 Caddy？这将中断所有服务！', api.caddyStop.bind(api)],
        ['caddyRestartBtn', 'caddyRestart', '确定重启 Caddy？', api.caddyRestart.bind(api)],
        ['caddyReloadBtn', 'caddyReload', '确定重载 Caddy 配置？', api.caddyReload.bind(api)],
    ];

    caddyActions.forEach(([btnId, , confirmMsg, fn]) => {
        document.getElementById(btnId).addEventListener('click', async () => {
            if (!confirm(confirmMsg)) return;
            const resp = await fn();
            const msgEl = document.getElementById('caddyMsg');
            if (resp.code === 0) {
                msgEl.className = 'success-msg';
                msgEl.textContent = resp.message;
            } else {
                msgEl.className = 'error-msg';
                msgEl.textContent = resp.message;
            }
            msgEl.style.display = 'block';
        });
    });
})();
