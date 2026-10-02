// Gyrus Agent Chat Controller: SSE Streaming, Tool Actions, & Session Manager
(function() {
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

        const escape = window.escapeHtml || function(s) { return s || ''; };

        fetch('/api/chat/sessions')
            .then(res => res.json())
            .then(sessions => {
                const activeId = sessionStorage.getItem('gyrus_chat_session') || localStorage.getItem('gyrus_active_chat_session');
                const hasActive = activeId && Array.isArray(sessions) && sessions.some(s => s.id === activeId);

                let html = `<option value="" disabled ${!hasActive ? 'selected' : ''}>${hasActive ? 'Select conversation...' : 'New Conversation'}</option>`;
                if (Array.isArray(sessions)) {
                    sessions.forEach(s => {
                        const isSelected = s.id === activeId ? 'selected' : '';
                        html += `<option value="${escape(s.id)}" ${isSelected}>${escape(s.title)} (${s.message_count})</option>`;
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

        const escape = window.escapeHtml || function(s) { return s || ''; };
        const renderMd = window.renderSimpleMarkdown || function(s) { return escape(s); };

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
                                <div class="chat-bubble user">${escape(msg.content)}</div>
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
                                    <div class="agent-text">${renderMd(msg.content)}</div>
                                </div>
                            `;
                            messagesContainer.appendChild(agentMsg);
                        }
                    });
                    messagesContainer.scrollTop = messagesContainer.scrollHeight;
                    if (typeof window.renderMermaidDiagrams === 'function') {
                        window.renderMermaidDiagrams(messagesContainer);
                    }
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

        const escape = window.escapeHtml || function(s) { return s || ''; };

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
                                <div class="session-modal-title" title="${escape(s.title)}">
                                    ${isActive ? '<span class="text-primary" style="margin-right: 4px;">●</span>' : ''}
                                    ${escape(s.title)}
                                </div>
                                <div class="session-modal-sub">
                                    <span>${s.message_count} ${s.message_count === 1 ? 'message' : 'messages'}</span>
                                    <span>•</span>
                                    <span>${dateStr}</span>
                                </div>
                            </div>
                            <div class="session-modal-actions">
                                <button type="button" class="btn btn-outline btn-xs" onclick="window.selectSessionFromModal('${escape(s.id)}')">
                                    Open
                                </button>
                                <button type="button" class="btn btn-outline btn-xs" onclick="window.deleteSession(event, '${escape(s.id)}', '${escape(s.title).replace(/'/g, "\\'")}')" style="color: #ef4444; border-color: rgba(239, 68, 68, 0.4);" title="Delete this conversation">
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
                listEl.innerHTML = `<div class="text-sm text-danger text-center p-4">Error loading conversations: ${escape(err.message)}</div>`;
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
        const escape = window.escapeHtml || function(s) { return s || ''; };
        const isDone = act.state === 'DONE';
        const iconHtml = isDone ? '<span class="action-icon" style="color: #10b981;">✓</span>' : '<span class="action-spinner"></span>';
        const durationHtml = act.duration ? `<span class="action-duration">${escape(act.duration)}</span>` : '';
        const detailsHtml = act.output ? `<div class="action-details" style="display: none;">${escape(act.output)}</div>` : '';
        const hasDetails = !!act.output;

        return `
            <div class="chat-action-badge ${isDone ? 'done' : 'active'}" id="action-step-${act.step_index}">
                <div class="action-header" ${hasDetails ? 'onclick="window.toggleActionDetails(this)"' : ''}>
                    ${iconHtml}
                    <span class="action-tool">${escape(act.tool_name)}</span>
                    <span class="action-param" title="${escape(act.summary)}">${escape(act.summary)}</span>
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

        const escape = window.escapeHtml || function(s) { return s || ''; };
        const renderMd = window.renderSimpleMarkdown || function(s) { return escape(s); };

        // Hide empty state
        const emptyState = document.getElementById('chat-empty-state');
        if (emptyState) emptyState.style.display = 'none';

        // Append User Message
        const userMsg = document.createElement('div');
        userMsg.className = 'chat-message user';
        userMsg.innerHTML = `
            <div class="chat-bubble user">${escape(prompt)}</div>
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
                                    textContainer.innerHTML = renderMd(accumulatedText);
                                    messagesContainer.scrollTop = messagesContainer.scrollHeight;
                                } else if (parsed.type === 'error') {
                                    if (statusBar) statusBar.style.display = 'none';
                                    if (cursor) cursor.style.display = 'none';
                                    accumulatedText += '\n\n**Error:** ' + parsed.data;
                                    textContainer.innerHTML = renderMd(accumulatedText);
                                    messagesContainer.scrollTop = messagesContainer.scrollHeight;
                                } else if (parsed.type === 'done') {
                                    // Turn complete
                                }
                            } catch (e) {
                                if (statusBar) statusBar.style.display = 'none';
                                accumulatedText += dataStr;
                                textContainer.innerHTML = renderMd(accumulatedText);
                            }
                        }
                    }

                    readChunk();
                }).catch(err => {
                    if (err.name !== 'AbortError') {
                        accumulatedText += `\n\n*Stream interrupted: ${escape(err.message)}*`;
                        textContainer.innerHTML = renderMd(accumulatedText);
                    }
                    finishStream();
                });
            }

            readChunk();
        })
        .catch(err => {
            if (err.name !== 'AbortError') {
                textContainer.innerHTML = `<span style="color: #ef4444;">Failed to communicate with agent: ${escape(err.message)}</span>`;
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
            if (typeof window.renderMermaidDiagrams === 'function') {
                window.renderMermaidDiagrams(messagesContainer);
            }
        }
    };

    window.initChatView = function() {
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
    };

    // Chat view listeners
    document.addEventListener('htmx:afterSwap', function() {
        window.initChatView();
    });
    document.addEventListener('DOMContentLoaded', function() {
        window.initChatView();
    });
})();
