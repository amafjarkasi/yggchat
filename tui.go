package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/yggdrasil-network/yggdrasil-go/src/core"
)

// Active views
const (
	ViewChat = iota
	ViewPeers
	ViewSettings
)

// Focused panels in ViewChat
const (
	PanelSidebar = iota
	PanelViewport
	PanelInput
)

// Ticker message
type PeerStatusTickMsg struct {
	Peers []core.PeerInfo
}

// Background retry tick message for offline queue
type OfflineRetryTickMsg struct{}

// File transmission progress message
type FileProgressMsg struct {
	ContactKey string
	Filename   string
	ChunkIdx   int
	Total      int
	Done       bool
	Err        error
	Data       []byte
}

type OfflineMessage struct {
	DestKey  string `json:"dest_key"`
	Text     string `json:"text"`
	TimeStr  string `json:"time_str"`
}

// Model representing TUI state
type Model struct {
	cfg          *AppConfig
	ygg          *YggManager
	discoveryMgr *DiscoveryManager
	styles       UIStyles
	width        int
	height       int
	themeName    string
	
	activeTab    int
	activePanel  int // focus for ViewChat
	
	// Chat View state
	selectedContactIdx int
	activeChatKey      string
	chatHistory        map[string][]string // key -> messages (strings formatted with Lip Gloss)
	viewport           viewport.Model
	textInput          textinput.Model
	unreadMessages     map[string]bool
	
	// Input history scroller
	inputHistory []string
	historyIdx   int

	// Peers Manager state
	selectedPeerIdx int
	peersInfo       []core.PeerInfo

	// Modal popups state
	modalActive       bool
	modalType         string // "add_contact", "add_peer", "error", "contact_request"
	contactNameInput  textinput.Model
	contactKeyInput   textinput.Model
	peerURIInput      textinput.Model
	errorMessage      string
	
	// Incoming contact request data
	incomingReqKey    string
	incomingReqName   string
	incomingReqPubKey string

	// Typing indicators and toggle settings
	lastTypingSent time.Time
	typingStatus   map[string]time.Time
	showTimestamps bool
}

func NewModel(cfg *AppConfig, ygg *YggManager) Model {
	themeName := "Catppuccin Mocha"
	styles := GetStyles(themeName)

	// Initialize widgets
	ti := textinput.New()
	ti.Placeholder = "Type a message or /command..."
	ti.Focus()
	ti.Prompt = " ❱ "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(styles.Colors.Info))
	ti.CharLimit = 1000

	vp := viewport.New(0, 0)
	vp.SetContent("Select a contact to start chatting.")

	// Modal inputs
	cName := textinput.New()
	cName.Placeholder = "Nickname (e.g. Alice)"
	cName.Focus()
	cName.Prompt = " 👤 "

	cKey := textinput.New()
	cKey.Placeholder = "Yggdrasil Public Key (Hex)"
	cKey.Prompt = " 🔑 "

	pURI := textinput.New()
	pURI.Placeholder = "Peer Address (e.g. tcp://1.2.3.4:9000)"
	pURI.Focus()
	pURI.Prompt = " 🔗 "

	// Start local subnet discovery
	dm := NewDiscoveryManager(ygg, cfg.Listeners)
	dm.Start()

	return Model{
		cfg:                cfg,
		ygg:                ygg,
		discoveryMgr:       dm,
		styles:             styles,
		themeName:          themeName,
		activeTab:          ViewChat,
		activePanel:        PanelInput,
		chatHistory:        LoadHistory(),
		viewport:           vp,
		textInput:          ti,
		selectedContactIdx: 0,
		selectedPeerIdx:    0,
		contactNameInput:   cName,
		contactKeyInput:    cKey,
		peerURIInput:       pURI,
		unreadMessages:     make(map[string]bool),
		inputHistory:       []string{},
		historyIdx:         0,
		typingStatus:       make(map[string]time.Time),
		showTimestamps:     true,
	}
}

func (m Model) Init() tea.Cmd {
	readCmd := func() tea.Msg {
		msg := <-m.ygg.MsgChan()
		return msg
	}

	tickCmd := tea.Tick(PeerStatusTickSec*time.Second, func(t time.Time) tea.Msg {
		return PeerStatusTickMsg{Peers: m.ygg.GetPeersInfo()}
	})

	retryCmd := tea.Tick(OfflineRetrySec*time.Second, func(t time.Time) tea.Msg {
		return OfflineRetryTickMsg{}
	})

	return tea.Batch(readCmd, tickCmd, retryCmd, textinput.Blink)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.recalculateLayouts()
		if m.activeChatKey != "" {
			m.viewport.SetContent(m.getWrappedHistory(m.activeChatKey))
			m.viewport.GotoBottom()
		}

	case IncomingMessage:
		senderKey := msg.SenderKey
		senderName := msg.Payload.SenderName
		
		if contact, ok := m.cfg.Contacts[senderKey]; ok {
			senderName = contact.Nickname
		} else {
			senderName = SafeSenderName(senderName, senderKey)
		}

		timeStr := time.Unix(msg.Payload.Timestamp, 0).Format("15:04:05")

		switch msg.Payload.Type {
		case "typing":
			m.typingStatus[senderKey] = time.Now()
			nextReadCmd := func() tea.Msg {
				incoming := <-m.ygg.MsgChan()
				return incoming
			}
			return m, nextReadCmd

		case "ping":
			_ = m.ygg.SendPingMessage(senderKey, m.cfg.Username, true)

		case "pong":
			m.appendSystemMessage(fmt.Sprintf("Ping response from %s: Successful!", senderName))

		case "shake":
			fmt.Print("\a") // Play terminal audible bell
			m.appendSystemMessage(fmt.Sprintf("⚡ %s sent you a Nudge/Shake! ⚡", senderName))

		case "contact_req":
			// Rate limit contact requests to prevent flooding
			if !IsContactRequestAllowed(senderKey) {
				m.appendSystemMessage(fmt.Sprintf("Contact request rate limit exceeded from %s", senderName))
				break
			}
			m.modalActive = true
			m.modalType = "contact_request"
			m.incomingReqKey = senderKey
			m.incomingReqName = msg.Payload.SenderName
			m.incomingReqPubKey = msg.Payload.ECDHPubKey

		case "contact_acc":
			// Process Bob's accepted request key
			aesKeyHex, err := DeriveSharedSecret(m.cfg.ECDHPrivateKey, msg.Payload.ECDHPubKey)
			if err == nil {
				if c, ok := m.cfg.Contacts[senderKey]; ok {
					c.SharedSecret = aesKeyHex
					m.cfg.Contacts[senderKey] = c
					_ = m.cfg.Save()
				}
			}
			m.appendSystemMessage(fmt.Sprintf("✓ %s accepted contact request. [⚡ E2EE Established]", senderName))

		case "read":
			// Update single ticks to double ticks
			history := m.chatHistory[senderKey]
			singleTick := lipgloss.NewStyle().Foreground(lipgloss.Color(m.styles.Colors.Subtext)).Render("✓")
			doubleTick := lipgloss.NewStyle().Foreground(lipgloss.Color(m.styles.Colors.Success)).Render("✓✓")

			for i, line := range history {
				if strings.HasSuffix(line, singleTick) {
					history[i] = strings.TrimSuffix(line, singleTick) + doubleTick
				}
			}
			_ = SaveHistory(m.chatHistory)

			if m.activeChatKey == senderKey {
				m.viewport.SetContent(m.getWrappedHistory(m.activeChatKey))
				m.viewport.GotoBottom()
			}

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
			
			if _, ok := m.chatHistory[senderKey]; !ok {
				m.chatHistory[senderKey] = []string{}
			}
			
			history := m.chatHistory[senderKey]
			progressStr := fmt.Sprintf("[📤 Receiving %s: %d/%d (%d%%)]", msg.Payload.FileName, msg.Payload.ChunkIdx+1, msg.Payload.TotalChunks, (msg.Payload.ChunkIdx+1)*100/msg.Payload.TotalChunks)
			
			if msg.Payload.ChunkIdx == msg.Payload.TotalChunks-1 {
				progressStr = fmt.Sprintf("[✓ Received %s saved to ./downloads/%s]", msg.Payload.FileName, msg.Payload.FileName)
			}
			
			if len(history) > 0 && strings.HasPrefix(history[len(history)-1], "[📤 Receiving") {
				history[len(history)-1] = progressStr
			} else {
				m.chatHistory[senderKey] = append(history, progressStr)
			}

			// Receive image preview check
			if msg.Payload.ChunkIdx == msg.Payload.TotalChunks-1 && isImageFile(msg.Payload.FileName) {
				previewWidth := m.viewport.Width - 8
				if previewWidth < 10 {
					previewWidth = 10
				}
				ansiArt, err := RenderImageFile(filePath, previewWidth)
				if err == nil {
					m.chatHistory[senderKey] = append(m.chatHistory[senderKey], "", "🖼️ [Image Preview]:", ansiArt)
				}
			}

			_ = SaveHistory(m.chatHistory)
			
			if m.activeChatKey == senderKey {
				m.viewport.SetContent(m.getWrappedHistory(m.activeChatKey))
				m.viewport.GotoBottom()
			}

		default: // "chat"
			text := msg.Payload.Text
			
			// E2EE Decrypt if encrypted
			if msg.Payload.IsEncrypted {
				if contact, ok := m.cfg.Contacts[senderKey]; ok && contact.SharedSecret != "" {
					decrypted, err := DecryptMessage(contact.SharedSecret, msg.Payload.Text, msg.Payload.Nonce)
					if err != nil {
						text = lipgloss.NewStyle().Foreground(lipgloss.Color(m.styles.Colors.Error)).Render("[🔒 Encrypted: Decryption failed]")
					} else {
						text = decrypted
					}
				} else {
					text = lipgloss.NewStyle().Foreground(lipgloss.Color(m.styles.Colors.Error)).Render("[🔒 Encrypted: Shared secret missing]")
				}
			}

			// Play audible bell notification if chat is in background or active chat is different
			if senderKey != m.activeChatKey || m.activeTab != ViewChat {
				fmt.Print("\a")
			}

			nameTag := lipgloss.NewStyle().
				Foreground(lipgloss.Color(m.styles.Colors.Accent)).
				Bold(true).
				Render(senderName)

			bubble := fmt.Sprintf("[%s] %s: %s", timeStr, nameTag, text)
			m.chatHistory[senderKey] = append(m.chatHistory[senderKey], bubble)
			_ = SaveHistory(m.chatHistory)

			// Send back read receipt immediately if active
			if m.activeChatKey == senderKey && m.activePanel == PanelViewport {
				_ = m.ygg.SendReadReceipt(senderKey, m.cfg.Username, msg.Payload.Timestamp)
			}

			if m.activeChatKey == senderKey {
				m.viewport.SetContent(m.getWrappedHistory(m.activeChatKey))
				m.viewport.GotoBottom()
			} else {
				m.unreadMessages[senderKey] = true
			}
		}

		nextReadCmd := func() tea.Msg {
			incoming := <-m.ygg.MsgChan()
			return incoming
		}
		cmds = append(cmds, nextReadCmd)

	case FileProgressMsg:
		if msg.Err != nil {
			m.appendSystemMessage(fmt.Sprintf("File send error: %v", msg.Err))
			return m, nil
		}

		history := m.chatHistory[msg.ContactKey]
		progressStr := fmt.Sprintf("[📥 Sending %s: %d/%d (%d%%)]", msg.Filename, msg.ChunkIdx+1, msg.Total, (msg.ChunkIdx+1)*100/msg.Total)
		
		if msg.Done {
			progressStr = fmt.Sprintf("[✓ Sent %s successfully!]", msg.Filename)
		}

		if len(history) > 0 && strings.HasPrefix(history[len(history)-1], "[📥 Sending") {
			history[len(history)-1] = progressStr
		} else {
			m.chatHistory[msg.ContactKey] = append(history, progressStr)
		}
		_ = SaveHistory(m.chatHistory)

		if m.activeChatKey == msg.ContactKey {
			m.viewport.SetContent(m.getWrappedHistory(m.activeChatKey))
			m.viewport.GotoBottom()
		}

		if !msg.Done {
			cmds = append(cmds, sendNextChunkCmd(msg.ContactKey, msg.Filename, msg.ChunkIdx+1, msg.Total, msg.Data, m.cfg.Username, m.ygg))
		}

	case PeerStatusTickMsg:
		m.peersInfo = msg.Peers
		
		tickCmd := tea.Tick(PeerStatusTickSec*time.Second, func(t time.Time) tea.Msg {
			return PeerStatusTickMsg{Peers: m.ygg.GetPeersInfo()}
		})
		cmds = append(cmds, tickCmd)

	case OfflineRetryTickMsg:
		// Offline queue scanner
		queue := m.loadOfflineQueue()
		if len(queue) > 0 {
			var newQueue []OfflineMessage
			for _, item := range queue {
				contact, hasContact := m.cfg.Contacts[item.DestKey]
				
				var err error
				if hasContact && contact.SharedSecret != "" {
					err = m.ygg.SendEncryptedChatMessage(item.DestKey, m.cfg.Username, item.Text, contact.SharedSecret)
				} else {
					err = m.ygg.SendChatMessage(item.DestKey, m.cfg.Username, item.Text)
				}

				if err == nil {
					// Delivered! Remove [⏳ PENDING] message block and append standard message
					history := m.chatHistory[item.DestKey]
					pendingPrefix := fmt.Sprintf("[%s] %s %s", item.TimeStr, m.cfg.Username, lipgloss.NewStyle().Foreground(lipgloss.Color(m.styles.Colors.Warning)).Render("(pending)"))
					
					nameTag := lipgloss.NewStyle().
						Foreground(lipgloss.Color(m.styles.Colors.Primary)).
						Bold(true).
						Render(m.cfg.Username)
					singleTick := lipgloss.NewStyle().Foreground(lipgloss.Color(m.styles.Colors.Subtext)).Render("✓")
					bubble := fmt.Sprintf("[%s] %s: %s %s", item.TimeStr, nameTag, item.Text, singleTick)

					for i, line := range history {
						if strings.HasPrefix(line, pendingPrefix) {
							history[i] = bubble
							break
						}
					}
					_ = SaveHistory(m.chatHistory)

					if m.activeChatKey == item.DestKey {
						m.viewport.SetContent(m.getWrappedHistory(m.activeChatKey))
						m.viewport.GotoBottom()
					}
				} else {
					newQueue = append(newQueue, item)
				}
			}
			m.saveOfflineQueue(newQueue)
		}

		retryCmd := tea.Tick(OfflineRetrySec*time.Second, func(t time.Time) tea.Msg {
			return OfflineRetryTickMsg{}
		})
		cmds = append(cmds, retryCmd)

	case tea.KeyMsg:
		if m.activeTab == ViewChat && m.activePanel == PanelInput && m.activeChatKey != "" && msg.Type != tea.KeyEnter && msg.Type != tea.KeyTab && !strings.HasPrefix(msg.String(), "ctrl+") {
			if time.Since(m.lastTypingSent) > TypingDebounceSec*time.Second {
				_ = m.ygg.SendTypingIndicator(m.activeChatKey, m.cfg.Username)
				m.lastTypingSent = time.Now()
			}
		}

		if m.modalActive {
			if m.modalType == "contact_request" {
				switch msg.String() {
				case "y", "Y":
					m.modalActive = false

					// Calculate shared secret key using helper
					aesKeyHex, err := DeriveSharedSecret(m.cfg.ECDHPrivateKey, m.incomingReqPubKey)
					if err != nil {
						m.appendSystemMessage(fmt.Sprintf("Failed to derive shared secret: %v", err))
						return m, nil
					}

					m.cfg.Contacts[m.incomingReqKey] = Contact{
						PublicKey:    m.incomingReqKey,
						Nickname:     m.incomingReqName,
						SharedSecret: aesKeyHex,
					}
					_ = m.cfg.Save()

					bobPubHex, _ := GetMyECDHPublicKeyHex(m.cfg.ECDHPrivateKey)
					_ = m.ygg.SendContactRequest(m.incomingReqKey, m.cfg.Username, true, bobPubHex)
					
					m.selectContact(m.cfg.Contacts[m.incomingReqKey])
					m.selectedContactIdx = len(m.getContactsList()) - 1
					m.appendSystemMessage(fmt.Sprintf("Accepted contact request. [⚡ E2EE Established]"))
				case "n", "N", "esc":
					m.modalActive = false
				}
				return m, nil
			}

			switch msg.String() {
			case "esc":
				m.modalActive = false
				m.textInput.Focus()
				return m, nil
			case "enter":
				m.submitModal()
				return m, nil
			case "tab":
				if m.modalType == "add_contact" {
					if m.contactNameInput.Focused() {
						m.contactNameInput.Blur()
						m.contactKeyInput.Focus()
					} else {
						m.contactKeyInput.Blur()
						m.contactNameInput.Focus()
					}
				}
			}

			var cmd tea.Cmd
			if m.modalType == "add_contact" {
				if m.contactNameInput.Focused() {
					m.contactNameInput, cmd = m.contactNameInput.Update(msg)
				} else {
					m.contactKeyInput, cmd = m.contactKeyInput.Update(msg)
				}
			} else if m.modalType == "add_peer" {
				m.peerURIInput, cmd = m.peerURIInput.Update(msg)
			}
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			return m, tea.Batch(cmds...)
		}

		switch msg.String() {
		case "ctrl+c":
			m.ygg.Stop()
			m.discoveryMgr.Stop()
			return m, tea.Quit
		case "ctrl+t":
			switch m.themeName {
			case "Catppuccin Mocha":
				m.themeName = "Nord"
			case "Nord":
				m.themeName = "Gruvbox"
			case "Gruvbox":
				m.themeName = "Dracula"
			case "Dracula":
				m.themeName = "Tokyo Night"
			default:
				m.themeName = "Catppuccin Mocha"
			}
			m.styles = GetStyles(m.themeName)
			m.textInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(m.styles.Colors.Info))
			m.recalculateLayouts()
			return m, nil
		case "ctrl+u":
			if m.activePanel == PanelInput {
				m.textInput.SetValue("")
				m.textInput.SetCursor(0)
			}
			return m, nil
		case "ctrl+y":
			var textToCopy string
			var desc string
			if m.activeTab == ViewSettings || m.activeChatKey == "" {
				textToCopy = fmt.Sprintf("%x", m.ygg.PublicKey())
				desc = "Your Public Key"
			} else {
				textToCopy = m.activeChatKey
				desc = "Contact's Public Key"
			}
			err := writeToClipboard(textToCopy)
			if err != nil {
				m.appendSystemMessage(fmt.Sprintf("Failed to copy: %v", err))
			} else {
				m.appendSystemMessage(fmt.Sprintf("✓ Copied %s to clipboard!", desc))
			}
			return m, nil
		case "ctrl+r":
			m.ygg.RetryPeersNow()
			return m, nil
		case "ctrl+d":
			m.showTimestamps = !m.showTimestamps
			if m.activeChatKey != "" {
				m.viewport.SetContent(m.getWrappedHistory(m.activeChatKey))
				m.viewport.GotoBottom()
			}
			return m, nil
		case "ctrl+p":
			m.activeTab = ViewPeers
			m.recalculateLayouts()
		case "ctrl+h":
			m.activeTab = ViewChat
			m.recalculateLayouts()
		case "ctrl+s":
			m.activeTab = ViewSettings
			m.recalculateLayouts()
		case "ctrl+n":
			m.modalActive = true
			m.modalType = "add_contact"
			m.contactNameInput.Reset()
			m.contactKeyInput.Reset()
			m.contactNameInput.Focus()
			return m, nil
		case "ctrl+a":
			m.modalActive = true
			m.modalType = "add_peer"
			m.peerURIInput.Reset()
			m.peerURIInput.Focus()
			return m, nil
		case "tab":
			if m.activeTab == ViewChat && m.activePanel == PanelInput && strings.HasPrefix(m.textInput.Value(), "/") {
				commands := []string{"/nick", "/peer", "/add", "/ping", "/send", "/clear", "/whois", "/shake", "/shout", "/help"}
				val := m.textInput.Value()
				var matches []string
				for _, cmd := range commands {
					if strings.HasPrefix(cmd, val) {
						matches = append(matches, cmd)
					}
				}
				if len(matches) > 0 {
					nextMatch := matches[0]
					if len(matches) > 1 && val == matches[0] {
						nextMatch = matches[1]
					}
					m.textInput.SetValue(nextMatch + " ")
					m.textInput.SetCursor(len(m.textInput.Value()))
				}
				return m, nil
			}

			if m.activeTab == ViewChat {
				m.activePanel = (m.activePanel + 1) % 3
				if m.activePanel == PanelInput {
					m.textInput.Focus()
				} else {
					m.textInput.Blur()
				}
				
				// Send read receipt if we tab into the active viewport
				if m.activePanel == PanelViewport && m.activeChatKey != "" {
					_ = m.ygg.SendReadReceipt(m.activeChatKey, m.cfg.Username, time.Now().Unix())
				}
			}
		case "shift+tab":
			if m.activeTab == ViewChat {
				m.activePanel = (m.activePanel - 1 + 3) % 3
				if m.activePanel == PanelInput {
					m.textInput.Focus()
				} else {
					m.textInput.Blur()
				}
			}
		case "up", "down":
			if m.activeTab == ViewChat {
				if m.activePanel == PanelInput {
					if len(m.inputHistory) > 0 {
						if msg.String() == "up" {
							if m.historyIdx > 0 {
								m.historyIdx--
							}
							m.textInput.SetValue(m.inputHistory[m.historyIdx])
						} else {
							if m.historyIdx < len(m.inputHistory)-1 {
								m.historyIdx++
								m.textInput.SetValue(m.inputHistory[m.historyIdx])
							} else {
								m.historyIdx = len(m.inputHistory)
								m.textInput.SetValue("")
							}
						}
						m.textInput.SetCursor(len(m.textInput.Value()))
					}
					return m, nil
				}
				
				contactsList := m.getContactsList()
				if m.activePanel == PanelSidebar && len(contactsList) > 0 {
					if msg.String() == "up" {
						m.selectedContactIdx = (m.selectedContactIdx - 1 + len(contactsList)) % len(contactsList)
					} else {
						m.selectedContactIdx = (m.selectedContactIdx + 1) % len(contactsList)
					}
					m.selectContact(contactsList[m.selectedContactIdx])
				}
			} else if m.activeTab == ViewPeers && len(m.peersInfo) > 0 {
				if msg.String() == "up" {
					m.selectedPeerIdx = (m.selectedPeerIdx - 1 + len(m.peersInfo)) % len(m.peersInfo)
				} else {
					m.selectedPeerIdx = (m.selectedPeerIdx + 1) % len(m.peersInfo)
				}
			}
		case "enter":
			if m.activeTab == ViewChat {
				if m.activePanel == PanelInput && m.textInput.Value() != "" {
					msgText := m.textInput.Value()
					m.textInput.SetValue("")
					m.inputHistory = append(m.inputHistory, msgText)
					m.historyIdx = len(m.inputHistory)
					
					if strings.HasPrefix(msgText, "/") {
						cmd := m.handleSlashCommand(msgText)
						if cmd != nil {
							cmds = append(cmds, cmd)
						}
						return m, tea.Batch(cmds...)
					}
					
					if m.activeChatKey != "" {
						contact, hasContact := m.cfg.Contacts[m.activeChatKey]
						timeStr := time.Now().Format(TimeFormat)
						nameTag := lipgloss.NewStyle().
							Foreground(lipgloss.Color(m.styles.Colors.Primary)).
							Bold(true).
							Render(m.cfg.Username)

						var err error
						if hasContact && contact.SharedSecret != "" {
							err = m.ygg.SendEncryptedChatMessage(m.activeChatKey, m.cfg.Username, msgText, contact.SharedSecret)
						} else {
							err = m.ygg.SendChatMessage(m.activeChatKey, m.cfg.Username, msgText)
						}

						var bubble string
						if err != nil {
							// Unreachable! Queue offline!
							bubble = fmt.Sprintf("[%s] %s %s: %s", timeStr, m.cfg.Username, lipgloss.NewStyle().Foreground(lipgloss.Color(m.styles.Colors.Warning)).Render("(pending)"), msgText)
							m.enqueueOfflineMessage(m.activeChatKey, msgText, timeStr)
						} else {
							singleTick := lipgloss.NewStyle().Foreground(lipgloss.Color(m.styles.Colors.Subtext)).Render("✓")
							bubble = fmt.Sprintf("[%s] %s: %s %s", timeStr, nameTag, msgText, singleTick)
						}
						
						m.chatHistory[m.activeChatKey] = append(m.chatHistory[m.activeChatKey], bubble)
						_ = SaveHistory(m.chatHistory)
						m.viewport.SetContent(m.getWrappedHistory(m.activeChatKey))
						m.viewport.GotoBottom()
					}
				} else if m.activePanel == PanelSidebar {
					contactsList := m.getContactsList()
					if len(contactsList) > 0 {
						m.selectContact(contactsList[m.selectedContactIdx])
						m.activePanel = PanelInput
						m.textInput.Focus()
					}
				}
			}
		case "delete", "backspace":
			if m.activeTab == ViewPeers && len(m.peersInfo) > 0 {
				peer := m.peersInfo[m.selectedPeerIdx]
				_ = m.ygg.RemovePeer(peer.URI)
				var newPeers []string
				for _, p := range m.cfg.Peers {
					if p != peer.URI {
						newPeers = append(newPeers, p)
					}
				}
				m.cfg.Peers = newPeers
				_ = m.cfg.Save()
				m.selectedPeerIdx = 0
			}
		}

		var cmd tea.Cmd
		if m.activeTab == ViewChat {
			if m.activePanel == PanelInput {
				m.textInput, cmd = m.textInput.Update(msg)
				cmds = append(cmds, cmd)
			} else if m.activePanel == PanelViewport {
				m.viewport, cmd = m.viewport.Update(msg)
				cmds = append(cmds, cmd)
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) recalculateLayouts() {
	headerHeight := 1
	footerHeight := 1
	mainHeight := m.height - headerHeight - footerHeight
	if mainHeight < 6 {
		mainHeight = 6
	}

	sidebarWidth := 26
	mainWidth := m.width - sidebarWidth
	if mainWidth < 20 {
		mainWidth = 20
	}

	// Update Sidebar frame size
	sbFrameHeight := m.styles.SidebarStyle.GetVerticalFrameSize()
	sbFrameWidth := m.styles.SidebarStyle.GetHorizontalFrameSize()
	m.styles.SidebarStyle = m.styles.SidebarStyle.
		Height(mainHeight - sbFrameHeight).
		Width(sidebarWidth - sbFrameWidth)

	// Update Input box size
	inputBoxOuterHeight := 3
	inputFrameHeight := m.styles.InputStyle.GetVerticalFrameSize()
	inputFrameWidth := m.styles.InputStyle.GetHorizontalFrameSize()
	m.styles.InputStyle = m.styles.InputStyle.
		Height(inputBoxOuterHeight - inputFrameHeight).
		Width(mainWidth - inputFrameWidth)
	m.styles.InputActiveStyle = m.styles.InputActiveStyle.
		Height(inputBoxOuterHeight - inputFrameHeight).
		Width(mainWidth - inputFrameWidth)

	// Update Viewport box size
	vpBoxOuterHeight := mainHeight - inputBoxOuterHeight - 1
	vpFrameHeight := m.styles.ChatViewportStyle.GetVerticalFrameSize()
	vpFrameWidth := m.styles.ChatViewportStyle.GetHorizontalFrameSize()
	m.styles.ChatViewportStyle = m.styles.ChatViewportStyle.
		Height(vpBoxOuterHeight - vpFrameHeight).
		Width(mainWidth - vpFrameWidth)

	// Inner widgets
	m.viewport.Width = mainWidth - vpFrameWidth - 2
	m.viewport.Height = vpBoxOuterHeight - vpFrameHeight - 3 // space for border, title line, and typing status
	m.viewport.Style = lipgloss.NewStyle().Background(lipgloss.Color(m.styles.Colors.Base))

	m.textInput.Width = mainWidth - inputFrameWidth - 6
}

func (m *Model) selectContact(contact Contact) {
	m.activeChatKey = contact.PublicKey
	m.unreadMessages[m.activeChatKey] = false
	if _, ok := m.chatHistory[m.activeChatKey]; !ok {
		m.chatHistory[m.activeChatKey] = []string{}
	}
	m.viewport.SetContent(m.getWrappedHistory(m.activeChatKey))
	m.viewport.GotoBottom()

	// Send read receipt when we click on the contact
	_ = m.ygg.SendReadReceipt(m.activeChatKey, m.cfg.Username, time.Now().Unix())
}

func (m *Model) submitModal() {
	m.modalActive = false
	defer m.textInput.Focus()

	switch m.modalType {
	case "add_contact":
			name := strings.TrimSpace(m.contactNameInput.Value())
			key := strings.TrimSpace(m.contactKeyInput.Value())
			if name == "" || key == "" {
				m.modalActive = true
				m.modalType = "error"
				m.errorMessage = "Both fields are required."
				return
			}

			// Validate key is 64-char hex (32 bytes)
			keyBytes, err := hex.DecodeString(key)
			if err != nil || len(keyBytes) != 32 {
				m.modalActive = true
				m.modalType = "error"
				m.errorMessage = "Public key must be a valid 64-character hexadecimal string."
				return
			}

			// ECDH Key derivation for E2EE contact request
			pubHex, err := GetMyECDHPublicKeyHex(m.cfg.ECDHPrivateKey)
			if err != nil {
				m.modalActive = true
				m.modalType = "error"
				m.errorMessage = "Failed to generate ECDH key."
				return
			}
			
			m.cfg.Contacts[key] = Contact{
				PublicKey: key,
				Nickname:  name,
			}
			_ = m.cfg.Save()

			_ = m.ygg.SendContactRequest(key, m.cfg.Username, false, pubHex)

		m.selectContact(m.cfg.Contacts[key])
		m.selectedContactIdx = len(m.getContactsList()) - 1
		m.appendSystemMessage("Sent secure contact request... [Key Exchange Pending]")

	case "add_peer":
		uri := strings.TrimSpace(m.peerURIInput.Value())
		if uri == "" {
			return
		}
		
		err := m.ygg.AddPeer(uri)
		if err != nil {
			m.modalActive = true
			m.modalType = "error"
			m.errorMessage = fmt.Sprintf("Failed to add peer: %v", err)
			return
		}

		found := false
		for _, p := range m.cfg.Peers {
			if p == uri {
				found = true
				break
			}
		}
		if !found {
			m.cfg.Peers = append(m.cfg.Peers, uri)
			_ = m.cfg.Save()
		}
	}
}

func (m Model) getContactsList() []Contact {
	var list []Contact
	for _, c := range m.cfg.Contacts {
		list = append(list, c)
	}
	return list
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing TUI..."
	}

	headerStr := m.renderHeader()
	footerStr := m.renderFooter()

	var mainStr string
	if m.modalActive {
		mainHeight := m.height - 2
		modalBox := m.renderModalBox()
		mainStr = lipgloss.NewStyle().
			Background(lipgloss.Color(m.styles.Colors.Base)).
			Width(m.width).
			Height(mainHeight).
			AlignHorizontal(lipgloss.Center).
			AlignVertical(lipgloss.Center).
			Render(modalBox)
	} else {
		mainStr = m.renderMainContent()
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		headerStr,
		mainStr,
		footerStr,
	)
}

func (m Model) renderHeader() string {
	ipStr := "Not Started"
	pkStr := "None"
	if m.ygg.PublicKey() != nil {
		pkStr = fmt.Sprintf("%x", m.ygg.PublicKey())
		ipStr = m.ygg.Address().String()
	}
	
	var meshOnline bool
	for _, p := range m.peersInfo {
		if p.Up {
			meshOnline = true
			break
		}
	}
	
	statusText := "🔴 OFFLINE"
	if meshOnline {
		statusText = "🌐 ONLINE"
	}
	
	// Left title
	left := lipgloss.NewStyle().
		Background(lipgloss.Color(m.styles.Colors.Accent)).
		Foreground(lipgloss.Color(m.styles.Colors.Crust)).
		Bold(true).
		Padding(0, 1).
		Render(fmt.Sprintf("⚡ YGGDRASIL (%s)", statusText))

	// Tabs
	tabs := []string{"CHAT [Ctrl+H]", "PEERS [Ctrl+P]", "SETTINGS [Ctrl+S]"}
	var renderedTabs []string
	for i, tab := range tabs {
		if i == m.activeTab {
			renderedTabs = append(renderedTabs, m.styles.ActiveTabStyle.Render(tab))
		} else {
			renderedTabs = append(renderedTabs, m.styles.InactiveTabStyle.Render(tab))
		}
	}
	tabsStr := lipgloss.JoinHorizontal(lipgloss.Left, renderedTabs...)
	
	var activePeers int
	for _, p := range m.peersInfo {
		if p.Up {
			activePeers++
		}
	}

	leftAndTabsWidth := lipgloss.Width(left) + lipgloss.Width(tabsStr) + 2
	maxRightWidth := m.width - leftAndTabsWidth
	if maxRightWidth < 10 {
		maxRightWidth = 10
	}

	var rightStr string
	
	// Format 1: Full info
	rightStr = fmt.Sprintf("Theme: %s | IP: %s | Peers: %d | User: %s | Key: %s...%s ", m.themeName, ipStr, activePeers, m.cfg.Username, pkStr[:4], pkStr[len(pkStr)-4:])
	
	// Format 2: Shortened IP
	if lipgloss.Width(rightStr) > maxRightWidth && len(ipStr) > 15 {
		shortIP := fmt.Sprintf("%s...%s", ipStr[:8], ipStr[len(ipStr)-4:])
		rightStr = fmt.Sprintf("Theme: %s | IP: %s | Peers: %d | User: %s | Key: %s...%s ", m.themeName, shortIP, activePeers, m.cfg.Username, pkStr[:4], pkStr[len(pkStr)-4:])
	}
	
	// Format 3: No Key
	if lipgloss.Width(rightStr) > maxRightWidth && len(ipStr) > 15 {
		shortIP := fmt.Sprintf("%s...%s", ipStr[:8], ipStr[len(ipStr)-4:])
		rightStr = fmt.Sprintf("Theme: %s | IP: %s | Peers: %d | User: %s ", m.themeName, shortIP, activePeers, m.cfg.Username)
	}

	// Format 4: No Key, No IP
	if lipgloss.Width(rightStr) > maxRightWidth {
		rightStr = fmt.Sprintf("Peers: %d | User: %s ", activePeers, m.cfg.Username)
	}

	// Format 5: Minimal
	if lipgloss.Width(rightStr) > maxRightWidth {
		rightStr = fmt.Sprintf("User: %s ", m.cfg.Username)
	}

	// Clip if still too long
	if lipgloss.Width(rightStr) > maxRightWidth {
		rightStr = rightStr[:maxRightWidth]
	}

	right := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.styles.Colors.Subtext)).
		Render(rightStr)

	spacesCount := m.width - leftAndTabsWidth - lipgloss.Width(right)
	if spacesCount < 0 {
		spacesCount = 0
	}
	spaces := strings.Repeat(" ", spacesCount)

	headerLine := left + " " + tabsStr + spaces + right
	
	return m.styles.HeaderStyle.
		Width(m.width).
		Render(headerLine)
}

func (m Model) renderMainContent() string {
	switch m.activeTab {
	case ViewChat:
		return m.renderChatView()
	case ViewPeers:
		return m.renderPeersView()
	case ViewSettings:
		return m.renderSettingsView()
	default:
		return ""
	}
}

func (m Model) renderChatView() string {
	contactsList := m.getContactsList()

	// Focus-driven sidebar borders
	var sbBorderColor string
	if m.activePanel == PanelSidebar {
		sbBorderColor = m.styles.Colors.Primary
	} else {
		sbBorderColor = m.styles.Colors.Muted
	}
	sidebarStyle := m.styles.SidebarStyle.Copy().
		BorderForeground(lipgloss.Color(sbBorderColor))

	// Sidebar contents
	var sidebarItems []string
	sidebarItems = append(sidebarItems, lipgloss.NewStyle().Foreground(lipgloss.Color(m.styles.Colors.Accent)).Bold(true).Render("👤 CONTACTS"))
	sidebarItems = append(sidebarItems, "")

	if len(contactsList) == 0 {
		sidebarItems = append(sidebarItems, lipgloss.NewStyle().Foreground(lipgloss.Color(m.styles.Colors.Subtext)).Render("No contacts."))
		sidebarItems = append(sidebarItems, lipgloss.NewStyle().Foreground(lipgloss.Color(m.styles.Colors.Muted)).Render("Type /add to request."))
	} else {
		for i, c := range contactsList {
			prefix := " "
			if m.selectedContactIdx == i {
				prefix = "❱"
			}

			displayName := fmt.Sprintf(" %s %s", prefix, c.Nickname)
			if m.unreadMessages[c.PublicKey] {
				displayName += " ●"
			}

			// Add visual indicator if contact is securely E2EE enabled!
			if c.SharedSecret != "" {
				displayName += " 🔒"
			}

			var contactStr string
			if m.selectedContactIdx == i {
				if m.activePanel == PanelSidebar {
					contactStr = m.styles.ContactActiveStyle.Render(displayName)
				} else {
					contactStr = m.styles.ContactSelectedInactiveStyle.Render(displayName)
				}
			} else {
				contactStr = m.styles.ContactInactiveStyle.Render(displayName)
			}
			sidebarItems = append(sidebarItems, contactStr)
		}
	}

	sidebarContent := strings.Join(sidebarItems, "\n")
	sidebar := sidebarStyle.Render(sidebarContent)

	// Focus-driven viewport borders
	var vpBorderColor string
	if m.activePanel == PanelViewport {
		vpBorderColor = m.styles.Colors.Primary
	} else {
		vpBorderColor = m.styles.Colors.Muted
	}
	viewportStyle := m.styles.ChatViewportStyle.Copy().
		BorderForeground(lipgloss.Color(vpBorderColor))

	var rightContent string
	if m.activeChatKey != "" {
		var activeContactName string
		var padlockBadge string
		if contact, ok := m.cfg.Contacts[m.activeChatKey]; ok {
			activeContactName = contact.Nickname
			if contact.SharedSecret != "" {
				padlockBadge = lipgloss.NewStyle().
					Foreground(lipgloss.Color(m.styles.Colors.Success)).
					Bold(true).
					Render(" [🛡️ E2EE SECURED]")
			} else {
				padlockBadge = lipgloss.NewStyle().
					Foreground(lipgloss.Color(m.styles.Colors.Warning)).
					Bold(true).
					Render(" [🔓 UNENCRYPTED]")
			}
		} else {
			activeContactName = m.activeChatKey[:12] + "..."
			padlockBadge = lipgloss.NewStyle().
				Foreground(lipgloss.Color(m.styles.Colors.Warning)).
				Bold(true).
				Render(" [🔓 UNENCRYPTED]")
		}

		chatHeader := lipgloss.NewStyle().
			Foreground(lipgloss.Color(m.styles.Colors.Accent)).
			Bold(true).
			Render("💬 Chatting with: " + activeContactName + padlockBadge)
			
		rightContent = chatHeader + "\n\n" + m.viewport.View()
	} else {
		rightContent = m.renderDashboard()
	}

	viewportRender := viewportStyle.Render(rightContent)

	// Focus-driven input borders
	var inputBorderColor string
	if m.activePanel == PanelInput {
		inputBorderColor = m.styles.Colors.Primary
	} else {
		inputBorderColor = m.styles.Colors.Muted
	}
	inputStyle := m.styles.InputStyle.Copy().
		BorderForeground(lipgloss.Color(inputBorderColor))

	inputRender := inputStyle.Render(m.textInput.View())

	// Typing Indicator Status Line
	var typingIndicator string
	if m.activeChatKey != "" {
		if t, ok := m.typingStatus[m.activeChatKey]; ok && time.Since(t) < TypingDisplaySec*time.Second {
			nickname := m.activeChatKey[:8]
			if contact, ok := m.cfg.Contacts[m.activeChatKey]; ok && contact.Nickname != "" {
				nickname = contact.Nickname
			}
			typingIndicator = lipgloss.NewStyle().
				Foreground(lipgloss.Color(m.styles.Colors.Subtext)).
				Italic(true).
				Render(fmt.Sprintf(" 💬 %s is typing...", nickname))
		}
	}

	statusLine := " "
	if typingIndicator != "" {
		statusLine = typingIndicator
	}

	rightPane := lipgloss.JoinVertical(
		lipgloss.Left,
		viewportRender,
		statusLine,
		inputRender,
	)

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		sidebar,
		rightPane,
	)
}

func (m Model) renderDashboard() string {
	var lines []string
	lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color(m.styles.Colors.Accent)).Bold(true).Render("⚡ YGGDRASIL NODE DASHBOARD"))
	lines = append(lines, "")
	
	ipStr := "Unknown"
	pkStr := "Unknown"
	if m.ygg.PublicKey() != nil {
		ipStr = m.ygg.Address().String()
		pkStr = fmt.Sprintf("%x", m.ygg.PublicKey())
	}
	
	lines = append(lines, fmt.Sprintf("  Username:     %s", lipgloss.NewStyle().Foreground(lipgloss.Color(m.styles.Colors.Text)).Bold(true).Render(m.cfg.Username)))
	lines = append(lines, fmt.Sprintf("  Node Status:  %s", m.styles.StatusOnlineStyle.Render("🟢 ACTIVE (Userspace overlay)")))
	lines = append(lines, fmt.Sprintf("  IPv6 Address: %s", lipgloss.NewStyle().Foreground(lipgloss.Color(m.styles.Colors.Info)).Render(ipStr)))
	lines = append(lines, fmt.Sprintf("  Public Key:   %s", lipgloss.NewStyle().Foreground(lipgloss.Color(m.styles.Colors.Subtext)).Render(pkStr)))
	lines = append(lines, "")
	lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color(m.styles.Colors.Accent)).Bold(true).Render("  QUICK SHORTCUTS:"))
	lines = append(lines, "  • [Ctrl+T]  Cycle TUI Themes dynamically (Mocha ➔ Nord ➔ Gruvbox).")
	lines = append(lines, "  • [Ctrl+Y]  Copy your Public Key (or contact's key) to Clipboard.")
	lines = append(lines, "  • [Ctrl+R]  Force-retry all Peering connections immediately.")
	lines = append(lines, "  • [Up/Down] Cycle Chat Input history (when focused in input box).")
	lines = append(lines, "")
	lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color(m.styles.Colors.Accent)).Bold(true).Render("  PROTOCOLS & SLASH COMMANDS:"))
	lines = append(lines, "  • /nick <username>        Change display name dynamically.")
	lines = append(lines, "  • /peer <tcp_uri>         Peers with remote node.")
	lines = append(lines, "  • /add <key> <nickname>   Initiate secure Contact request.")
	lines = append(lines, "  • /ping                   Measure route latency to active contact.")
	lines = append(lines, "  • /send <filepath>        Transfer file P2P over mesh overlay.")
	
	return strings.Join(lines, "\n")
}

func (m Model) renderPeersView() string {
	var peersItems []string
	peersItems = append(peersItems, lipgloss.NewStyle().Foreground(lipgloss.Color(m.styles.Colors.Accent)).Bold(true).Render("🔗 CONFIGURED PEERS"))
	peersItems = append(peersItems, "")

	if len(m.peersInfo) == 0 {
		peersItems = append(peersItems, lipgloss.NewStyle().Foreground(lipgloss.Color(m.styles.Colors.Subtext)).Render("No connected peers."))
		peersItems = append(peersItems, lipgloss.NewStyle().Foreground(lipgloss.Color(m.styles.Colors.Muted)).Render("Press Ctrl+A or use /peer to connect."))
	} else {
		for i, p := range m.peersInfo {
			prefix := " "
			if m.selectedPeerIdx == i {
				prefix = "❱"
			}

			status := m.styles.StatusOfflineStyle.Render("🔴 OFFLINE")
			if p.Up {
				status = m.styles.StatusOnlineStyle.Render("🟢 ONLINE")
			}

			peerStr := fmt.Sprintf(" %s %s  (%s)", prefix, p.URI, status)
			if m.selectedPeerIdx == i {
				peerStr = lipgloss.NewStyle().Background(lipgloss.Color(m.styles.Colors.Overlay)).Bold(true).Render(peerStr)
			}
			peersItems = append(peersItems, peerStr)
		}
		peersItems = append(peersItems, "")
		peersItems = append(peersItems, lipgloss.NewStyle().Foreground(lipgloss.Color(m.styles.Colors.Subtext)).Render("Press [Backspace] or [Delete] to remove selected peer."))
	}

	content := strings.Join(peersItems, "\n")
	
	mainHeight := m.height - 2
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(m.styles.Colors.Muted)).
		Background(lipgloss.Color(m.styles.Colors.Base)).
		Padding(1, 2)

	boxWidth := m.width - boxStyle.GetHorizontalFrameSize()
	boxHeight := mainHeight - boxStyle.GetVerticalFrameSize()

	return boxStyle.Width(boxWidth).Height(boxHeight).Render(content)
}

func (m Model) renderSettingsView() string {
	var settingsItems []string
	settingsItems = append(settingsItems, lipgloss.NewStyle().Foreground(lipgloss.Color(m.styles.Colors.Accent)).Bold(true).Render("⚙️ SYSTEM SETTINGS"))
	settingsItems = append(settingsItems, "")

	ipStr := "Not Started"
	pkStr := "None"
	if m.ygg.PublicKey() != nil {
		pkStr = fmt.Sprintf("%x", m.ygg.PublicKey())
		ipStr = m.ygg.Address().String()
	}

	settingsItems = append(settingsItems, fmt.Sprintf("  Username:        %s", m.cfg.Username))
	settingsItems = append(settingsItems, fmt.Sprintf("  IPv6 Address:    %s", lipgloss.NewStyle().Foreground(lipgloss.Color(m.styles.Colors.Success)).Bold(true).Render(ipStr)))
	settingsItems = append(settingsItems, fmt.Sprintf("  Public Key:      %s", lipgloss.NewStyle().Foreground(lipgloss.Color(m.styles.Colors.Info)).Render(pkStr)))
	settingsItems = append(settingsItems, "")
	settingsItems = append(settingsItems, "  Yggdrasil Listeners:")
	if len(m.cfg.Listeners) == 0 {
		settingsItems = append(settingsItems, "    - None")
	} else {
		for _, l := range m.cfg.Listeners {
			settingsItems = append(settingsItems, fmt.Sprintf("    - %s", l))
		}
	}

	content := strings.Join(settingsItems, "\n")

	mainHeight := m.height - 2
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(m.styles.Colors.Muted)).
		Background(lipgloss.Color(m.styles.Colors.Base)).
		Padding(1, 2)

	boxWidth := m.width - boxStyle.GetHorizontalFrameSize()
	boxHeight := mainHeight - boxStyle.GetVerticalFrameSize()

	return boxStyle.Width(boxWidth).Height(boxHeight).Render(content)
}

func (m Model) renderFooter() string {
	var activeHelp string
	if m.activeTab == ViewChat {
		switch m.activePanel {
		case PanelSidebar:
			activeHelp = " ↑/↓: Navigate • Enter: Select Chat • Ctrl+N: Add Contact • Tab: Focus Input"
		case PanelViewport:
			activeHelp = " ↑/↓: Scroll • Tab: Focus Input"
		case PanelInput:
			activeHelp = " Enter: Send Message • Tab: Focus Sidebar"
		}
	} else if m.activeTab == ViewPeers {
		activeHelp = " Ctrl+A: Add Peer • Del: Remove Peer • Up/Down: Navigate"
	} else {
		activeHelp = " Settings Tab"
	}

	rightStr := "Ctrl+C: Quit "
	right := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.styles.Colors.Accent)).
		Bold(true).
		Render(rightStr)

	maxLeftWidth := m.width - lipgloss.Width(right) - 2
	if maxLeftWidth < 5 {
		maxLeftWidth = 5
	}

	if len(activeHelp) > maxLeftWidth {
		activeHelp = activeHelp[:maxLeftWidth-3] + "..."
	}

	left := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.styles.Colors.Subtext)).
		Render(" " + activeHelp)

	spacesCount := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if spacesCount < 0 {
		spacesCount = 0
	}
	spaces := strings.Repeat(" ", spacesCount)

	footerLine := left + spaces + right

	return m.styles.FooterStyle.
		Width(m.width).
		Render(footerLine)
}

func (m Model) renderModalBox() string {
	var modalContent string

	switch m.modalType {
	case "add_contact":
		modalContent = lipgloss.JoinVertical(
			lipgloss.Center,
			m.styles.ModalHeaderStyle.Render("👤 ADD NEW CONTACT"),
			"Enter contact nickname:",
			m.contactNameInput.View(),
			"",
			"Enter contact public key (Hex):",
			m.contactKeyInput.View(),
			"",
			"[Enter] Submit  •  [Esc] Cancel",
		)
	case "add_peer":
		modalContent = lipgloss.JoinVertical(
			lipgloss.Center,
			m.styles.ModalHeaderStyle.Render("🔗 ADD PEER CONNECTION"),
			"Enter Peer URI (e.g. tcp://host:port):",
			m.peerURIInput.View(),
			"",
			"[Enter] Submit  •  [Esc] Cancel",
		)
	case "contact_request":
		modalContent = lipgloss.JoinVertical(
			lipgloss.Center,
			m.styles.ModalHeaderStyle.Render("👤 CONTACT REQUEST"),
			fmt.Sprintf("Accept contact request from %s?", m.incomingReqName),
			fmt.Sprintf("Key: %s...%s", m.incomingReqKey[:8], m.incomingReqKey[len(m.incomingReqKey)-8:]),
			"",
			"[Y] Accept  •  [N] Decline",
		)
	case "error":
		modalContent = lipgloss.JoinVertical(
			lipgloss.Center,
			m.styles.ModalHeaderStyle.Copy().Foreground(lipgloss.Color(m.styles.Colors.Error)).Render("⚠️ ERROR"),
			m.errorMessage,
			"",
			"[Esc] Close",
		)
	}

	return m.styles.ModalStyle.Render(modalContent)
}

func (m *Model) handleSlashCommand(cmdStr string) tea.Cmd {
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return nil
	}

	cmd := parts[0]

	switch cmd {
	case "/nick":
		if len(parts) < 2 {
			m.appendSystemMessage("Usage: /nick <new_username>")
			return nil
		}
		newName := strings.Join(parts[1:], " ")
		m.cfg.Username = newName
		_ = m.cfg.Save()
		m.appendSystemMessage(fmt.Sprintf("Username changed to: %s", newName))
		return nil

	case "/peer":
		if len(parts) < 2 {
			m.appendSystemMessage("Usage: /peer <tcp_uri>")
			return nil
		}
		uri := parts[1]
		err := m.ygg.AddPeer(uri)
		if err != nil {
			m.appendSystemMessage(fmt.Sprintf("Failed to peer: %v", err))
		} else {
			found := false
			for _, p := range m.cfg.Peers {
				if p == uri {
					found = true
					break
				}
			}
			if !found {
				m.cfg.Peers = append(m.cfg.Peers, uri)
				_ = m.cfg.Save()
			}
			m.appendSystemMessage(fmt.Sprintf("Peering connection initiated: %s", uri))
		}
		return nil

	case "/add":
			if len(parts) < 3 {
				m.appendSystemMessage("Usage: /add <key> <nickname>")
				return nil
			}
			key := parts[1]
			nickname := parts[2]
			
			keyBytes, err := hex.DecodeString(key)
			if err != nil || len(keyBytes) != 32 {
				m.appendSystemMessage("Error: Key must be a 64-character hexadecimal string.")
				return nil
			}

			pubHex, err := GetMyECDHPublicKeyHex(m.cfg.ECDHPrivateKey)
			if err != nil {
				m.appendSystemMessage(fmt.Sprintf("Failed to generate ECDH key: %v", err))
				return nil
			}

			err = m.ygg.SendContactRequest(key, m.cfg.Username, false, pubHex)
			if err != nil {
				m.appendSystemMessage(fmt.Sprintf("Contact request failed: %v", err))
			} else {
				m.appendSystemMessage(fmt.Sprintf("Sent contact request to %s (Pending response...)", nickname))
				m.cfg.Contacts[key] = Contact{PublicKey: key, Nickname: nickname}
				_ = m.cfg.Save()
			}
			return nil

	case "/ping":
		if m.activeChatKey == "" {
			m.appendSystemMessage("Error: Select a contact in the sidebar first.")
			return nil
		}
		err := m.ygg.SendPingMessage(m.activeChatKey, m.cfg.Username, false)
		if err != nil {
			m.appendSystemMessage(fmt.Sprintf("Ping failed: %v", err))
		} else {
			m.appendSystemMessage("Sent ping...")
		}
		return nil

	case "/send":
		if len(parts) < 2 {
			m.appendSystemMessage("Usage: /send <filepath>")
			return nil
		}
		if m.activeChatKey == "" {
			m.appendSystemMessage("Error: Select a contact in the sidebar first.")
			return nil
		}
		
		path := strings.Join(parts[1:], " ")
		path = strings.Trim(path, "\"'")
		data, err := os.ReadFile(path)
		if err != nil {
			m.appendSystemMessage(fmt.Sprintf("Failed to read file: %v", err))
			return nil
		}

		filename := filepath.Base(path)
		chunkSize := ChunkSize
		totalChunks := (len(data) + chunkSize - 1) / chunkSize
		
		m.appendSystemMessage(fmt.Sprintf("[📥 Sending %s: 0/%d (0%%)]", filename, totalChunks))
		return sendNextChunkCmd(m.activeChatKey, filename, 0, totalChunks, data, m.cfg.Username, m.ygg)

	case "/clear":
		if m.activeChatKey != "" {
			m.chatHistory[m.activeChatKey] = []string{}
			_ = SaveHistory(m.chatHistory)
			m.viewport.SetContent("")
		}
		return nil

	case "/search":
		if len(parts) < 2 {
			m.appendSystemMessage("Usage: /search <query>")
			return nil
		}
		if m.activeChatKey == "" {
			m.appendSystemMessage("Error: Select a contact in the sidebar first.")
			return nil
		}
		query := strings.ToLower(strings.Join(parts[1:], " "))
		m.appendSystemMessage(fmt.Sprintf("🔍 Searching chat history for: '%s'", query))
		
		found := false
		for _, line := range m.chatHistory[m.activeChatKey] {
			plainLine := stripANSI(line)
			if strings.Contains(strings.ToLower(plainLine), query) {
				m.appendSystemMessage(" ❱ " + line)
				found = true
			}
		}
		if !found {
			m.appendSystemMessage("  No matching messages found in history.")
		}
		return nil

	case "/help":
		m.appendSystemMessage("Available Commands:")
		m.appendSystemMessage("  /nick <username>      - Change display name dynamically")
		m.appendSystemMessage("  /peer <tcp_uri>       - Peer with a remote node")
		m.appendSystemMessage("  /add <key> <nickname> - Send E2EE contact request")
		m.appendSystemMessage("  /ping                 - Measure connection RTT latency")
		m.appendSystemMessage("  /send <filepath>      - Send file P2P with inline preview")
		m.appendSystemMessage("  /clear                - Clear chat window history locally")
		m.appendSystemMessage("  /whois                - Show contact details (Key, E2EE)")
		m.appendSystemMessage("  /shake                - Send a screen nudge/beep to contact")
		m.appendSystemMessage("  /shout <message>      - Broadcast message to all contacts")
		return nil

	case "/whois":
		if m.activeChatKey == "" {
			m.appendSystemMessage("Error: Select a contact in the sidebar first.")
			return nil
		}
		contact, ok := m.cfg.Contacts[m.activeChatKey]
		if ok {
			m.appendSystemMessage(fmt.Sprintf("Contact Nickname: %s", contact.Nickname))
			m.appendSystemMessage(fmt.Sprintf("Yggdrasil Key:    %s", contact.PublicKey))
			isE2EE := "No (Key exchange pending)"
			if contact.SharedSecret != "" {
				isE2EE = "Yes (AES-GCM Tunnel)"
			}
			m.appendSystemMessage(fmt.Sprintf("E2EE Encryption:  %s", isE2EE))
		}
		return nil

	case "/shake":
		if m.activeChatKey == "" {
			m.appendSystemMessage("Error: Select a contact in the sidebar first.")
			return nil
		}
		err := m.ygg.SendShakeMessage(m.activeChatKey, m.cfg.Username)
		if err != nil {
			m.appendSystemMessage(fmt.Sprintf("Nudge/Shake failed: %v", err))
		} else {
			m.appendSystemMessage("Sent a Nudge/Shake nudge!")
		}
		return nil

	case "/shout":
		if len(parts) < 2 {
			m.appendSystemMessage("Usage: /shout <message>")
			return nil
		}
		shoutMsg := strings.Join(parts[1:], " ")
		m.appendSystemMessage(fmt.Sprintf("📢 Shouting to all contacts: %s", shoutMsg))
		
		for key, contact := range m.cfg.Contacts {
			var err error
			if contact.SharedSecret != "" {
				err = m.ygg.SendEncryptedChatMessage(key, m.cfg.Username, shoutMsg, contact.SharedSecret)
			} else {
				err = m.ygg.SendChatMessage(key, m.cfg.Username, shoutMsg)
			}
			
			timeStr := time.Now().Format(TimeFormat)
			nameTag := lipgloss.NewStyle().Foreground(lipgloss.Color(m.styles.Colors.Primary)).Bold(true).Render(m.cfg.Username)
			var bubble string
			if err != nil {
				bubble = fmt.Sprintf("[%s] %s %s: %s", timeStr, m.cfg.Username, lipgloss.NewStyle().Foreground(lipgloss.Color(m.styles.Colors.Warning)).Render("(pending)"), shoutMsg)
				m.enqueueOfflineMessage(key, shoutMsg, timeStr)
			} else {
				singleTick := lipgloss.NewStyle().Foreground(lipgloss.Color(m.styles.Colors.Subtext)).Render("✓")
				bubble = fmt.Sprintf("[%s] %s: %s %s", timeStr, nameTag, shoutMsg, singleTick)
			}
			m.chatHistory[key] = append(m.chatHistory[key], bubble)
		}
		_ = SaveHistory(m.chatHistory)
		if m.activeChatKey != "" {
			m.viewport.SetContent(m.getWrappedHistory(m.activeChatKey))
			m.viewport.GotoBottom()
		}
		return nil
	}

	m.appendSystemMessage(fmt.Sprintf("Unknown command: %s", cmd))
	return nil
}

func (m *Model) appendSystemMessage(text string) {
	if m.activeChatKey == "" {
		return
	}
	timeStr := time.Now().Format(TimeFormat)
	sysTag := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.styles.Colors.Accent)).
		Bold(true).
		Render("SYSTEM")
		
	bubble := fmt.Sprintf("[%s] %s: %s", timeStr, sysTag, text)
	m.chatHistory[m.activeChatKey] = append(m.chatHistory[m.activeChatKey], bubble)
	_ = SaveHistory(m.chatHistory)
	m.viewport.SetContent(m.getWrappedHistory(m.activeChatKey))
	m.viewport.GotoBottom()
}

func (m Model) getWrappedHistory(key string) string {
	history, ok := m.chatHistory[key]
	if !ok || len(history) == 0 {
		return ""
	}

	width := m.viewport.Width
	if width <= 0 {
		width = 80
	}

	style := lipgloss.NewStyle().Width(width)
	
	var wrapped []string
	for _, line := range history {
		displayLine := line
		if !m.showTimestamps {
			displayLine = stripTimestamp(line)
		}
		wrapped = append(wrapped, style.Render(displayLine))
	}

	var lines []string
	for _, w := range wrapped {
		lines = append(lines, strings.Split(w, "\n")...)
	}

	// Prepend newlines to push chat to bottom
	if len(lines) < m.viewport.Height {
		paddingCount := m.viewport.Height - len(lines)
		padding := make([]string, paddingCount)
		for i := range padding {
			padding[i] = ""
		}
		lines = append(padding, lines...)
	}

	return strings.Join(lines, "\n")
}

func stripTimestamp(line string) string {
	if len(line) >= 11 && line[0] == '[' && line[9] == ']' {
		return line[11:]
	}
	return line
}

func stripANSI(str string) string {
	var sb strings.Builder
	inEscape := false
	for _, r := range str {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		sb.WriteRune(r)
	}
	return sb.String()
}

func (m *Model) enqueueOfflineMessage(destKey string, text string, timeStr string) {
	queue := m.loadOfflineQueue()
	queue = append(queue, OfflineMessage{
		DestKey: destKey,
		Text:    text,
		TimeStr: timeStr,
	})
	m.saveOfflineQueue(queue)
}

func (m *Model) saveOfflineQueue(queue []OfflineMessage) {
	cfgName := getConfigFilename()
	ext := filepath.Ext(cfgName)
	base := cfgName[:len(cfgName)-len(ext)]
	path := filepath.Join(".", base+"_pending.json")
	data, _ := json.MarshalIndent(queue, "", "  ")
	_ = os.WriteFile(path, data, 0600)
}

func (m *Model) loadOfflineQueue() []OfflineMessage {
	cfgName := getConfigFilename()
	ext := filepath.Ext(cfgName)
	base := cfgName[:len(cfgName)-len(ext)]
	path := filepath.Join(".", base+"_pending.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return []OfflineMessage{}
	}
	var queue []OfflineMessage
	_ = json.Unmarshal(data, &queue)
	return queue
}

func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return ext == ".png" || ext == ".jpg" || ext == ".jpeg"
}

func sendNextChunkCmd(destKey string, filename string, chunkIdx int, totalChunks int, fullData []byte, username string, ygg *YggManager) tea.Cmd {
	return func() tea.Msg {
		chunkSize := ChunkSize
		start := chunkIdx * chunkSize
		end := start + chunkSize
		if end > len(fullData) {
			end = len(fullData)
		}

		chunkData := fullData[start:end]

		err := ygg.SendFileChunk(destKey, username, filename, chunkIdx, totalChunks, chunkData)
		if err != nil {
			return FileProgressMsg{
				ContactKey: destKey,
				Filename:   filename,
				Err:        err,
			}
		}

		done := (chunkIdx + 1) >= totalChunks
		return FileProgressMsg{
			ContactKey: destKey,
			Filename:   filename,
			ChunkIdx:   chunkIdx,
			Total:      totalChunks,
			Done:       done,
			Data:       fullData,
		}
	}
}

func writeToClipboard(text string) error {
	cmd := exec.Command("clip")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	go func() {
		defer stdin.Close()
		_, _ = io.WriteString(stdin, text)
	}()
	return cmd.Run()
}
