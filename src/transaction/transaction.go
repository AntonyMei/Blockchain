package transaction

type Tx struct {
	// Value: #coins used
	// PubKey: TODO: finish this
	Value  int
	PubKey string
}

type SourcePointer struct {
	// SourcePacketID: ID of source TxPacket
	// TxIdx: index of source Tx in source TxPacket
	// Sig: TODO: finish this
	SourcePacketID []byte
	TxIdx          int
	Sig            string
}

type TxPacket struct {
	PacketID   []byte
	SourceList []SourcePointer
	TxList     []Tx
}
