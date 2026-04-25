# vsmprotocol

## [GitHub: cezarpena/vsm-protocol](https://github.com/cezarpena/vsm-protocol)

[![Crates.io](https://img.shields.io/crates/v/vsmprotocol.svg)](https://crates.io/crates/vsmprotocol)

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

```toml
[dependencies]
vsmprotocol = "0.1"
```

Requires `sudo` — raw packet injection needs root.

On first `cargo build`, the correct precompiled binary for your platform is automatically downloaded from GitHub Releases.

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

```rust
use vsmprotocol::VSMProtocol;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let vsm = VSMProtocol::load("./vsmprotocol.dylib")?;

    let identity = vsm.generate_identity()?;

    // Dial a server on an invisible port
    let sid = vsm.dial("lo0", "127.0.0.1", 9999, &identity);

    if sid != -1 {
        vsm.send_message(sid, "Message through an invisible port.")?;
    }

    Ok(())
}
```

---

## What's Inside

- **Core**: Go shared library using `libpcap` for raw packet injection/sniffing
- **FFI**: `libloading` for safe dynamic library calls from Rust
- **Encryption**: Noise XX handshake → ChaCha20-Poly1305
- **Build**: `build.rs` auto-downloads the correct binary (macOS ARM/Intel, Linux, Windows)

---

## License
MIT · © 2026 Cezar Pena
