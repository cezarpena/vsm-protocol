package main

import (
	"fmt"
	"vsm-protocol/core/stealth"
)

func main() {
	// 1. Generate the raw keys
	pub, priv, err := stealth.GenerateKey()
	if err != nil {
		panic(err)
	}
	// 2. Wrap them into our identity struct
	id := stealth.VsmIdentity{
		ID:      stealth.EncodeKey(pub), // Your PeerID is your encoded PubKey
		PubKey:  stealth.EncodeKey(pub),
		PrivKey: stealth.EncodeKey(priv),
	}
	// 3. Save it!
	err = stealth.SaveIdentity("my_stealth_id.json", id)
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully generated Stealth Identity!")
	fmt.Println("PeerID:", id.ID)
}
