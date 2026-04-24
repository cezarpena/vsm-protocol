package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"os"
	"vsm-protocol/core/stealth"
	"vsm-protocol/core/stealth/transport"
)

func main() {
	// 1. Load the same identity used for the server
	data, _ := os.ReadFile("my_stealth_id.json")
	var id stealth.VsmIdentity
	json.Unmarshal(data, &id)

	// 2. Decode the Private Key
	privKeyBytes, _ := base64.RawURLEncoding.DecodeString(id.PrivKey)
	privKey := ed25519.PrivateKey(privKeyBytes)

	// 3. Dial
	transport.StealthDial("lo0", "127.0.0.1", 9999, privKey)
}
