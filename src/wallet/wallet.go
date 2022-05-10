package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"github.com/AntonyMei/Blockchain/config"
	"github.com/AntonyMei/Blockchain/src/utils"
	"golang.org/x/crypto/ripemd160"
)

type Wallet struct {
	// PrivateKey is generated using Elliptic Curve Digital Signature Algorithm
	// PublicKeyHash goes through a more complicated process, similar to Bitcoin, see also
	// https://dev.to/nheindev/building-a-blockchain-in-go-pt-v-wallets-12na
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

func GenerateKeyPair() (ecdsa.PrivateKey, []byte) {
	// Generate KeyPair based on Elliptic Curve Digital Signature Algorithm
	curve := elliptic.P256()
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	utils.Handle(err)
	// public key, ... means append after
	publicKey := append(privateKey.PublicKey.X.Bytes(), privateKey.PublicKey.Y.Bytes()...)
	return *privateKey, publicKey
}

func PublicKeyHash(publicKey []byte) []byte {
	// Step1: public key is privateKey.PublicKey.X + privateKey.PublicKey.Y
	// Step2: sha256 hash -> ripemd160 hash
	publicKeyAfterSHA256 := sha256.Sum256(publicKey)
	ripemdHasher := ripemd160.New()
	_, err := ripemdHasher.Write(publicKeyAfterSHA256[:])
	utils.Handle(err)
	publicKeyAfterRipemd := ripemdHasher.Sum(nil)
	return publicKeyAfterRipemd
}

func Checksum(ripemdHash []byte) []byte {
	firstHash := sha256.Sum256(ripemdHash)
	secondHash := sha256.Sum256(firstHash[:])
	return secondHash[:config.ChecksumLength]
}

func CreateWallet() *Wallet {
	privateKey, publicKey := GenerateKeyPair()
	newWallet := Wallet{privateKey, publicKey}
	return &newWallet
}

func (w *Wallet) Address() []byte {
	pubHash := PublicKeyHash(w.PublicKey)
	versionedHash := append([]byte{config.WalletVersion}, pubHash...)
	checksum := Checksum(versionedHash)
	finalHash := append(versionedHash, checksum...)
	address := utils.Base58Encode(finalHash)
	return address
}
