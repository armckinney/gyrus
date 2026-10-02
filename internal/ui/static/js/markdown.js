// Gyrus Markdown & Mermaid Visualizer: Streaming Parser, Links, and Diagrams
(function() {
    // Mermaid Theme Initializer
    window.initMermaidTheme = function() {
        if (typeof mermaid === 'undefined') return;
        const current = document.documentElement.getAttribute('data-theme') || 'light';
        const m = (window.mermaid && window.mermaid.default) ? window.mermaid.default : window.mermaid;
        m.initialize({
            startOnLoad: false,
            theme: current === 'dark' ? 'dark' : 'neutral',
            securityLevel: 'loose'
        });
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
        const escape = window.escapeHtml || function(s) { return s || ''; };
        const rawHref = href.replace(/&amp;/g, '&');

        // Check if link points to an internal Gyrus document
        const gyrusDocMatch = rawHref.match(/(?:\.gyrus|docs)\/([a-z0-9-_]+)(?:\.md)?(?:#.*)?$/i);
        if (gyrusDocMatch) {
            const docId = gyrusDocMatch[1];
            return `<a href="/docs/${escape(docId)}" class="chat-link chat-doc-link" onclick="window.handleChatDocLink(event, '${escape(docId)}')" title="View document ${escape(docId)}"><span style="margin-right: 2px;">📄</span>${linkText}</a>`;
        }

        if (rawHref.startsWith('/docs/')) {
            const docId = rawHref.replace('/docs/', '').split('#')[0].split('?')[0];
            return `<a href="${escape(rawHref)}" class="chat-link chat-doc-link" onclick="window.handleChatDocLink(event, '${escape(docId)}')" title="View document ${escape(docId)}"><span style="margin-right: 2px;">📄</span>${linkText}</a>`;
        }

        // Check if file:/// URL
        if (rawHref.startsWith('file://') || rawHref.startsWith('file:/')) {
            const cleanPath = rawHref.replace(/^file:\/\//, '').replace(/^file:\//, '/');
            return `<a href="${escape(rawHref)}" class="chat-link chat-file-link" onclick="window.handleChatFileLink(event, '${escape(rawHref)}')" title="Click to copy path: ${escape(cleanPath)}">${linkText} <span class="link-icon">↗</span></a>`;
        }

        // Web URL
        if (rawHref.startsWith('http://') || rawHref.startsWith('https://')) {
            return `<a href="${escape(rawHref)}" class="chat-link" target="_blank" rel="noopener noreferrer">${linkText} <span class="link-icon">↗</span></a>`;
        }

        return `<a href="${escape(rawHref)}" class="chat-link">${linkText}</a>`;
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

        const escape = window.escapeHtml || function(s) { return s || ''; };
        const outputEl = card.querySelector('.mermaid-output');
        const rawCode = card.getAttribute('data-code') || '';
        const fsContent = document.getElementById('mermaid-fullscreen-content');
        const fsSource = document.getElementById('mermaid-fullscreen-source');
        const fsSourceBtn = document.getElementById('mermaid-fs-source-btn');

        if (fsContent && outputEl) {
            fsContent.innerHTML = outputEl.innerHTML;
        }

        if (fsSource) {
            fsSource.innerHTML = `<pre style="margin: 0; padding: 1.5rem; height: 100%; overflow: auto; background: var(--bg-card);"><code class="language-mermaid">${escape(rawCode)}</code></pre>`;
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
        const escape = window.escapeHtml || function(s) { return s || ''; };

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
                                <div class="text-xs mb-1" style="color: #f59e0b; font-weight: 500;">Diagram parse notice: ${escape(err.message || 'Syntax error')}</div>
                                <pre style="margin: 0; background: var(--bg-card);"><code class="language-mermaid">${escape(rawCode)}</code></pre>
                            </div>
                        `;
                    });
            } catch (e) {
                console.warn("Mermaid synchronous error:", e);
                const orphanedError = document.getElementById('d' + diagramId);
                if (orphanedError) orphanedError.remove();

                outputEl.innerHTML = `
                    <div style="width: 100%;">
                        <div class="text-xs mb-1" style="color: #f59e0b; font-weight: 500;">Diagram parse notice: ${escape(e.message || 'Syntax error')}</div>
                        <pre style="margin: 0; background: var(--bg-card);"><code class="language-mermaid">${escape(rawCode)}</code></pre>
                    </div>
                `;
            }
        });
    };

    // Client-side lightweight markdown formatter for streamed chunks
    window.renderSimpleMarkdown = function(raw) {
        if (!raw) return '';
        const escape = window.escapeHtml || function(s) { return s || ''; };

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
        html = escape(html);

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
            return `<code>${escape(code)}</code>`;
        });

        // 7. Restore code blocks (with Mermaid diagram support)
        html = html.replace(/__CODE_BLOCK_(\d+)__/g, function(match, idx) {
            const block = codeBlocks[parseInt(idx, 10)];
            const lang = (block.lang || '').toLowerCase();
            const code = block.code;

            if (lang === 'mermaid') {
                return `
                    <div class="mermaid-diagram-card" data-code="${escape(code)}">
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
                            <pre><code class="language-mermaid">${escape(code)}</code></pre>
                        </div>
                    </div>
                `;
            }

            return `<pre><code class="language-${escape(lang)}">${escape(code)}</code></pre>`;
        });

        // Clean up paragraph wrappers around pre & mermaid blocks
        html = html.replace(/<p><pre>/g, '<pre>');
        html = html.replace(/<\/pre><\/p>/g, '</pre>');
        html = html.replace(/<p>\s*<div class="mermaid-diagram-card"/g, '<div class="mermaid-diagram-card"');
        html = html.replace(/<\/div>\s*<\/p>/g, '</div>');

        return html;
    };

    // Diagram renderer listeners
    document.addEventListener('htmx:afterSwap', function() {
        window.renderMermaidDiagrams();
    });
    document.addEventListener('DOMContentLoaded', function() {
        window.initMermaidTheme();
        window.renderMermaidDiagrams();
    });
})();
