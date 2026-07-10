(async function() {
    const mc = document.getElementById('mainContent');

    const resp = await api.getCaddyfile();
    const content = resp.code === 0 ? (resp.data.content || '') : '';

    mc.innerHTML = `
        <h2 class="page-title">Caddyfile 编辑器</h2>
        <div class="card">
            <textarea class="caddyfile-editor" id="caddyfileContent" spellcheck="false">${escapeHtml(content)}</textarea>
            <div style="margin-top:12px;">
                <button class="btn btn-primary" id="saveCaddyfileBtn">保存并重载</button>
                <button class="btn" id="reloadCaddyfileBtn" style="margin-left:8px;">重新读取</button>
            </div>
            <div id="caddyfileMsg" style="margin-top:8px;display:none;"></div>
        </div>
    `;

    document.getElementById('saveCaddyfileBtn').addEventListener('click', async () => {
        const content = document.getElementById('caddyfileContent').value;
        const msgEl = document.getElementById('caddyfileMsg');

        const resp = await api.saveCaddyfile(content);
        if (resp.code === 0) {
            msgEl.className = 'success-msg';
            msgEl.textContent = 'Caddyfile 保存成功，Caddy 已重载';
        } else {
            msgEl.className = 'error-msg';
            msgEl.textContent = resp.message;
        }
        msgEl.style.display = 'block';
        setTimeout(() => { msgEl.style.display = 'none'; }, 5000);
    });

    document.getElementById('reloadCaddyfileBtn').addEventListener('click', async () => {
        const resp = await api.getCaddyfile();
        if (resp.code === 0) {
            document.getElementById('caddyfileContent').value = resp.data.content || '';
        }
    });
})();
