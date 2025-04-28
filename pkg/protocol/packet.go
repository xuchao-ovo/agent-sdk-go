package protocol

// PacketHeader 数据包头部
type PacketHeader struct {
	MagicNumber uint16
	Version     byte
	SeqNum      uint32
	PacketType  byte
	Status      byte
	TaskID      byte
	DataLen     uint16
}

type PacketBuffer struct {
	Data     []byte
	LastSeq  uint32
	Complete bool
}

// DataPacket 数据包
type DataPacket struct {
	Header PacketHeader
	Data   []byte
	CRC32  uint32
}

// 数据状态
var (
	DataStart    = 0
	DataTransfer = 1
	DataEnd      = 2
)
