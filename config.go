package main

import (
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
)

type Contact struct {
	PublicKey    string `json:"publicKey"`
	Nickname     string `json:"nickname"`
	SharedSecret string `json:"sharedSecret,omitempty"` // hex-encoded AES key for E2EE
	Group        string `json:"group,omitempty"`         // Contact group (e.g., "Friends", "Work")
	Blocked      bool   `json:"blocked,omitempty"`       // Whether contact is blocked
}

type AppConfig struct {
	PrivateKey     string             `json:"privateKey"`     // hex-encoded ed25519 private key
	ECDHPrivateKey string             `json:"ecdhPrivateKey"` // hex-encoded Curve25519 ecdh key
	Peers          []string           `json:"peers"`
	Listeners      []string           `json:"listeners"`
	Contacts       map[string]Contact `json:"contacts"`
	Username       string             `json:"username"`
	
	// Feature settings
	AutoDeleteDays int                `json:"autoDeleteDays,omitempty"`  // Auto-delete messages after N days (0 = disabled)
	ShowTyping     bool               `json:"showTyping,omitempty"`      // Show typing indicators
	ShowReadReceipts bool             `json:"showReadReceipts,omitempty"` // Show read receipts
	NotificationSound string          `json:"notificationSound,omitempty"` // Notification sound theme
	CustomTheme    *CustomThemeConfig `json:"customTheme,omitempty"`     // Custom theme colors
	
	// Privacy settings
	BurnAfterRead  bool  `json:"burnAfterRead,omitempty"`   // Delete messages after they are read
	BurnTimeoutSec int   `json:"burnTimeoutSec,omitempty"`  // Seconds to wait before burning (0 = immediate)
	StripMetadata  bool  `json:"stripMetadata,omitempty"`   // Strip metadata from messages
	DecoyTraffic   bool  `json:"decoyTraffic,omitempty"`    // Generate decoy traffic to prevent traffic analysis
}

type CustomThemeConfig struct {
	Name    string `json:"name"`
	Base    string `json:"base"`
	Mantle  string `json:"mantle"`
	Crust   string `json:"crust"`
	Text    string `json:"text"`
	Subtext string `json:"subtext"`
	Muted   string `json:"muted"`
	Overlay string `json:"overlay"`
	Primary string `json:"primary"`
	Accent  string `json:"accent"`
	Success string `json:"success"`
	Warning string `json:"warning"`
	Error   string `json:"error"`
	Info    string `json:"info"`
}

var configFilename = "yggchat.json"

func SetConfigFilename(name string) {
	ConfigMutex.Lock()
	defer ConfigMutex.Unlock()
	configFilename = name
}

func getConfigFilename() string {
	ConfigMutex.RLock()
	defer ConfigMutex.RUnlock()
	return configFilename
}

func generateECDHPrivateKey() (string, error) {
	key, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(key.Bytes()), nil
}

func LoadConfig() (*AppConfig, error) {
	configPath := filepath.Join(".", getConfigFilename())
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return CreateDefaultConfig()
		}
		return nil, err
	}

	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	
	if cfg.Contacts == nil {
		cfg.Contacts = make(map[string]Contact)
	}

	// Dynamic migration: generate ECDH private key if missing in existing config
	if cfg.ECDHPrivateKey == "" {
		ecdhPriv, err := generateECDHPrivateKey()
		if err == nil {
			cfg.ECDHPrivateKey = ecdhPriv
			_ = cfg.Save()
		}
	}

	return &cfg, nil
}

func (c *AppConfig) Save() error {
	configPath := filepath.Join(".", getConfigFilename())
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0600)
}

func CreateDefaultConfig() (*AppConfig, error) {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	ecdhPriv, err := generateECDHPrivateKey()
	if err != nil {
		return nil, err
	}

	cfg := &AppConfig{
		PrivateKey:     hex.EncodeToString(priv),
		ECDHPrivateKey: ecdhPriv,
		Peers:          []string{},
		Listeners:      []string{"tcp://0.0.0.0:9000"}, // Default listener port
		Contacts:       make(map[string]Contact),
		Username:       "YggUser",
	}

	if err := cfg.Save(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func GetHistoryFilename() string {
	cfgName := getConfigFilename()
	ext := filepath.Ext(cfgName)
	base := cfgName[:len(cfgName)-len(ext)]
	return base + "_history.json"
}

func LoadHistory() map[string][]string {
	historyPath := filepath.Join(".", GetHistoryFilename())
	data, err := os.ReadFile(historyPath)
	if err != nil {
		return make(map[string][]string)
	}

	var hist map[string][]string
	if err := json.Unmarshal(data, &hist); err != nil {
		return make(map[string][]string)
	}
	return hist
}

func SaveHistory(hist map[string][]string) error {
	historyPath := filepath.Join(".", GetHistoryFilename())
	tmpPath := historyPath + ".tmp"
	
	data, err := json.MarshalIndent(hist, "", "  ")
	if err != nil {
		return err
	}
	
	// Write to temp file first, then atomic rename to prevent corruption
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return err
	}
	
	return os.Rename(tmpPath, historyPath)
}
