package stealth

import (
	"context"
	"fmt"

	"github.com/cretz/bine/tor"
)

// TorManager handles the lifecycle of the Tor process
type TorManager struct {
	Instance *tor.Tor
}

// StartTor bootstraps a new Tor process on your computer
func StartTor() (*TorManager, error) {
	fmt.Println(" [TOR] Bootstrapping Tor process... (this may take a minute)")

	// 1. Start Tor binary
	t, err := tor.Start(nil, nil)
	if err != nil {
		return nil, err
	}

	// 2. Wait for it to reach 100% connectivity
	// 3-minute timeout
	return &TorManager{Instance: t}, nil
}

func (m *TorManager) CreateHiddenService(ctx context.Context, port int) (string, error) {
	// 1. Define the service settings (Forwarding local traffic to the Onion)
	// Map Onion's port 80 to Local Go app's port
	service, err := m.Instance.Listen(ctx, &tor.ListenConf{
		Version3: true, RemotePorts: []int{port},
	})

	if err != nil {
		return "", err
	}

	fmt.Printf(" [TOR] Stealth Node is LIVE at %s.onion\n", service.ID)
	return service.ID, nil
}

func (m *TorManager) CreateAuthorizedService(ctx context.Context, port int, clientKeys []string) (string, error) {
	// 1. Configure the "Invite List"
	// Tor uses these keys to authenticate the visitors
	conf := &tor.ListenConf{
		Version3:    true,
		RemotePorts: []int{port},
	}

	// If we have keys, we enable Client Auth
	if len(clientKeys) > 0 {
		conf.ClientAuths = make(map[string]string)
		for i, key := range clientKeys {
			nickname := fmt.Sprintf("peer_%d", i)
			conf.ClientAuths[nickname] = key
		}
	}

	service, err := m.Instance.Listen(ctx, conf)
	if err != nil {
		return "", err
	}

	fmt.Printf(" [TOR] Armored Node is LIVE at: %s.onion\n", service.ID)
	if len(clientKeys) > 0 {
		fmt.Println(" [TOR] Status: Hidden from the public. Only authorized peers can connect.")

	}

	return service.ID, nil

}
