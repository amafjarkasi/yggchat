# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2024-01-15

### Added

#### Core Messaging Features
- **Message Replies**: Reply to specific messages with context
- **Message Reactions**: React with emojis (👍❤️😂😮😢🔥👏🎉)
- **Message Editing**: Edit sent messages after sending
- **Message Deletion**: Delete messages with confirmation
- **Markdown Support**: Bold, italic, code, strikethrough, links
- **Emoji Picker**: Built-in emoji selector (32 emojis)
- **Broadcast Lists**: Send message to all contacts at once

#### Security Features
- **Contact Blocking**: Block/unblock contacts
- **Panic Button**: Emergency data wipe with triple confirmation
- **Auto-Delete**: Configurable message expiry in settings

#### UI/UX Features
- **Desktop Notifications**: Browser notifications for new messages
- **Drag & Drop Files**: Drop files to send directly
- **Custom Themes**: User-defined color schemes via config
- **Message Search**: Search and highlight messages
- **Typing Toggle**: Enable/disable typing indicators
- **Read Receipt Toggle**: Enable/disable read receipts
- **Peer Statistics Dashboard**: Connection quality metrics
- **Bandwidth Monitor**: Track data usage
- **Connection Graph**: Visual mesh network map
- **Auto-Reconnect**: Automatic reconnection with exponential backoff
- **Video Previews**: Support for MP4/WebM/OGG files

#### Infrastructure
- GitHub Actions CI/CD pipeline
- Dockerfile and docker-compose.yml
- Makefile with 30+ build targets
- MIT LICENSE file
- CONTRIBUTING.md guidelines
- SECURITY.md policy
- CHANGELOG.md (this file)
- Example configuration files

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
