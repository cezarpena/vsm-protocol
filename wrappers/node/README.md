# vsmprotocol

## [GitHub: cezarpena/vsm-protocol](https://github.com/cezarpena/vsm-protocol)

[![npm version](https://badge.fury.io/js/vsmprotocol.svg)](https://www.npmjs.com/package/vsmprotocol)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Make a port invisible to every scanner on the internet — then talk through it anyway.

VSM does not open sockets. There is no `listen()`, no `accept()`, no entry in `netstat` or `lsof`. The port does not exist as far as the operating system is concerned. But authorized peers can still connect, because the server is sniffing raw packets with `libpcap` and looking for a specific cryptographic signature hidden inside the TCP header.

### The Trick: ISN Steganography

Every TCP connection starts with a SYN packet that carries a 32-bit **Initial Sequence Number**. Normally this is random noise. VSM replaces it with an **HMAC-SHA256 signature** derived from the peer's private key and the current 30-second time window.

The server extracts the ISN from every incoming SYN, computes what the correct value should be, and compares:
- **Mismatch** → packet is silently dropped. No response. Port appears dead.
- **Match** → server injects a raw SYN-ACK, completing a handshake entirely outside the kernel.

After the stealth handshake, a Noise XX key exchange provides mutual authentication and ChaCha20-Poly1305 encryption.

---

## Install

```bash
npm install vsmprotocol
```

Requires `sudo` — raw packet injection needs root.

---

## Firewall Setup (Required)

The kernel will send RST packets for traffic on ports it doesn't know about. You must silence them:

**macOS:**
```bash
echo "block drop out proto tcp from any to any port 9999" | sudo pfctl -a "com.apple/vsm" -f - && sudo pfctl -e
```

**Linux:**
```bash
sudo iptables -A OUTPUT -p tcp --tcp-flags RST RST --sport 9999 -j DROP
```

---

## Usage (ESM)

### Server
```javascript
import { startServer, generateIdentity } from 'vsmprotocol';

const identity = generateIdentity();

startServer(9999, identity,
  (sessionId, peerId) => {
    console.log(`Stealth connection from ${peerId}`);
  },
  (sessionId, peerId, message) => {
    console.log(`${peerId}: ${message}`);
  }
);
```

### Client
```javascript
import { dial, sendMessage } from 'vsmprotocol';

const identity = generateIdentity();
const sessionId = dial('lo0', '127.0.0.1', 9999, identity);

if (sessionId !== -1) {
  sendMessage(sessionId, "Message through an invisible port.");
}
```

---

## What's Inside

- **Core**: Go shared library using `libpcap` for raw packet injection/sniffing
- **FFI**: `koffi` for thread-safe native calls from Node.js
- **Encryption**: Noise XX handshake → ChaCha20-Poly1305
- **Bundled binaries**: macOS (ARM + Intel), Linux (AMD64), Windows (AMD64)

---

## License
MIT · © 2026 Cezar Pena
