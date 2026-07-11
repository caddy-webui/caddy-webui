(async function() {
    const mc = document.getElementById('mainContent');
    const hash = window.location.hash;
    const isEdit = hash.includes('/edit');
    const isNew = hash.includes('/new');
    let siteId = null;
    let site = null;

    if (isEdit) {
        const parts = hash.split('/');
        siteId = parts[parts.length - 2];
        const resp = await api.getSite(siteId);
        if (router.currentPage !== 'site-edit') return;
        if (resp.code === 0) site = resp.data;
    }

    const d = site || { domain: '', proxy_target: '', cert_mode: 'auto', proxy_config: { routes: [], websocket: false } };

    mc.innerHTML = `
        <h2 class="page-title">${isNew ? '新增站点' : '编辑站点'}</h2>
        <div class="card">
            <div class="form-group">
                <label>域名</label>
                <input type="text" id="siteDomain" value="${escapeHtml(d.domain)}" placeholder="example.com" ${isEdit ? '' : ''}>
            </div>
            <div class="form-group">
                <label>代理目标 URL</label>
                <input type="text" id="siteProxyTarget" value="${escapeHtml(d.proxy_target || '')}" placeholder="http://127.0.0.1:8080">
            </div>

            <h4 style="margin:16px 0 8px;">路径路由</h4>
            <div id="routesContainer"></div>
            <button class="btn btn-sm btn-primary" id="addRouteBtn">添加路由</button>

            <div class="cert-section">
                <h4>SSL 证书配置</h4>
                <div class="cert-mode-group">
                    <label><input type="radio" name="certMode" value="auto" ${d.cert_mode === 'auto' ? 'checked' : ''}> 自动申请（Let's Encrypt）</label>
                    <label><input type="radio" name="certMode" value="custom" ${d.cert_mode === 'custom' ? 'checked' : ''}> 自定义上传</label>
                </div>
                <div id="certCustomArea" style="display:${d.cert_mode === 'custom' ? 'block' : 'none'};">
                    <div class="cert-upload-area">
                        <div class="form-group">
                            <label>证书文件（.pem/.crt）</label>
                            <input type="file" id="certFile" accept=".pem,.crt">
                        </div>
                        <div class="form-group">
                            <label>私钥文件（.key）</label>
                            <input type="file" id="keyFile" accept=".key">
                        </div>
                        <button class="btn btn-sm btn-primary" id="uploadCertBtn">${isEdit && d.cert_file_path ? '更新证书' : '上传证书'}</button>
                    </div>
                    ${isEdit && d.cert_file_path ? `
                    <div class="cert-info ${d.cert_status === 'expired' ? 'cert-error' : d.cert_status === 'expiring' ? 'cert-warning' : ''}">
                        当前证书：${statusBadge(d.cert_status)} 有效期至 ${formatDate(d.cert_expires_at)}<br>
                        <small>${d.cert_file_path}</small>
                    </div>` : ''}
                </div>
                <div id="certAutoArea" style="display:${d.cert_mode === 'auto' ? 'block' : 'none'};">
                    ${isEdit && d.cert_status !== 'none' ? `
                    <div class="cert-info">
                        证书状态：${statusBadge(d.cert_status)} 有效期至 ${formatDate(d.cert_expires_at)}
                        ${d.cert_mode === 'auto' ? '<button class="btn btn-sm btn-warning" id="renewCertBtn" style="margin-left:12px;">续期证书</button>' : ''}
                    </div>` : '<p style="color:#95a5a6;font-size:13px;">启用站点后，Caddy 将自动申请 SSL 证书</p>'}
                </div>
            </div>

            <div style="margin-top:20px;">
                <button class="btn btn-primary" id="saveSiteBtn">保存</button>
                <a href="#/sites" class="btn" style="margin-left:8px;">取消</a>
            </div>
        </div>
    `;

    let routes = (d.proxy_config && d.proxy_config.routes) || [];
    renderRoutes();

    function renderRoutes() {
        const container = document.getElementById('routesContainer');
        container.innerHTML = routes.map((r, i) => `
            <div class="route-item">
                <div class="route-header">
                    <span>路由 #${i + 1}</span>
                    <button class="btn btn-sm btn-danger" onclick="removeRoute(${i})">删除</button>
                </div>
                <div class="form-group"><label>路径</label><input type="text" value="${escapeHtml(r.path || '')}" onchange="routes[${i}].path=this.value" placeholder="/api/*"></div>
                <div class="form-group"><label>后端地址（多个用空格分隔）</label><input type="text" value="${(r.backends || []).join(' ')}" onchange="routes[${i}].backends=this.value.split(' ').filter(Boolean)" placeholder="http://127.0.0.1:3000"></div>
            </div>
        `).join('');
    }

    window.removeRoute = function(i) {
        routes.splice(i, 1);
        renderRoutes();
    };

    document.getElementById('addRouteBtn').addEventListener('click', () => {
        routes.push({ path: '', backends: [], headers: {} });
        renderRoutes();
    });

    document.querySelectorAll('input[name="certMode"]').forEach(r => {
        r.addEventListener('change', function() {
            document.getElementById('certCustomArea').style.display = this.value === 'custom' ? 'block' : 'none';
            document.getElementById('certAutoArea').style.display = this.value === 'auto' ? 'block' : 'none';
        });
    });

    if (isEdit) {
        const renewBtn = document.getElementById('renewCertBtn');
        if (renewBtn) {
            renewBtn.addEventListener('click', async () => {
                const resp = await api.renewCert(siteId);
                alert(resp.code === 0 ? '续期请求已发送' : resp.message);
            });
        }

        document.getElementById('uploadCertBtn').addEventListener('click', async () => {
            const certFile = document.getElementById('certFile').files[0];
            const keyFile = document.getElementById('keyFile').files[0];
            if (!certFile || !keyFile) {
                alert('请同时选择证书文件和私钥文件');
                return;
            }
            const resp = await api.uploadCert(siteId, certFile, keyFile);
            alert(resp.code === 0 ? '证书上传成功' : resp.message);
        });
    }

    document.getElementById('saveSiteBtn').addEventListener('click', async () => {
        const data = {
            domain: document.getElementById('siteDomain').value.trim(),
            proxy_target: document.getElementById('siteProxyTarget').value.trim(),
            cert_mode: document.querySelector('input[name="certMode"]:checked').value,
            enabled: true,
            proxy_config: { routes: routes, websocket: false }
        };

        if (!data.domain) {
            alert('域名不能为空');
            return;
        }

        let resp;
        if (isNew) {
            resp = await api.createSite(data);
        } else {
            resp = await api.updateSite(siteId, data);
        }

        if (resp.code === 0) {
            router.navigate('#/sites');
        } else {
            alert(resp.message);
        }
    });
})();
