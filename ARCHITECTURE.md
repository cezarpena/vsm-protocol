# VSM Stealth Architecture: Invisibility via ISN Injection

The primary goal of VSM is to provide a communication channel that is **digitally non-existent** to unauthorized observers. It achieves this by bypassing the standard OS Network Stack and using cryptographic values hidden in the standard TCP handshake fields.

## 1. The Core Objective: Total Port Invisibility

A standard port scanner (like `nmap` or `masscan`) works by sending a TCP `SYN` packet to a port and observing the response:
- **SYN-ACK**: Port is Open.
- **RST**: Port is Closed.
- **No Response**: Port is Filtered (Firewalled).

VSM makes a port appear **Closed or Filtered** to all standard probes. It does this by:
1. **Not opening a system socket**: The server does NOT use `listen()` or `accept()`. Netstat/lsof will show nothing.
2. **Raw Sniffing**: Using `libpcap` to "sniff" every packet entering the interface at the driver level.
3. **Silent Drop**: If a packet doesn't contain the correct cryptographic "knock", the server logic ignores it completely. The OS kernel, seeing no application listening on that port, would normally send an `RST` packet, but VSM requires the user to silence these via a firewall rule (iptables/pf), creating a "black hole".

## 2. The Cryptographic Knock (ISN Steganography)

The "Knock" is not a payload; it is the **Initial Sequence Number (ISN)** of the first TCP SYN packet. VSM repurposes the `Sequence Number` field (32 bits) as a time-rotating cryptographic passkey.

### The Calculation
The knock is calculated using a shared Ed25519 Private Key and the current system time:

1. **Time Windowing**: The current Unix timestamp is divided into 30-second blocks. This prevents "replay attacks" where a captured knock remains valid indefinitely.
   ```go
   window := time.Now().Unix() / 30
   ```
2. **HMAC Signing**: The time window is signed using the peer's private key via HMAC-SHA256.
   ```go
   mac := hmac.New(sha256.New, privKey)
   mac.Write(windowBytes)
   hash := mac.Sum(nil)
   ```
3. **ISN Extraction**: The first 4 bytes (32 bits) of the resulting hash are used as the TCP Sequence Number.
   ```go
   knock := binary.BigEndian.Uint32(hash[:4])
   ```

## 3. The Stealth Handshake Flow

Unlike a standard TCP handshake, VSM injects raw packets directly into the wire using BPF (Berkeley Packet Filter) bytecodes.

### Step A: The Initiator (Dialer)
Instead of a standard `connect()`, the dialer:
1. Calculates the **Knock** for the current 30s window.
2. Constructs a raw IPv4/TCP SYN packet.
3. Sets `tcp.Seq = Knock`.
4. Injects the packet onto the wire.

### Step B: The Responder (Sniffer)
The server sits in a tight loop watching raw packets:
1. It extracts the `Seq` from every incoming `SYN` packet.
2. It calculates the **Expected Knock** for the current (and sometimes previous) 30s window.
3. **Comparison**:
   - If `incoming.Seq != expected_knock`: The loop `continues`. No `SYN-ACK` is sent.
   - If `incoming.Seq == expected_knock`: The server manually constructs and injects a `SYN-ACK` back to the peer.

## 4. Secure Tunneling (Secondary)

Only after the secret knock has been verified does the protocol move to the **Noise XX** handshake. This provides:
- **Mutual Authentication**: Both sides verify identities.
- **Perfect Forward Secrecy**: Even if keys are stolen later, past traffic cannot be decrypted.
- **Encryption**: All application data is wrapped in ChaCha20-Poly1305.

## 5. Security Implications

- **Anti-Scanning**: Because the "listening" port never responds to standard SYN packets, it is impossible to discover the service via horizontal network scanning.
- **Replay Protection**: Because the knock changes every 30 seconds based on a hash of the time, an attacker who captures a valid SYN packet only has a few seconds to use it before it becomes invalid.
- **Kernel Bypass**: Traffic is invisible to standard diagnostic tools that rely on the kernel's socket table (`netstat`, `ss`, `lsof`).
