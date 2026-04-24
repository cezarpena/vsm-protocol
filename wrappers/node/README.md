# vsmprotocol

**VSM Stealth Protocol** — A high-performance, invisible P2P encrypted transport for Node.js.

[![npm version](https://badge.fury.io/js/vsmprotocol.svg)](https://www.npmjs.com/package/vsmprotocol)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Enable bidirectional encrypted communication without opening a single port. `vsmprotocol` uses raw packet injection to create stealth tunnels that bypass the standard OS socket table, making your traffic invisible to `netstat`, `lsof`, and port scanners.

---

## 🚀 Quick Start

### 1. Installation
```bash
npm install vsmprotocol
```

*Note: Requires `sudo` (root) permissions to inject and sniff raw packets at the network interface level.*

### 2. Setup Firewalls (Mandatory)
Because `vsmprotocol` operates outside the kernel's knowledge, the OS will try to reset "unknown" incoming traffic. You must tell your firewall to silence RST packets on your chosen port.

**macOS:**
```bash
echo "block drop out proto tcp from any to any port 9999" | sudo pfctl -a "com.apple/vsm" -f - && sudo pfctl -e
```

**Linux:**
```bash
sudo iptables -A OUTPUT -p tcp --tcp-flags RST RST -j DROP
```

---

## 💻 Usage (ESM)

### Start a Stealth Server
```javascript
import { startServer, generateIdentity } from 'vsmprotocol';

const identity = generateIdentity(); // Or load one from JSON

startServer(9999, identity, 
  (sessionId, peerId) => {
    console.log(`[SRV] Stealth connection from ${peerId}`);
  }, 
  (sessionId, peerId, message) => {
    console.log(`[MSG] ${peerId}: ${message}`);
  }
);
```

### Dial a Server
```javascript
import { dial, sendMessage } from 'vsmprotocol';

const identity = { ... }; // Your identity
const sessionId = dial('lo0', '127.0.0.1', 9999, identity);

if (sessionId !== -1) {
  sendMessage(sessionId, "Hello from the ghost network.");
}
```

---

## 🛡️ Under the Hood

- **Core**: Written in Go using `libpcap` for raw packet control.
- **Handshake**: Noise XX (Mutual Authentication + Forward Secrecy).
- **Encryption**: ChaCha20-Poly1305.
- **FFI**: Powered by `koffi` for high-performance, thread-safe library calls.
- **Bundled**: Includes precompiled binaries for macOS (ARM/Intel), Linux (AMD64), and Windows (AMD64).

---

## 📜 License
MIT

---
*Powered by the VSM-Cell Project.*
