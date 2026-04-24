package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"vsm-protocol/core/stealth"
	"vsm-protocol/core/stealth/transport"
)

func main() {
	data, err := os.ReadFile("my_stealth_id.json")
	if err != nil {
		fmt.Println("Error reading identity file:", err)
		return
	}

	var id stealth.VsmIdentity
	if err := json.Unmarshal(data, &id); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	privKeyBytes, err := base64.RawURLEncoding.DecodeString(id.PrivKey)
	if err != nil {
		fmt.Println("Error decoding key:", err)
		return
	}

	privKey := ed25519.PrivateKey(privKeyBytes)

	transport.ListenForKnock("lo0", 9999, privKey)
}
