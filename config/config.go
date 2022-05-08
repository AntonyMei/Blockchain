package config

// InitialChainDifficulty is equal to four times the number of zeros at hash value head.
const InitialChainDifficulty = 16

// MiningReward is the number of coins given to each block
const MiningReward = 100

// PersistentStoragePath is where we store the chain on disk
const PersistentStoragePath = "./tmp/blocks"
