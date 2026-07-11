# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2024-01-15

### Added

#### Core Features
- Yggdrasil IPv6 overlay network integration
- End-to-end encryption (Curve25519 ECDH + AES-256-GCM)
- Automatic UDP multicast peer discovery
- Asynchronous P2P file transfers with chunking
- Read receipts (✓ sent, ✓✓ read)
- Typing indicators
- Nudge/shake notifications
- Offline message queueing with auto-retry
- Message history persistence

#### Web Console
- Glassmorphic UI with Tokyo Night theme
- Real-time updates via Server-Sent Events (SSE)
- Audio notifications (Web Audio API)
- Screen shake effects (CSS animations)
- Command autocomplete (Tab key)
- Responsive design

#### Terminal TUI
- Bubble Tea framework integration
- 5 built-in themes (Catppuccin Mocha, Nord, Gruvbox, Dracula, Tokyo Night)
- Split-pane layout (Sidebar, Viewport, Input)
- Inline image previews (ANSI half-blocks)
- Keyboard shortcuts
- Input history (Up/Down arrows)

#### Security
- Path traversal protection
- XSS prevention (HTML escaping)
- Rate limiting on contact requests
- Atomic file writes
- Thread-safe configuration
- Input validation

#### Commands
- `/help` - Show available commands
- `/nick` - Change username
- `/peer` - Add peer connection
- `/add` - Send contact request
- `/ping` - Measure latency
- `/send` - Send file
- `/shake` - Send nudge
- `/whois` - Show contact info
- `/search` - Search history
- `/clear` - Clear chat
- `/shout` - Broadcast message

#### Configuration
- JSON-based configuration
- Auto-generated keys on first run
- Multiple config file support
- Command-line flags

#### Documentation
- Comprehensive README with Mermaid diagrams
- Installation guide (multiple platforms)
- Deployment guide (Docker, systemd, Raspberry Pi)
- Use cases (10 scenarios)
- Security architecture documentation
- API reference

### Security

- Curve25519 ECDH key exchange
- AES-256-GCM message encryption
- SHA-256 key derivation
- Ed25519 identity keys
- Path traversal mitigation
- XSS prevention
- Rate limiting

## [Unreleased]

### Planned
- Group chat support
- Message reactions
- Voice messages
- Message editing/deletion
- Custom theme editor
- Plugin system
- Mobile app

---

## Version History

| Version | Date | Description |
|---------|------|-------------|
| 1.0.0 | 2024-01-15 | Initial release |

---

## How to Update

```bash
# Pull latest changes
git pull origin main

# Rebuild
make build

# Or with Docker
docker pull amafjarkasi/yggchat:latest
```
