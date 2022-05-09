package config

const (
	// InitialChainDifficulty is equal to four times the number of zeros at hash value head.
	InitialChainDifficulty = 16
	// MiningReward is the number of coins given to each block
	MiningReward = 100

	// PersistentStoragePath is where we store the chain on disk
	PersistentStoragePath = "./tmp/blocks"

	// GenesisData is contained in Data field of genesis block
	GenesisData = "Genesis"
	// CoinbaseSig is signature of coinbase transactions
	CoinbaseSig = "Coinbase Signature"
)