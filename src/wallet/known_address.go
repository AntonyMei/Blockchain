package wallet

import "crypto/ecdsa"

type KnownAddress struct {
	publicKey ecdsa.PublicKey
	address   []byte
}
