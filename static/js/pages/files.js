(async function() {
    const mc = document.getElementById('mainContent');

    const sitesResp = await api.getSites(1, 100);
    if (router.currentPage !== 'files') return;
    const sites = sitesResp.code === 0 ? sitesResp.data.items : [];

    mc.innerHTML = `
        <h2 class="page-title">文件管理</h2>
        <div class="card">
            <div class="form-group">
                <label>目标站点</label>
                <select id="fileSiteDomain">
                    <option value="">请选择站点</option>
                    ${sites.map(s => `<option value="${escapeHtml(s.domain)}">${escapeHtml(s.domain)}</option>`).join('')}
                </select>
            </div>
            <div class="form-group">
                <label>目标路径（相对于站点根目录）</label>
                <input type="text" id="filePath" value="/" placeholder="/">
            </div>
            <div class="form-group">
                <label>选择文件</label>
                <input type="file" id="uploadFileInput">
            </div>
            <button class="btn btn-primary" id="uploadFileBtn">上传</button>
            <div id="uploadMsg" style="margin-top:8px;display:none;"></div>
        </div>
    `;

    document.getElementById('uploadFileBtn').addEventListener('click', async () => {
        const domain = document.getElementById('fileSiteDomain').value;
        const file = document.getElementById('uploadFileInput').files[0];
        const path = document.getElementById('filePath').value;
        const msgEl = document.getElementById('uploadMsg');

        if (!domain) {
            alert('请选择目标站点');
            return;
        }
        if (!file) {
            alert('请选择文件');
            return;
        }

        const resp = await api.uploadFile(domain, file, path);
        if (resp.code === 0) {
            msgEl.className = 'success-msg';
            msgEl.textContent = `上传成功: ${resp.data.path} (${resp.data.size} bytes)`;
        } else {
            msgEl.className = 'error-msg';
            msgEl.textContent = resp.message;
        }
        msgEl.style.display = 'block';
    });
})();
