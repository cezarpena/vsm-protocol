# vsmprotocol (Python SDK)

## 🔗 [GitHub Repository: cezarpena/vsm-protocol](https://github.com/cezarpena/vsm-protocol)

**VSM Stealth Protocol** — Invisible P2P encrypted transport for Python.

Enable bidirectional encrypted communication without opening a single port. `vsmprotocol` uses raw packet injection to create stealth tunnels that bypass the standard OS socket table.

---

## 🚀 Quick Start

### 1. Installation
```bash
pip install vsmprotocol
```

*Note: Requires `sudo` (root) permissions to inject and sniff raw packets.*

### 2. Setup Firewalls (Mandatory)
Silence outgoing RST packets on your chosen port:

**macOS:**
```bash
echo "block drop out proto tcp from any to any port 9999" | sudo pfctl -a "com.apple/vsm" -f - && sudo pfctl -e
```

**Linux:**
```bash
sudo iptables -A OUTPUT -p tcp --tcp-flags RST RST -j DROP
```

---

## 💻 Usage

### Start a Stealth Server
```python
import vsmprotocol as vsm
import time

def on_knock(sid, peer):
    print(f" [SRV] Knock from {peer.decode()}")

def on_msg(sid, peer, msg):
    print(f" [SRV] Message: {msg.decode()}")

server_id = vsm.generate_identity()

# KEEP THE RETURN VALUE IN A VARIABLE to prevent Garbage Collection
refs = vsm.start_server(9999, server_id, on_knock, on_msg)

while True:
    time.sleep(1)
```

### Dial a Server
```python
import vsmprotocol as vsm

my_id = vsm.generate_identity()
sid = vsm.dial("lo0", "127.0.0.1", 9999, my_id)

if sid != -1:
    vsm.send_message(sid, "Hello from the ghost network.")
```

---

## 📜 License
MIT
