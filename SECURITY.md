# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| 1.0.x | ✅ Yes |
| < 1.0 | ❌ No |

## Reporting a Vulnerability

**Please do NOT report security vulnerabilities through public GitHub issues.**

Instead, please report them via email to: **security@example.com** (replace with actual email)

### What to Include

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

### Response Timeline

| Action | Timeframe |
|--------|-----------|
| Acknowledgment | 48 hours |
| Initial assessment | 1 week |
| Fix development | 2-4 weeks |
| Public disclosure | After fix is released |

## Security Features

### End-to-End Encryption

- **Key Exchange**: Curve25519 ECDH
- **Message Encryption**: AES-256-GCM
- **Key Derivation**: SHA-256
- **Identity Keys**: Ed25519

### Network Security

- All Yggdrasil traffic encrypted with Noise Protocol
- No plaintext transmission of keys or messages
- UDP multicast limited to local network

### Application Security

- Path traversal protection (`filepath.Base()`)
- XSS prevention (HTML escaping)
- Rate limiting on contact requests
- Atomic file writes
- Mutex-protected configuration
- Input validation and sanitization

## Security Best Practices for Users

1. **Verify Contact Keys**: Always verify public keys through a separate channel
2. **Use Strong Usernames**: Avoid impersonation
3. **Keep Software Updated**: Pull latest security patches
4. **Firewall Configuration**: Block unnecessary ports
5. **Separate Configs**: Use different configs for different identities
6. **Secure Storage**: Protect config files containing private keys

## Known Limitations

1. **No Forward Secrecy**: Compromised private key exposes all messages
2. **No Deniability**: Messages are authenticated
3. **Metadata**: Network-level metadata visible to ISP (but not message content)
4. **Trust on First Use**: Initial key exchange not verified by default

## Security Audits

This project has not been formally audited. Use at your own risk.

## Cryptographic Libraries

| Library | Purpose |
|---------|---------|
| `crypto/ecdh` | Curve25519 key exchange |
| `crypto/aes` | AES encryption |
| `crypto/cipher` | GCM mode |
| `crypto/ed25519` | Identity keys |
| `crypto/sha256` | Key derivation |

All cryptographic operations use Go's standard library, which is well-maintained and audited.

## Disclosure Policy

- Vulnerabilities disclosed after fix is released
- Credit given to reporters (unless anonymous)
- CVE assigned for significant vulnerabilities

## Contact

For security concerns: **security@example.com** (replace with actual email)

For general questions: Use GitHub Issues
