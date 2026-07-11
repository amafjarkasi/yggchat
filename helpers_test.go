package main

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestSafeSenderName(t *testing.T) {
	tests := []struct {
		name       string
		senderName string
		senderKey  string
		expected   string
	}{
		{
			name:       "Normal case with long names",
			senderName: "Alice",
			senderKey:  "abcdef1234567890abcdef1234567890",
			expected:   "Alic...7890",
		},
		{
			name:       "Short sender name (< 4 chars)",
			senderName: "Al",
			senderKey:  "abcdef1234567890abcdef1234567890",
			expected:   "abcdef12...",
		},
		{
			name:       "Short sender key (< 4 chars)",
			senderName: "Alice",
			senderKey:  "ab",
			expected:   "Unknown",
		},
		{
			name:       "Empty sender name",
			senderName: "",
			senderKey:  "abcdef1234567890",
			expected:   "abcdef12...",
		},
		{
			name:       "Both empty",
			senderName: "",
			senderKey:  "",
			expected:   "Unknown",
		},
		{
			name:       "Exactly 3 chars - fallback to key",
			senderName: "Ali",
			senderKey:  "1234567890",
			expected:   "12345678...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SafeSenderName(tt.senderName, tt.senderKey)
			if result != tt.expected {
				t.Errorf("SafeSenderName(%q, %q) = %q, want %q", tt.senderName, tt.senderKey, result, tt.expected)
			}
		})
	}
}

func TestEscapeHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No special chars",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "Script tag",
			input:    "<script>alert('xss')</script>",
			expected: "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
		},
		{
			name:     "HTML entities",
			input:    "Tom & Jerry <b>bold</b>",
			expected: "Tom &amp; Jerry &lt;b&gt;bold&lt;/b&gt;",
		},
		{
			name:     "Quotes",
			input:    `He said "hello"`,
			expected: "He said &#34;hello&#34;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EscapeHTML(tt.input)
			if result != tt.expected {
				t.Errorf("EscapeHTML(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal filename",
			input:    "document.txt",
			expected: "document.txt",
		},
		{
			name:     "Path traversal attack",
			input:    "../../etc/passwd",
			expected: "passwd",
		},
		{
			name:     "Windows path traversal",
			input:    "..\\..\\windows\\system32\\config\\sam",
			expected: "sam",
		},
		{
			name:     "Null byte injection",
			input:    "file.txt\x00.jpg",
			expected: "file.txt.jpg",
		},
		{
			name:     "Dot only",
			input:    ".",
			expected: "unnamed_file",
		},
		{
			name:     "Double dot only",
			input:    "..",
			expected: "unnamed_file",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "unnamed_file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDeriveSharedSecret(t *testing.T) {
	// Generate two key pairs
	privKey1, err := generateECDHPrivateKey()
	if err != nil {
		t.Fatalf("Failed to generate key 1: %v", err)
	}

	privKey2, err := generateECDHPrivateKey()
	if err != nil {
		t.Fatalf("Failed to generate key 2: %v", err)
	}

	// Get public keys
	pubKey1, err := GetMyECDHPublicKeyHex(privKey1)
	if err != nil {
		t.Fatalf("Failed to get public key 1: %v", err)
	}

	pubKey2, err := GetMyECDHPublicKeyHex(privKey2)
	if err != nil {
		t.Fatalf("Failed to get public key 2: %v", err)
	}

	// Derive shared secrets from both sides
	secret1, err := DeriveSharedSecret(privKey1, pubKey2)
	if err != nil {
		t.Fatalf("Failed to derive secret 1: %v", err)
	}

	secret2, err := DeriveSharedSecret(privKey2, pubKey1)
	if err != nil {
		t.Fatalf("Failed to derive secret 2: %v", err)
	}

	// Both sides should derive the same secret
	if secret1 != secret2 {
		t.Errorf("Shared secrets don't match: %q != %q", secret1, secret2)
	}

	// Secret should be 64 hex characters (32 bytes)
	if len(secret1) != 64 {
		t.Errorf("Expected 64 char hex secret, got %d chars", len(secret1))
	}
}

func TestDeriveSharedSecretInvalidInput(t *testing.T) {
	// Test with invalid private key
	_, err := DeriveSharedSecret("invalid", "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890")
	if err == nil {
		t.Error("Expected error for invalid private key, got nil")
	}

	// Test with invalid public key
	privKey, _ := generateECDHPrivateKey()
	_, err = DeriveSharedSecret(privKey, "invalid")
	if err == nil {
		t.Error("Expected error for invalid public key, got nil")
	}
}

func TestDecryptMessage(t *testing.T) {
	// Generate shared secret
	privKey1, _ := generateECDHPrivateKey()
	privKey2, _ := generateECDHPrivateKey()
	pubKey2, _ := GetMyECDHPublicKeyHex(privKey2)
	sharedSecret, _ := DeriveSharedSecret(privKey1, pubKey2)

	// Test that DecryptMessage with invalid input returns error
	_, err := DecryptMessage("invalid", "invalid", "invalid")
	if err == nil {
		t.Error("Expected error for invalid input, got nil")
	}

	// Test with invalid hex
	_, err = DecryptMessage(sharedSecret, "nothex", "nothex")
	if err == nil {
		t.Error("Expected error for invalid hex, got nil")
	}
}

func TestIsContactRequestAllowed(t *testing.T) {
	// Reset rate limiter for test
	contactReqRateLimiter.mu.Lock()
	contactReqRateLimiter.counts = make(map[string][]time.Time)
	contactReqRateLimiter.mu.Unlock()

	senderKey := "test_sender_key"

	// First request should be allowed
	if !IsContactRequestAllowed(senderKey) {
		t.Error("First request should be allowed")
	}

	// Send up to the limit
	for i := 1; i < MaxContactReqRate; i++ {
		if !IsContactRequestAllowed(senderKey) {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// Next request should be blocked
	if IsContactRequestAllowed(senderKey) {
		t.Error("Request should be blocked after rate limit exceeded")
	}

	// Different sender should be allowed
	if !IsContactRequestAllowed("different_sender") {
		t.Error("Different sender should be allowed")
	}
}

func TestConstants(t *testing.T) {
	// Verify constants have reasonable values
	if ChunkSize != 8192 {
		t.Errorf("ChunkSize should be 8192, got %d", ChunkSize)
	}
	if TimeFormat != "15:04:05" {
		t.Errorf("TimeFormat should be '15:04:05', got %q", TimeFormat)
	}
	if TypingDebounceSec != 2 {
		t.Errorf("TypingDebounceSec should be 2, got %d", TypingDebounceSec)
	}
	if TypingDisplaySec != 3 {
		t.Errorf("TypingDisplaySec should be 3, got %d", TypingDisplaySec)
	}
	if OfflineRetrySec != 5 {
		t.Errorf("OfflineRetrySec should be 5, got %d", OfflineRetrySec)
	}
	if PeerStatusTickSec != 2 {
		t.Errorf("PeerStatusTickSec should be 2, got %d", PeerStatusTickSec)
	}
	if DiscoveryBeaconSec != 5 {
		t.Errorf("DiscoveryBeaconSec should be 5, got %d", DiscoveryBeaconSec)
	}
	if MaxSenderNameLen != 50 {
		t.Errorf("MaxSenderNameLen should be 50, got %d", MaxSenderNameLen)
	}
	if MaxContactReqRate != 5 {
		t.Errorf("MaxContactReqRate should be 5, got %d", MaxContactReqRate)
	}
}

func TestStripTimestamp(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "With timestamp",
			input:    "[15:04:05] Hello World",
			expected: "Hello World",
		},
		{
			name:     "Without timestamp",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Short string",
			input:    "[15:04]",
			expected: "[15:04]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripTimestamp(tt.input)
			if result != tt.expected {
				t.Errorf("stripTimestamp(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestStripANSI(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No ANSI codes",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "With ANSI color codes",
			input:    "\x1b[31mRed Text\x1b[0m",
			expected: "Red Text",
		},
		{
			name:     "Multiple ANSI codes",
			input:    "\x1b[1m\x1b[32mBold Green\x1b[0m Normal",
			expected: "Bold Green Normal",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripANSI(tt.input)
			if result != tt.expected {
				t.Errorf("stripANSI(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsImageFile(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{name: "PNG", input: "image.png", expected: true},
		{name: "JPG", input: "photo.jpg", expected: true},
		{name: "JPEG", input: "photo.jpeg", expected: true},
		{name: "Uppercase PNG", input: "IMAGE.PNG", expected: true},
		{name: "Text file", input: "document.txt", expected: false},
		{name: "No extension", input: "file", expected: false},
		{name: "GIF not supported", input: "animation.gif", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isImageFile(tt.input)
			if result != tt.expected {
				t.Errorf("isImageFile(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestHistorySaveLoad(t *testing.T) {
	// Create a temporary history
	history := map[string][]string{
		"key1": {"message1", "message2"},
		"key2": {"message3"},
	}

	// Save history
	err := SaveHistory(history)
	if err != nil {
		t.Fatalf("Failed to save history: %v", err)
	}

	// Load history
	loaded := LoadHistory()

	// Verify contents
	if len(loaded) != len(history) {
		t.Errorf("Expected %d keys, got %d", len(history), len(loaded))
	}

	for key, messages := range history {
		loadedMessages, ok := loaded[key]
		if !ok {
			t.Errorf("Missing key %q in loaded history", key)
			continue
		}
		if len(loadedMessages) != len(messages) {
			t.Errorf("Key %q: expected %d messages, got %d", key, len(messages), len(loadedMessages))
			continue
		}
		for i, msg := range messages {
			if loadedMessages[i] != msg {
				t.Errorf("Key %q, message %d: expected %q, got %q", key, i, msg, loadedMessages[i])
			}
		}
	}

	// Cleanup
	historyPath := GetHistoryFilename()
	_ = os.Remove(historyPath)
	_ = os.Remove(historyPath + ".tmp")
}

func TestGetHistoryFilename(t *testing.T) {
	// Default config filename
	SetConfigFilename("yggchat.json")
	result := GetHistoryFilename()
	if !strings.HasSuffix(result, "_history.json") {
		t.Errorf("Expected history filename to end with '_history.json', got %q", result)
	}

	// Custom config filename
	SetConfigFilename("alice.json")
	result = GetHistoryFilename()
	if !strings.Contains(result, "alice") {
		t.Errorf("Expected history filename to contain 'alice', got %q", result)
	}

	// Reset to default
	SetConfigFilename("yggchat.json")
}
