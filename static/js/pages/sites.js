(async function() {
    const mc = document.getElementById('mainContent');
    let page = 1;

    async function loadSites() {
        try {
            const resp = await api.getSites(page, 20);
            if (resp.code !== 0) return;
            const data = resp.data;

            let rows = '';
            data.items.forEach(site => {
                rows += `
                    <tr>
                        <td>${escapeHtml(site.domain)}</td>
                        <td>${site.enabled ? '<span class="badge badge-success">启用</span>' : '<span class="badge badge-danger">禁用</span>'}</td>
                        <td>${escapeHtml(site.proxy_target || '-')}</td>
                        <td>${statusBadge(site.cert_status)}</td>
                        <td>
                            <div class="site-list-actions">
                                <button class="btn btn-sm btn-${site.enabled ? 'warning' : 'success'}" onclick="toggleSite(${site.id}, ${!site.enabled})">${site.enabled ? '禁用' : '启用'}</button>
                                <a href="#/sites/${site.id}/edit" class="btn btn-sm btn-primary">编辑</a>
                                <button class="btn btn-sm btn-danger" onclick="deleteSite(${site.id}, '${escapeHtml(site.domain)}')">删除</button>
                            </div>
                        </td>
                    </tr>
                `;
            });

            mc.innerHTML = `
                <div class="toolbar">
                    <h2 class="page-title" style="margin-bottom:0;">站点管理</h2>
                    <a href="#/sites/new" class="btn btn-primary">新增站点</a>
                </div>
                <div class="card">
                    <table>
                        <thead>
                            <tr><th>域名</th><th>状态</th><th>代理目标</th><th>证书</th><th>操作</th></tr>
                        </thead>
                        <tbody>${rows || '<tr><td colspan="5" style="text-align:center;">暂无站点</td></tr>'}</tbody>
                    </table>
                </div>
                <div style="text-align:center;margin-top:12px;color:#95a5a6;font-size:13px;">
                    共 ${data.total} 条，第 ${data.page} / ${Math.ceil(data.total / data.page_size)} 页
                </div>
            `;
        } catch (e) {
            console.error('加载站点列表失败:', e);
        }
    }

    window.toggleSite = async function(id, enabled) {
        if (!confirm(enabled ? '确定启用该站点？' : '确定禁用该站点？')) return;
        const resp = await api.toggleSite(id, enabled);
        if (resp.code === 0) loadSites();
        else alert(resp.message);
    };

    window.deleteSite = async function(id, domain) {
        if (!confirm('确定删除站点 ' + domain + '？此操作不可恢复。')) return;
        const resp = await api.deleteSite(id);
        if (resp.code === 0) loadSites();
        else alert(resp.message);
    };

    await loadSites();
})();
