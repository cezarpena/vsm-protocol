# vsmprotocol (Rust SDK)

**VSM Stealth Protocol** — High-performance, invisible P2P encrypted transport for Rust.

[![Crates.io](https://img.shields.io/crates/v/vsmprotocol.svg)](https://crates.io/crates/vsmprotocol)
[![Documentation](https://docs.rs/vsmprotocol/badge.svg)](https://docs.rs/vsmprotocol)

Enable bidirectional encrypted communication without opening a single port. `vsmprotocol` uses raw packet injection to create stealth tunnels that bypass the standard OS socket table.

For the full protocol documentation and other language SDKs, visit the [Main Repository](https://github.com/cezarpena/vsm-protocol).

---

## 🚀 Quick Start

### 1. Installation
Add this to your `Cargo.toml`:
```toml
[dependencies]
vsmprotocol = "0.1"
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

### Create a Stealth Session
```rust
use vsmprotocol::VSMProtocol;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let vsm = VSMProtocol::load("./vsmprotocol.dylib")?;
    
    // Generate a fresh identity
    let id = vsm.generate_identity()?;
    println!("My Identity: {}", id);

    // Dial a remote stealth server
    let sid = vsm.dial("lo0", "127.0.0.1", 9999, &id);
    
    if sid != -1 {
        vsm.send_message(sid, "Hello from Rust!")?;
    }

    Ok(())
}
```

---

## ⚙️ How it works
This crate is a lightweight wrapper around the Go-compiled `vsmprotocol` shared library.
- On first build, the `build.rs` script automatically downloads the correct precompiled binary for your architecture (macOS ARM/Intel, Linux, Windows) from the GitHub releases.
- It links dynamically to provide a safe, idiomatic Rust API.

---

## 📜 License
MIT

---
*Main Repository: [https://github.com/cezarpena/vsm-protocol](https://github.com/cezarpena/vsm-protocol)*
