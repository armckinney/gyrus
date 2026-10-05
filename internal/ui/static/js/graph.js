// Gyrus Knowledge Graph Visualizer: Cytoscape.js Integration, Layouts, & Inspector
(function() {
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
                padding: 60,
                nodeRepulsion: function(node) { return 2400000 * Math.max(1, node.degree()); },
                idealEdgeLength: function(edge) {
                    const maxDeg = Math.max(edge.source().degree(), edge.target().degree());
                    return 140 + maxDeg * 15;
                },
                edgeElasticity: function(edge) { return 4; },
                componentSpacing: 160,
                gravity: 0.04,
                nodeOverlap: 60,
                numIter: 1000
            };
        } else if (layoutName === 'breadthfirst') {
            layoutOpts = {
                name: 'breadthfirst',
                directed: true,
                padding: 60,
                spacingFactor: 1.1,
                grid: true,
                animate: true,
                animationDuration: 400
            };
        } else if (layoutName === 'concentric') {
            layoutOpts = {
                name: 'concentric',
                padding: 60,
                spacingFactor: 1.1,
                minNodeSpacing: 45,
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
                padding: 60,
                radius: Math.max(220, cy.nodes().length * 11),
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
    let _lastGraphInitKey = '';
    let _lastGraphInitTime = 0;

    window.initGraph = function(containerId, dataUrl, focusDocId) {
        const container = document.getElementById(containerId);
        if (!container) return;

        const initKey = containerId + '|' + (dataUrl || '') + '|' + (focusDocId || '');
        const now = Date.now();
        if (initKey === _lastGraphInitKey && (now - _lastGraphInitTime) < 400) {
            return;
        }
        _lastGraphInitKey = initKey;
        _lastGraphInitTime = now;

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

                (data.nodes || []).forEach(item => {
                    const node = item.data || item;
                    if (!node || !node.id) return;
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
                (data.edges || []).forEach(item => {
                    const edge = item.data || item;
                    if (!edge || !edge.source || !edge.target) return;
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
                        name: 'concentric',
                        animate: false,
                        padding: 60,
                        spacingFactor: 1.1,
                        minNodeSpacing: 45,
                        concentric: function(node) {
                            return node.degree();
                        },
                        levelWidth: function(nodes) {
                            return 2;
                        }
                    }
                });

                window._currentCy = cy;

                // Adjust viewport once layout ready
                cy.ready(function() {
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
                    setTimeout(function() {
                        if (window._currentCy === cy) {
                            cy.resize();
                            if (!focusDocId) {
                                cy.fit(null, 60);
                            }
                        }
                    }, 100);
                    setTimeout(function() {
                        if (window._currentCy === cy) {
                            cy.resize();
                            if (!focusDocId) {
                                cy.fit(null, 60);
                            }
                        }
                    }, 250);
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
                    const escape = window.escapeHtml || function(s) { return s || ''; };
                    if (inspector) {
                        inspector.innerHTML = `
                            <div class="card p-4">
                                <h4 class="font-bold text-base mb-1">${escape(node.data('fullLabel'))}</h4>
                                <div class="text-xs text-muted mb-2"><code>${escape(docId)}</code></div>
                                <div class="flex gap-2 mb-3">
                                    <span class="badge badge-category">${escape(node.data('category'))}</span>
                                    <span class="badge badge-type">${escape(node.data('type'))}</span>
                                    <span class="badge badge-${escape(node.data('status'))}">${escape(node.data('status'))}</span>
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

    // Auto-trigger graph initialization on HTMX swaps and DOMContentLoaded
    function checkAndInitGraph() {
        const cyContainer = document.getElementById('cy-container');
        if (cyContainer) {
            const focusId = cyContainer.getAttribute('data-focus') || '';
            const dataUrl = cyContainer.getAttribute('data-url') || (window.location.search ? '/api/graph' + window.location.search : '/api/graph');
            window.initGraph('cy-container', dataUrl, focusId);
        }
    }

    document.addEventListener('htmx:afterSwap', checkAndInitGraph);
    document.addEventListener('DOMContentLoaded', checkAndInitGraph);
})();
