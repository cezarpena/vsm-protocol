# 🌌 VSM Stealth Protocol

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/cezarpena/vsm-protocol)](https://goreportcard.com/report/github.com/cezarpena/vsm-protocol)
[![Platform](https://img.shields.io/badge/Platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey)](#)

**VSM Stealth Protocol** is a high-performance, cross-language P2P transport library designed for network invisibility. It enables fully encrypted, bidirectional communication by injecting raw TCP packets directly into the network interface, bypassing the standard OS socket table.

> [!IMPORTANT]
> This protocol does not open any listening ports. It uses a "Silent Knock" mechanism to wake up listeners through raw packet sniffing, making it invisible to `netstat`, `lsof`, and standard port scanners.

---

## 🚀 Key Features

*   **Invisible Transport**: Zero open ports. No entry in the kernel's socket table.
*   **Cryptographic Knocks**: High-entropy HMAC knocks embedded in TCP sequence numbers.
*   **Noise XX Handshake**: Full mutual authentication and forward secrecy using the Noise Protocol Framework (ChaCha20-Poly1305).
*   **Multi-Language SDKs**: Native wrappers for **Python, Node.js, C++, Java, C#, and Rust**.
*   **Zero-Copy Core**: Built in Go for memory safety and concurrency, compiled to a single shared library.

---

## 🛠 Architecture

The protocol operates outside the traditional TCP/IP stack behavior:

1.  **The Knock**: The client injects a raw `SYN` packet with a calculated ISN (Initial Sequence Number) derived from an HMAC of the current timestamp and a shared secret.
2.  **The Sniffer**: The server uses `libpcap` (or BPF) to monitor raw traffic. When it detects a valid knock, it responds with a raw `SYN-ACK`.
3.  **Experimental Tunnel**: Once handshaking completes, all data is exchanged via raw `PSH/ACK` packets. These packets are "stolen" from the kernel or ignored by it using local firewall rules (iptables/pf).

---

## 📦 Installation & SDKs

### Precompiled Binaries
The core logic is compiled into `vsmprotocol.dylib` (macOS), `vsmprotocol.so` (Linux), or `vsmprotocol.dll` (Windows). These are located in the `dist/` directory.

### Python
```bash
# In your project
from vsmprotocol import start_server, dial, send_message
```

### Node.js (ESM)
```javascript
import { startServer, dial, sendMessage } from './wrappers/node/index.js';
```

### C++ (Header-only)
```cpp
#include "wrappers/cpp/vsm_protocol.hpp"
vsm::Protocol vsm("./dist/vsmprotocol.dylib");
```

---

## 🛡️ Firewall Configuration (CRITICAL)

Because the OS kernel does not know about the stealth connection, it will attempt to send a `RST` (Reset) packet when it sees an unexpected TCP packet. You **must** silence the kernel for the specific port you are using.

### macOS (pf)
Add this to your `/etc/pf.conf`:
```text
block drop out proto tcp from any to any port 9999
```
Then reload: `sudo pfctl -f /etc/pf.conf && sudo pfctl -e`

### Linux (iptables)
```bash
sudo iptables -A OUTPUT -p tcp --tcp-flags RST RST -j DROP
```

---

## 🔧 Building from Source

Requires **Go 1.21+** and a C compiler (gcc/clang) for CGO.

```bash
# Build for current platform
./build.sh current

# Cross-compile macOS targets
./build.sh all
```

---

## 📜 License
MIT License - See [LICENSE](LICENSE) for details.

---
