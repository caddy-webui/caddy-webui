(async function() {
    const mc = document.getElementById('mainContent');

    async function loadDashboard() {
        try {
            const resp = await api.getDashboard();
            if (resp.code !== 0) return;
            const d = resp.data;

            mc.innerHTML = `
                <h2 class="page-title">仪表盘</h2>
                <div class="grid-4" style="margin-bottom:20px;">
                    <div class="card">
                        <div class="card-title">CPU 使用率</div>
                        <div class="card-value">${d.system.cpu_usage !== null ? d.system.cpu_usage.toFixed(1) + '%' : 'N/A'}</div>
                    </div>
                    <div class="card">
                        <div class="card-title">内存使用率</div>
                        <div class="card-value">${d.system.memory_usage !== null ? d.system.memory_usage.toFixed(1) + '%' : 'N/A'}</div>
                        <div style="font-size:12px;color:#95a5a6;margin-top:4px;">
                            ${d.system.memory_used_mb !== null ? d.system.memory_used_mb.toFixed(0) + ' / ' + d.system.memory_total_mb.toFixed(0) + ' MB' : ''}
                        </div>
                    </div>
                    <div class="card">
                        <div class="card-title">磁盘使用率</div>
                        <div class="card-value">${d.system.disk_usage !== null ? d.system.disk_usage.toFixed(1) + '%' : 'N/A'}</div>
                        <div style="font-size:12px;color:#95a5a6;margin-top:4px;">
                            ${d.system.disk_used_gb !== null ? d.system.disk_used_gb.toFixed(1) + ' / ' + d.system.disk_total_gb.toFixed(1) + ' GB' : ''}
                        </div>
                    </div>
                    <div class="card">
                        <div class="card-title">Caddy 状态</div>
                        <div class="card-value status-${d.caddy.status}">${d.caddy.status === 'running' ? '运行中' : d.caddy.status === 'stopped' ? '已停止' : '未知'}</div>
                        <div style="font-size:12px;color:#95a5a6;margin-top:4px;">v${d.caddy.version}</div>
                    </div>
                </div>
                <div class="grid-2">
                    <div class="card">
                        <div class="card-title">站点统计</div>
                        <div style="display:flex;gap:20px;margin-top:8px;">
                            <div><span style="font-size:24px;font-weight:600;">${d.sites.total}</span><br><span style="font-size:12px;color:#95a5a6;">总数</span></div>
                            <div><span style="font-size:24px;font-weight:600;color:#27ae60;">${d.sites.enabled}</span><br><span style="font-size:12px;color:#95a5a6;">启用</span></div>
                            <div><span style="font-size:24px;font-weight:600;color:#e74c3c;">${d.sites.disabled}</span><br><span style="font-size:12px;color:#95a5a6;">禁用</span></div>
                        </div>
                    </div>
                    <div class="card">
                        <div class="card-title">证书概览</div>
                        <div style="display:flex;gap:20px;margin-top:8px;">
                            <div><span style="font-size:24px;font-weight:600;color:#27ae60;">${d.certificates.valid}</span><br><span style="font-size:12px;color:#95a5a6;">有效</span></div>
                            <div><span style="font-size:24px;font-weight:600;color:#f39c12;">${d.certificates.expiring_soon}</span><br><span style="font-size:12px;color:#95a5a6;">即将过期</span></div>
                            <div><span style="font-size:24px;font-weight:600;color:#e74c3c;">${d.certificates.expired}</span><br><span style="font-size:12px;color:#95a5a6;">已过期</span></div>
                        </div>
                    </div>
                </div>
            `;
        } catch (e) {
            console.error('仪表盘加载失败:', e);
        }
    }

    await loadDashboard();
    dashboardTimer = setInterval(loadDashboard, 30000);
})();
