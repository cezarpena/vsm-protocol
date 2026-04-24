package stealth

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"

	"github.com/flynn/noise"
)

// NoiseSession handles the encryption state for a stealth connection
type NoiseSession struct {
	config  noise.Config
	state   *noise.HandshakeState
	encrypt *noise.CipherState
	decrypt *noise.CipherState
}

// InitializeNoise sets up the handshake logic
func InitializeNoise(isInitiator bool, privKey ed25519.PrivateKey) (*NoiseSession, error) {
	// 1. Pick "XX" handshake pattern.

	suite := noise.NewCipherSuite(noise.DH25519, noise.CipherChaChaPoly, noise.HashSHA256)

	// 2. Generate a static Curve25519 keypair for the Noise handshake
	staticKey, err := suite.GenerateKeypair(rand.Reader)
	if err != nil {
		return nil, err
	}

	// 3. Setup Configuration
	config := noise.Config{
		CipherSuite:   suite,
		Pattern:       noise.HandshakeXX,
		Initiator:     isInitiator,
		StaticKeypair: staticKey,
	}

	// 4. Create the handshake state
	state, err := noise.NewHandshakeState(config)
	if err != nil {
		return nil, err
	}

	return &NoiseSession{
		config: config,
		state:  state,
	}, nil
}

func (s *NoiseSession) HandshakeStep(input []byte, outgoing bool) ([]byte, error) {
	if outgoing {
		out, cs1, cs2, err := s.state.WriteMessage(nil, nil)

		if cs1 != nil {
			s.assignCipherStates(cs1, cs2)
			fmt.Println(" [NOISE] Handshake complete! Secure tunnel established.")
		}

		return out, err
	} else {
		_, cs1, cs2, err := s.state.ReadMessage(nil, input)

		if cs1 != nil {
			s.assignCipherStates(cs1, cs2)
			fmt.Println(" [NOISE] Handshake complete! Secure tunnel established")
		}

		return nil, err
	}
}

// assignCipherStates handles the Noise spec ordering:
// cs1 always encrypts initiator→responder, cs2 encrypts responder→initiator.
// So the initiator encrypts with cs1, but the responder encrypts with cs2.
func (s *NoiseSession) assignCipherStates(cs1, cs2 *noise.CipherState) {
	if s.config.Initiator {
		s.encrypt = cs1
		s.decrypt = cs2
	} else {
		s.encrypt = cs2
		s.decrypt = cs1
	}
}

func (s *NoiseSession) Encrypt(plaintext []byte) ([]byte, error) {
	if s.encrypt == nil {
		return nil, fmt.Errorf("Handshake not complete")
	}

	// Encrypt(out, ad, plaintext)
	// ad is "Additional Data"
	return s.encrypt.Encrypt(nil, nil, plaintext)
}

func (s *NoiseSession) Decrypt(ciphertext []byte) ([]byte, error) {
	if s.decrypt == nil {
		return nil, fmt.Errorf("Handshake not complete")
	}

	return s.decrypt.Decrypt(nil, nil, ciphertext)
}
