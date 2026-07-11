package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/url"
	"sync"
	"time"

	golog "github.com/gologme/log"
	"github.com/yggdrasil-network/yggdrasil-go/src/config"
	"github.com/yggdrasil-network/yggdrasil-go/src/core"
	iwt "github.com/Arceliar/ironwood/types"
)

type ChatPayload struct {
	SenderName  string `json:"sender_name"`
	Text        string `json:"text"`
	Timestamp   int64  `json:"timestamp"`
	Type        string `json:"type,omitempty"` // "chat", "ping", "pong", "contact_req", "contact_acc", "file_chunk", "read"
	FileName    string `json:"filename,omitempty"`
	ChunkIdx    int    `json:"chunk_idx,omitempty"`
	TotalChunks int    `json:"total_chunks,omitempty"`
	Data        []byte `json:"data,omitempty"`
	
	// E2EE fields
	ECDHPubKey  string `json:"ecdh_pubkey,omitempty"`
	Nonce       string `json:"nonce,omitempty"`
	IsEncrypted bool   `json:"is_encrypted,omitempty"`
}

type IncomingMessage struct {
	SenderKey string
	Payload   ChatPayload
}

type YggManager struct {
	node      *core.Core
	msgChan   chan IncomingMessage
	logger    *golog.Logger
	mu        sync.RWMutex
	listeners []string
	peers     []string
	running   bool
	wg        sync.WaitGroup
	quit      chan struct{}
}

func NewYggManager() *YggManager {
	return &YggManager{
		msgChan: make(chan IncomingMessage, 100),
		quit:    make(chan struct{}),
		logger:  golog.New(io.Discard, "", 0), // Silence standard Yggdrasil logs
	}
}

func (y *YggManager) Start(privKeyHex string, listenAddrs []string, peerAddrs []string) error {
	privKeyBytes, err := hex.DecodeString(privKeyHex)
	if err != nil {
		return fmt.Errorf("invalid private key: %w", err)
	}

	if len(privKeyBytes) != ed25519.PrivateKeySize {
		return fmt.Errorf("invalid private key size: expected %d, got %d", ed25519.PrivateKeySize, len(privKeyBytes))
	}

	// Build Yggdrasil config
	cfg := &config.NodeConfig{
		PrivateKey: config.KeyBytes(privKeyBytes),
		IfName:     "none", // Disable TUN/TAP (user-space only)
	}

	if err := cfg.GenerateSelfSignedCertificate(); err != nil {
		return fmt.Errorf("failed to generate self-signed certificate: %w", err)
	}

	// Compile setup options
	var opts []core.SetupOption
	for _, listenStr := range listenAddrs {
		opts = append(opts, core.ListenAddress(listenStr))
	}
	for _, peerStr := range peerAddrs {
		opts = append(opts, core.Peer{URI: peerStr})
	}

	// Create and start core
	node, err := core.New(cfg.Certificate, y.logger, opts...)
	if err != nil {
		return fmt.Errorf("failed to initialize Yggdrasil core: %w", err)
	}

	y.mu.Lock()
	y.node = node
	y.listeners = listenAddrs
	y.peers = peerAddrs
	y.running = true
	y.mu.Unlock()

	// Start packet reading goroutine
	y.wg.Add(1)
	go y.readPackets()

	return nil
}

func (y *YggManager) Stop() {
	y.mu.Lock()
	if !y.running {
		y.mu.Unlock()
		return
	}
	y.running = false
	close(y.quit)
	if y.node != nil {
		y.node.Stop()
	}
	y.mu.Unlock()

	y.wg.Wait()
}

func (y *YggManager) PublicKey() ed25519.PublicKey {
	y.mu.RLock()
	defer y.mu.RUnlock()
	if y.node == nil {
		return nil
	}
	return y.node.PublicKey()
}

func (y *YggManager) Address() net.IP {
	y.mu.RLock()
	defer y.mu.RUnlock()
	if y.node == nil {
		return nil
	}
	return y.node.Address()
}

func (y *YggManager) AddPeer(peerURI string) error {
	u, err := url.Parse(peerURI)
	if err != nil {
		return err
	}

	y.mu.Lock()
	defer y.mu.Unlock()
	if y.node == nil {
		return fmt.Errorf("node not started")
	}

	if err := y.node.AddPeer(u, ""); err != nil {
		return err
	}

	// Track peer list
	found := false
	for _, p := range y.peers {
		if p == peerURI {
			found = true
			break
		}
	}
	if !found {
		y.peers = append(y.peers, peerURI)
	}
	return nil
}

func (y *YggManager) RemovePeer(peerURI string) error {
	u, err := url.Parse(peerURI)
	if err != nil {
		return err
	}

	y.mu.Lock()
	defer y.mu.Unlock()
	if y.node == nil {
		return fmt.Errorf("node not started")
	}

	if err := y.node.RemovePeer(u, ""); err != nil {
		return err
	}

	// Untrack peer list
	var newPeers []string
	for _, p := range y.peers {
		if p != peerURI {
			newPeers = append(newPeers, p)
		}
	}
	y.peers = newPeers
	return nil
}

func (y *YggManager) Listen(listenURI string) error {
	u, err := url.Parse(listenURI)
	if err != nil {
		return err
	}

	y.mu.Lock()
	defer y.mu.Unlock()
	if y.node == nil {
		return fmt.Errorf("node not started")
	}

	if _, err := y.node.Listen(u, ""); err != nil {
		return err
	}

	found := false
	for _, l := range y.listeners {
		if l == listenURI {
			found = true
			break
		}
	}
	if !found {
		y.listeners = append(y.listeners, listenURI)
	}
	return nil
}

func (y *YggManager) sendPacket(destKeyHex string, payload ChatPayload) error {
	destKeyBytes, err := hex.DecodeString(destKeyHex)
	if err != nil {
		return fmt.Errorf("invalid destination public key: %w", err)
	}

	if len(destKeyBytes) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid key size")
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Prefix packet with 4-byte magic header "YGGC"
	packet := append([]byte("YGGC"), payloadBytes...)

	y.mu.RLock()
	node := y.node
	y.mu.RUnlock()

	if node == nil {
		return fmt.Errorf("node not initialized")
	}

	addr := iwt.Addr(destKeyBytes)
	_, err = node.WriteTo(packet, addr)
	return err
}

func (y *YggManager) SendChatMessage(destKeyHex string, senderName string, text string) error {
	payload := ChatPayload{
		SenderName: senderName,
		Text:       text,
		Timestamp:  time.Now().Unix(),
		Type:        "chat",
	}
	return y.sendPacket(destKeyHex, payload)
}

func (y *YggManager) SendPingMessage(destKeyHex string, senderName string, isPong bool) error {
	msgType := "ping"
	if isPong {
		msgType = "pong"
	}
	payload := ChatPayload{
		SenderName: senderName,
		Timestamp:  time.Now().Unix(),
		Type:        msgType,
	}
	return y.sendPacket(destKeyHex, payload)
}

func (y *YggManager) SendContactRequest(destKeyHex string, senderName string, isAccept bool, ecdhPubKey string) error {
	msgType := "contact_req"
	if isAccept {
		msgType = "contact_acc"
	}
	payload := ChatPayload{
		SenderName: senderName,
		Timestamp:  time.Now().Unix(),
		Type:       msgType,
		ECDHPubKey: ecdhPubKey,
	}
	return y.sendPacket(destKeyHex, payload)
}

func (y *YggManager) SendEncryptedChatMessage(destKeyHex string, senderName string, text string, sharedSecretHex string) error {
	secretBytes, err := hex.DecodeString(sharedSecretHex)
	if err != nil {
		return err
	}

	block, err := aes.NewCipher(secretBytes)
	if err != nil {
		return err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	ciphertext := aesgcm.Seal(nil, nonce, []byte(text), nil)

	payload := ChatPayload{
		SenderName:  senderName,
		Text:        hex.EncodeToString(ciphertext),
		Timestamp:   time.Now().Unix(),
		Type:        "chat",
		IsEncrypted: true,
		Nonce:       hex.EncodeToString(nonce),
	}
	return y.sendPacket(destKeyHex, payload)
}

func (y *YggManager) SendReadReceipt(destKeyHex string, senderName string, lastMsgTimestamp int64) error {
	payload := ChatPayload{
		SenderName: senderName,
		Timestamp:  time.Now().Unix(),
		Type:       "read",
		Text:       fmt.Sprintf("%d", lastMsgTimestamp),
	}
	return y.sendPacket(destKeyHex, payload)
}

func (y *YggManager) SendFileChunk(destKeyHex string, senderName string, filename string, chunkIdx int, totalChunks int, data []byte) error {
	payload := ChatPayload{
		SenderName:  senderName,
		Timestamp:   time.Now().Unix(),
		Type:        "file_chunk",
		FileName:    filename,
		ChunkIdx:    chunkIdx,
		TotalChunks: totalChunks,
		Data:        data,
	}
	return y.sendPacket(destKeyHex, payload)
}

func (y *YggManager) GetPeersInfo() []core.PeerInfo {
	y.mu.RLock()
	defer y.mu.RUnlock()
	if y.node == nil {
		return nil
	}
	return y.node.GetPeers()
}

func (y *YggManager) readPackets() {
	defer y.wg.Done()

	buf := make([]byte, 65535)
	for {
		y.mu.RLock()
		node := y.node
		running := y.running
		y.mu.RUnlock()

		if !running || node == nil {
			return
		}

		n, from, err := node.ReadFrom(buf)
		if err != nil {
			select {
			case <-y.quit:
				return
			default:
				time.Sleep(100 * time.Millisecond)
				continue
			}
		}

		if n < 4 {
			continue
		}

		// Check magic header
		if string(buf[:4]) != "YGGC" {
			continue
		}

		var payload ChatPayload
		if err := json.Unmarshal(buf[4:n], &payload); err != nil {
			continue
		}
		if payload.Type == "" {
			payload.Type = "chat"
		}

		senderHex := hex.EncodeToString(from.(iwt.Addr))
		y.msgChan <- IncomingMessage{
			SenderKey: senderHex,
			Payload:   payload,
		}
	}
}

func (y *YggManager) MsgChan() <-chan IncomingMessage {
	return y.msgChan
}

func (y *YggManager) RetryPeersNow() {
	y.mu.RLock()
	defer y.mu.RUnlock()
	if y.node != nil {
		y.node.RetryPeersNow()
	}
}

func (y *YggManager) SendShakeMessage(destKeyHex string, senderName string) error {
	payload := ChatPayload{
		SenderName: senderName,
		Timestamp:  time.Now().Unix(),
		Type:       "shake",
	}
	return y.sendPacket(destKeyHex, payload)
}

func (y *YggManager) SendTypingIndicator(destKeyHex string, senderName string) error {
	payload := ChatPayload{
		SenderName: senderName,
		Timestamp:  time.Now().Unix(),
		Type:       "typing",
	}
	return y.sendPacket(destKeyHex, payload)
}
