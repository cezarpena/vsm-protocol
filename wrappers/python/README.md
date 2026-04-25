# vsmprotocol

## [GitHub: cezarpena/vsm-protocol](https://github.com/cezarpena/vsm-protocol)

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
pip install vsmprotocol
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

## Usage

### Server
```python
import vsmprotocol as vsm
import time

def on_knock(sid, peer):
    print(f"Stealth connection from {peer.decode()}")

def on_msg(sid, peer, msg):
    print(f"{peer.decode()}: {msg.decode()}")

identity = vsm.generate_identity()

# Store the return value — prevents garbage collection of C callbacks
refs = vsm.start_server(9999, identity, on_knock, on_msg)

while True:
    time.sleep(1)
```

### Client
```python
import vsmprotocol as vsm

identity = vsm.generate_identity()
sid = vsm.dial("lo0", "127.0.0.1", 9999, identity)

if sid != -1:
    vsm.send_message(sid, "Message through an invisible port.")
```

---

## What's Inside

- **Core**: Go shared library using `libpcap` for raw packet injection/sniffing
- **FFI**: `ctypes` for zero-dependency native calls from Python
- **Encryption**: Noise XX handshake → ChaCha20-Poly1305
- **Bundled binaries**: macOS (ARM + Intel), Linux (AMD64), Windows (AMD64)

---

## License
MIT · © 2026 Cezar Pena
