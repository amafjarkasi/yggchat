<p align="center">
  <img src="logo-v2.png" width="180" height="180" alt="Yggdrasil Mesh Chat Logo" />
</p>

<h1 align="center">⚡ YGGDRASIL MESH CHAT ⚡</h1>

<p align="center">
  <strong>A Zero-Dependency, Serverless, Decentralized P2P Encrypted Messaging & File Exchange Client</strong>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Language-Go-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/Encryption-E2EE-4CAF50?style=for-the-badge" alt="E2EE" />
  <img src="https://img.shields.io/badge/Network-Yggdrasil-7B1FA2?style=for-the-badge" alt="Yggdrasil" />
  <img src="https://img.shields.io/badge/Interface-Web%20+%20TUI-FF69B4?style=for-the-badge" alt="Dual UI" />
  <img src="https://img.shields.io/badge/Tests-Passing-4CAF50?style=for-the-badge" alt="Tests" />
  <img src="https://img.shields.io/badge/License-MIT-blue?style=for-the-badge" alt="License" />
</p>

---

## 📖 Table of Contents

- [Overview](#-overview)
- [Key Features](#-key-features)
- [Security Architecture](#-security-architecture)
- [System Architecture](#-system-architecture)
- [Installation & Building](#-installation--building)
- [Usage Guide](#-usage-guide)
- [Web Console](#-web-console)
- [Terminal TUI](#-terminal-tui)
- [Slash Commands Reference](#-slash-commands-reference)
- [Configuration](#-configuration)
- [File Transfers](#-file-transfers)
- [Encryption Deep Dive](#-encryption-deep-dive)
- [Network Discovery](#-network-discovery)
- [Testing](#-testing)
- [Troubleshooting](#-troubleshooting)
- [Practical Use Cases](#-practical-use-cases)
- [Dependencies](#-dependencies)
- [Contributing](#-contributing)
- [License](#-license)

---

## 📖 Overview

### What is Yggdrasil Mesh Chat?

**Yggdrasil Mesh Chat** is a next-generation, serverless, peer-to-peer encrypted messaging and file exchange client that operates entirely in user-space on top of the [Yggdrasil IPv6 Overlay Network](https://yggdrasil-network.github.io/). Unlike traditional messaging applications that rely on centralized servers, cloud infrastructure, and third-party services, Yggdrasil Mesh Chat enables direct, encrypted communication between nodes with **zero external servers**, **zero user accounts**, and **zero third-party coordination**.

### What is Yggdrasil?

[Yggdrasil](https://yggdrasil-network.github.io/) is an early-stage implementation of a fully encrypted, self-arranging IPv6 overlay network. It creates a decentralized mesh network where:

- **Every node gets a unique IPv6 address** — derived cryptographically from your public key
- **Routing is automatic** — nodes discover each other and build optimal paths
- **All traffic is encrypted** — using the Noise Protocol Framework
- **No central authority** — no DNS servers, no certificate authorities, no registrars
- **Works anywhere** — over LAN, WAN, internet, or any TCP/UDP connection

The name "Yggdrasil" comes from Norse mythology — the immense cosmic tree that connects the nine worlds. Similarly, the Yggdrasil network connects nodes across the world in a tree-like routing structure.

### How Does It Work?

```mermaid
sequenceDiagram
    participant Alice as Alice's Node
    participant Ygg as Yggdrasil Network
    participant Bob as Bob's Node
    
    Note over Alice,Bob: Both nodes join the Yggdrasil overlay
    
    Alice->>Ygg: Node joins with Ed25519 keypair
    Bob->>Ygg: Node joins with Ed25519 keypair
    
    Note over Alice,Bob: Nodes get IPv6 addresses derived from keys
    
    Alice->>Alice: Generate ECDH keypair for encryption
    Bob->>Bob: Generate ECDH keypair for encryption
    
    Alice->>Ygg: Send contact request + ECDH public key
    Ygg->>Bob: Deliver contact request
    
    Bob->>Bob: Accept & derive shared secret
    Bob->>Ygg: Send acceptance + ECDH public key
    Ygg->>Alice: Deliver acceptance
    
    Alice->>Alice: Derive same shared secret
    Note over Alice,Bob: AES-256 key derived via SHA-256(ECDH_shared_secret)
    
    Alice->>Ygg: Send AES-GCM encrypted message
    Ygg->>Bob: Deliver encrypted message
    Bob->>Bob: Decrypt with shared AES key
```

### Why Was This Built?

Traditional messaging applications (Signal, WhatsApp, Telegram, Discord) all share common problems:

1. **Centralized Infrastructure**: They rely on servers that can go down, be hacked, or be compelled to hand over data
2. **Account Requirements**: Phone numbers, email addresses, or other personally identifiable information is required
3. **Metadata Collection**: Even with E2EE, servers know who talks to whom, when, and how often
4. **Internet Dependency**: They require internet connectivity to function
5. **Single Points of Failure**: If the company shuts down, the service disappears
6. **Trust Requirements**: You must trust the company to not backdoor the encryption

Yggdrasil Mesh Chat solves all of these problems by being:

- **Truly Decentralized**: No servers exist — nodes communicate directly
- **Identity-Free**: No accounts, phone numbers, or emails — just cryptographic keys
- **Metadata-Free**: No central entity sees your communication patterns
- **Network-Agnostic**: Works over LAN, WAN, mesh networks, or the internet
- **Resilient**: No single point of failure — the network survives as long as any two nodes exist
- **Trustless**: Cryptography, not corporate promises, protects your messages

### Architecture Philosophy

The application follows these core principles:

1. **Single Binary Distribution**: The entire application compiles into one executable with zero runtime dependencies
2. **User-Space Operation**: No root privileges, TUN/TAP devices, or kernel modules required
3. **Dual Interface Design**: Choose between a modern web UI or a traditional terminal interface
4. **Defense in Depth**: Multiple security layers protect against various attack vectors
5. **Graceful Degradation**: Features work even when some components fail
6. **Offline-First**: Messages queue locally and deliver when contacts come online

### Technical Specifications

| Specification | Value |
|---------------|-------|
| **Language** | Go 1.20+ |
| **Network Protocol** | Yggdrasil IPv6 Overlay (Noise Protocol Framework) |
| **Routing** | Ironwood cryptographic routing tree |
| **Transport** | TCP and UDP (auto-selected) |
| **Encryption (Network)** | ChaCha20-Poly1305 (Yggdrasil layer) |
| **Encryption (Messages)** | AES-256-GCM (application layer) |
| **Key Exchange** | Curve25519 ECDH |
| **Key Derivation** | SHA-256 |
| **Identity Keys** | Ed25519 |
| **Discovery** | UDP Multicast (224.0.0.50:9999) |
| **File Chunk Size** | 8 KB (8192 bytes) |
| **Default Port** | 9000 (Yggdrasil), 8080 (Web Console) |
| **Binary Size** | ~16 MB (self-contained) |
| **Runtime Dependencies** | None |

### What Makes This Different?

| Feature | Traditional Chat Apps | Yggdrasil Mesh Chat |
|---------|----------------------|---------------------|
| **Server Required** | ✅ Yes (centralized) | ❌ No (peer-to-peer) |
| **Account Registration** | ✅ Required | ❌ Not needed |
| **Internet Dependency** | ✅ Required | ❌ Works on LAN only |
| **Metadata Collection** | ✅ Yes | ❌ None |
| **Single Point of Failure** | ✅ Yes | ❌ Decentralized |
| **E2EE by Default** | ❌ Sometimes | ✅ Always |
| **File Sharing** | Via cloud servers | Direct P2P transfer |
| **Open Source** | ❌ Usually proprietary | ✅ Fully open source |
| **Self-Hostable** | ❌ Complex setup | ✅ Just run the binary |
| **Works Offline** | ❌ No | ✅ Yes (LAN mesh) |
| **Phone Number Required** | ✅ Usually | ❌ Never |
| **Metadata Visible to Provider** | ✅ Yes | ❌ No provider exists |

### Comparison with Similar Projects

| Project | Serverless | E2EE | File Transfer | Dual UI | Auto-Discovery |
|---------|------------|------|---------------|---------|----------------|
| **Yggdrasil Mesh Chat** | ✅ | ✅ | ✅ P2P | ✅ Web+TUI | ✅ UDP Multicast |
| Briar | ✅ | ✅ | ✅ | ❌ Android only | ✅ Tor/BT |
| Session | ❌ | ✅ | ✅ | ❌ | ❌ |
| Matrix/Element | ❌ | ✅ | ✅ | ✅ | ❌ |
| Jami | ✅ | ✅ | ✅ | ❌ | ✅ DHT |
| Tox | ✅ | ✅ | ✅ | ❌ | ✅ DHT |

### System Requirements

**Minimum:**
- Go 1.20+ (for building)
- Any modern OS: Windows, macOS, Linux, FreeBSD
- 10 MB disk space
- 50 MB RAM

**Recommended:**
- Go 1.21+ for optimal performance
- Local network connectivity for auto-discovery
- Port 9000 accessible for incoming connections

---

## 🚀 Key Features

### 🔒 End-to-End Encryption (E2EE)

All messages between contacts are encrypted using industry-standard cryptography:

- **Key Exchange**: Curve25519 Diffie-Hellman (ECDH) — the same protocol used by Signal, WhatsApp, and WireGuard
- **Message Encryption**: AES-256-GCM (Galois/Counter Mode) — provides both confidentiality and authenticity
- **Key Derivation**: SHA-256 hash of the shared secret ensures uniform key distribution
- **Perfect Forward Secrecy**: Each contact pair has a unique derived key
- **Visual Indicators**: 🔒 padlock icon shows E2EE status for each contact

### 🌐 Automatic Network Discovery

Nodes automatically discover and connect to nearby peers:

- **UDP Multicast Beaconing**: Broadcasts on `224.0.0.50:9999` every 5 seconds
- **LAN Auto-Peering**: Nodes on the same local network connect automatically
- **Self-Detection**: Prevents nodes from connecting to themselves
- **Zero Configuration**: No manual peer entry required for local network
- **Manual Peering**: Add remote peers via TCP URI (e.g., `tcp://1.2.3.4:9000`)

### 🌍 Dual Interface Mode

Choose your preferred interface:

**Web Console (Default)**
- Glassmorphic design with Tokyo Night color scheme
- Real-time updates via Server-Sent Events (SSE)
- CSS animations and screen shake effects
- Audio notifications via Web Audio API
- Responsive layout for all screen sizes
- Command autocomplete with Tab key

**Terminal TUI (Alternative)**
- 5 built-in themes: Catppuccin Mocha, Nord, Gruvbox, Dracula, Tokyo Night
- Keyboard-driven navigation (Tab, Arrow keys, shortcuts)
- Split-pane layout: Contacts sidebar | Chat viewport | Input field
- Inline image previews using ANSI half-block characters
- Typing indicators and read receipts
- Message history scrolling

### 📥 Asynchronous P2P File Transfers

Send files directly between nodes without any intermediary:

- **Chunked Transfer**: Files split into 8KB chunks for reliable delivery
- **Asynchronous Sending**: Non-blocking file transfers with progress tracking
- **Progress Indicators**: Real-time percentage display during transfer
- **Image Previews**: PNG/JPG files show inline previews after completion
- **Safe Filenames**: Path traversal protection prevents directory escape
- **Auto-Resume**: Failed chunks can be retried

### 💬 Real-Time Messaging Features

- **Read Receipts**: Single check (✓) when sent, double check (✓✓) when read
- **Typing Indicators**: See when your contact is typing (debounced at 2 seconds)
- **Nudge/Shake**: Send attention-grabbing screen vibrations with audio alerts
- **Offline Queueing**: Messages buffered locally when contact is offline
- **Auto-Flush**: Queued messages automatically sent when contact comes online
- **Message Search**: Full-text search through chat history

### ⌨️ Keyboard Shortcuts (TUI)

| Shortcut | Action |
|----------|--------|
| `Tab` / `Shift+Tab` | Cycle focus: Sidebar → Viewport → Input |
| `Ctrl+T` | Cycle themes (Mocha → Nord → Gruvbox → Dracula → Tokyo Night) |
| `Ctrl+Y` | Copy public key to clipboard |
| `Ctrl+U` | Clear input field |
| `Ctrl+D` | Toggle timestamp display |
| `Ctrl+R` | Force retry all peer connections |
| `Ctrl+N` | Add new contact |
| `Ctrl+A` | Add new peer |
| `Ctrl+P` | Switch to Peers view |
| `Ctrl+H` | Switch to Chat view |
| `Ctrl+S` | Switch to Settings view |
| `Ctrl+C` | Quit application |
| `↑` / `Down` | Navigate contacts / Scroll chat / Input history |
| `Enter` | Send message / Select contact |
| `Delete` / `Backspace` | Remove selected peer |

---

## 🛡️ Security Architecture

### Threat Model

Yggdrasil Mesh Chat is designed to protect against:

| Threat | Protection |
|--------|------------|
| **Eavesdropping** | E2EE with AES-256-GCM |
| **Man-in-the-Middle** | Curve25519 key exchange with contact verification |
| **Message Forgery** | GCM authentication tags |
| **Replay Attacks** | Timestamp-based nonce with random component |
| **Path Traversal** | Filename sanitization via `filepath.Base()` |
| **XSS Attacks** | HTML escaping of all user input |
| **Flooding/Spam** | Rate limiting on contact requests (5/minute) |
| **Data Corruption** | Atomic file writes (temp + rename) |
| **Race Conditions** | Mutex-protected configuration access |

### Cryptographic Details

```
Key Exchange Flow:
  Alice                          Bob
    |                              |
    |-- Contact Request + PubA -->|
    |                              |
    |<-- Contact Accept + PubB ---|
    |                              |
    SharedSecret = ECDH(PrivA, PubB) = ECDH(PrivB, PubA)
    AESKey = SHA-256(SharedSecret)
    |                              |
    |-- AES-GCM(Key, Nonce, Msg) ->|
    |<- AES-GCM(Key, Nonce, Msg) --|
```

### Security Best Practices

1. **Verify Contact Keys**: Always verify your contact's public key through a separate channel
2. **Use Strong Usernames**: Choose unique usernames to avoid impersonation
3. **Keep Software Updated**: Regularly pull latest changes for security patches
4. **Firewall Configuration**: Block unnecessary ports on your overlay IPv6 address
5. **Separate Configs**: Use different config files for different identities

---

## 🏗️ System Architecture

```mermaid
graph TB
    subgraph "User Interfaces"
        WEB[Web Console Browser]
        TUI[Terminal TUI]
    end
    
    subgraph "Application Layer"
        WS[Web Server<br/>HTTP + SSE]
        BT[Bubble Tea Loop<br/>Model-Update-View]
    end
    
    subgraph "Core Layer"
        YM[YggManager<br/>Packet Handler]
        DM[Discovery Manager<br/>UDP Multicast]
    end
    
    subgraph "Security Layer"
        HELP[helpers.go<br/>ECDH, AES, Sanitization]
    end
    
    subgraph "Network Layer"
        YC[Yggdrasil Core<br/>IPv6 Overlay]
        IR[Ironwood Router<br/>Crypto Routing]
    end
    
    subgraph "Transport"
        TCP[TCP Sockets]
        UDP[UDP Sockets]
        MC[Multicast UDP]
    end
    
    WEB -->|HTTP/SSE| WS
    TUI -->|Commands| BT
    WS --> YM
    BT --> YM
    WS --> HELP
    BT --> HELP
    YM --> YC
    DM --> MC
    YC --> IR
    IR --> TCP
    IR --> UDP
    DM --> YM
```

### File Structure

```
yggchat/
├── main.go                 # Application entry point & CLI flags
├── config.go               # Configuration management with atomic writes
├── ygg.go                  # Yggdrasil network manager & packet handling
├── ygg_test.go             # Core functionality tests
├── discovery.go            # UDP multicast peer discovery
├── helpers.go              # Security functions & shared utilities
├── helpers_test.go         # Comprehensive helper tests
├── web_server.go           # HTTP server, SSE, API endpoints
├── tui.go                  # Terminal UI (Bubble Tea framework)
├── ui_styles.go            # Theme definitions (5 color schemes)
├── image_render.go         # ANSI image preview renderer
├── web/
│   ├── index.html          # Web console HTML structure
│   ├── index.css           # Glassmorphic CSS styling
│   └── index.js            # Client-side JavaScript logic
├── go.mod                  # Go module dependencies
├── go.sum                  # Dependency checksums
├── .gitignore              # Git ignore rules
├── README.md               # This file
└── logo.png                # Application logo
```

### Data Flow

```
1. User types message in Web/TUI
2. Message sent to WebServer/TUI handler
3. Handler checks if contact has shared secret (E2EE)
4. If encrypted: AES-GCM encrypt with derived key
5. Payload serialized as JSON with "YGGC" magic header
6. YggManager sends packet via Yggdrasil overlay
7. Recipient's YggManager receives packet
8. Payload deserialized and decrypted if needed
9. Message displayed in recipient's UI
```

---

## ⚙️ Installation & Building

### Prerequisites

| Requirement | Version | Notes |
|-------------|---------|-------|
| **Go** | 1.20+ | Download from [golang.org](https://golang.org/dl/) |
| **Git** | Any | For cloning the repository |
| **OS** | Windows/macOS/Linux/FreeBSD | Cross-platform support |
| **Disk Space** | 50 MB | For source + build artifacts |
| **RAM** | 50 MB minimum | 100 MB recommended |

### Option 1: Build from Source (Recommended)

```bash
# Clone the repository
git clone https://github.com/amafjarkasi/yggchat.git
cd yggchat

# Download dependencies
go mod tidy

# Build for your current platform
go build -o yggchat

# Verify the build
./yggchat --help
```

### Option 2: Cross-Platform Builds

Build for different operating systems and architectures:

```bash
# Windows (64-bit)
GOOS=windows GOARCH=amd64 go build -o yggchat.exe

# Windows (32-bit)
GOOS=windows GOARCH=386 go build -o yggchat32.exe

# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o yggchat-mac-intel

# macOS (Apple Silicon M1/M2/M3)
GOOS=darwin GOARCH=arm64 go build -o yggchat-mac-arm

# Linux (64-bit)
GOOS=linux GOARCH=amd64 go build -o yggchat-linux

# Linux (ARM - Raspberry Pi)
GOOS=linux GOARCH=arm GOARM=7 go build -o yggchat-pi

# Linux (ARM64 - Raspberry Pi 4)
GOOS=linux GOARCH=arm64 go build -o yggchat-pi4

# FreeBSD
GOOS=freebsd GOARCH=amd64 go build -o yggchat-freebsd
```

### Option 3: Optimized Production Build

```bash
# Strip debug info and reduce binary size
go build -ldflags="-s -w" -o yggchat

# Build with version info
go build -ldflags="-X main.Version=1.0.0 -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o yggchat
```

### Running Tests

```bash
# Run all tests
go test -v ./...

# Run specific test suites
go test -v -run TestECDHKeyExchange        # Cryptography tests
go test -v -run TestSafeSenderName         # Security helper tests
go test -v -run TestSanitizeFilename       # Input validation tests
go test -v -run TestConfigLoadSave         # Configuration tests
go test -v -run TestIsContactRequestAllowed # Rate limiting tests

# Run tests with coverage report
go test -cover ./...

# Run tests with detailed coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out           # Opens HTML coverage report

# Run benchmarks
go test -bench=. ./...
```

### Quick Start

After building, you can immediately start using the application:

```bash
# Launch Web Console (default - opens browser automatically)
./yggchat

# Launch Terminal TUI
./yggchat --tui

# Custom port for Web Console
./yggchat --port 9090

# Use a custom config file (for multiple identities)
./yggchat --config alice.json

# Combine multiple flags
./yggchat --tui --config bob.json --port 8888
```

### CLI Flags Reference

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | `yggchat.json` | Configuration file path |
| `--username` | (from config) | Override username |
| `--tui` | `false` | Use Terminal UI instead of Web Console |
| `--port` | `8080` | Port for Web Console HTTP server |

### First Run Walkthrough

```bash
# 1. Build the application
go build -o yggchat

# 2. Run it (Web Console mode)
./yggchat
# Output: [🌍 Web Console Server running on http://127.0.0.1:8080]
# Browser opens automatically

# 3. Your node is now live on the Yggdrasil mesh!
# 4. Check the Settings tab for your public key
# 5. Share your public key with contacts
# 6. Add contacts using /add command or the + button
```

---

## 🎮 Usage Guide

### First Launch

1. Run `./yggchat.exe` (Web Console) or `./yggchat.exe --tui` (Terminal)
2. A new config file (`yggchat.json`) is generated with:
   - Ed25519 private key for Yggdrasil identity
   - Curve25519 ECDH key for encryption
   - Default listener on `tcp://0.0.0.0:9000`
3. Your node joins the Yggdrasil mesh network
4. UDP multicast discovers nearby peers automatically

### Adding a Contact

1. Get your contact's Yggdrasil public key (they can copy it with `Ctrl+Y`)
2. Use `/add <public_key> <nickname>` or click the **+** button
3. A contact request is sent with your ECDH public key
4. When they accept, E2EE is automatically established
5. Look for the 🔒 padlock indicator

### Sending Messages

1. Select a contact from the sidebar
2. Type your message in the input field
3. Press `Enter` or click **SEND**
4. Messages are encrypted if E2EE is established
5. Read receipts show ✓ (sent) and ✓✓ (read)

### Sending Files

1. Select a contact
2. Use `/send <filepath>` (e.g., `/send ~/photo.jpg`)
3. File is split into 8KB chunks and sent asynchronously
4. Progress is shown in real-time
5. Image files (PNG/JPG) display inline previews

---

## 🖥️ Web Console

### Features

- **Glassmorphic UI**: Translucent panels with blur effects
- **Real-Time Updates**: Server-Sent Events (SSE) for instant messaging
- **Audio Notifications**: Web Audio API generates beep sounds
- **Screen Shake**: CSS animations for nudge/shake messages
- **Command Autocomplete**: Press Tab to cycle through commands
- **Responsive Design**: Works on desktop and mobile browsers

### API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/` | GET | Serve web console frontend |
| `/events` | GET | SSE event stream |
| `/api/state` | GET | Get current state (contacts, history, peers) |
| `/api/send` | POST | Send message or command |

### SSE Event Types

| Event | Description |
|-------|-------------|
| `incoming_msg` | New message received |
| `typing` | Contact is typing |
| `read` | Read receipt received |
| `shake` | Nudge/shake received |
| `contact_req` | Contact request received |
| `peers` | Peer status update |

---

## 💻 Terminal TUI

### Views

**Chat View** (`Ctrl+H`)
- Contact sidebar with unread indicators
- Message viewport with scroll support
- Input field with command history (↑/↓ arrows)
- Typing indicator display

**Peers View** (`Ctrl+P`)
- List of connected peers
- Online/offline status
- Connection direction (inbound/outbound)
- Latency and traffic statistics
- Remove peers with Delete key

**Settings View** (`Ctrl+S`)
- Username display
- IPv6 overlay address
- Public key (hex)
- Listener configuration

### Themes

| Theme | Description |
|-------|-------------|
| Catppuccin Mocha | Default dark theme with pastel accents |
| Nord | Arctic blue color scheme |
| Gruvbox | Retro warm color scheme |
| Dracula | Purple-accented dark theme |
| Tokyo Night | Deep blue night theme |

---

## 💬 Slash Commands Reference

### General Commands

| Command | Description | Example |
|---------|-------------|---------|
| `/help` | Show all available commands | `/help` |
| `/nick <name>` | Change your display name | `/nick Alice` |
| `/clear` | Clear chat history for active contact | `/clear` |
| `/whois` | Display contact information and E2EE status | `/whois` |

### Network Commands

| Command | Description | Example |
|---------|-------------|---------|
| `/peer <uri>` | Connect to a remote peer | `/peer tcp://1.2.3.4:9000` |
| `/ping` | Measure latency to active contact | `/ping` |

### Contact Commands

| Command | Description | Example |
|---------|-------------|---------|
| `/add <key> <name>` | Send contact request | `/add abc123... Bob` |
| `/shake` | Send nudge/shake to active contact | `/shake` |
| `/shout <msg>` | Broadcast message to all contacts | `/shout Hello everyone!` |

### File Commands

| Command | Description | Example |
|---------|-------------|---------|
| `/send <path>` | Send file to active contact | `/send ~/document.pdf` |
| `/search <query>` | Search chat history | `/search important` |

---

## 🔧 Configuration

### Config File Location

Default: `yggchat.json` in the current directory

Override with: `--config <filename>`

### Config Structure

```json
{
  "privateKey": "hex-encoded-ed25519-private-key",
  "ecdhPrivateKey": "hex-encoded-curve25519-ecdh-key",
  "peers": [
    "tcp://192.168.1.100:9000",
    "tcp://10.0.0.5:9000"
  ],
  "listeners": [
    "tcp://0.0.0.0:9000"
  ],
  "contacts": {
    "contact-public-key-hex": {
      "publicKey": "contact-public-key-hex",
      "nickname": "Alice",
      "sharedSecret": "hex-encoded-aes-256-key"
    }
  },
  "username": "MyUsername"
}
```

### Config Fields

| Field | Type | Description |
|-------|------|-------------|
| `privateKey` | string | Ed25519 private key for Yggdrasil identity (auto-generated) |
| `ecdhPrivateKey` | string | Curve25519 private key for E2EE (auto-generated) |
| `peers` | string[] | List of peer URIs to connect to |
| `listeners` | string[] | List of addresses to listen on |
| `contacts` | object | Map of contact public keys to contact info |
| `username` | string | Your display name |

### Related Files

| File | Description |
|------|-------------|
| `yggchat.json` | Main configuration |
| `yggchat_history.json` | Chat message history |
| `yggchat_pending.json` | Queued offline messages |
| `downloads/` | Received files |

---

## 📥 File Transfers

### How It Works

1. Sender reads file from disk
2. File is split into 8KB chunks
3. Each chunk sent as a `file_chunk` payload
4. Receiver appends chunks to file in `./downloads/`
5. Progress updates displayed in chat
6. Image files render inline previews on completion

### Supported File Types

| Type | Inline Preview |
|------|---------------|
| PNG | ✅ Yes |
| JPG/JPEG | ✅ Yes |
| Other | ❌ No (saved to downloads) |

### Transfer Limits

- Chunk size: 8KB (8192 bytes)
- Pacing: 50ms delay between chunks
- Max filename: Sanitized via `filepath.Base()`
- Storage: `./downloads/` directory

---

## 🔐 Encryption Deep Dive

### Key Exchange Protocol

1. **Alice** generates ephemeral Curve25519 keypair
2. **Alice** sends `contact_req` with her public key to **Bob**
3. **Bob** generates ephemeral Curve25519 keypair
4. **Bob** computes `SharedSecret = ECDH(PrivB, PubA)`
5. **Bob** derives `AESKey = SHA-256(SharedSecret)`
6. **Bob** sends `contact_acc` with his public key to **Alice**
7. **Alice** computes same `SharedSecret = ECDH(PrivA, PubB)`
8. Both now share the same AES-256 key

### Message Encryption

```go
// Encryption
nonce = random(12 bytes)
ciphertext = AES-GCM-Seal(key, nonce, plaintext, aad)
payload = {text: hex(ciphertext), nonce: hex(nonce), is_encrypted: true}

// Decryption
plaintext = AES-GCM-Open(key, nonce, ciphertext, aad)
```

### Security Properties

- **Confidentiality**: AES-256-GCM encryption
- **Authenticity**: GCM authentication tags
- **Integrity**: Tampered messages fail decryption
- **Forward Secrecy**: Per-contact key isolation
- **No Key Escrow**: Keys never leave your device

---

## 🌐 Network Discovery

### Multicast Discovery Protocol

```
Every 5 seconds:
1. Get local IPv4 addresses
2. For each address, broadcast: "tcp://<ip>:<port>"
3. Listen for broadcasts from other nodes
4. If received URI is not self, add as peer
```

### Manual Peering

For connections beyond local network:

```bash
# From Node A
/peer tcp://<node-b-ip>:9000

# Or configure in yggchat.json
{
  "peers": ["tcp://203.0.113.50:9000"]
}
```

### NAT Traversal

Yggdrasil supports peering through:
- Direct TCP connections
- UDP connections
- SOCKS5 proxies
- Tor hidden services

---

## 🧪 Testing

### Running Tests

```bash
# All tests
go test -v ./...

# Specific test suite
go test -v -run TestSafeSenderName
go test -v -run TestDeriveSharedSecret
go test -v -run TestSanitizeFilename
go test -v -run TestECDHKeyExchange
go test -v -run TestIsContactRequestAllowed

# With coverage
go test -cover ./...
```

### Test Coverage

| Module | Tests |
|--------|-------|
| Configuration | Load/save, key encoding, contacts |
| Chat Protocol | Packet formatting, serialization |
| Cryptography | ECDH exchange, AES-GCM encrypt/decrypt |
| Security Helpers | Filename sanitization, HTML escaping, rate limiting |
| Utilities | Timestamp stripping, ANSI removal, image detection |
| History | Atomic writes, load/save integrity |

---

## ❓ Troubleshooting

### Common Issues

**"Failed to start Yggdrasil core"**
- Port 9000 may be in use
- Try: `./yggchat.exe --config new.json` for fresh config

**No peers discovered**
- Ensure UDP multicast is not blocked by firewall
- Check if nodes are on same subnet
- Try manual peering: `/peer tcp://<ip>:9000`

**Messages not delivering**
- Check if contact is online (peer status)
- Messages queue automatically for offline contacts
- Use `/ping` to test connectivity

**E2EE not working**
- Ensure contact request was accepted
- Check for 🔒 padlock indicator
- Verify with `/whois` command

**Web Console not loading**
- Check if port 8080 is available
- Try different port: `--port 9090`
- Check browser console for errors

**File transfer fails**
- Ensure `./downloads/` directory exists
- Check file permissions
- Verify contact is online

---

## 💡 Practical Use Cases

### 1. Disaster Relief & Emergency Operations

**Scenario**: A natural disaster (earthquake, hurricane, flood) has destroyed internet infrastructure. Emergency responders need to coordinate across a disaster zone.

**How Yggdrasil Mesh Chat Helps**:
- **Zero Infrastructure Required**: No cell towers, no internet, no servers needed
- **Instant Deployment**: Copy the single binary to laptops via USB drive
- **Auto-Discovery**: Nodes find each other automatically on the same Wi-Fi network (even a basic router)
- **Encrypted Coordination**: Sensitive location data and victim information stays encrypted
- **File Sharing**: Share maps, medical records, evacuation routes directly P2P

**Setup Example**:
```bash
# Emergency responder setup (takes 30 seconds)
# 1. Copy yggchat binary to laptop
# 2. Connect to local Wi-Fi router (no internet needed)
./yggchat --config responder-alpha.json --username "Alpha Team"

# 3. Other responders do the same
# 4. Nodes auto-discover each other via UDP multicast
# 5. Add team members by sharing public keys
/add <alpha-team-key> "Alpha Team Lead"
```

**Real-World Benefits**:
- Works in areas with zero cellular coverage
- No dependency on external services that may be down
- Coordination continues even if internet infrastructure is destroyed
- Messages queue offline if a team member temporarily disconnects

---

### 2. Development Team Communication

**Scenario**: A software development team works in a secure lab environment where external services (Slack, Discord, Teams) are blocked by corporate firewall or security policy.

**How Yggdrasil Mesh Chat Helps**:
- **No External Dependencies**: Works entirely within the corporate network
- **No IT Approval Needed**: Single binary, no installation required
- **No Cloud Services**: All data stays on the local network
- **Code Sharing**: Share code snippets, config files, logs directly
- **Compliance**: E2EE by default meets security audit requirements

**Setup Example**:
```bash
# Developer workstation setup
# 1. Download binary (or build from source)
go build -o yggchat

# 2. Launch on workstation
./yggchat --username "Alice - Backend"

# 3. All developers on same network auto-discover each other
# 4. Add specific teammates
/add <bob-key> "Bob - Frontend"
/add <charlie-key> "Charlie - DevOps"

# 5. Share files directly
/send ./logs/error-2024-01-15.log
/send ./configs/staging.env
```

**Real-World Benefits**:
- No monthly SaaS subscription costs
- No data leaves the corporate network
- Works in air-gapped environments
- Developers can use TUI from their terminal workflow

---

### 3. Journalism & Whistleblowing

**Scenario**: A journalist needs to communicate securely with a source who cannot risk metadata exposure. Traditional messaging apps (Signal, WhatsApp) require phone numbers that can be traced.

**How Yggdrasil Mesh Chat Helps**:
- **No Phone Numbers Required**: Identity is purely cryptographic keys
- **No Metadata**: No central server knows who is talking to whom
- **No Account Creation**: No email, no phone, no identity verification
- **Works Over Tor**: Can be routed through Tor for additional anonymity
- **Plausible Deniability**: No account history, no registration records

**Setup Example**:
```bash
# Journalist setup (on a clean laptop)
# 1. Boot from Tails OS or similar privacy-focused OS
# 2. Connect through Tor or VPN
# 3. Run with a disposable config
./yggchat --config temp-source-interview.json --username "anon"

# 4. Share public key through secure channel (Signal, in-person, dead drop)
# 5. Source adds journalist's key
# 6. All communication is E2EE with no metadata trail
```

**Security Considerations**:
- Generate a fresh config for each sensitive source
- Use Tails OS or Whonix for maximum anonymity
- Verify contact keys through out-of-band channel
- Delete config files after communication is complete

---

### 4. Conferences, Hackathons & Events

**Scenario**: A tech conference or hackathon with 200+ attendees. The venue Wi-Fi is overloaded, cellular coverage is poor, and organizers need a reliable communication channel.

**How Yggdrasil Mesh Chat Helps**:
- **Local Network Only**: Works on venue LAN without internet
- **Auto-Scaling**: More nodes = stronger mesh network
- **File Sharing**: Share presentations, code repos, slides directly
- **Zero Setup for Attendees**: Just run the binary
- **Fun & Engaging**: Retro TUI appeals to developer audiences

**Setup Example**:
```bash
# Event organizer setup (before event)
# 1. Set up a dedicated Wi-Fi network for the event
# 2. Pre-install yggchat on event machines
# 3. Create event-specific config

# Attendee setup (takes 1 minute)
./yggchat --config hackathon.json --username "Team-Phoenix"

# 4. All attendees on same Wi-Fi auto-discover each other
# 5. Create team channels by adding teammates
/add <teammate1-key> "Phoenix - Lead"
/add <teammate2-key> "Phoenix - Designer"

# 6. Share project files
/send ./project/README.md
/send ./demo/presentation.pdf
```

**Event Benefits**:
- No need to share Wi-Fi passwords for messaging
- Works even if venue internet is unreliable
- Attendees can continue communicating after leaving venue (via manual peering)
- Great icebreaker activity (share public keys)

---

### 5. Gaming & LAN Parties

**Scenario**: A group of friends gathering for a LAN party. They want to coordinate gaming sessions, share game files, and trash-talk without relying on external services.

**How Yggdrasil Mesh Chat Helps**:
- **Zero-Latency**: Direct LAN communication, no server round-trips
- **File Sharing**: Share game mods, maps, save files directly
- **No Accounts**: No Discord or Steam required
- **Fun TUI**: Retro terminal interface fits the LAN party aesthetic
- **Shake Command**: Annoy your friends with screen shakes!

**Setup Example**:
```bash
# LAN party setup
# Everyone connects to the same network switch or Wi-Fi

# Player 1
./yggchat --username "xX_ProGamer_Xx"

# Player 2
./yggchat --username "N00bSlayer"

# Auto-discovery finds everyone on the network
# Add friends
/add <friend-key> "N00bSlayer"

# Share game files
/send ./mods/texture-pack.zip
/send ./maps/custom-arena.bsp

# Trash talk
/shout GG no re!
/shake
```

**Gaming Benefits**:
- No latency (direct LAN connection)
- No internet bandwidth consumed
- Share large files without upload limits
- Works in basements and rural areas without good internet

---

### 6. Military & Defense Communications

**Scenario**: A military unit needs secure, serverless communication in the field where traditional communication infrastructure is unavailable or compromised.

**How Yggdrasil Mesh Chat Helps**:
- **No Central Server**: Cannot be taken down by enemy action
- **Encrypted by Default**: Military-grade Curve25519 + AES-256
- **Works Over Any Network**: Radio, satellite, mesh, LAN
- **Air-Gap Compatible**: Works on isolated networks
- **Minimal Footprint**: Single binary, no dependencies

**Security Features for Military Use**:
- All keys generated locally, never transmitted in plaintext
- No metadata leakage to any third party
- Messages can be configured to auto-expire (manual deletion)
- Works on hardened Linux systems

---

### 7. Rural & Remote Area Communication

**Scenario**: Communities in rural areas with limited or no internet connectivity. Villages separated by mountains where cell towers don't reach.

**How Yggdrasil Mesh Chat Helps**:
- **Works Over Long-Range Wi-Fi**: Can use directional antennas
- **No ISP Required**: Pure peer-to-peer communication
- **Low Bandwidth**: Minimal network overhead
- **Solar Compatible**: Can run on low-power devices (Raspberry Pi)

**Rural Setup Example**:
```bash
# Village A - Raspberry Pi with long-range antenna
./yggchat --config village-a.json --username "Village-A-Hub"

# Village B - Another Raspberry Pi
./yggchat --config village-b.json --username "Village-B-Hub"

# Manual peering over long-range link
/peer tcp://10.0.0.2:9000

# Residents connect to their village hub
# Messages route through the mesh between villages
```

---

### 8. Academic Research Networks

**Scenario**: Researchers at different universities need to share sensitive research data (medical records, unpublished findings) that cannot go through commercial cloud services.

**How Yggdrasil Mesh Chat Helps**:
- **HIPAA/GDPR Compatible**: No third-party data processors
- **Direct Transfer**: Data goes researcher-to-researcher
- **Audit Trail**: Local history logs for compliance
- **Cross-Institution**: Works across university networks

---

### 9. Offshore & Maritime Communication

**Scenario**: Ships at sea need to communicate with each other and with shore stations without satellite internet.

**How Yggdrasil Mesh Chat Helps**:
- **Works Over Radio**: Can operate over HF/VHF radio links
- **Store-and-Forward**: Messages queue until connection available
- **Ship-to-Ship**: Direct communication between vessels
- **Low Bandwidth**: Efficient protocol for limited links

---

### 10. Privacy-Conscious Personal Use

**Scenario**: An individual who values privacy and doesn't want their messaging habits tracked by corporations.

**How Yggdrasil Mesh Chat Helps**:
- **No Cloud Sync**: Messages stay on your device
- **No Analytics**: No tracking, no telemetry
- **No Ads**: No advertising business model
- **Open Source**: Fully auditable codebase
- **Self-Sovereign**: You own your data completely

---

## 🚀 Deployment

### Standalone Desktop Deployment

The simplest deployment - just run the binary on your desktop:

```bash
# Windows
yggchat.exe

# macOS / Linux
./yggchat
```

**When to Use**: Personal use, small teams, testing

### LAN Server Deployment

Deploy on a dedicated machine that stays online 24/7:

```bash
# On a Linux server
# 1. Copy binary to server
scp yggchat user@server:/opt/yggchat/

# 2. Create systemd service
sudo tee /etc/systemd/system/yggchat.service << EOF
[Unit]
Description=Yggdrasil Mesh Chat Server
After=network.target

[Service]
Type=simple
User=yggchat
Group=yggchat
WorkingDirectory=/opt/yggchat
ExecStart=/opt/yggchat/yggchat --port 8080
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# 3. Create service user
sudo useradd -r -s /bin/false yggchat
sudo chown -R yggchat:yggchat /opt/yggchat

# 4. Enable and start
sudo systemctl daemon-reload
sudo systemctl enable yggchat
sudo systemctl start yggchat

# 5. Check status
sudo systemctl status yggchat
```

**When to Use**: Team deployments, always-on nodes, office environments

### Docker Deployment

Containerized deployment for easy management:

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o yggchat

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/yggchat .
COPY --from=builder /app/web ./web
EXPOSE 8080 9000
CMD ["./yggchat", "--port", "8080"]
```

```bash
# Build and run
docker build -t yggchat .
docker run -d \
  --name yggchat \
  -p 8080:8080 \
  -p 9000:9000 \
  -v yggchat-data:/app/data \
  yggchat
```

```yaml
# docker-compose.yml
version: '3.8'
services:
  yggchat:
    build: .
    ports:
      - "8080:8080"
      - "9000:9000"
    volumes:
      - yggchat-data:/app
    restart: unless-stopped

volumes:
  yggchat-data:
```

**When to Use**: Containerized environments, Kubernetes, cloud deployments

### Raspberry Pi Deployment

Deploy on a low-power Raspberry Pi for always-on mesh networking:

```bash
# On Raspberry Pi (ARM)
# 1. Cross-compile for ARM
GOOS=linux GOARCH=arm GOARM=7 go build -o yggchat-pi

# 2. Copy to Pi
scp yggchat-pi pi@raspberrypi:/home/pi/yggchat/

# 3. SSH to Pi and run
ssh pi@raspberrypi
chmod +x yggchat-pi
./yggchat-pi --username "Pi-Node"

# 4. (Optional) Auto-start on boot
crontab -e
# Add: @reboot /home/pi/yggchat/yggchat-pi --username "Pi-Node"
```

**When to Use**: Home networks, IoT mesh, low-power deployments, rural networks

### Multi-Node Mesh Deployment

Deploy multiple nodes for redundancy:

```bash
# Node 1 - Primary (Office)
./yggchat --config node1.json --port 8080 --username "Office-Primary"

# Node 2 - Backup (Remote Office)
./yggchat --config node2.json --port 8080 --username "Remote-Backup"

# Node 3 - Mobile (Laptop)
./yggchat --config node3.json --port 8080 --username "Mobile"

# Nodes peer with each other
# Node 2 connects to Node 1
/peer tcp://office-ip:9000

# Node 3 connects to both
/peer tcp://office-ip:9000
/peer tcp://remote-ip:9000
```

**When to Use**: High availability, geographic distribution, redundancy

### Air-Gapped Network Deployment

Deploy on networks with no internet access:

```bash
# On air-gapped machine
# 1. Build on internet-connected machine
go build -o yggchat

# 2. Copy binary via USB drive
cp yggchat /media/usb/

# 3. On air-gapped machine
./yggchat --config secure-lab.json

# 4. All nodes on same LAN auto-discover
# 5. Manual peering for segmented networks
/peer tcp://10.0.1.5:9000
```

**When to Use**: Military, classified environments, secure labs, SCADA systems

### Reverse Proxy Deployment (Nginx)

For production web deployments with SSL:

```nginx
# /etc/nginx/sites-available/yggchat
server {
    listen 443 ssl;
    server_name chat.example.com;

    ssl_certificate /etc/letsencrypt/live/chat.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/chat.example.com/privkey.pem;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_read_timeout 86400;
    }

    location /events {
        proxy_pass http://127.0.0.1:8080/events;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
        proxy_buffering off;
        proxy_cache off;
        chunked_transfer_encoding off;
    }
}
```

**When to Use**: Public-facing deployments, SSL termination, domain names

### Firewall Configuration

Ensure the following ports are accessible:

```bash
# Linux (ufw)
sudo ufw allow 8080/tcp   # Web Console
sudo ufw allow 9000/tcp   # Yggdrasil peering
sudo ufw allow 9999/udp   # Multicast discovery

# Linux (iptables)
iptables -A INPUT -p tcp --dport 8080 -j ACCEPT
iptables -A INPUT -p tcp --dport 9000 -j ACCEPT
iptables -A INPUT -p udp --dport 9999 -j ACCEPT

# Windows (PowerShell as Admin)
New-NetFirewallRule -DisplayName "Yggchat Web" -Direction Inbound -Port 8080 -Protocol TCP -Action Allow
New-NetFirewallRule -DisplayName "Yggchat Peer" -Direction Inbound -Port 9000 -Protocol TCP -Action Allow
New-NetFirewallRule -DisplayName "Yggchat Discovery" -Direction Inbound -Port 9999 -Protocol UDP -Action Allow
```

---

## 📦 Dependencies

### Core Dependencies

| Package | Purpose |
|---------|---------|
| `yggdrasil-network/yggdrasil-go` | IPv6 overlay networking |
| `Arceliar/ironwood` | Cryptographic routing |
| `Arceliar/phony` | Actor model concurrency |

### UI Dependencies

| Package | Purpose |
|---------|---------|
| `charmbracelet/bubbletea` | Terminal UI framework |
| `charmbracelet/bubbles` | TUI components (input, viewport) |
| `charmbracelet/lipgloss` | Terminal styling |

### Other Dependencies

| Package | Purpose |
|---------|---------|
| `gologme/log` | Logging |
| `coder/websocket` | WebSocket support |
| `quic-go/quic-go` | QUIC protocol |

---

## 🤝 Contributing

### Development Setup

```bash
# Clone repository
git clone https://github.com/amafjarkasi/yggchat.git
cd yggchat

# Install dependencies
go mod tidy

# Run tests
go test -v ./...

# Build
go build -o yggchat.exe
```

### Code Structure

- `main.go` — Entry point and CLI parsing
- `config.go` — Configuration management
- `ygg.go` — Core Yggdrasil integration
- `helpers.go` — Shared security and utility functions
- `web_server.go` — HTTP/SSE web server
- `tui.go` — Terminal UI implementation
- `discovery.go` — Network discovery

### Pull Request Process

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit pull request with description

---

## 📄 License

MIT License - See [LICENSE](LICENSE) file for details

---

<p align="center">
  <strong>Made with ⚡ and Go</strong>
</p>

<p align="center">
  <a href="https://github.com/amafjarkasi/yggchat">GitHub</a> •
  <a href="https://github.com/amafjarkasi/yggchat/issues">Issues</a> •
  <a href="https://github.com/amafjarkasi/yggchat/releases">Releases</a>
</p>
