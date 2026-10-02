// Gyrus Web UI Client Scripts
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
        initMermaidTheme();
    }

    function initMermaidTheme() {
        if (typeof mermaid === 'undefined') return;
        const current = document.documentElement.getAttribute('data-theme') || 'light';
        const m = (window.mermaid && window.mermaid.default) ? window.mermaid.default : window.mermaid;
        m.initialize({
            startOnLoad: false,
            theme: current === 'dark' ? 'dark' : 'neutral',
            securityLevel: 'loose'
        });
    }

    window.toggleTheme = function() {
        const current = document.documentElement.getAttribute('data-theme') || 'light';
        const next = current === 'dark' ? 'light' : 'dark';
        document.documentElement.setAttribute('data-theme', next);
        localStorage.setItem('gyrus-theme', next);
        initMermaidTheme();
        document.querySelectorAll('.mermaid-diagram-card').forEach(card => {
            card.removeAttribute('data-rendered');
        });
        window.renderMermaidDiagrams();
    };

    // Synchronize active states in navigation sidebar and header
    function updateActiveNav() {
        const path = window.location.pathname;
        const search = window.location.search;
        const currentTarget = path + search;
        const isGraph = path === '/graph' || path.startsWith('/graph');
        const navChat = document.getElementById('nav-agent-chat');
        const isChat = (path === '/' && navChat) || path === '/chat';

        // Header graph button
        const headerGraphBtn = document.getElementById('header-graph-btn');
        if (headerGraphBtn) {
            headerGraphBtn.classList.toggle('active', isGraph);
        }

        const navGraph = document.getElementById('nav-knowledge-graph');
        const navAll = document.getElementById('nav-all-docs');

        // Clean and update sidebar links: only the exact matching target gets 'active'
        document.querySelectorAll('.app-sidebar .nav-list a').forEach(a => {
            const href = a.getAttribute('href');
            let shouldBeActive = false;

            if (isChat) {
                shouldBeActive = (a === navChat);
            } else if (isGraph) {
                shouldBeActive = (a === navGraph);
            } else if ((path === '/docs' || path === '/') && !search) {
                shouldBeActive = (a === navAll);
            } else if (href === currentTarget) {
                shouldBeActive = true;
            }

            a.classList.toggle('active', shouldBeActive);
        });
    }

    // Keyboard shortcuts
    document.addEventListener('keydown', function(e) {
        if (e.key === '/' && document.activeElement.tagName !== 'INPUT' && document.activeElement.tagName !== 'TEXTAREA') {
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

    // Helper to produce short, uncluttered node labels
    function getShortLabel(id, title) {
        if (!id) return '';
        const parts = id.split('-');
        if (parts.length >= 2 && ['adr', 'prd', 'spec', 'guide', 'ip', 'rfc', 'tech'].includes(parts[0])) {
            return parts.slice(0, 2).join('-');
        }
        if (parts.length >= 3 && parts[0] === 'ticket') {
            return parts.slice(0, 3).join('-');
        }
        if (title && title.length <= 16) {
            return title;
        }
        if (title) {
            return title.substring(0, 14) + '…';
        }
        return id.length > 15 ? id.substring(0, 13) + '…' : id;
    }

    // Post-layout Anti-Collision & Radial Dispersion Engine
    function disperseClusteredNodes(cy) {
        const nodes = cy.nodes();
        if (nodes.length === 0) return;

        const minDistance = 140; // minimum distance between ANY two nodes
        const iterations = 60;

        // Pass 1: Global pairwise collision repulsion
        for (let iter = 0; iter < iterations; iter++) {
            let moved = false;
            for (let i = 0; i < nodes.length; i++) {
                for (let j = i + 1; j < nodes.length; j++) {
                    const n1 = nodes[i];
                    const n2 = nodes[j];
                    const p1 = n1.position();
                    const p2 = n2.position();
                    let dx = p2.x - p1.x;
                    let dy = p2.y - p1.y;
                    let dist = Math.sqrt(dx * dx + dy * dy);

                    // If either is a hub node with multiple edges, give even more breathing room
                    const reqDist = (n1.degree() >= 3 || n2.degree() >= 3) ? 180 : minDistance;

                    if (dist < reqDist) {
                        moved = true;
                        if (dist < 0.001) {
                            dx = (Math.random() - 0.5) * 20;
                            dy = (Math.random() - 0.5) * 20;
                            dist = Math.sqrt(dx * dx + dy * dy);
                        }
                        const overlap = (reqDist - dist) / 2;
                        const nx = dx / dist;
                        const ny = dy / dist;

                        n1.position({ x: p1.x - nx * overlap, y: p1.y - ny * overlap });
                        n2.position({ x: p2.x + nx * overlap, y: p2.y + ny * overlap });
                    }
                }
            }
            if (!moved) break;
        }

        // Pass 2: Radial dispersion around high-degree hub nodes (e.g. prd-001, guide-004)
        nodes.forEach(hub => {
            const deg = hub.degree();
            if (deg >= 3) {
                const hubPos = hub.position();
                const neighbors = hub.neighborhood().nodes();
                const count = neighbors.length;
                if (count === 0) return;

                // Generous radial spacing: at least 220px, expanding with neighbor count
                const radius = Math.max(240, count * 40);
                const angleStep = (2 * Math.PI) / count;

                // Deterministic ordering to prevent crossing
                const sortedNeighbors = neighbors.toArray().sort((a, b) => a.id().localeCompare(b.id()));

                sortedNeighbors.forEach((nbr, idx) => {
                    // Only reposition leaves or lower-degree nodes
                    if (nbr.degree() <= 3) {
                        const targetAngle = idx * angleStep;
                        nbr.position({
                            x: hubPos.x + Math.cos(targetAngle) * radius,
                            y: hubPos.y + Math.sin(targetAngle) * radius
                        });
                    }
                });
            }
        });

        // Pass 3: Final relaxation pass to eliminate any residual leaf-on-leaf overlaps
        for (let iter = 0; iter < 30; iter++) {
            let moved = false;
            for (let i = 0; i < nodes.length; i++) {
                for (let j = i + 1; j < nodes.length; j++) {
                    const n1 = nodes[i];
                    const n2 = nodes[j];
                    const p1 = n1.position();
                    const p2 = n2.position();
                    const dx = p2.x - p1.x;
                    const dy = p2.y - p1.y;
                    const dist = Math.sqrt(dx * dx + dy * dy);

                    if (dist < minDistance && dist > 0.001) {
                        moved = true;
                        const overlap = (minDistance - dist) / 2;
                        const nx = dx / dist;
                        const ny = dy / dist;
                        n1.position({ x: p1.x - nx * overlap, y: p1.y - ny * overlap });
                        n2.position({ x: p2.x + nx * overlap, y: p2.y + ny * overlap });
                    }
                }
            }
            if (!moved) break;
        }
    }

    // Interactive Graph Controls
    window.zoomGraph = function(factor) {
        if (!window._currentCy) return;
        const cy = window._currentCy;
        cy.zoom({
            level: cy.zoom() * factor,
            renderedPosition: { x: cy.width() / 2, y: cy.height() / 2 }
        });
    };

    window.fitGraph = function() {
        if (!window._currentCy) return;
        window._currentCy.fit(null, 50);
    };

    window.scaleGraph = function(factor) {
        if (!window._currentCy) return;
        const cy = window._currentCy;
        const nodes = cy.nodes();
        if (nodes.length === 0) return;

        const bb = nodes.boundingBox();
        const cx = (bb.x1 + bb.x2) / 2;
        const cyCenter = (bb.y1 + bb.y2) / 2;

        nodes.forEach(node => {
            const p = node.position();
            node.position({
                x: cx + (p.x - cx) * factor,
                y: cyCenter + (p.y - cyCenter) * factor
            });
        });
        cy.fit(null, 50);
    };

    window.switchGraphLayout = function(layoutName) {
        if (!window._currentCy) return;
        const cy = window._currentCy;
        let layoutOpts = { name: layoutName, animate: true, animationDuration: 400 };

        if (layoutName === 'cose') {
            layoutOpts = {
                name: 'cose',
                animate: true,
                animationDuration: 500,
                padding: 80,
                nodeRepulsion: function(node) { return 4000000 * Math.max(1, node.degree()); },
                idealEdgeLength: function(edge) {
                    const maxDeg = Math.max(edge.source().degree(), edge.target().degree());
                    return 240 + maxDeg * 25;
                },
                edgeElasticity: function(edge) { return 4; },
                componentSpacing: 260,
                gravity: 0.02,
                nodeOverlap: 80,
                numIter: 1000
            };
        } else if (layoutName === 'breadthfirst') {
            layoutOpts = {
                name: 'breadthfirst',
                directed: true,
                padding: 80,
                spacingFactor: 1.8,
                grid: true,
                animate: true,
                animationDuration: 400
            };
        } else if (layoutName === 'concentric') {
            layoutOpts = {
                name: 'concentric',
                padding: 80,
                spacingFactor: 1.8,
                minNodeSpacing: 80,
                concentric: function(node) {
                    return node.degree();
                },
                levelWidth: function(nodes) {
                    return 2;
                },
                animate: true,
                animationDuration: 400
            };
        } else if (layoutName === 'circle') {
            layoutOpts = {
                name: 'circle',
                padding: 80,
                radius: Math.max(340, cy.nodes().length * 18),
                animate: true,
                animationDuration: 400
            };
        }

        const layout = cy.layout(layoutOpts);
        layout.on('layoutstop', function() {
            if (layoutName === 'cose') {
                disperseClusteredNodes(cy);
            }
            cy.fit(null, 60);
        });
        layout.run();
    };


    // Initialize Cytoscape Graph
    window.initGraph = function(containerId, dataUrl, focusDocId) {
        const container = document.getElementById(containerId);
        if (!container) return;

        // Ensure container has visible dimensions
        if (container.clientHeight === 0) {
            container.style.height = '600px';
        }

        fetch(dataUrl)
            .then(res => res.json())
            .then(data => {
                if (!window.cytoscape) {
                    console.error("Cytoscape.js not loaded on window");
                    return;
                }

                const isDark = document.documentElement.getAttribute('data-theme') === 'dark';
                const nodeColor = isDark ? '#60a5fa' : '#2563eb';
                const nodeTextColor = isDark ? '#f8fafc' : '#0f172a';
                const edgeColor = isDark ? '#475569' : '#cbd5e1';
                const pillBg = isDark ? '#1e293b' : '#ffffff';
                const pillBorder = isDark ? '#334155' : '#e2e8f0';

                const elements = [];
                const nodeIds = new Set();

                (data.nodes || []).forEach(node => {
                    if (!node.id) return;
                    nodeIds.add(node.id);
                    const shortLabel = getShortLabel(node.id, node.title);
                    const fullLabel = node.title || node.id;

                    elements.push({
                        group: 'nodes',
                        data: {
                            id: node.id,
                            shortLabel: shortLabel,
                            fullLabel: fullLabel,
                            label: shortLabel,
                            category: node.category || '',
                            type: node.type || '',
                            status: node.status || ''
                        }
                    });
                });

                // Guarantee any dangling edge source/target gets an external node
                (data.edges || []).forEach(edge => {
                    if (!edge.source || !edge.target) return;
                    if (!nodeIds.has(edge.source)) {
                        nodeIds.add(edge.source);
                        const shortLabel = getShortLabel(edge.source, edge.source);
                        elements.push({
                            group: 'nodes',
                            data: {
                                id: edge.source,
                                shortLabel: shortLabel,
                                fullLabel: edge.source,
                                label: shortLabel,
                                category: 'reference',
                                type: 'external',
                                status: 'external'
                            }
                        });
                    }
                    if (!nodeIds.has(edge.target)) {
                        nodeIds.add(edge.target);
                        const shortLabel = getShortLabel(edge.target, edge.target);
                        elements.push({
                            group: 'nodes',
                            data: {
                                id: edge.target,
                                shortLabel: shortLabel,
                                fullLabel: edge.target,
                                label: shortLabel,
                                category: 'reference',
                                type: 'external',
                                status: 'external'
                            }
                        });
                    }
                    elements.push({
                        group: 'edges',
                        data: {
                            id: edge.source + '->' + edge.target + ':' + edge.relationship_type,
                            source: edge.source,
                            target: edge.target,
                            label: edge.relationship_type || 'links'
                        }
                    });
                });

                // Destroy existing instance if container already populated
                if (window._currentCy) {
                    try { window._currentCy.destroy(); } catch (e) {}
                    window._currentCy = null;
                }

                const cy = window.cytoscape({
                    container: container,
                    elements: elements,
                    style: [
                        {
                            selector: 'node',
                            style: {
                                'background-color': nodeColor,
                                'label': 'data(shortLabel)',
                                'color': nodeTextColor,
                                'font-size': '10px',
                                'font-family': 'ui-sans-serif, system-ui, sans-serif',
                                'text-valign': 'bottom',
                                'text-margin-y': 6,
                                'text-background-color': pillBg,
                                'text-background-opacity': 0.88,
                                'text-background-padding': '3px',
                                'text-background-shape': 'roundrectangle',
                                'text-border-width': 1,
                                'text-border-color': pillBorder,
                                'width': 26,
                                'height': 26,
                                'border-width': 2,
                                'border-color': isDark ? '#1e293b' : '#ffffff',
                                'transition-property': 'opacity, border-color, background-color',
                                'transition-duration': '0.2s'
                            }
                        },
                        {
                            selector: 'node[type = "adr"]',
                            style: { 'background-color': '#8b5cf6' }
                        },
                        {
                            selector: 'node[type = "prd"]',
                            style: { 'background-color': '#3b82f6' }
                        },
                        {
                            selector: 'node[type = "specification"]',
                            style: { 'background-color': '#06b6d4' }
                        },
                        {
                            selector: 'node[status = "superseded"], node[status = "deprecated"]',
                            style: { 'opacity': 0.5, 'background-color': '#9ca3af' }
                        },
                        {
                            selector: 'node[status = "external"]',
                            style: {
                                'background-color': isDark ? '#334155' : '#e2e8f0',
                                'border-style': 'dashed',
                                'border-color': isDark ? '#64748b' : '#94a3b8',
                                'color': isDark ? '#94a3b8' : '#64748b'
                            }
                        },
                        {
                            selector: 'edge',
                            style: {
                                'width': 1.5,
                                'line-color': edgeColor,
                                'target-arrow-color': edgeColor,
                                'target-arrow-shape': 'triangle',
                                'curve-style': 'bezier',
                                'arrow-scale': 0.75,
                                'label': '',
                                'font-size': '9px',
                                'color': isDark ? '#9ca3af' : '#64748b',
                                'text-rotation': 'autorotate',
                                'text-margin-y': -8,
                                'transition-property': 'opacity, line-color, width',
                                'transition-duration': '0.2s'
                            }
                        },
                        // Highlighted state for selected node and its direct neighbors
                        {
                            selector: 'node.highlighted',
                            style: {
                                'label': 'data(fullLabel)',
                                'font-size': '11px',
                                'font-weight': '600',
                                'z-index': 999,
                                'text-border-color': isDark ? '#60a5fa' : '#2563eb',
                                'text-border-width': 1.5,
                                'text-background-opacity': 0.96
                            }
                        },
                        {
                            selector: 'node:selected',
                            style: {
                                'border-width': 3,
                                'border-color': '#f59e0b',
                                'background-color': '#f59e0b',
                                'label': 'data(fullLabel)',
                                'font-size': '12px',
                                'font-weight': '700',
                                'z-index': 1000,
                                'text-border-color': '#f59e0b',
                                'text-border-width': 2,
                                'text-background-opacity': 0.98
                            }
                        },
                        {
                            selector: 'edge.highlighted',
                            style: {
                                'width': 2.5,
                                'line-color': isDark ? '#60a5fa' : '#2563eb',
                                'target-arrow-color': isDark ? '#60a5fa' : '#2563eb',
                                'arrow-scale': 0.9,
                                'label': 'data(label)',
                                'text-background-color': pillBg,
                                'text-background-opacity': 0.92,
                                'text-background-padding': '2px',
                                'text-background-shape': 'roundrectangle',
                                'text-border-width': 1,
                                'text-border-color': isDark ? '#3b82f6' : '#93c5fd',
                                'color': isDark ? '#93c5fd' : '#1d4ed8',
                                'font-weight': '600',
                                'z-index': 998
                            }
                        },
                        // Dimmed state for irrelevant nodes and edges when a selection is active
                        {
                            selector: 'node.dimmed',
                            style: {
                                'opacity': 0.2,
                                'label': ''
                            }
                        },
                        {
                            selector: 'edge.dimmed',
                            style: {
                                'opacity': 0.08
                            }
                        }
                    ],
                    layout: {
                        name: 'cose',
                        animate: false,
                        padding: 80,
                        nodeRepulsion: function(node) { return 4000000 * Math.max(1, node.degree()); },
                        idealEdgeLength: function(edge) {
                            const maxDeg = Math.max(edge.source().degree(), edge.target().degree());
                            return 240 + maxDeg * 25;
                        },
                        edgeElasticity: function(edge) { return 4; },
                        componentSpacing: 260,
                        gravity: 0.02,
                        nodeOverlap: 80,
                        numIter: 1000
                    }
                });

                window._currentCy = cy;

                // Adjust viewport once layout ready
                cy.ready(function() {
                    disperseClusteredNodes(cy);
                    cy.resize();
                    if (focusDocId) {
                        const targetNode = cy.getElementById(focusDocId);
                        if (targetNode && targetNode.length > 0) {
                            targetNode.select();
                            applyHighlight(targetNode);
                            cy.center(targetNode);
                            cy.zoom({ level: 1.4, position: targetNode.position() });
                        } else {
                            cy.fit(null, 60);
                        }
                    } else {
                        cy.fit(null, 60);
                    }
                });

                // Window resize adjustment
                window.addEventListener('resize', function() {
                    if (window._currentCy) {
                        window._currentCy.resize();
                    }
                });

                function applyHighlight(node) {
                    cy.elements().removeClass('highlighted dimmed');

                    const neighborhood = node.neighborhood();
                    const connectedEdges = node.connectedEdges();

                    // Dim all elements first
                    cy.elements().addClass('dimmed');

                    // Highlight selected node, adjacent neighbor nodes, and connecting edges
                    node.removeClass('dimmed').addClass('highlighted');
                    neighborhood.removeClass('dimmed').addClass('highlighted');
                    connectedEdges.removeClass('dimmed').addClass('highlighted');
                }

                function clearHighlight() {
                    cy.elements().removeClass('highlighted dimmed');
                }

                // Node selection navigation & neighborhood highlighting
                cy.on('tap', 'node', function(evt) {
                    const node = evt.target;
                    const docId = node.id();

                    applyHighlight(node);

                    const inspector = document.getElementById('node-inspector');
                    if (inspector) {
                        inspector.innerHTML = `
                            <div class="card p-4">
                                <h4 class="font-bold text-base mb-1">${escapeHtml(node.data('fullLabel'))}</h4>
                                <div class="text-xs text-muted mb-2"><code>${escapeHtml(docId)}</code></div>
                                <div class="flex gap-2 mb-3">
                                    <span class="badge badge-category">${escapeHtml(node.data('category'))}</span>
                                    <span class="badge badge-type">${escapeHtml(node.data('type'))}</span>
                                    <span class="badge badge-${escapeHtml(node.data('status'))}">${escapeHtml(node.data('status'))}</span>
                                </div>
                                <a href="/docs/${encodeURIComponent(docId)}" class="btn btn-primary btn-sm block text-center" hx-get="/docs/${encodeURIComponent(docId)}" hx-target="#main-content" hx-push-url="true">View Document</a>
                            </div>
                        `;
                        if (window.htmx) {
                            window.htmx.process(inspector);
                        }
                    }
                });

                // Tap background to reset focus
                cy.on('tap', function(evt) {
                    if (evt.target === cy) {
                        clearHighlight();
                        const inspector = document.getElementById('node-inspector');
                        if (inspector) {
                            inspector.innerHTML = '';
                        }
                    }
                });
            })
            .catch(err => {
                console.error("Failed loading graph topology:", err);
            });
    };

    // Auto-trigger graph initialization and nav synchronization on HTMX swaps & history
    document.addEventListener('htmx:afterSwap', function(evt) {
        updateActiveNav();
        const cyContainer = document.getElementById('cy-container');
        if (cyContainer) {
            const focusId = cyContainer.getAttribute('data-focus') || '';
            window.initGraph('cy-container', '/api/graph', focusId);
        }
        initChatView();
        window.renderMermaidDiagrams();
    });

    window.addEventListener('popstate', updateActiveNav);
    document.addEventListener('htmx:historyRestore', updateActiveNav);

    // Agent Chat Client Logic
    let currentChatAbort = null;

    function getOrCreateSessionId() {
        let sid = sessionStorage.getItem('gyrus_chat_session') || localStorage.getItem('gyrus_active_chat_session');
        if (!sid) {
            sid = 'session-' + Date.now();
            sessionStorage.setItem('gyrus_chat_session', sid);
            localStorage.setItem('gyrus_active_chat_session', sid);
        }
        return sid;
    }

    window.autoResizeTextarea = function(el) {
        if (!el) return;
        el.style.height = 'auto';
        el.style.height = Math.min(el.scrollHeight, 160) + 'px';
    };

    window.handleChatKeyDown = function(evt) {
        if (evt.key === 'Enter' && !evt.shiftKey) {
            evt.preventDefault();
            window.handleChatSubmit(evt);
        }
    };

    window.sendStarterPrompt = function(promptText) {
        const input = document.getElementById('chat-input');
        if (input) {
            input.value = promptText;
            window.handleChatSubmit(new Event('submit'));
        }
    };

    window.onConversationSelect = function(val) {
        if (!val) return;
        window.selectSession(val);
    };

    window.deleteActiveSession = function(evt) {
        const select = document.getElementById('chat-conversation-select');
        if (!select || !select.value) return;
        const opt = select.querySelector(`option[value="${select.value}"]`);
        const title = opt ? opt.textContent : '';
        window.deleteSession(evt, select.value, title);
    };

    window.onModelChange = function(model) {
        localStorage.setItem('gyrus_chat_model', model);
        window.updateEffortOptionsForModel(model);
    };

    window.onEffortChange = function(effort) {
        localStorage.setItem('gyrus_chat_effort', effort);
    };

    window.updateEffortOptionsForModel = function(modelId) {
        const modelSelect = document.getElementById('chat-model-select');
        const effortSelect = document.getElementById('chat-effort-select');
        if (!effortSelect) return;

        let efforts = "auto,low,medium,high";
        if (modelSelect) {
            const opt = modelSelect.querySelector(`option[value="${modelId}"]`);
            if (opt && opt.getAttribute('data-efforts')) {
                efforts = opt.getAttribute('data-efforts');
            }
        }

        const currentEffort = localStorage.getItem('gyrus_chat_effort') || effortSelect.value || 'default';

        if (efforts === 'none' || efforts === '') {
            effortSelect.innerHTML = `<option value="default" selected>N/A (Disabled)</option>`;
            effortSelect.disabled = true;
            effortSelect.style.opacity = '0.6';
            return;
        }

        effortSelect.disabled = false;
        effortSelect.style.opacity = '1';

        const parts = efforts.split(',').map(s => s.trim().toLowerCase());
        let html = '';
        const effortLabels = {
            'default': 'Auto',
            'auto': 'Auto',
            'low': 'Low',
            'medium': 'Medium',
            'high': 'High',
            'xhigh': 'Extra High',
            'max': 'Max'
        };

        let selectedFound = false;
        parts.forEach(eff => {
            const val = eff === 'auto' ? 'default' : eff;
            const label = effortLabels[eff] || eff.charAt(0).toUpperCase() + eff.slice(1);
            const isSel = (val === currentEffort || (currentEffort === 'default' && val === 'default'));
            if (isSel) selectedFound = true;
            html += `<option value="${val}" ${isSel ? 'selected' : ''}>${label}</option>`;
        });

        effortSelect.innerHTML = html;
        if (!selectedFound && parts.length > 0) {
            const firstVal = parts[0] === 'auto' ? 'default' : parts[0];
            effortSelect.value = firstVal;
            localStorage.setItem('gyrus_chat_effort', firstVal);
        }
    };

    window.startNewChat = function() {
        if (currentChatAbort) {
            currentChatAbort.abort();
            currentChatAbort = null;
        }
        const newSid = 'session-' + Date.now();
        sessionStorage.setItem('gyrus_chat_session', newSid);
        localStorage.setItem('gyrus_active_chat_session', newSid);

        const select = document.getElementById('chat-conversation-select');
        if (select) select.value = '';
        const delBtn = document.getElementById('btn-delete-session');
        if (delBtn) delBtn.style.display = 'none';

        const messages = document.getElementById('chat-messages');
        if (messages) {
            const bubbles = messages.querySelectorAll('.chat-message');
            bubbles.forEach(b => b.remove());
            const emptyState = document.getElementById('chat-empty-state');
            if (emptyState) emptyState.style.display = 'flex';
        }

        const input = document.getElementById('chat-input');
        if (input) {
            input.value = '';
            input.focus();
            window.autoResizeTextarea(input);
        }
        const stopBtn = document.getElementById('btn-chat-stop');
        const sendBtn = document.getElementById('btn-chat-send');
        if (stopBtn) stopBtn.style.display = 'none';
        if (sendBtn) sendBtn.disabled = false;
    };

    window.stopChatStream = function() {
        if (currentChatAbort) {
            currentChatAbort.abort();
            currentChatAbort = null;
        }
        const cursors = document.querySelectorAll('.cursor-blink');
        cursors.forEach(c => c.remove());
        const stopBtn = document.getElementById('btn-chat-stop');
        const sendBtn = document.getElementById('btn-chat-send');
        if (stopBtn) stopBtn.style.display = 'none';
        if (sendBtn) sendBtn.disabled = false;
    };

    window.loadSessionsList = function() {
        const modal = document.getElementById('chat-sessions-modal');
        if (modal && modal.style.display !== 'none') {
            window.populateSessionsModal();
        }

        const select = document.getElementById('chat-conversation-select');
        if (!select) return;

        fetch('/api/chat/sessions')
            .then(res => res.json())
            .then(sessions => {
                const activeId = sessionStorage.getItem('gyrus_chat_session') || localStorage.getItem('gyrus_active_chat_session');
                const hasActive = activeId && Array.isArray(sessions) && sessions.some(s => s.id === activeId);

                let html = `<option value="" disabled ${!hasActive ? 'selected' : ''}>${hasActive ? 'Select conversation...' : 'New Conversation'}</option>`;
                if (Array.isArray(sessions)) {
                    sessions.forEach(s => {
                        const isSelected = s.id === activeId ? 'selected' : '';
                        html += `<option value="${escapeHtml(s.id)}" ${isSelected}>${escapeHtml(s.title)} (${s.message_count})</option>`;
                    });
                }
                select.innerHTML = html;

                const delBtn = document.getElementById('btn-delete-session');
                if (hasActive) {
                    select.value = activeId;
                    if (delBtn) delBtn.style.display = 'inline-flex';
                } else {
                    select.value = '';
                    if (delBtn) delBtn.style.display = 'none';
                }
            })
            .catch(err => {
                console.error("Failed loading sessions:", err);
            });
    };

    window.selectSession = function(sessionId) {
        if (!sessionId) {
            window.startNewChat();
            return;
        }
        if (currentChatAbort) {
            currentChatAbort.abort();
            currentChatAbort = null;
        }

        sessionStorage.setItem('gyrus_chat_session', sessionId);
        localStorage.setItem('gyrus_active_chat_session', sessionId);

        const select = document.getElementById('chat-conversation-select');
        if (select) select.value = sessionId;
        const delBtn = document.getElementById('btn-delete-session');
        if (delBtn) delBtn.style.display = 'inline-flex';

        const messagesContainer = document.getElementById('chat-messages');
        const emptyState = document.getElementById('chat-empty-state');
        if (!messagesContainer) return;

        fetch('/api/chat/sessions/' + encodeURIComponent(sessionId))
            .then(res => {
                if (!res.ok) throw new Error("Session not found");
                return res.json();
            })
            .then(session => {
                // Clear existing bubbles
                messagesContainer.querySelectorAll('.chat-message').forEach(b => b.remove());

                if (!session.messages || session.messages.length === 0) {
                    if (emptyState) emptyState.style.display = 'flex';
                } else {
                    if (emptyState) emptyState.style.display = 'none';
                    session.messages.forEach(msg => {
                        if (msg.role === 'user') {
                            const userMsg = document.createElement('div');
                            userMsg.className = 'chat-message user';
                            userMsg.innerHTML = `
                                <div class="chat-bubble user">${escapeHtml(msg.content)}</div>
                                <div class="chat-avatar user">U</div>
                            `;
                            messagesContainer.appendChild(userMsg);
                        } else {
                            const agentMsg = document.createElement('div');
                            agentMsg.className = 'chat-message agent';

                            let actionsHtml = '';
                            if (msg.actions && msg.actions.length > 0) {
                                actionsHtml = `<div class="agent-actions-container">` + msg.actions.map(act => renderActionBadgeHtml(act)).join('') + `</div>`;
                            }

                            agentMsg.innerHTML = `
                                <div class="chat-avatar agent">
                                    <img src="/static/img/logo.png" alt="Gyrus Logo" width="20" height="20" style="border-radius: 4px;">
                                </div>
                                <div class="chat-bubble agent">
                                    ${actionsHtml}
                                    <div class="agent-text">${renderSimpleMarkdown(msg.content)}</div>
                                </div>
                            `;
                            messagesContainer.appendChild(agentMsg);
                        }
                    });
                    messagesContainer.scrollTop = messagesContainer.scrollHeight;
                    window.renderMermaidDiagrams(messagesContainer);
                }

                if (session.model) {
                    const modelSelect = document.getElementById('chat-model-select');
                    if (modelSelect) {
                        modelSelect.value = session.model;
                        window.updateEffortOptionsForModel(session.model);
                    }
                }
                if (session.effort) {
                    const effortSelect = document.getElementById('chat-effort-select');
                    if (effortSelect && !effortSelect.disabled) effortSelect.value = session.effort;
                }
            })
            .catch(err => {
                console.error("Failed loading session:", err);
            });
    };

    window.deleteSession = function(evt, sessionId, sessionTitle) {
        if (evt) evt.stopPropagation();
        const promptMsg = sessionTitle
            ? `Are you sure you want to delete "${sessionTitle}"?`
            : "Are you sure you want to delete this conversation?";
        if (!confirm(promptMsg)) return;

        fetch('/api/chat/sessions/' + encodeURIComponent(sessionId), { method: 'DELETE' })
            .then(res => res.json())
            .then(() => {
                const currentSid = sessionStorage.getItem('gyrus_chat_session') || localStorage.getItem('gyrus_active_chat_session');
                if (currentSid === sessionId) {
                    window.startNewChat();
                }
                window.loadSessionsList();
                const modal = document.getElementById('chat-sessions-modal');
                if (modal && modal.style.display !== 'none') {
                    window.populateSessionsModal();
                }
            })
            .catch(err => {
                console.error("Failed deleting session:", err);
            });
    };

    window.openSessionsModal = function() {
        const modal = document.getElementById('chat-sessions-modal');
        if (!modal) return;
        modal.style.display = 'flex';
        window.populateSessionsModal();
    };

    window.closeSessionsModal = function() {
        const modal = document.getElementById('chat-sessions-modal');
        if (modal) modal.style.display = 'none';
    };

    window.closeSessionsModalOnBackdrop = function(evt) {
        if (evt.target && evt.target.id === 'chat-sessions-modal') {
            window.closeSessionsModal();
        }
    };

    window.populateSessionsModal = function() {
        const listEl = document.getElementById('chat-modal-sessions-list');
        if (!listEl) return;
        listEl.innerHTML = '<div class="text-sm text-muted text-center p-4">Loading conversations...</div>';

        fetch('/api/chat/sessions')
            .then(res => res.json())
            .then(sessions => {
                if (!Array.isArray(sessions) || sessions.length === 0) {
                    listEl.innerHTML = '<div class="text-sm text-muted text-center p-6">No saved conversations yet.</div>';
                    return;
                }
                const activeId = sessionStorage.getItem('gyrus_chat_session') || localStorage.getItem('gyrus_active_chat_session');
                let html = '';
                sessions.forEach(s => {
                    const isActive = s.id === activeId;
                    const d = new Date(s.updated_at || s.created_at);
                    const dateStr = d.toLocaleDateString() + ' ' + d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
                    html += `
                        <div class="session-modal-item ${isActive ? 'active-item' : ''}">
                            <div class="session-modal-meta">
                                <div class="session-modal-title" title="${escapeHtml(s.title)}">
                                    ${isActive ? '<span class="text-primary" style="margin-right: 4px;">●</span>' : ''}
                                    ${escapeHtml(s.title)}
                                </div>
                                <div class="session-modal-sub">
                                    <span>${s.message_count} ${s.message_count === 1 ? 'message' : 'messages'}</span>
                                    <span>•</span>
                                    <span>${dateStr}</span>
                                </div>
                            </div>
                            <div class="session-modal-actions">
                                <button type="button" class="btn btn-outline btn-xs" onclick="window.selectSessionFromModal('${escapeHtml(s.id)}')">
                                    Open
                                </button>
                                <button type="button" class="btn btn-outline btn-xs" onclick="window.deleteSession(event, '${escapeHtml(s.id)}', '${escapeHtml(s.title).replace(/'/g, "\\'")}')" style="color: #ef4444; border-color: rgba(239, 68, 68, 0.4);" title="Delete this conversation">
                                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                        <polyline points="3 6 5 6 21 6"></polyline>
                                        <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
                                    </svg>
                                </button>
                            </div>
                        </div>
                    `;
                });
                listEl.innerHTML = html;
            })
            .catch(err => {
                listEl.innerHTML = `<div class="text-sm text-danger text-center p-4">Error loading conversations: ${escapeHtml(err.message)}</div>`;
            });
    };

    window.selectSessionFromModal = function(sessionId) {
        window.selectSession(sessionId);
        window.closeSessionsModal();
    };

    window.clearAllSessions = function() {
        if (!confirm("Are you sure you want to delete ALL saved conversations? This cannot be undone.")) return;

        fetch('/api/chat/sessions', { method: 'DELETE' })
            .then(res => res.json())
            .then(() => {
                window.startNewChat();
                window.loadSessionsList();
                window.closeSessionsModal();
            })
            .catch(err => {
                console.error("Failed clearing all sessions:", err);
            });
    };

    function renderActionBadgeHtml(act) {
        const isDone = act.state === 'DONE';
        const iconHtml = isDone ? '<span class="action-icon" style="color: #10b981;">✓</span>' : '<span class="action-spinner"></span>';
        const durationHtml = act.duration ? `<span class="action-duration">${escapeHtml(act.duration)}</span>` : '';
        const detailsHtml = act.output ? `<div class="action-details" style="display: none;">${escapeHtml(act.output)}</div>` : '';
        const hasDetails = !!act.output;

        return `
            <div class="chat-action-badge ${isDone ? 'done' : 'active'}" id="action-step-${act.step_index}">
                <div class="action-header" ${hasDetails ? 'onclick="toggleActionDetails(this)"' : ''}>
                    ${iconHtml}
                    <span class="action-tool">${escapeHtml(act.tool_name)}</span>
                    <span class="action-param" title="${escapeHtml(act.summary)}">${escapeHtml(act.summary)}</span>
                    ${durationHtml}
                </div>
                ${detailsHtml}
            </div>
        `;
    }

    window.toggleActionDetails = function(headerEl) {
        const details = headerEl.nextElementSibling;
        if (details && details.classList.contains('action-details')) {
            details.style.display = details.style.display === 'none' ? 'block' : 'none';
        }
    };

    window.handleChatSubmit = function(evt) {
        if (evt && evt.preventDefault) evt.preventDefault();
        const input = document.getElementById('chat-input');
        if (!input) return;
        const prompt = input.value.trim();
        if (!prompt) return;

        input.value = '';
        window.autoResizeTextarea(input);

        const messagesContainer = document.getElementById('chat-messages');
        if (!messagesContainer) return;

        // Hide empty state
        const emptyState = document.getElementById('chat-empty-state');
        if (emptyState) emptyState.style.display = 'none';

        // Append User Message
        const userMsg = document.createElement('div');
        userMsg.className = 'chat-message user';
        userMsg.innerHTML = `
            <div class="chat-bubble user">${escapeHtml(prompt)}</div>
            <div class="chat-avatar user">U</div>
        `;
        messagesContainer.appendChild(userMsg);

        // Append Agent Message Placeholder
        const agentMsg = document.createElement('div');
        agentMsg.className = 'chat-message agent';
        agentMsg.innerHTML = `
            <div class="chat-avatar agent">
                <img src="/static/img/logo.png" alt="Gyrus Logo" width="20" height="20" style="border-radius: 4px;">
            </div>
            <div class="chat-bubble agent">
                <div class="agent-actions-container"></div>
                <div class="agent-text"></div>
                <div class="agent-status-bar"><span class="thinking-spinner"></span> <span class="status-text">Thinking...</span></div>
                <span class="cursor-blink" style="display: none;"></span>
            </div>
        `;
        messagesContainer.appendChild(agentMsg);
        messagesContainer.scrollTop = messagesContainer.scrollHeight;

        const actionsContainer = agentMsg.querySelector('.agent-actions-container');
        const textContainer = agentMsg.querySelector('.agent-text');
        const statusBar = agentMsg.querySelector('.agent-status-bar');
        const statusText = agentMsg.querySelector('.status-text');
        const cursor = agentMsg.querySelector('.cursor-blink');
        const stopBtn = document.getElementById('btn-chat-stop');
        const sendBtn = document.getElementById('btn-chat-send');

        if (stopBtn) stopBtn.style.display = 'inline-flex';
        if (sendBtn) sendBtn.disabled = true;

        const sessionId = getOrCreateSessionId();
        const modelSelect = document.getElementById('chat-model-select');
        const effortSelect = document.getElementById('chat-effort-select');
        const model = modelSelect ? modelSelect.value : (localStorage.getItem('gyrus_chat_model') || 'default');
        const effort = effortSelect ? effortSelect.value : (localStorage.getItem('gyrus_chat_effort') || 'default');

        const abortController = new AbortController();
        currentChatAbort = abortController;

        let accumulatedText = '';

        fetch('/api/chat/stream', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                message: prompt,
                session_id: sessionId,
                model: model,
                effort: effort
            }),
            signal: abortController.signal
        })
        .then(response => {
            if (!response.ok) {
                throw new Error(`Server returned HTTP ${response.status}`);
            }
            const reader = response.body.getReader();
            const decoder = new TextDecoder();
            let buffer = '';

            function readChunk() {
                reader.read().then(({ done, value }) => {
                    if (done) {
                        finishStream();
                        return;
                    }

                    buffer += decoder.decode(value, { stream: true });
                    const lines = buffer.split('\n');
                    buffer = lines.pop(); // keep remainder

                    for (const line of lines) {
                        const trimmed = line.trim();
                        if (trimmed.startsWith('data:')) {
                            const dataStr = trimmed.slice(5).trim();
                            if (!dataStr) continue;
                            try {
                                const parsed = JSON.parse(dataStr);
                                if (parsed.type === 'action' && parsed.action) {
                                    const act = parsed.action;
                                    let actionBadge = actionsContainer.querySelector(`#action-step-${act.step_index}`);
                                    if (!actionBadge) {
                                        actionBadge = document.createElement('div');
                                        actionsContainer.appendChild(actionBadge);
                                    }
                                    actionBadge.outerHTML = renderActionBadgeHtml(act);
                                    messagesContainer.scrollTop = messagesContainer.scrollHeight;
                                } else if (parsed.type === 'status') {
                                    if (statusBar) statusBar.style.display = 'inline-flex';
                                    if (statusText) statusText.textContent = parsed.data || 'Thinking...';
                                    if (cursor) cursor.style.display = 'none';
                                    messagesContainer.scrollTop = messagesContainer.scrollHeight;
                                } else if (parsed.type === 'token') {
                                    if (statusBar) statusBar.style.display = 'none';
                                    if (cursor) cursor.style.display = 'inline-block';
                                    accumulatedText += parsed.data;
                                    textContainer.innerHTML = renderSimpleMarkdown(accumulatedText);
                                    messagesContainer.scrollTop = messagesContainer.scrollHeight;
                                } else if (parsed.type === 'error') {
                                    if (statusBar) statusBar.style.display = 'none';
                                    if (cursor) cursor.style.display = 'none';
                                    accumulatedText += '\n\n**Error:** ' + parsed.data;
                                    textContainer.innerHTML = renderSimpleMarkdown(accumulatedText);
                                    messagesContainer.scrollTop = messagesContainer.scrollHeight;
                                } else if (parsed.type === 'done') {
                                    // Turn complete
                                }
                            } catch (e) {
                                if (statusBar) statusBar.style.display = 'none';
                                accumulatedText += dataStr;
                                textContainer.innerHTML = renderSimpleMarkdown(accumulatedText);
                            }
                        }
                    }

                    readChunk();
                }).catch(err => {
                    if (err.name !== 'AbortError') {
                        accumulatedText += `\n\n*Stream interrupted: ${escapeHtml(err.message)}*`;
                        textContainer.innerHTML = renderSimpleMarkdown(accumulatedText);
                    }
                    finishStream();
                });
            }

            readChunk();
        })
        .catch(err => {
            if (err.name !== 'AbortError') {
                textContainer.innerHTML = `<span style="color: #ef4444;">Failed to communicate with agent: ${escapeHtml(err.message)}</span>`;
            }
            finishStream();
        });

        function finishStream() {
            if (statusBar) statusBar.remove();
            if (cursor) cursor.remove();
            if (stopBtn) stopBtn.style.display = 'none';
            if (sendBtn) sendBtn.disabled = false;
            currentChatAbort = null;
            if (input) input.focus();
            window.loadSessionsList();
            window.renderMermaidDiagrams(messagesContainer);
        }
    };

    function initChatView() {
        const chatContainer = document.querySelector('.chat-page-container');
        if (!chatContainer) return;

        // Restore model and effort preferences
        const savedModel = localStorage.getItem('gyrus_chat_model');
        const modelSelect = document.getElementById('chat-model-select');
        if (savedModel && modelSelect) {
            modelSelect.value = savedModel;
        }

        const selectedModel = modelSelect ? modelSelect.value : 'default';
        window.updateEffortOptionsForModel(selectedModel);

        const savedEffort = localStorage.getItem('gyrus_chat_effort');
        const effortSelect = document.getElementById('chat-effort-select');
        if (savedEffort && effortSelect && !effortSelect.disabled) {
            effortSelect.value = savedEffort;
        }

        // Load sessions list into dropdown
        window.loadSessionsList();

        // Restore active session if exists
        const activeSid = sessionStorage.getItem('gyrus_chat_session') || localStorage.getItem('gyrus_active_chat_session');
        if (activeSid) {
            window.selectSession(activeSid);
        }
    }

    // Toast notification manager
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

    // Chat link click handlers
    window.handleChatFileLink = function(evt, fileUrl) {
        evt.preventDefault();
        let cleanPath = fileUrl.replace(/^file:\/\//, '').replace(/^file:\//, '/');
        if (navigator.clipboard && navigator.clipboard.writeText) {
            navigator.clipboard.writeText(cleanPath).then(() => {
                window.showToast('Copied file path: ' + cleanPath);
            }).catch(() => {
                window.showToast(cleanPath);
            });
        } else {
            window.showToast(cleanPath);
        }
    };

    window.handleChatDocLink = function(evt, docId) {
        evt.preventDefault();
        if (window.htmx) {
            window.htmx.ajax('GET', '/docs/' + encodeURIComponent(docId), { target: '#main-content', pushUrl: true });
        } else {
            window.location.href = '/docs/' + encodeURIComponent(docId);
        }
    };

    // Helper to format any URL into safe, enhanced HTML
    function formatLink(linkText, href) {
        if (!href) return linkText;
        href = href.trim();
        const rawHref = href.replace(/&amp;/g, '&');

        // Check if link points to an internal Gyrus document
        const gyrusDocMatch = rawHref.match(/(?:\.gyrus|docs)\/([a-z0-9-_]+)(?:\.md)?(?:#.*)?$/i);
        if (gyrusDocMatch) {
            const docId = gyrusDocMatch[1];
            return `<a href="/docs/${escapeHtml(docId)}" class="chat-link chat-doc-link" onclick="window.handleChatDocLink(event, '${escapeHtml(docId)}')" title="View document ${escapeHtml(docId)}"><span style="margin-right: 2px;">📄</span>${linkText}</a>`;
        }

        if (rawHref.startsWith('/docs/')) {
            const docId = rawHref.replace('/docs/', '').split('#')[0].split('?')[0];
            return `<a href="${escapeHtml(rawHref)}" class="chat-link chat-doc-link" onclick="window.handleChatDocLink(event, '${escapeHtml(docId)}')" title="View document ${escapeHtml(docId)}"><span style="margin-right: 2px;">📄</span>${linkText}</a>`;
        }

        // Check if file:/// URL
        if (rawHref.startsWith('file://') || rawHref.startsWith('file:/')) {
            const cleanPath = rawHref.replace(/^file:\/\//, '').replace(/^file:\//, '/');
            return `<a href="${escapeHtml(rawHref)}" class="chat-link chat-file-link" onclick="window.handleChatFileLink(event, '${escapeHtml(rawHref)}')" title="Click to copy path: ${escapeHtml(cleanPath)}">${linkText} <span class="link-icon">↗</span></a>`;
        }

        // Web URL
        if (rawHref.startsWith('http://') || rawHref.startsWith('https://')) {
            return `<a href="${escapeHtml(rawHref)}" class="chat-link" target="_blank" rel="noopener noreferrer">${linkText} <span class="link-icon">↗</span></a>`;
        }

        return `<a href="${escapeHtml(rawHref)}" class="chat-link">${linkText}</a>`;
    }

    // Toggle raw mermaid code view
    window.toggleMermaidSource = function(btn) {
        const card = btn.closest('.mermaid-diagram-card');
        if (!card) return;
        const raw = card.querySelector('.mermaid-raw');
        const output = card.querySelector('.mermaid-output');
        if (!raw || !output) return;
        if (raw.style.display === 'none') {
            raw.style.display = 'block';
            output.style.display = 'none';
            btn.textContent = 'View Diagram';
        } else {
            raw.style.display = 'none';
            output.style.display = 'flex';
            btn.textContent = 'View Source';
        }
    };

    // Fullscreen view for mermaid diagrams
    window.openMermaidFullscreen = function(btn) {
        const card = btn.closest('.mermaid-diagram-card');
        if (!card) return;

        let modal = document.getElementById('mermaid-fullscreen-modal');
        if (!modal) {
            modal = document.createElement('div');
            modal.id = 'mermaid-fullscreen-modal';
            modal.className = 'mermaid-fullscreen-modal';
            modal.onclick = function(e) {
                if (e.target === modal) window.closeMermaidFullscreen();
            };
            modal.innerHTML = `
                <div class="mermaid-fullscreen-dialog">
                    <div class="mermaid-fullscreen-header">
                        <span class="mermaid-label" style="font-size: 0.95rem;">
                            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right: 6px;">
                                <circle cx="18" cy="5" r="3"></circle>
                                <circle cx="6" cy="12" r="3"></circle>
                                <circle cx="18" cy="19" r="3"></circle>
                                <line x1="8.59" y1="13.51" x2="15.42" y2="17.49"></line>
                                <line x1="15.41" y1="6.51" x2="8.59" y2="10.49"></line>
                            </svg>
                            Mermaid Diagram
                        </span>
                        <div class="flex items-center gap-2">
                            <button type="button" class="btn btn-outline btn-xs" id="mermaid-fs-source-btn" onclick="window.toggleFsMermaidSource()">
                                View Source
                            </button>
                            <button type="button" class="chat-modal-close-btn" style="position: static;" onclick="window.closeMermaidFullscreen()" title="Close (Esc)">✕</button>
                        </div>
                    </div>
                    <div class="mermaid-fullscreen-body" id="mermaid-fullscreen-content"></div>
                    <div class="mermaid-fullscreen-source" id="mermaid-fullscreen-source" style="display: none;"></div>
                </div>
            `;
            document.body.appendChild(modal);

            document.addEventListener('keydown', function(e) {
                if (e.key === 'Escape' && modal.style.display !== 'none') {
                    window.closeMermaidFullscreen();
                }
            });
        }

        const outputEl = card.querySelector('.mermaid-output');
        const rawCode = card.getAttribute('data-code') || '';
        const fsContent = document.getElementById('mermaid-fullscreen-content');
        const fsSource = document.getElementById('mermaid-fullscreen-source');
        const fsSourceBtn = document.getElementById('mermaid-fs-source-btn');

        if (fsContent && outputEl) {
            fsContent.innerHTML = outputEl.innerHTML;
        }

        if (fsSource) {
            fsSource.innerHTML = `<pre style="margin: 0; padding: 1.5rem; height: 100%; overflow: auto; background: var(--bg-card);"><code class="language-mermaid">${escapeHtml(rawCode)}</code></pre>`;
            fsSource.style.display = 'none';
        }

        if (fsContent) {
            fsContent.style.display = 'flex';
        }
        if (fsSourceBtn) {
            fsSourceBtn.textContent = 'View Source';
        }

        modal.style.display = 'flex';
        document.body.style.overflow = 'hidden';
    };

    window.closeMermaidFullscreen = function() {
        const modal = document.getElementById('mermaid-fullscreen-modal');
        if (modal) {
            modal.style.display = 'none';
        }
        document.body.style.overflow = '';
    };

    window.toggleFsMermaidSource = function() {
        const fsContent = document.getElementById('mermaid-fullscreen-content');
        const fsSource = document.getElementById('mermaid-fullscreen-source');
        const fsSourceBtn = document.getElementById('mermaid-fs-source-btn');
        if (!fsContent || !fsSource || !fsSourceBtn) return;

        if (fsSource.style.display === 'none') {
            fsSource.style.display = 'block';
            fsContent.style.display = 'none';
            fsSourceBtn.textContent = 'View Diagram';
        } else {
            fsSource.style.display = 'none';
            fsContent.style.display = 'flex';
            fsSourceBtn.textContent = 'View Source';
        }
    };

    // Render pending mermaid diagrams in container
    let mermaidRenderCounter = 0;
    window.renderMermaidDiagrams = function(container) {
        if (typeof mermaid === 'undefined') return;
        const root = container || document;

        // Upgrade any raw markdown pre code.language-mermaid if present
        const rawPreMermaid = root.querySelectorAll('pre code.language-mermaid:not([data-processed="true"])');
        rawPreMermaid.forEach(codeEl => {
            codeEl.setAttribute('data-processed', 'true');
            const preEl = codeEl.parentElement;
            if (!preEl || preEl.closest('.mermaid-diagram-card')) return;
            const rawCode = (codeEl.textContent || codeEl.innerText || '').trim();
            if (!rawCode) return;

            const wrapper = document.createElement('div');
            wrapper.className = 'mermaid-diagram-card';
            wrapper.setAttribute('data-code', rawCode);
            wrapper.innerHTML = `
                <div class="mermaid-diagram-header">
                    <span class="mermaid-label">
                        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right: 5px;">
                            <circle cx="18" cy="5" r="3"></circle>
                            <circle cx="6" cy="12" r="3"></circle>
                            <circle cx="18" cy="19" r="3"></circle>
                            <line x1="8.59" y1="13.51" x2="15.42" y2="17.49"></line>
                            <line x1="15.41" y1="6.51" x2="8.59" y2="10.49"></line>
                        </svg>
                        Mermaid Diagram
                    </span>
                    <div class="flex items-center gap-1">
                        <button type="button" class="btn btn-outline btn-xs" onclick="window.toggleMermaidSource(this)" style="font-size: 0.72rem; padding: 2px 7px;">
                            View Source
                        </button>
                        <button type="button" class="btn btn-outline btn-xs" onclick="window.openMermaidFullscreen(this)" title="Open in Fullscreen" style="font-size: 0.72rem; padding: 2px 7px; display: inline-flex; align-items: center; gap: 3px;">
                            <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                <polyline points="15 3 21 3 21 9"></polyline>
                                <polyline points="9 21 3 21 3 15"></polyline>
                                <line x1="21" y1="3" x2="14" y2="10"></line>
                                <line x1="3" y1="21" x2="10" y2="14"></line>
                            </svg>
                            Fullscreen
                        </button>
                    </div>
                </div>
                <div class="mermaid-output"><span class="text-xs text-muted">Rendering diagram...</span></div>
                <div class="mermaid-raw" style="display: none;"></div>
            `;
            preEl.parentNode.insertBefore(wrapper, preEl);
            wrapper.querySelector('.mermaid-raw').appendChild(preEl);
        });

        // Render diagram cards
        const cards = root.querySelectorAll('.mermaid-diagram-card:not([data-rendered="true"])');
        if (!cards || cards.length === 0) return;

        const m = (window.mermaid && window.mermaid.default) ? window.mermaid.default : window.mermaid;

        cards.forEach(card => {
            const outputEl = card.querySelector('.mermaid-output');
            const rawCode = card.getAttribute('data-code');
            if (!outputEl || !rawCode) return;

            card.setAttribute('data-rendered', 'true');
            const diagramId = 'mermaid-svg-' + Date.now() + '-' + (++mermaidRenderCounter);

            try {
                m.render(diagramId, rawCode.trim())
                    .then(result => {
                        outputEl.innerHTML = result.svg;
                    })
                    .catch(err => {
                        console.warn("Mermaid rendering failed:", err);
                        const orphanedError = document.getElementById('d' + diagramId);
                        if (orphanedError) orphanedError.remove();

                        outputEl.innerHTML = `
                            <div style="width: 100%;">
                                <div class="text-xs mb-1" style="color: #f59e0b; font-weight: 500;">Diagram parse notice: ${escapeHtml(err.message || 'Syntax error')}</div>
                                <pre style="margin: 0; background: var(--bg-card);"><code class="language-mermaid">${escapeHtml(rawCode)}</code></pre>
                            </div>
                        `;
                    });
            } catch (e) {
                console.warn("Mermaid synchronous error:", e);
                const orphanedError = document.getElementById('d' + diagramId);
                if (orphanedError) orphanedError.remove();

                outputEl.innerHTML = `
                    <div style="width: 100%;">
                        <div class="text-xs mb-1" style="color: #f59e0b; font-weight: 500;">Diagram parse notice: ${escapeHtml(e.message || 'Syntax error')}</div>
                        <pre style="margin: 0; background: var(--bg-card);"><code class="language-mermaid">${escapeHtml(rawCode)}</code></pre>
                    </div>
                `;
            }
        });
    };

    // Client-side lightweight markdown formatter for streamed chunks
    function renderSimpleMarkdown(raw) {
        if (!raw) return '';

        // 1. Extract fenced code blocks
        const codeBlocks = [];
        let html = raw.replace(/```([a-z0-9_-]*)\n([\s\S]*?)```/g, function(match, lang, code) {
            const idx = codeBlocks.length;
            codeBlocks.push({ lang: lang ? lang.trim() : '', code: code });
            return `__CODE_BLOCK_${idx}__`;
        });

        // 2. Extract inline code
        const inlineCodes = [];
        html = html.replace(/`([^`\n]+)`/g, function(match, code) {
            const idx = inlineCodes.length;
            inlineCodes.push(code);
            return `__INLINE_CODE_${idx}__`;
        });

        // 3. HTML escape remaining text
        html = escapeHtml(html);

        // 4. Parse Links:
        // 4a. Inverted links: (url)[link-name] e.g. (file:///...)[name]
        html = html.replace(/\(((?:https?:\/\/|file:\/\/|\/|\.\.?\/|#)[^\s)]+)\)\[([^\]]+)\]/g, function(match, href, text) {
            return formatLink(text, href);
        });

        // 4b. Standard markdown links: [link-name](url)
        html = html.replace(/\[([^\]]+)\]\(((?:https?:\/\/|file:\/\/|\/|\.\.?\/|#)[^\s)]+)\)/g, function(match, text, href) {
            return formatLink(text, href);
        });

        // 4c. Angle-bracket autolinks: <url>
        html = html.replace(/&lt;((?:https?|file):\/\/[^\s&]+)&gt;/g, function(match, href) {
            return formatLink(href, href);
        });

        // 5. Text formatting
        // Bold: **text**
        html = html.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');

        // Italic: *text*
        html = html.replace(/(^|[^*])\*([^*]+)\*/g, '$1<em>$2</em>');

        // Headers
        html = html.replace(/^### (.*$)/gim, '<h4 style="margin: 0.5rem 0 0.25rem 0; font-size: 1rem;">$1</h4>');
        html = html.replace(/^## (.*$)/gim, '<h3 style="margin: 0.75rem 0 0.25rem 0; font-size: 1.1rem;">$1</h3>');
        html = html.replace(/^# (.*$)/gim, '<h2 style="margin: 1rem 0 0.5rem 0; font-size: 1.25rem;">$1</h2>');

        // Unordered lists
        html = html.replace(/^\s*[-*]\s+(.*)$/gim, '<li style="margin-left: 1.25rem;">$1</li>');

        // Paragraph breaks
        html = html.replace(/\n\n+/g, '</p><p>');
        html = '<p>' + html + '</p>';
        html = html.replace(/<p><\/p>/g, '');

        // 6. Restore inline code
        html = html.replace(/__INLINE_CODE_(\d+)__/g, function(match, idx) {
            const code = inlineCodes[parseInt(idx, 10)];
            return `<code>${escapeHtml(code)}</code>`;
        });

        // 7. Restore code blocks (with Mermaid diagram support)
        html = html.replace(/__CODE_BLOCK_(\d+)__/g, function(match, idx) {
            const block = codeBlocks[parseInt(idx, 10)];
            const lang = (block.lang || '').toLowerCase();
            const code = block.code;

            if (lang === 'mermaid') {
                return `
                    <div class="mermaid-diagram-card" data-code="${escapeHtml(code)}">
                        <div class="mermaid-diagram-header">
                            <span class="mermaid-label">
                                <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right: 5px;">
                                    <circle cx="18" cy="5" r="3"></circle>
                                    <circle cx="6" cy="12" r="3"></circle>
                                    <circle cx="18" cy="19" r="3"></circle>
                                    <line x1="8.59" y1="13.51" x2="15.42" y2="17.49"></line>
                                    <line x1="15.41" y1="6.51" x2="8.59" y2="10.49"></line>
                                </svg>
                                Mermaid Diagram
                            </span>
                            <div class="flex items-center gap-1">
                                <button type="button" class="btn btn-outline btn-xs" onclick="window.toggleMermaidSource(this)" style="font-size: 0.72rem; padding: 2px 7px;">
                                    View Source
                                </button>
                                <button type="button" class="btn btn-outline btn-xs" onclick="window.openMermaidFullscreen(this)" title="Open in Fullscreen" style="font-size: 0.72rem; padding: 2px 7px; display: inline-flex; align-items: center; gap: 3px;">
                                    <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                        <polyline points="15 3 21 3 21 9"></polyline>
                                        <polyline points="9 21 3 21 3 15"></polyline>
                                        <line x1="21" y1="3" x2="14" y2="10"></line>
                                        <line x1="3" y1="21" x2="10" y2="14"></line>
                                    </svg>
                                    Fullscreen
                                </button>
                            </div>
                        </div>
                        <div class="mermaid-output"><span class="text-xs text-muted">Rendering diagram...</span></div>
                        <div class="mermaid-raw" style="display: none;">
                            <pre><code class="language-mermaid">${escapeHtml(code)}</code></pre>
                        </div>
                    </div>
                `;
            }

            return `<pre><code class="language-${escapeHtml(lang)}">${escapeHtml(code)}</code></pre>`;
        });

        // Clean up paragraph wrappers around pre & mermaid blocks
        html = html.replace(/<p><pre>/g, '<pre>');
        html = html.replace(/<\/pre><\/p>/g, '</pre>');
        html = html.replace(/<p>\s*<div class="mermaid-diagram-card"/g, '<div class="mermaid-diagram-card"');
        html = html.replace(/<\/div>\s*<\/p>/g, '</div>');

        return html;
    }

    function escapeHtml(str) {
        if (!str) return '';
        return String(str).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
    }

    // Initialize theme and nav states on page load
    initTheme();
    document.addEventListener('DOMContentLoaded', function() {
        initTheme();
        updateActiveNav();
        initChatView();
        window.renderMermaidDiagrams();
    });
})();
