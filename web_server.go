package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

//go:embed web/*
var webAssets embed.FS

type WebServer struct {
	ygg     *YggManager
	cfg     *AppConfig
	port    int
	clients map[chan string]bool
	mu      sync.Mutex
}

type SendRequest struct {
	Type       string `json:"type"`
	Dest       string `json:"dest"`
	Text       string `json:"text,omitempty"`
	Name       string `json:"name,omitempty"`
	PublicKey  string `json:"publicKey,omitempty"`
	PeerURI    string `json:"peerURI,omitempty"`
	SenderKey  string `json:"senderKey,omitempty"`
	SenderName string `json:"senderName,omitempty"`
	ECDHPubKey string `json:"ecdhPubKey,omitempty"`
	
	// New feature fields
	ReplyTo    int64  `json:"replyTo,omitempty"`
	Reaction   string `json:"reaction,omitempty"`
	EditID     int64  `json:"editID,omitempty"`
	DeleteID   int64  `json:"deleteID,omitempty"`
	MessageID  int64  `json:"messageID,omitempty"`
}

// PeerSummary is a JSON-safe version of core.PeerInfo
type PeerSummary struct {
	URI      string `json:"URI"`
	Up       bool   `json:"Up"`
	Inbound  bool   `json:"Inbound"`
	RXBytes  uint64 `json:"RXBytes"`
	TXBytes  uint64 `json:"TXBytes"`
	LatencyMs int64 `json:"LatencyMs"`
	UptimeSec int64 `json:"UptimeSec"`
}

func NewWebServer(ygg *YggManager, cfg *AppConfig, port int) *WebServer {
	return &WebServer{
		ygg:     ygg,
		cfg:     cfg,
		port:    port,
		clients: make(map[chan string]bool),
	}
}

func (s *WebServer) Start() error {
	// Sub-filesystem to serve web folder
	subFS, err := fs.Sub(webAssets, "web")
	if err != nil {
		return err
	}

	// Security middleware wrapper with hardened headers
	securityMiddleware := func(handler http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Security headers
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
			
			// Content Security Policy - restrict resource loading
			csp := "default-src 'self'; " +
				"script-src 'self' 'unsafe-inline'; " +
				"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; " +
				"font-src 'self' https://fonts.gstatic.com; " +
				"img-src 'self' data: blob:; " +
				"connect-src 'self'; " +
				"frame-ancestors 'none'; " +
				"base-uri 'self'; " +
				"form-action 'self'"
			w.Header().Set("Content-Security-Policy", csp)
			
			// CORS - restrict to localhost only (not wildcard)
			origin := r.Header.Get("Origin")
			if origin == "" || origin == fmt.Sprintf("http://127.0.0.1:%d", s.port) || 
			   origin == fmt.Sprintf("http://localhost:%d", s.port) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-CSRF-Token")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			handler(w, r)
		}
	}

	// CSRF-protected middleware for POST endpoints
	csrfProtected := func(handler http.HandlerFunc) http.HandlerFunc {
		return securityMiddleware(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				// Verify CSRF token from header
				csrfToken := r.Header.Get("X-CSRF-Token")
				if csrfToken == "" {
					http.Error(w, "CSRF token missing", http.StatusForbidden)
					return
				}
				// In production, validate token against session
				// For now, we accept any non-empty token as the app runs locally
			}
			handler(w, r)
		})
	}

	// Rate limiting for API endpoints
	apiRateLimiter := NewRateLimiter(30, time.Minute) // 30 requests per minute

	http.Handle("/", securityMiddleware(http.FileServer(http.FS(subFS)).ServeHTTP))
	http.HandleFunc("/events", securityMiddleware(s.handleEvents))
	http.HandleFunc("/api/state", csrfProtected(s.handleState))
	http.HandleFunc("/api/send", csrfProtected(func(w http.ResponseWriter, r *http.Request) {
		// Apply rate limiting
		clientIP := r.RemoteAddr
		if !apiRateLimiter.Allow(clientIP) {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		s.handleSend(w, r)
	}))

	// Background ticker to broadcast peer status changes to web clients
	go s.peerStatusBroadcastLoop()

	// Background loop to process incoming packets and broadcast over SSE
	go s.startIncomingPacketLoop()

	// Launch default web browser automatically
	go func() {
		time.Sleep(500 * time.Millisecond)
		openBrowser(fmt.Sprintf("http://127.0.0.1:%d", s.port))
	}()

	addr := fmt.Sprintf("127.0.0.1:%d", s.port)
	fmt.Printf("[🌍 Web Console Server running on http://%s]\n", addr)
	return http.ListenAndServe(addr, nil)
}

func (s *WebServer) handleEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	ch := make(chan string, 100)
	s.mu.Lock()
	s.clients[ch] = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.clients, ch)
		s.mu.Unlock()
		close(ch)
	}()

	// Signal connection open
	fmt.Fprintf(w, "data: connected\n\n")
	flusher.Flush()

	for {
		select {
		case msg, open := <-ch:
			if !open {
				return
			}
			fmt.Fprintf(w, "%s\n\n", msg)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func (s *WebServer) BroadcastEvent(event string, data string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Format Server-Sent Event block
	payload := fmt.Sprintf("event: %s\ndata: %s", event, data)
	for ch := range s.clients {
		select {
		case ch <- payload:
		default:
		}
	}
}

func (s *WebServer) handleState(w http.ResponseWriter, r *http.Request) {
	// Collect data without holding the web mutex, to avoid deadlock
	ipStr := "None"
	pkStr := "None"
	if s.ygg.PublicKey() != nil {
		ipStr = s.ygg.Address().String()
		pkStr = fmt.Sprintf("%x", s.ygg.PublicKey())
	}

	s.mu.Lock()
	username := s.cfg.Username
	contacts := s.cfg.Contacts
	s.mu.Unlock()

	if contacts == nil {
		contacts = make(map[string]Contact)
	}

	// Convert peer list to JSON-safe structs
	rawPeers := s.ygg.GetPeersInfo()
	peerSummaries := make([]PeerSummary, 0, len(rawPeers))
	for _, p := range rawPeers {
		peerSummaries = append(peerSummaries, PeerSummary{
			URI:       p.URI,
			Up:        p.Up,
			Inbound:   p.Inbound,
			RXBytes:   p.RXBytes,
			TXBytes:   p.TXBytes,
			LatencyMs: p.Latency.Milliseconds(),
			UptimeSec: int64(p.Uptime.Seconds()),
		})
	}

	history := LoadHistory()

	state := map[string]interface{}{
		"username":  username,
		"publicKey": pkStr,
		"ipv6":      ipStr,
		"contacts":  contacts,
		"history":   history,
		"peers":     peerSummaries,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(state)
}

func (s *WebServer) handleSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	timeStr := time.Now().Format(TimeFormat)

	switch req.Type {
	case "chat":
			contact, hasContact := s.cfg.Contacts[req.Dest]
			var err error
			if hasContact && contact.SharedSecret != "" {
				err = s.ygg.SendEncryptedChatMessage(req.Dest, s.cfg.Username, req.Text, contact.SharedSecret)
			} else {
				err = s.ygg.SendChatMessage(req.Dest, s.cfg.Username, req.Text)
			}

			var bubble string
			if err != nil {
				// Queue offline message
				bubble = fmt.Sprintf("[%s] %s (pending): %s", timeStr, EscapeHTML(s.cfg.Username), EscapeHTML(req.Text))
				s.enqueueOfflineMsg(req.Dest, req.Text, timeStr)
			} else {
				nameTag := fmt.Sprintf(`<span style="color: #7aa2f7; font-weight: bold;">%s</span>`, EscapeHTML(s.cfg.Username))
				bubble = fmt.Sprintf("[%s] %s: %s✓", timeStr, nameTag, EscapeHTML(req.Text))
			}

		s.mu.Lock()
		history := LoadHistory()
		history[req.Dest] = append(history[req.Dest], bubble)
		_ = SaveHistory(history)
		s.mu.Unlock()

	case "typing":
		_ = s.ygg.SendTypingIndicator(req.Dest, s.cfg.Username)

	case "read":
		_ = s.ygg.SendReadReceipt(req.Dest, s.cfg.Username, time.Now().Unix())
		
		// Burn after read: delete messages from history after they are read
		if s.cfg.BurnAfterRead {
			go func() {
				// Wait for burn timeout (default 5 seconds)
				burnDelay := 5 * time.Second
				if s.cfg.BurnTimeoutSec > 0 {
					burnDelay = time.Duration(s.cfg.BurnTimeoutSec) * time.Second
				}
				time.Sleep(burnDelay)
				
				s.mu.Lock()
				history := LoadHistory()
				if _, ok := history[req.Dest]; ok {
					// Keep only system messages, remove chat messages
					var kept []string
					for _, msg := range history[req.Dest] {
						if strings.Contains(msg, "SYSTEM:") {
							kept = append(kept, msg)
						}
					}
					history[req.Dest] = kept
					_ = SaveHistory(history)
				}
				s.mu.Unlock()
				
				// Broadcast update to web clients
				s.BroadcastEvent("burn", fmt.Sprintf(`{"sender_key":"%s"}`, req.Dest))
			}()
		}

	case "clear":
		s.mu.Lock()
		history := LoadHistory()
		history[req.Dest] = []string{}
		_ = SaveHistory(history)
		s.mu.Unlock()

	case "add_contact":
			pubHex, err := GetMyECDHPublicKeyHex(s.cfg.ECDHPrivateKey)
			if err != nil {
				http.Error(w, "Failed to generate ECDH public key", http.StatusInternalServerError)
				return
			}

			s.cfg.Contacts[req.PublicKey] = Contact{
				PublicKey: req.PublicKey,
				Nickname:  req.Name,
			}
			_ = s.cfg.Save()

			_ = s.ygg.SendContactRequest(req.PublicKey, s.cfg.Username, false, pubHex)

	case "add_peer":
		err := s.ygg.AddPeer(req.PeerURI)
		if err == nil {
			found := false
			for _, p := range s.cfg.Peers {
				if p == req.PeerURI {
					found = true
					break
				}
			}
			if !found {
				s.cfg.Peers = append(s.cfg.Peers, req.PeerURI)
				_ = s.cfg.Save()
			}
		}

	case "delete_peer":
		_ = s.ygg.RemovePeer(req.PeerURI)
		var newPeers []string
		for _, p := range s.cfg.Peers {
			if p != req.PeerURI {
				newPeers = append(newPeers, p)
			}
		}
		s.cfg.Peers = newPeers
		_ = s.cfg.Save()

	case "contact_req_accept":
			aesKeyHex, err := DeriveSharedSecret(s.cfg.ECDHPrivateKey, req.ECDHPubKey)
			if err != nil {
				http.Error(w, "Failed to derive shared secret", http.StatusBadRequest)
				return
			}

			s.cfg.Contacts[req.SenderKey] = Contact{
				PublicKey:    req.SenderKey,
				Nickname:     req.SenderName,
				SharedSecret: aesKeyHex,
			}
			_ = s.cfg.Save()

			bobPubHex, _ := GetMyECDHPublicKeyHex(s.cfg.ECDHPrivateKey)
			_ = s.ygg.SendContactRequest(req.SenderKey, s.cfg.Username, true, bobPubHex)

		bubble := fmt.Sprintf("[%s] SYSTEM: Handshake completed. [⚡ E2EE Established]", timeStr)
		s.appendAndBroadcastSystemMsg(req.SenderKey, bubble)

	case "command":
		s.handleSlashCommand(req.Text, req.Dest)
		
	case "reply":
		contact, hasContact := s.cfg.Contacts[req.Dest]
		var err error
		if hasContact && contact.SharedSecret != "" {
			err = s.ygg.SendReplyMessage(req.Dest, s.cfg.Username, req.Text, req.ReplyTo)
		} else {
			err = s.ygg.SendReplyMessage(req.Dest, s.cfg.Username, req.Text, req.ReplyTo)
		}
		
		if err == nil {
			nameTag := fmt.Sprintf(`<span style="color: #7aa2f7; font-weight: bold;">%s</span>`, EscapeHTML(s.cfg.Username))
			bubble := fmt.Sprintf("[%s] %s (reply): %s✓", timeStr, nameTag, EscapeHTML(req.Text))
			s.mu.Lock()
			history := LoadHistory()
			history[req.Dest] = append(history[req.Dest], bubble)
			_ = SaveHistory(history)
			s.mu.Unlock()
		}
		
	case "reaction":
		_ = s.ygg.SendReaction(req.Dest, s.cfg.Username, req.Reaction, req.ReplyTo)
		
	case "edit":
		_ = s.ygg.SendEditMessage(req.Dest, s.cfg.Username, req.Text, req.EditID)
		// Update local history
		s.mu.Lock()
		history := LoadHistory()
		if msgs, ok := history[req.Dest]; ok {
			for i, msg := range msgs {
				if strings.Contains(msg, fmt.Sprintf("%d", req.EditID)) {
					history[req.Dest][i] = fmt.Sprintf("[%s] %s (edited): %s", timeStr, EscapeHTML(s.cfg.Username), EscapeHTML(req.Text))
					break
				}
			}
		}
		_ = SaveHistory(history)
		s.mu.Unlock()
		
	case "delete":
		_ = s.ygg.SendDeleteMessage(req.Dest, s.cfg.Username, req.DeleteID)
		// Remove from local history
		s.mu.Lock()
		history := LoadHistory()
		if msgs, ok := history[req.Dest]; ok {
			for i, msg := range msgs {
				if strings.Contains(msg, fmt.Sprintf("%d", req.DeleteID)) {
					history[req.Dest] = append(msgs[:i], msgs[i+1:]...)
					break
				}
			}
		}
		_ = SaveHistory(history)
		s.mu.Unlock()
		
	case "block":
		if contact, ok := s.cfg.Contacts[req.Dest]; ok {
			contact.Blocked = true
			s.cfg.Contacts[req.Dest] = contact
			_ = s.cfg.Save()
			s.appendAndBroadcastSystemMsg(req.Dest, fmt.Sprintf("[%s] SYSTEM: Contact blocked", timeStr))
		}
		
	case "unblock":
		if contact, ok := s.cfg.Contacts[req.Dest]; ok {
			contact.Blocked = false
			s.cfg.Contacts[req.Dest] = contact
			_ = s.cfg.Save()
			s.appendAndBroadcastSystemMsg(req.Dest, fmt.Sprintf("[%s] SYSTEM: Contact unblocked", timeStr))
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"success"}`))
}

func (s *WebServer) handleSlashCommand(cmdStr string, activeKey string) {
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return
	}

	cmd := parts[0]
	timeStr := time.Now().Format(TimeFormat)

	switch cmd {
	case "/nick":
		if len(parts) < 2 {
			s.appendAndBroadcastSystemMsg(activeKey, "["+timeStr+"] SYSTEM: Usage: /nick <new_username>")
			return
		}
		newName := strings.Join(parts[1:], " ")
		s.cfg.Username = newName
		_ = s.cfg.Save()
		s.appendAndBroadcastSystemMsg(activeKey, fmt.Sprintf("[%s] SYSTEM: Username changed to: %s", timeStr, newName))

	case "/peer":
		if len(parts) < 2 {
			s.appendAndBroadcastSystemMsg(activeKey, "["+timeStr+"] SYSTEM: Usage: /peer <tcp_uri>")
			return
		}
		uri := parts[1]
		err := s.ygg.AddPeer(uri)
		if err != nil {
			s.appendAndBroadcastSystemMsg(activeKey, fmt.Sprintf("[%s] SYSTEM: Peering failed: %v", timeStr, err))
		} else {
			s.appendAndBroadcastSystemMsg(activeKey, fmt.Sprintf("[%s] SYSTEM: Peering connection initiated: %s", timeStr, uri))
		}

	case "/add":
			if len(parts) < 3 {
				s.appendAndBroadcastSystemMsg(activeKey, "["+timeStr+"] SYSTEM: Usage: /add <key> <nickname>")
				return
			}
			key := parts[1]
			nickname := parts[2]

			pubHex, err := GetMyECDHPublicKeyHex(s.cfg.ECDHPrivateKey)
			if err != nil {
				s.appendAndBroadcastSystemMsg(activeKey, fmt.Sprintf("[%s] SYSTEM: Failed to generate ECDH key: %v", timeStr, err))
				return
			}

			err = s.ygg.SendContactRequest(key, s.cfg.Username, false, pubHex)
			if err != nil {
				s.appendAndBroadcastSystemMsg(activeKey, fmt.Sprintf("[%s] SYSTEM: Contact request failed: %v", timeStr, err))
			} else {
				s.appendAndBroadcastSystemMsg(activeKey, fmt.Sprintf("[%s] SYSTEM: Contact request sent to %s (Pending key exchange...)", timeStr, nickname))
				s.cfg.Contacts[key] = Contact{PublicKey: key, Nickname: nickname}
				_ = s.cfg.Save()
			}

	case "/ping":
		if activeKey == "" {
			return
		}
		err := s.ygg.SendPingMessage(activeKey, s.cfg.Username, false)
		if err != nil {
			s.appendAndBroadcastSystemMsg(activeKey, fmt.Sprintf("[%s] SYSTEM: Ping failed: %v", timeStr, err))
		} else {
			s.appendAndBroadcastSystemMsg(activeKey, fmt.Sprintf("[%s] SYSTEM: Sent Ping request...", timeStr))
		}

	case "/send":
		if len(parts) < 2 || activeKey == "" {
			return
		}
		path := strings.Join(parts[1:], " ")
		path = strings.Trim(path, "\"'")
		data, err := os.ReadFile(path)
		if err != nil {
			s.appendAndBroadcastSystemMsg(activeKey, fmt.Sprintf("[%s] SYSTEM: Failed to read file: %v", timeStr, err))
			return
		}

		filename := filepath.Base(path)
		chunkSize := ChunkSize
		totalChunks := (len(data) + chunkSize - 1) / chunkSize
		
		s.appendAndBroadcastSystemMsg(activeKey, fmt.Sprintf("[%s] SYSTEM: Initiating file transfer of %s (%d chunks)...", timeStr, filename, totalChunks))
		
		// Run file transfer asynchronously
		go func() {
			for i := 0; i < totalChunks; i++ {
				start := i * chunkSize
				end := start + chunkSize
				if end > len(data) {
					end = len(data)
				}
				err := s.ygg.SendFileChunk(activeKey, s.cfg.Username, filename, i, totalChunks, data[start:end])
				if err != nil {
					s.appendAndBroadcastSystemMsg(activeKey, fmt.Sprintf("[%s] SYSTEM: File transmission failed: %v", timeStr, err))
					return
				}
				time.Sleep(50 * time.Millisecond) // lightweight pacing
			}
			s.appendAndBroadcastSystemMsg(activeKey, fmt.Sprintf("[%s] SYSTEM: Sent file %s successfully!", timeStr, filename))
		}()

	case "/shake":
		if activeKey == "" {
			return
		}
		err := s.ygg.SendShakeMessage(activeKey, s.cfg.Username)
		if err != nil {
			s.appendAndBroadcastSystemMsg(activeKey, fmt.Sprintf("[%s] SYSTEM: Shake failed: %v", timeStr, err))
		} else {
			s.appendAndBroadcastSystemMsg(activeKey, fmt.Sprintf("[%s] SYSTEM: Sent a Nudge/Shake nudge!", timeStr))
		}

	case "/whois":
		if activeKey == "" {
			return
		}
		contact, ok := s.cfg.Contacts[activeKey]
		if ok {
			s.appendAndBroadcastSystemMsg(activeKey, fmt.Sprintf("[%s] SYSTEM: Nickname: %s | Hex Key: %s", timeStr, contact.Nickname, contact.PublicKey))
			isE2EE := "No"
			if contact.SharedSecret != "" {
				isE2EE = "Yes (AES-GCM Tunnel)"
			}
			s.appendAndBroadcastSystemMsg(activeKey, fmt.Sprintf("[%s] SYSTEM: E2EE Tunnel Active: %s", timeStr, isE2EE))
		}

	case "/search":
		if len(parts) < 2 || activeKey == "" {
			return
		}
		query := strings.ToLower(strings.Join(parts[1:], " "))
		s.appendAndBroadcastSystemMsg(activeKey, fmt.Sprintf("[%s] SYSTEM: Searching for '%s'...", timeStr, query))
		
		s.mu.Lock()
		history := LoadHistory()
		chatLog := history[activeKey]
		s.mu.Unlock()

		found := false
		for _, line := range chatLog {
			plainLine := stripANSI(line)
			if strings.Contains(strings.ToLower(plainLine), query) {
				s.appendAndBroadcastSystemMsg(activeKey, " ❱ "+line)
				found = true
			}
		}
		if !found {
			s.appendAndBroadcastSystemMsg(activeKey, "  No matching history results.")
		}

	case "/help":
		s.appendAndBroadcastSystemMsg(activeKey, fmt.Sprintf("[%s] SYSTEM: Web CLI Slash Commands:", timeStr))
		s.appendAndBroadcastSystemMsg(activeKey, "  /nick <new_name>     - Change display name")
		s.appendAndBroadcastSystemMsg(activeKey, "  /peer <tcp_uri>      - Dial external TCP peer connection")
		s.appendAndBroadcastSystemMsg(activeKey, "  /add <key> <nickname>- Initiate contact key exchange handshake")
		s.appendAndBroadcastSystemMsg(activeKey, "  /ping                - Send latency check ping")
		s.appendAndBroadcastSystemMsg(activeKey, "  /send <file_path>    - P2P file transfer")
		s.appendAndBroadcastSystemMsg(activeKey, "  /shake               - Send screen vibration nudge")
		s.appendAndBroadcastSystemMsg(activeKey, "  /whois               - Output contact diagnostic metadata")
		s.appendAndBroadcastSystemMsg(activeKey, "  /search <keyword>    - Query local chat history logs")
		s.appendAndBroadcastSystemMsg(activeKey, "  /clear               - Wipe current console history cache")
	}
}

func (s *WebServer) appendAndBroadcastSystemMsg(activeKey string, text string) {
	s.mu.Lock()
	history := LoadHistory()
	history[activeKey] = append(history[activeKey], text)
	_ = SaveHistory(history)
	s.mu.Unlock()

	evtData, _ := json.Marshal(map[string]string{
		"sender_key": activeKey,
		"bubble":     text,
	})
	s.BroadcastEvent("incoming_msg", string(evtData))
}

func (s *WebServer) peerStatusBroadcastLoop() {
	ticker := time.NewTicker(PeerStatusTickSec * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		rawPeers := s.ygg.GetPeersInfo()
		summaries := make([]PeerSummary, 0, len(rawPeers))
		for _, p := range rawPeers {
			summaries = append(summaries, PeerSummary{
				URI:       p.URI,
				Up:        p.Up,
				Inbound:   p.Inbound,
				RXBytes:   p.RXBytes,
				TXBytes:   p.TXBytes,
				LatencyMs: p.Latency.Milliseconds(),
				UptimeSec: int64(p.Uptime.Seconds()),
			})
		}
		data, err := json.Marshal(summaries)
		if err == nil {
			s.BroadcastEvent("peers", string(data))
		}
	}
}

func (s *WebServer) startIncomingPacketLoop() {
	for {
		msg, ok := <-s.ygg.MsgChan()
		if !ok {
			return
		}

		senderKey := msg.SenderKey
		senderName := msg.Payload.SenderName

		if contact, ok := s.cfg.Contacts[senderKey]; ok {
			senderName = contact.Nickname
		} else {
			senderName = SafeSenderName(senderName, senderKey)
		}

		timeStr := time.Unix(msg.Payload.Timestamp, 0).Format("15:04:05")

		switch msg.Payload.Type {
		case "typing":
			evtData, _ := json.Marshal(map[string]string{
				"sender_key":  senderKey,
				"sender_name": senderName,
			})
			s.BroadcastEvent("typing", string(evtData))

		case "shake":
			evtData, _ := json.Marshal(map[string]string{
				"sender_key":  senderKey,
				"sender_name": senderName,
			})
			s.BroadcastEvent("shake", string(evtData))

		case "read":
			evtData, _ := json.Marshal(map[string]string{
				"sender_key": senderKey,
			})
			s.BroadcastEvent("read", string(evtData))

		case "contact_req":
			// Rate limit contact requests to prevent flooding
			if !IsContactRequestAllowed(senderKey) {
				s.appendAndBroadcastSystemMsg(senderKey, fmt.Sprintf("[%s] SYSTEM: Contact request rate limit exceeded from %s", timeStr, senderName))
				break
			}
			evtData, _ := json.Marshal(map[string]string{
				"sender_key":  senderKey,
				"sender_name": msg.Payload.SenderName,
				"ecdh_pubkey":  msg.Payload.ECDHPubKey,
			})
			s.BroadcastEvent("contact_req", string(evtData))

		case "contact_acc":
			aesKeyHex, err := DeriveSharedSecret(s.cfg.ECDHPrivateKey, msg.Payload.ECDHPubKey)
			if err == nil {
				if c, ok := s.cfg.Contacts[senderKey]; ok {
					c.SharedSecret = aesKeyHex
					s.cfg.Contacts[senderKey] = c
					_ = s.cfg.Save()
				}
			}

			bubble := fmt.Sprintf("[%s] SYSTEM: %s accepted contact request. [⚡ E2EE Established]", timeStr, senderName)
			s.appendAndBroadcastSystemMsg(senderKey, bubble)

		case "file_chunk":
			_ = os.MkdirAll("./downloads", 0755)
			safeFilename := SanitizeFilename(msg.Payload.FileName)
			filePath := filepath.Join(".", "downloads", safeFilename)

			var flags int
			if msg.Payload.ChunkIdx == 0 {
				flags = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
			} else {
				flags = os.O_WRONLY | os.O_APPEND
			}

			f, err := os.OpenFile(filePath, flags, 0644)
			if err == nil {
				_, _ = f.Write(msg.Payload.Data)
				_ = f.Close()
			}

			progressStr := fmt.Sprintf("[%s] SYSTEM: Receiving %s: %d/%d (%d%%)", timeStr, msg.Payload.FileName, msg.Payload.ChunkIdx+1, msg.Payload.TotalChunks, (msg.Payload.ChunkIdx+1)*100/msg.Payload.TotalChunks)
			if msg.Payload.ChunkIdx == msg.Payload.TotalChunks-1 {
				progressStr = fmt.Sprintf("[%s] SYSTEM: Received file %s saved to ./downloads/", timeStr, msg.Payload.FileName)
			}

			s.mu.Lock()
			history := LoadHistory()
			chatLog := history[senderKey]
			if len(chatLog) > 0 && strings.Contains(chatLog[len(chatLog)-1], "SYSTEM: Receiving") {
				chatLog[len(chatLog)-1] = progressStr
			} else {
				chatLog = append(chatLog, progressStr)
			}

			// Add image preview if file completed and is image
			if msg.Payload.ChunkIdx == msg.Payload.TotalChunks-1 && isImageFile(msg.Payload.FileName) {
				previewWidth := 50
				ansiArt, err := RenderImageFile(filePath, previewWidth)
				if err == nil {
					chatLog = append(chatLog, "🖼️ [Image Preview]:", fmt.Sprintf("<pre>%s</pre>", ansiArt))
				}
			}

			history[senderKey] = chatLog
			_ = SaveHistory(history)
			s.mu.Unlock()

			evtData, _ := json.Marshal(map[string]string{
				"sender_key": senderKey,
				"bubble":     progressStr,
			})
			s.BroadcastEvent("incoming_msg", string(evtData))

		case "reply":
			// Handle reply message
			text := msg.Payload.Text
			if msg.Payload.IsEncrypted {
				if contact, ok := s.cfg.Contacts[senderKey]; ok && contact.SharedSecret != "" {
					decrypted, err := DecryptMessage(contact.SharedSecret, msg.Payload.Text, msg.Payload.Nonce)
					if err == nil {
						text = decrypted
					}
				}
			}
			nameTag := fmt.Sprintf(`<span style="color: #cba6f7; font-weight: bold;">%s</span>`, EscapeHTML(senderName))
			bubble := fmt.Sprintf("[%s] %s (reply): %s", timeStr, nameTag, EscapeHTML(text))
			s.mu.Lock()
			history := LoadHistory()
			history[senderKey] = append(history[senderKey], bubble)
			_ = SaveHistory(history)
			s.mu.Unlock()
			replyEvtData, _ := json.Marshal(map[string]interface{}{
				"sender_key": senderKey,
				"bubble":     bubble,
				"reply_to":   msg.Payload.ReplyTo,
			})
			s.BroadcastEvent("incoming_msg", string(replyEvtData))

		case "reaction":
			evtData, _ := json.Marshal(map[string]interface{}{
				"sender_key": senderKey,
				"emoji":      msg.Payload.Reaction,
				"message_id": msg.Payload.ReplyTo,
			})
			s.BroadcastEvent("reaction", string(evtData))

		case "edit":
			s.mu.Lock()
			history := LoadHistory()
			if msgs, ok := history[senderKey]; ok {
				for i, msgLine := range msgs {
					if strings.Contains(msgLine, fmt.Sprintf("%d", msg.Payload.EditID)) {
						history[senderKey][i] = fmt.Sprintf("[%s] %s (edited): %s", timeStr, EscapeHTML(senderName), EscapeHTML(msg.Payload.Text))
						break
					}
				}
			}
			_ = SaveHistory(history)
			s.mu.Unlock()
			editEvtData, _ := json.Marshal(map[string]interface{}{
				"sender_key": senderKey,
				"edit_id":    msg.Payload.EditID,
				"new_text":   msg.Payload.Text,
			})
			s.BroadcastEvent("edit", string(editEvtData))

		case "delete":
			s.mu.Lock()
			history := LoadHistory()
			if msgs, ok := history[senderKey]; ok {
				for i, msgLine := range msgs {
					if strings.Contains(msgLine, fmt.Sprintf("%d", msg.Payload.DeleteID)) {
						history[senderKey] = append(msgs[:i], msgs[i+1:]...)
						break
					}
				}
			}
			_ = SaveHistory(history)
			s.mu.Unlock()
			deleteEvtData, _ := json.Marshal(map[string]interface{}{
				"sender_key": senderKey,
				"delete_id":  msg.Payload.DeleteID,
			})
			s.BroadcastEvent("delete", string(deleteEvtData))

		default:
			text := msg.Payload.Text
			if msg.Payload.IsEncrypted {
				if contact, ok := s.cfg.Contacts[senderKey]; ok && contact.SharedSecret != "" {
					decrypted, err := DecryptMessage(contact.SharedSecret, msg.Payload.Text, msg.Payload.Nonce)
					if err == nil {
						text = decrypted
					} else {
						text = "[🔒 Encrypted: Decryption failed]"
					}
				} else {
					text = "[🔒 Encrypted: Shared secret missing]"
				}
			}

			nameTag := fmt.Sprintf(`<span style="color: #cba6f7; font-weight: bold;">%s</span>`, EscapeHTML(senderName))
			bubble := fmt.Sprintf("[%s] %s: %s", timeStr, nameTag, EscapeHTML(text))

			s.mu.Lock()
			history := LoadHistory()
			history[senderKey] = append(history[senderKey], bubble)
			_ = SaveHistory(history)
			s.mu.Unlock()

			evtData, _ := json.Marshal(map[string]string{
				"sender_key": senderKey,
				"bubble":     bubble,
			})
			s.BroadcastEvent("incoming_msg", string(evtData))
		}
	}
}

func (s *WebServer) enqueueOfflineMsg(destKey string, text string, timeStr string) {
	// Re-uses local file pending mechanism
	cfgName := getConfigFilename()
	ext := filepath.Ext(cfgName)
	base := cfgName[:len(cfgName)-len(ext)]
	path := filepath.Join(".", base+"_pending.json")
	
	var queue []OfflineMessage
	data, err := os.ReadFile(path)
	if err == nil {
		_ = json.Unmarshal(data, &queue)
	}

	queue = append(queue, OfflineMessage{
		DestKey: destKey,
		Text:    text,
		TimeStr: timeStr,
	})

	bytesData, _ := json.MarshalIndent(queue, "", "  ")
	_ = os.WriteFile(path, bytesData, 0600)
}

func openBrowser(url string) {
	var cmd string
	var args []string
	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default:
		cmd = "xdg-open"
		args = []string{url}
	}
	_ = exec.Command(cmd, args...).Start()
}
