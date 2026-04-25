# VSM Protocol — Invisible Ports

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/cezarpena/vsm-protocol)](https://goreportcard.com/report/github.com/cezarpena/vsm-protocol)
[![npm](https://img.shields.io/npm/v/vsmprotocol)](https://www.npmjs.com/package/vsmprotocol)
[![PyPI](https://img.shields.io/pypi/v/vsmprotocol)](https://pypi.org/project/vsmprotocol/)
[![Crates.io](https://img.shields.io/crates/v/vsmprotocol)](https://crates.io/crates/vsmprotocol)
[![Platform](https://img.shields.io/badge/Platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey)](#)

VSM makes ports invisible. Not filtered, not closed — **nonexistent**.

A port scanner like `nmap` determines port state by sending a TCP SYN and reading the response: SYN-ACK means open, RST means closed, silence means filtered. VSM ensures that no scanner ever gets a SYN-ACK, because the server **never opens a socket**. There is no `listen()`, no `accept()`, no entry in the kernel socket table. Tools like `netstat`, `ss`, and `lsof` show nothing. The port does not exist as far as the operating system is concerned.

But authorized peers can still connect. Here's how.

---

## How It Works: ISN Steganography

The TCP header has a 32-bit field called the **Initial Sequence Number (ISN)**. Every SYN packet carries one. Operating systems normally fill it with a pseudo-random value. VSM replaces it with a **time-rotating HMAC signature** derived from the peer's private key.

### The Knock Calculation

```
1. Divide current Unix time by 30        → Time Window (uint64)
2. HMAC-SHA256(privateKey, timeWindow)    → 32-byte Hash
3. Take first 4 bytes of hash            → 32-bit Knock Value
4. Set tcp.Seq = Knock Value              → Inject raw SYN packet
```

The result: a standard-looking TCP SYN packet where the sequence number is actually a cryptographic proof of identity that expires every 30 seconds.

### The Handshake

```
             Scanner                    VSM Server
               │                            │
          SYN (random ISN) ──────────────►  │  ← HMAC check fails → DROPPED (no response)
               │                            │
               │                            │
             Peer                      VSM Server
               │                            │
          SYN (HMAC ISN) ────────────────►  │  ← HMAC check passes
               │                   SYN-ACK  │
               │  ◄──────────────────────── │  ← Raw injected, no kernel involvement
               │                            │
               │     [Noise XX Handshake]   │  ← Mutual auth + forward secrecy
               │     [Encrypted Tunnel]     │  ← ChaCha20-Poly1305
```

The server sits in a raw packet sniffing loop using `libpcap`. For every incoming SYN on the target port, it extracts the ISN and computes what the correct knock *should* be for the current 30-second window. If the values don't match, the packet is silently ignored. If they match, the server constructs a raw SYN-ACK and injects it back onto the wire — entirely outside the kernel's TCP stack.

### Why a Firewall Rule Is Required

Because the kernel has no socket bound to the target port, it will respond to the peer's SYN with a RST ("connection refused"). This RST would kill the stealth handshake before it starts. VSM requires a local firewall rule to silence these kernel-generated RSTs:

**macOS (pf):**
```bash
echo "block drop out proto tcp from any to any port 9999" | sudo pfctl -a "com.apple/vsm" -f - && sudo pfctl -e
```

**Linux (iptables):**
```bash
sudo iptables -A OUTPUT -p tcp --tcp-flags RST RST --sport 9999 -j DROP
```

With the RST silenced, the port becomes a **black hole** to unauthorized traffic and a functioning endpoint for authorized peers.

---

## Security Properties

| Property | Mechanism |
|---|---|
| **Port Invisibility** | No system socket. No kernel state. Nothing to scan. |
| **Replay Protection** | Knock rotates every 30 seconds via HMAC over time window. |
| **Kernel Bypass** | All packets are raw-injected via BPF. Invisible to `netstat`, `ss`, `lsof`. |
| **Forward Secrecy** | Post-knock encryption uses Noise XX (ephemeral key exchange). |
| **Mutual Authentication** | Both peers verify identity during the Noise handshake. |

---

## SDKs

The core is written in Go and compiled to a shared library (`vsmprotocol.dylib` / `.so` / `.dll`). Language wrappers load it via FFI:

| Language | Package | Install |
|---|---|---|
| **Node.js** | [npmjs.com/package/vsmprotocol](https://www.npmjs.com/package/vsmprotocol) | `npm install vsmprotocol` |
| **Python** | [pypi.org/project/vsmprotocol](https://pypi.org/project/vsmprotocol/) | `pip install vsmprotocol` |
| **Rust** | [crates.io/crates/vsmprotocol](https://crates.io/crates/vsmprotocol) | `cargo add vsmprotocol` |
| **Java/JVM** | [JitPack](https://jitpack.io/#cezarpena/vsm-protocol) | See Gradle config below |
| **C++** | Header-only | `#include "vsm_protocol.hpp"` |
| **C#** | P/Invoke wrapper | See `wrappers/csharp/` |

### Java/Gradle
```gradle
repositories { maven { url 'https://jitpack.io' } }
dependencies { implementation 'com.github.cezarpena:vsm-protocol:v0.1.5' }
```

---

## Building from Source

Requires **Go 1.21+** and a C compiler for CGO.

```bash
./build.sh current    # Build for current platform
./build.sh all        # Cross-compile all targets
```

Binaries land in `dist/{os}-{arch}/`.

---

## Full Technical Specification

See [ARCHITECTURE.md](ARCHITECTURE.md) for the complete protocol breakdown including packet construction, ISN calculation internals, and the Noise XX handshake sequence.

---

## License
MIT — See [LICENSE](LICENSE)

© 2026 Cezar Pena
