package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
)

func TestConfigLoadSave(t *testing.T) {
	tempFile := "yggchat_test_config.json"
	
	cfg := &AppConfig{
		PrivateKey: "00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff",
		Peers:      []string{"tcp://127.0.0.1:9000"},
		Listeners:  []string{"tcp://0.0.0.0:9000"},
		Contacts:   make(map[string]Contact),
		Username:   "TestUser",
	}

	cfg.Contacts["testkey"] = Contact{
		PublicKey: "testkey",
		Nickname:  "TestFriend",
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	err = os.WriteFile(tempFile, data, 0600)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	defer os.Remove(tempFile)

	readData, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var loadedCfg AppConfig
	if err := json.Unmarshal(readData, &loadedCfg); err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	if loadedCfg.Username != "TestUser" {
		t.Errorf("Expected username 'TestUser', got '%s'", loadedCfg.Username)
	}
	if len(loadedCfg.Peers) != 1 || loadedCfg.Peers[0] != "tcp://127.0.0.1:9000" {
		t.Errorf("Peers list mismatch")
	}
	if contact, ok := loadedCfg.Contacts["testkey"]; !ok || contact.Nickname != "TestFriend" {
		t.Errorf("Contacts map mismatch")
	}
}

func TestChatPacketFormatting(t *testing.T) {
	payload := ChatPayload{
		SenderName: "Alice",
		Text:       "Hello, Bob!",
		Timestamp:  1700000000,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	packet := append([]byte("YGGC"), payloadBytes...)

	if !bytes.Equal(packet[:4], []byte("YGGC")) {
		t.Errorf("Expected packet prefix to be 'YGGC'")
	}

	var decodedPayload ChatPayload
	if err := json.Unmarshal(packet[4:], &decodedPayload); err != nil {
		t.Fatalf("Failed to decode payload: %v", err)
	}

	if decodedPayload.SenderName != "Alice" {
		t.Errorf("Expected sender 'Alice', got '%s'", decodedPayload.SenderName)
	}
	if decodedPayload.Text != "Hello, Bob!" {
		t.Errorf("Expected text 'Hello, Bob!', got '%s'", decodedPayload.Text)
	}
	if decodedPayload.Timestamp != 1700000000 {
		t.Errorf("Expected timestamp 1700000000, got %d", decodedPayload.Timestamp)
	}
}

func TestKeyDecoding(t *testing.T) {
	hexKey := "1a69d075c7b16d9d869fa61df34294bab5b104d7e3f0f0fb45a69082d5ab5ce12d6658195e9aa24cae0a8a53cfaad24375bfd78850237e1bfb660323e94b1b4a"
	
	privKeyBytes, err := hex.DecodeString(hexKey)
	if err != nil {
		t.Fatalf("Failed to decode hex private key: %v", err)
	}

	if len(privKeyBytes) != 64 {
		t.Errorf("Expected key length of 64 bytes, got %d", len(privKeyBytes))
	}
}

func TestECDHKeyExchange(t *testing.T) {
	alicePrivHex, err := generateECDHPrivateKey()
	if err != nil {
		t.Fatalf("Failed to generate Alice ECDH key: %v", err)
	}
	alicePrivBytes, _ := hex.DecodeString(alicePrivHex)
	alicePriv, err := ecdh.X25519().NewPrivateKey(alicePrivBytes)
	if err != nil {
		t.Fatalf("Failed to parse Alice ECDH private key: %v", err)
	}
	alicePub := alicePriv.PublicKey()

	bobPrivHex, err := generateECDHPrivateKey()
	if err != nil {
		t.Fatalf("Failed to generate Bob ECDH key: %v", err)
	}
	bobPrivBytes, _ := hex.DecodeString(bobPrivHex)
	bobPriv, err := ecdh.X25519().NewPrivateKey(bobPrivBytes)
	if err != nil {
		t.Fatalf("Failed to parse Bob ECDH private key: %v", err)
	}
	bobPub := bobPriv.PublicKey()

	secretAlice, err := alicePriv.ECDH(bobPub)
	if err != nil {
		t.Fatalf("Alice ECDH failed: %v", err)
	}
	secretBob, err := bobPriv.ECDH(alicePub)
	if err != nil {
		t.Fatalf("Bob ECDH failed: %v", err)
	}

	keyAlice := sha256.Sum256(secretAlice)
	keyBob := sha256.Sum256(secretBob)

	if !bytes.Equal(keyAlice[:], keyBob[:]) {
		t.Fatalf("Derived AES keys do not match!")
	}

	msg := "Top Secret E2EE Chat Message"
	block, err := aes.NewCipher(keyAlice[:])
	if err != nil {
		t.Fatalf("NewCipher failed: %v", err)
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatalf("NewGCM failed: %v", err)
	}

	nonce := make([]byte, aesgcm.NonceSize())
	ciphertext := aesgcm.Seal(nil, nonce, []byte(msg), nil)

	blockBob, err := aes.NewCipher(keyBob[:])
	if err != nil {
		t.Fatalf("Bob NewCipher failed: %v", err)
	}
	aesgcmBob, err := cipher.NewGCM(blockBob)
	if err != nil {
		t.Fatalf("Bob NewGCM failed: %v", err)
	}

	decrypted, err := aesgcmBob.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if string(decrypted) != msg {
		t.Errorf("Expected decrypted message '%s', got '%s'", msg, string(decrypted))
	}
}
