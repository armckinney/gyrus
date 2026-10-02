// Gyrus Core Client Scripts: Theme, Navigation, Shortcuts, and Toast Notifications
(function() {
    // Theme Switcher Logic
    function initTheme() {
        const savedTheme = localStorage.getItem('gyrus-theme');
        if (savedTheme) {
            document.documentElement.setAttribute('data-theme', savedTheme);
        } else if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
            document.documentElement.setAttribute('data-theme', 'dark');
        } else {
            document.documentElement.setAttribute('data-theme', 'light');
        }
        if (typeof window.initMermaidTheme === 'function') {
            window.initMermaidTheme();
        }
    }

    window.toggleTheme = function() {
        const current = document.documentElement.getAttribute('data-theme') || 'light';
        const next = current === 'dark' ? 'light' : 'dark';
        document.documentElement.setAttribute('data-theme', next);
        localStorage.setItem('gyrus-theme', next);

        if (typeof window.initMermaidTheme === 'function') {
            window.initMermaidTheme();
        }
        document.querySelectorAll('.mermaid-diagram-card').forEach(card => {
            card.removeAttribute('data-rendered');
        });
        if (typeof window.renderMermaidDiagrams === 'function') {
            window.renderMermaidDiagrams();
        }
    };

    // Tenant / Owner Switcher
    window.switchTenant = function(tenant) {
        localStorage.setItem('gyrus-tenant', tenant || '');
        const url = new URL(window.location.href);
        if (tenant) {
            url.searchParams.set('tenant', tenant);
        } else {
            url.searchParams.delete('tenant');
        }
        window.location.href = url.pathname + url.search;
    };

    // Synchronize active states in navigation sidebar and header
    window.updateActiveNav = function() {
        const path = window.location.pathname;
        const search = window.location.search;
        const currentTarget = path + search;
        const isGraph = path === '/graph' || path.startsWith('/graph');
        const navChat = document.getElementById('nav-agent-chat');
        const isChat = (path === '/' && navChat) || path === '/chat';

        const navGraph = document.getElementById('nav-knowledge-graph');
        const navAll = document.getElementById('nav-all-docs');

        const params = new URLSearchParams(search);
        const currentScope = params.get('scope');
        const currentType = params.get('type');
        const isAllDocs = (path === '/docs' || path === '/') && !currentScope && !currentType;

        // Clean and update sidebar links
        document.querySelectorAll('.app-sidebar .nav-list a').forEach(a => {
            const href = a.getAttribute('href');
            let shouldBeActive = false;

            if (isChat) {
                shouldBeActive = (a === navChat);
            } else if (isGraph) {
                shouldBeActive = (a === navGraph);
            } else if (isAllDocs) {
                shouldBeActive = (a === navAll);
            } else if (href === currentTarget) {
                shouldBeActive = true;
            } else if (href) {
                try {
                    const linkUrl = new URL(href, window.location.origin);
                    if (linkUrl.pathname === path) {
                        const linkScope = linkUrl.searchParams.get('scope');
                        const linkType = linkUrl.searchParams.get('type');
                        if (currentScope && linkScope === currentScope) {
                            shouldBeActive = true;
                        } else if (currentType && linkType === currentType) {
                            shouldBeActive = true;
                        }
                    }
                } catch(e) {}
            }

            a.classList.toggle('active', shouldBeActive);
        });

        // Sync tenant dropdown with current URL
        const tenantSelect = document.getElementById('tenant-select');
        if (tenantSelect) {
            const tenantInUrl = params.get('tenant') || params.get('owner');
            if (tenantInUrl !== null) {
                tenantSelect.value = tenantInUrl;
            }
        }
    };

    // Sidebar Collapse / Expand Controller
    function initSidebarState() {
        const isCollapsed = localStorage.getItem('gyrus-sidebar-collapsed') === 'true';
        if (isCollapsed) {
            document.body.classList.add('sidebar-collapsed');
        }
        updateSidebarToggleBtn(isCollapsed);
    }

    function updateSidebarToggleBtn(isCollapsed) {
        const btn = document.getElementById('sidebar-toggle-btn');
        if (btn) {
            btn.title = isCollapsed ? 'Expand navigation (Ctrl+B)' : 'Collapse navigation (Ctrl+B)';
        }
    }

    window.toggleSidebar = function() {
        const isCollapsed = document.body.classList.toggle('sidebar-collapsed');
        localStorage.setItem('gyrus-sidebar-collapsed', isCollapsed ? 'true' : 'false');
        updateSidebarToggleBtn(isCollapsed);

        // Adjust Cytoscape canvas if graph is active
        if (window._currentCy) {
            setTimeout(function() {
                if (window._currentCy) {
                    window._currentCy.resize();
                }
            }, 250);
        }
    };

    // Keyboard shortcuts
    document.addEventListener('keydown', function(e) {
        // Press Ctrl+B or Cmd+B to toggle sidebar
        if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'b') {
            e.preventDefault();
            window.toggleSidebar();
        }
        // Press '/' to focus global search
        if (e.key === '/' && !['INPUT', 'TEXTAREA'].includes(document.activeElement.tagName)) {
            e.preventDefault();
            const searchInput = document.getElementById('search-input');
            if (searchInput) {
                searchInput.focus();
                searchInput.select();
            }
        }
        if (e.key === 'Escape') {
            const searchInput = document.getElementById('search-input');
            if (searchInput && document.activeElement === searchInput) {
                searchInput.blur();
            }
        }
    });

    // Global HTML escaping helper
    window.escapeHtml = function(str) {
        if (!str) return '';
        return String(str).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
    };

    // Global Toast Notification Manager
    window.showToast = function(message, durationMs) {
        let toast = document.getElementById('gyrus-toast');
        if (!toast) {
            toast = document.createElement('div');
            toast.id = 'gyrus-toast';
            toast.className = 'gyrus-toast';
            document.body.appendChild(toast);
        }
        toast.textContent = message;
        toast.classList.add('show');
        clearTimeout(toast._timeout);
        toast._timeout = setTimeout(() => {
            toast.classList.remove('show');
        }, durationMs || 2500);
    };

    // Navigation & theme listeners
    document.addEventListener('htmx:afterSwap', function() {
        window.updateActiveNav();
    });
    window.addEventListener('popstate', function() {
        window.updateActiveNav();
    });
    document.addEventListener('htmx:historyRestore', function() {
        window.updateActiveNav();
    });

    initTheme();
    initSidebarState();
    document.addEventListener('DOMContentLoaded', function() {
        initTheme();
        initSidebarState();
        window.updateActiveNav();
    });
})();
