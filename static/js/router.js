const router = {
    routes: {},
    currentRoute: null,
    dashboardTimer: null,

    register(path, handler) {
        this.routes[path] = handler;
    },

    navigate(hash) {
        if (window.location.hash === hash) return;
        window.location.hash = hash;
    },

    init() {
        window.addEventListener('hashchange', () => this.handleRoute());
        this.handleRoute();
    },

    handleRoute() {
        const hash = window.location.hash || '#/login';
        const path = hash.substring(1);

        if (path === '/login' || path === '/setup') {
            document.getElementById('app').style.display = 'none';
            document.getElementById('loginPage').style.display = path === '/login' ? 'block' : 'none';
            document.getElementById('setupPage').style.display = path === '/setup' ? 'block' : 'none';
            return;
        }

        if (!api.token) {
            if (path !== '/login' && path !== '/setup') {
                this.navigate('#/login');
            }
            return;
        }

        document.getElementById('app').style.display = 'block';
        document.getElementById('loginPage').style.display = 'none';
        document.getElementById('setupPage').style.display = 'none';

        const route = this.matchRoute(path);
        if (route) {
            this.currentRoute = path;
            this.updateNav(path);
            route.handler(route.params);
        }
    },

    matchRoute(path) {
        if (this.routes[path]) {
            return { handler: this.routes[path], params: {} };
        }

        for (const routePath in this.routes) {
            const routeParts = routePath.split('/');
            const pathParts = path.split('/');

            if (routeParts.length !== pathParts.length) continue;

            const params = {};
            let match = true;

            for (let i = 0; i < routeParts.length; i++) {
                if (routeParts[i].startsWith(':')) {
                    params[routeParts[i].substring(1)] = pathParts[i];
                } else if (routeParts[i] !== pathParts[i]) {
                    match = false;
                    break;
                }
            }

            if (match) {
                return { handler: this.routes[routePath], params };
            }
        }

        return null;
    },

    updateNav(path) {
        const page = path.split('/')[1];
        document.querySelectorAll('.nav-item').forEach(item => {
            item.classList.toggle('active', item.dataset.page === page);
        });
    }
};
