package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Constants for magic numbers used across the codebase
const (
	ChunkSize          = 8192
	TimeFormat         = "15:04:05"
	TypingDebounceSec  = 2
	TypingDisplaySec   = 3
	OfflineRetrySec    = 5
	PeerStatusTickSec  = 2
	DiscoveryBeaconSec = 5
	MaxSenderNameLen   = 50
	MaxContactReqRate  = 5 // max contact requests per minute per sender
)

// ConfigMutex protects concurrent access to configFilename and config operations
var ConfigMutex sync.RWMutex

// contactReqRateLimiter tracks contact request rates per sender
var contactReqRateLimiter = struct {
	mu      sync.Mutex
	counts  map[string][]time.Time
}{
	counts: make(map[string][]time.Time),
}

// SafeSenderName formats a sender name safely, preventing index-out-of-range panics
func SafeSenderName(senderName string, senderKey string) string {
	if len(senderName) > MaxSenderNameLen {
		senderName = senderName[:MaxSenderNameLen]
	}
	
	if len(senderName) >= 4 && len(senderKey) >= 4 {
		return fmt.Sprintf("%s...%s", senderName[:4], senderKey[len(senderKey)-4:])
	}
	if len(senderKey) >= 8 {
		return senderKey[:8] + "..."
	}
	return "Unknown"
}

// EscapeHTML escapes user input for safe HTML insertion in web UI
func EscapeHTML(s string) string {
	return html.EscapeString(s)
}

// SanitizeFilename removes path traversal components from filenames
func SanitizeFilename(filename string) string {
	// filepath.Base strips directory components, preventing path traversal
	safe := filepath.Base(filename)
	// Additional safety: remove null bytes and other problematic characters
	safe = strings.ReplaceAll(safe, "\x00", "")
	if safe == "." || safe == ".." || safe == "" {
		return "unnamed_file"
	}
	return safe
}

// DeriveSharedSecret performs ECDH key exchange and derives an AES-256 key
func DeriveSharedSecret(myPrivKeyHex string, theirPubKeyHex string) (string, error) {
	privBytes, err := hex.DecodeString(myPrivKeyHex)
	if err != nil {
		return "", fmt.Errorf("invalid private key: %w", err)
	}
	
	priv, err := ecdh.X25519().NewPrivateKey(privBytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}
	
	pubBytes, err := hex.DecodeString(theirPubKeyHex)
	if err != nil {
		return "", fmt.Errorf("invalid public key: %w", err)
	}
	
	pub, err := ecdh.X25519().NewPublicKey(pubBytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse public key: %w", err)
	}
	
	sharedSecret, err := priv.ECDH(pub)
	if err != nil {
		return "", fmt.Errorf("ECDH failed: %w", err)
	}
	
	aesKey := sha256.Sum256(sharedSecret)
	return hex.EncodeToString(aesKey[:]), nil
}

// GetMyECDHPublicKeyHex returns the public key hex for our ECDH private key
func GetMyECDHPublicKeyHex(privKeyHex string) (string, error) {
	privBytes, err := hex.DecodeString(privKeyHex)
	if err != nil {
		return "", err
	}
	priv, err := ecdh.X25519().NewPrivateKey(privBytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(priv.PublicKey().Bytes()), nil
}

// DecryptMessage decrypts an AES-GCM encrypted message
func DecryptMessage(sharedSecretHex string, ciphertextHex string, nonceHex string) (string, error) {
	secretBytes, err := hex.DecodeString(sharedSecretHex)
	if err != nil {
		return "", err
	}
	
	block, err := aes.NewCipher(secretBytes)
	if err != nil {
		return "", err
	}
	
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	
	nonceBytes, err := hex.DecodeString(nonceHex)
	if err != nil {
		return "", err
	}
	
	cipherBytes, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return "", err
	}
	
	plainBytes, err := aesgcm.Open(nil, nonceBytes, cipherBytes, nil)
	if err != nil {
		return "", err
	}
	
	return string(plainBytes), nil
}

// SaveHistoryAtomic writes history to disk atomically using a temp file + rename
func SaveHistoryAtomic(hist map[string][]string) error {
	ConfigMutex.Lock()
	defer ConfigMutex.Unlock()
	
	historyPath := filepath.Join(".", GetHistoryFilename())
	tmpPath := historyPath + ".tmp"
	
	data, err := marshalJSON(hist)
	if err != nil {
		return err
	}
	
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return err
	}
	
	return os.Rename(tmpPath, historyPath)
}

// IsContactRequestAllowed checks rate limiting for contact requests
func IsContactRequestAllowed(senderKey string) bool {
	contactReqRateLimiter.mu.Lock()
	defer contactReqRateLimiter.mu.Unlock()
	
	now := time.Now()
	cutoff := now.Add(-time.Minute)
	
	// Clean old entries
	times := contactReqRateLimiter.counts[senderKey]
	var valid []time.Time
	for _, t := range times {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	
	if len(valid) >= MaxContactReqRate {
		contactReqRateLimiter.counts[senderKey] = valid
		return false
	}
	
	contactReqRateLimiter.counts[senderKey] = append(valid, now)
	return true
}

// marshalJSON is a helper that marshals with indentation
func marshalJSON(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}
