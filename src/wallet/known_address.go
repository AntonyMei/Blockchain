package wallet

import "crypto/ecdsa"

type KnownAddress struct {
	PublicKey ecdsa.PublicKey
	Address   []byte
}
