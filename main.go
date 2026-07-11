package main

import (
	"flag"
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	configFlag := flag.String("config", "yggchat.json", "Config file name")
	usernameFlag := flag.String("username", "", "Optional username override")
	tuiFlag := flag.Bool("tui", false, "Start in Terminal TUI mode instead of Web Console")
	portFlag := flag.Int("port", 8080, "Port for Web Console Daemon")
	flag.Parse()

	SetConfigFilename(*configFlag)

	// Load configuration
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load application configuration: %v", err)
	}

	if *usernameFlag != "" {
		cfg.Username = *usernameFlag
		_ = cfg.Save()
	}

	// Initialize Yggdrasil Manager
	ygg := NewYggManager()
	
	// Start Yggdrasil Core
	err = ygg.Start(cfg.PrivateKey, cfg.Listeners, cfg.Peers)
	if err != nil {
		log.Fatalf("Failed to start Yggdrasil core: %v", err)
	}
	defer ygg.Stop()

	if *tuiFlag {
		// Start TUI Bubble Tea Program
		p := tea.NewProgram(NewModel(cfg, ygg), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Bubble Tea application error: %v\n", err)
		}
	} else {
		// Default: Start Web Server daemon and auto-launch browser
		server := NewWebServer(ygg, cfg, *portFlag)
		if err := server.Start(); err != nil {
			log.Fatalf("Web Server failed to start: %v", err)
		}
	}
}
