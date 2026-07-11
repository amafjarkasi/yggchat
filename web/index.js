// State variables
let activeTab = 'chat';
let activeChatKey = '';
let myUsername = 'YggUser';
let myPubKey = '';
let myIPv6 = '';
let contacts = {};
let history = {};
let peers = [];
let typingTimers = {};
let unreadCounts = {};

// EventSource stream connection
let eventSource = null;

document.addEventListener('DOMContentLoaded', () => {
    initApp();
    setupEventListeners();
});

// Fetch configuration state and render
async function initApp() {
    try {
        const res = await fetch('/api/state');
        const state = await res.json();
        
        myUsername = state.username || 'YggUser';
        myPubKey = state.publicKey || '';
        myIPv6 = state.ipv6 || '';
        contacts = state.contacts || {};
        history = state.history || {};
        peers = state.peers || [];

        // Update Header and Diagnostic screen badges
        document.getElementById('username-badge').textContent = `Username: ${myUsername}`;
        document.getElementById('node-pubkey-lbl').textContent = myPubKey;
        document.getElementById('overlay-ipv6-lbl').textContent = myIPv6;
        
        renderContactsList();
        renderPeersList();

        // Activate default tab so panels get pointer-events
        switchTab('chat');
        
        // Start EventSource listener
        connectEventSource();
    } catch (err) {
        console.error("Failed to initialize state", err);
    }
}

function connectEventSource() {
    eventSource = new EventSource('/events');

    eventSource.onopen = () => {
        document.getElementById('mesh-status-dot').className = 'pulse-dot active';
        document.getElementById('header-mesh-status').textContent = '(ONLINE)';
    };

    eventSource.onerror = (e) => {
        document.getElementById('mesh-status-dot').className = 'pulse-dot';
        document.getElementById('header-mesh-status').textContent = '(DISCONNECTED)';
        console.log("SSE disconnected, retrying in 3s...", e);
        setTimeout(connectEventSource, 3000);
    };

    eventSource.addEventListener('incoming_msg', (e) => {
        const data = JSON.parse(e.data);
        const senderKey = data.sender_key;
        
        // If not active, increment unread count and trigger sound notification
        if (senderKey !== activeChatKey || activeTab !== 'chat') {
            unreadCounts[senderKey] = (unreadCounts[senderKey] || 0) + 1;
            renderContactsList();
            playNotificationSound();
        }

        // Save incoming to client-side local history cache
        if (!history[senderKey]) {
            history[senderKey] = [];
        }
        history[senderKey].push(data.bubble);
        
        if (senderKey === activeChatKey) {
            renderChatHistory();
        }
    });

    eventSource.addEventListener('typing', (e) => {
        const data = JSON.parse(e.data);
        const senderKey = data.sender_key;
        const senderName = data.sender_name;

        if (senderKey === activeChatKey) {
            const bar = document.getElementById('typing-status-bar');
            bar.innerHTML = `💬 <strong>${senderName}</strong> is typing...`;
            
            // Clear status after 3 seconds of inactivity
            if (typingTimers[senderKey]) clearTimeout(typingTimers[senderKey]);
            typingTimers[senderKey] = setTimeout(() => {
                bar.innerHTML = '&nbsp;';
            }, 3000);
        }
    });

    eventSource.addEventListener('read', (e) => {
        const data = JSON.parse(e.data);
        const senderKey = data.sender_key;

        // Strip single check ✓ to double check ✓✓ in current chat viewport cache
        if (history[senderKey]) {
            history[senderKey] = history[senderKey].map(line => {
                if (line.endsWith('✓')) {
                    return line.substring(0, line.length - 1) + '<span style="color: #9ece6a;">✓✓</span>';
                }
                return line;
            });
        }

        if (senderKey === activeChatKey) {
            renderChatHistory();
        }
    });

    eventSource.addEventListener('peers', (e) => {
        peers = JSON.parse(e.data);
        renderPeersList();
        
        let onlineCount = peers.filter(p => p.Up).length;
        document.getElementById('peer-count-badge').textContent = `Peers: ${onlineCount}`;
    });

    eventSource.addEventListener('contact_req', (e) => {
        const data = JSON.parse(e.data);
        openContactRequestModal(data.sender_key, data.sender_name, data.ecdh_pubkey);
    });

    eventSource.addEventListener('shake', (e) => {
        const data = JSON.parse(e.data);
        playNotificationSound();
        
        // CSS Vibration screen shake effect!
        const body = document.body;
        body.classList.add('shake-effect');
        setTimeout(() => body.classList.remove('shake-effect'), 500);

        if (!history[data.sender_key]) history[data.sender_key] = [];
        history[data.sender_key].push(`⚡ SYSTEM: Nudge/Shake received from ${data.sender_name}`);
        if (data.sender_key === activeChatKey) {
            renderChatHistory();
        }
    });
}

// Render contacts sidebar
function renderContactsList() {
    const list = document.getElementById('contacts-list-container');
    list.innerHTML = '';

    const sortedContacts = Object.values(contacts);
    if (sortedContacts.length === 0) {
        list.innerHTML = `<div class="text-muted" style="padding: 1rem; font-size: 0.85rem;">No contacts. Click ＋ to add.</div>`;
        return;
    }

    sortedContacts.forEach(c => {
        const div = document.createElement('div');
        div.className = `contact-item ${c.publicKey === activeChatKey ? 'active' : ''}`;
        div.onclick = () => selectContact(c.publicKey);

        const unreadBadge = unreadCounts[c.publicKey] 
            ? `<span class="unread-dot"></span>` 
            : '';

        const padlock = c.sharedSecret ? '🔒' : '🔓';

        div.innerHTML = `
            <div class="contact-info">
                <span class="contact-name">${c.nickname} ${padlock}</span>
                <span class="contact-key">${c.publicKey.substring(0, 12)}...</span>
            </div>
            <div class="contact-meta">
                ${unreadBadge}
            </div>
        `;
        list.appendChild(div);
    });
}

// Select a contact and update chat panels
function selectContact(key) {
    activeChatKey = key;
    unreadCounts[key] = 0;
    
    renderContactsList();

    const contact = contacts[key];
    document.getElementById('chat-title-name').textContent = contact.nickname;
    
    const security = document.getElementById('chat-security-badge');
    if (contact.sharedSecret) {
        security.style.display = 'inline-block';
        security.textContent = '[🛡️ E2EE SECURED]';
        security.className = 'badge-secure';
    } else {
        security.style.display = 'inline-block';
        security.textContent = '[🔓 UNENCRYPTED]';
        security.className = 'badge-secure text-muted';
    }

    // Toggle viewport action headers
    document.getElementById('whois-contact-btn').style.display = 'inline-block';
    document.getElementById('nudge-contact-btn').style.display = 'inline-block';
    document.getElementById('clear-chat-btn').style.display = 'inline-block';
    document.getElementById('chat-input-container').style.display = 'flex';

    renderChatHistory();

    // Send read receipt
    postAPI('/api/send', { type: 'read', dest: key });
}

// Render messages history in viewport
function renderChatHistory() {
    const list = document.getElementById('chat-messages-container');
    list.innerHTML = '';

    const lines = history[activeChatKey] || [];
    if (lines.length === 0) {
        list.innerHTML = `
            <div class="welcome-screen">
                <h2>No messages yet</h2>
                <p>Start the conversation securely. Messages are fully E2EE-secured if lock is verified.</p>
            </div>
        `;
        return;
    }

    lines.forEach(line => {
        const div = document.createElement('div');
        
        // Parse basic tags or SYSTEM styling
        if (line.includes('SYSTEM:') || line.includes('[📤') || line.includes('[✓')) {
            div.className = 'msg-row system';
            div.innerHTML = line;
        } else {
            // Check message direction
            // Format: [15:04:05] Name: Text
            const match = line.match(/^\[(.*?)\]\s+(.*?):\s+(.*)$/);
            if (match) {
                const time = match[1];
                const name = match[2];
                let text = match[3];

                const isMe = name.includes(myUsername);
                div.className = `msg-row ${isMe ? 'me' : 'peer'}`;

                // Extract tick status if me
                let tickStr = '';
                if (isMe) {
                    if (text.endsWith('✓✓')) {
                        tickStr = '<span class="msg-tick read">✓✓</span>';
                        text = text.substring(0, text.length - 2);
                    } else if (text.endsWith('✓')) {
                        tickStr = '<span class="msg-tick">✓</span>';
                        text = text.substring(0, text.length - 1);
                    }
                }

                // Check for embedded image previews (ANSI conversion fallback or description)
                if (text.includes('🖼️ [Image Preview]')) {
                    div.className = 'msg-row system';
                    div.innerHTML = text;
                } else {
                    div.innerHTML = `
                        <div class="msg-bubble">${text}</div>
                        <div class="msg-meta">
                            <span class="msg-time">${time}</span>
                            ${tickStr}
                        </div>
                    `;
                }
            } else {
                div.className = 'msg-row peer';
                div.innerHTML = `<div class="msg-bubble">${line}</div>`;
            }
        }
        list.appendChild(div);
    });

    // Auto scroll bottom
    list.scrollTop = list.scrollHeight;
}

// Render connected peer lists
function renderPeersList() {
    const tbody = document.getElementById('peers-table-body');
    tbody.innerHTML = '';

    if (peers.length === 0) {
        tbody.innerHTML = `<tr><td colspan="6" class="text-muted">No peers configured. Click ＋ to add.</td></tr>`;
        return;
    }

    peers.forEach((p) => {
        const tr = document.createElement('tr');
        const status = p.Up
            ? `<span style="color: #9ece6a;">🟢 Connected</span>`
            : `<span style="color: #f7768e;">🔴 Offline</span>`;
        const direction = p.Inbound ? '⬇️ Inbound' : '⬆️ Outbound';
        const latency = p.LatencyMs > 0 ? `${p.LatencyMs}ms` : '—';
        const rx = p.RXBytes > 0 ? formatBytes(p.RXBytes) : '—';
        const tx = p.TXBytes > 0 ? formatBytes(p.TXBytes) : '—';
        tr.innerHTML = `
            <td><code>${p.URI || '—'}</code></td>
            <td>${direction}</td>
            <td>${latency}</td>
            <td><span style="color: #9ece6a;">${rx}</span> / <span style="color: #7aa2f7;">${tx}</span></td>
            <td>${status}</td>
            <td><button class="btn btn-sm error" onclick="deletePeer('${p.URI}')">Remove</button></td>
        `;
        tbody.appendChild(tr);
    });
}

function formatBytes(bytes) {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1048576) return (bytes / 1024).toFixed(1) + ' KB';
    if (bytes < 1073741824) return (bytes / 1048576).toFixed(1) + ' MB';
    return (bytes / 1073741824).toFixed(2) + ' GB';
}

// Command autocompletion engine
function setupEventListeners() {
    const input = document.getElementById('chat-input-field');
    
    // Command autocompletion on Tab
    input.addEventListener('keydown', (e) => {
        if (e.key === 'Tab') {
            e.preventDefault();
            const val = input.value;
            if (val.startsWith('/')) {
                const commands = ['/nick', '/peer', '/add', '/ping', '/send', '/clear', '/whois', '/shake', '/shout', '/help'];
                const matches = commands.filter(c => c.startsWith(val));
                if (matches.length > 0) {
                    input.value = matches[0] + ' ';
                }
            }
        }
    });

    // Send typing event indicator when input changes
    let lastTypingTime = 0;
    input.addEventListener('input', () => {
        const now = Date.now();
        if (now - lastTypingTime > 2000 && activeChatKey) {
            postAPI('/api/send', { type: 'typing', dest: activeChatKey });
            lastTypingTime = now;
        }
    });

    // Send msg on enter key
    input.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            sendChatMessage();
        }
    });

    document.getElementById('send-msg-btn').onclick = sendChatMessage;

    // View Tabs Switching
    document.getElementById('tab-chat-btn').onclick = () => switchTab('chat');
    document.getElementById('tab-settings-btn').onclick = () => switchTab('settings');

    // Modals Hooks
    const cModal = document.getElementById('add-contact-modal');
    document.getElementById('open-add-contact-btn').onclick = () => cModal.classList.add('active');
    document.getElementById('close-contact-modal-btn').onclick = () => cModal.classList.remove('active');
    document.getElementById('cancel-contact-modal-btn').onclick = () => cModal.classList.remove('active');

    const pModal = document.getElementById('add-peer-modal');
    document.getElementById('open-add-peer-btn').onclick = () => pModal.classList.add('active');
    document.getElementById('close-peer-modal-btn').onclick = () => pModal.classList.remove('active');
    document.getElementById('cancel-peer-modal-btn').onclick = () => pModal.classList.remove('active');

    // Submit actions
    document.getElementById('save-contact-btn').onclick = saveContact;
    document.getElementById('save-peer-btn').onclick = savePeer;

    // Viewport shortcuts
    document.getElementById('whois-contact-btn').onclick = () => executeSlashCommand(`/whois`);
    document.getElementById('nudge-contact-btn').onclick = () => executeSlashCommand(`/shake`);
    document.getElementById('clear-chat-btn').onclick = () => executeSlashCommand(`/clear`);

    // File Uploader
    const fileSelector = document.getElementById('file-uploader');
    document.getElementById('send-file-btn').onclick = () => fileSelector.click();
    fileSelector.onchange = () => {
        if (fileSelector.files.length > 0) {
            const file = fileSelector.files[0];
            executeSlashCommand(`/send ${file.name}`); // Note: Backend reads from local folder or runs transfer
            fileSelector.value = '';
        }
    };
}

// Switch Active Navigation tabs
function switchTab(tab) {
    activeTab = tab;
    document.getElementById('tab-chat-btn').className = `footer-tab ${tab === 'chat' ? 'active' : ''}`;
    document.getElementById('tab-settings-btn').className = `footer-tab ${tab === 'settings' ? 'active' : ''}`;

    document.getElementById('chat-panel').className = `panel chat-panel ${tab === 'chat' ? 'active' : ''}`;
    document.getElementById('dashboard-panel').className = `panel dashboard-panel ${tab === 'settings' ? 'active' : ''}`;
}

// Send Message Payload
async function sendChatMessage() {
    const input = document.getElementById('chat-input-field');
    const text = input.value.trim();
    if (!text) return;

    input.value = '';
    
    if (text.startsWith('/')) {
        executeSlashCommand(text);
        return;
    }

    if (!activeChatKey) return;

    // Optimistically render text locally
    const timeStr = new Date().toLocaleTimeString();
    const myNameTag = `<span style="color: #7aa2f7; font-weight: bold;">${myUsername}</span>`;
    const bubble = `[${timeStr}] ${myNameTag}: ${text}✓`;
    
    if (!history[activeChatKey]) history[activeChatKey] = [];
    history[activeChatKey].push(bubble);
    renderChatHistory();

    await postAPI('/api/send', {
        type: 'chat',
        dest: activeChatKey,
        text: text
    });
}

function executeSlashCommand(commandStr) {
    if (commandStr === '/clear') {
        history[activeChatKey] = [];
        renderChatHistory();
        postAPI('/api/send', { type: 'clear', dest: activeChatKey });
        return;
    }

    postAPI('/api/send', {
        type: 'command',
        text: commandStr,
        dest: activeChatKey
    });
}

// Add Contact Handshake
async function saveContact() {
    const nameInput = document.getElementById('new-contact-name');
    const keyInput = document.getElementById('new-contact-key');
    const name = nameInput.value.trim();
    const key = keyInput.value.trim();

    if (!name || !key) return;

    await postAPI('/api/send', {
        type: 'add_contact',
        name: name,
        publicKey: key
    });

    // Close and reload config state
    document.getElementById('add-contact-modal').classList.remove('active');
    setTimeout(initApp, 500);
}

// Add Peer connection
async function savePeer() {
    const uriInput = document.getElementById('new-peer-uri');
    const uri = uriInput.value.trim();
    if (!uri) return;

    await postAPI('/api/send', {
        type: 'add_peer',
        peerURI: uri
    });

    document.getElementById('add-peer-modal').classList.remove('active');
    setTimeout(initApp, 500);
}

async function deletePeer(uri) {
    await postAPI('/api/send', {
        type: 'delete_peer',
        peerURI: uri
    });
    setTimeout(initApp, 500);
}

// Modal Handshakes
let incomingRequestData = null;
function openContactRequestModal(key, name, ecdhPub) {
    incomingRequestData = { key, name, ecdhPub };
    document.getElementById('incoming-request-text').textContent = `Accept E2EE contact request from ${name}? Key: ${key.substring(0, 16)}...`;
    document.getElementById('contact-request-modal').classList.add('active');

    document.getElementById('accept-request-btn').onclick = async () => {
        await postAPI('/api/send', {
            type: 'contact_req_accept',
            senderKey: key,
            senderName: name,
            ecdhPubKey: ecdhPub
        });
        document.getElementById('contact-request-modal').classList.remove('active');
        setTimeout(initApp, 500);
    };

    document.getElementById('decline-request-btn').onclick = () => {
        document.getElementById('contact-request-modal').classList.remove('active');
    };
}

// Notification Alert sound
function playNotificationSound() {
    const audio = document.getElementById('notify-sound');
    if (audio) {
        // Quick synthetic beep sound generator in browser using Web Audio API!
        const context = new (window.AudioContext || window.webkitAudioContext)();
        const osc = context.createOscillator();
        const gain = context.createGain();
        osc.type = 'sine';
        osc.frequency.setValueAtTime(800, context.currentTime);
        gain.gain.setValueAtTime(0.1, context.currentTime);
        osc.connect(gain);
        gain.connect(context.destination);
        osc.start();
        osc.stop(context.currentTime + 0.15);
    }
}

// POST API Wrapper helper
async function postAPI(url, data) {
    try {
        const res = await fetch(url, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data)
        });
        return await res.json();
    } catch (err) {
        console.error("POST failed", err);
    }
}
