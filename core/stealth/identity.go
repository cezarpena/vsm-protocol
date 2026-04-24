package stealth

import (
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"os"
	"time"
)

type VsmIdentity struct {
	ID      string `json:"id"`
	PrivKey string `json:"privKey"`
	PubKey  string `json:"pubKey"`
}

func GenerateKey() (ed25519.PublicKey, ed25519.PrivateKey, error) {

	// rand.Reader is a cryptographically secure random number generator
	// provided by the OS

	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)

	if err != nil {
		return nil, nil, err
	}

	return pubKey, privKey, nil
}

func EncodeKey(rawKey []byte) string {
	return base64.RawURLEncoding.EncodeToString(rawKey)
}

func GenerateNewIdentity() (VsmIdentity, error) {
	pub, priv, err := GenerateKey()
	if err != nil {
		return VsmIdentity{}, err
	}

	pubStr := EncodeKey(pub)

	return VsmIdentity{
		ID:      pubStr,
		PubKey:  pubStr,
		PrivKey: EncodeKey(priv),
	}, nil
}

func SaveIdentity(path string, id VsmIdentity) error {
	vsm, err := json.MarshalIndent(id, "", "	")
	if err != nil {
		return err
	}

	return os.WriteFile(path, vsm, 0600)
}

// GenerateKnock creates a 32-bit secret signature that changes every 30 seconds
func GenerateKnock(privKey ed25519.PrivateKey) uint32 {
	// 1. Get the current time window (30 second blocks)
	window := time.Now().Unix() / 30
	windowBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(windowBytes, uint64(window))

	// 2. Hash the time window using private key as secret
	mac := hmac.New(sha256.New, privKey)
	mac.Write(windowBytes)
	hash := mac.Sum(nil)

	// 3. Take first 4 bytes and turn to uint32
	return binary.BigEndian.Uint32(hash[:4])
}

func VerifyKnock(privKey ed25519.PrivateKey, receivedKnock uint32) bool {
	// 1. Calculate what the knock SHOULD be right now
	expected := GenerateKnock(privKey)

	// 2. Do they match?
	if receivedKnock == expected {
		return true
	}

	return false
}

func DecodeKey(encoded string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(encoded)
}
