package metrics

import (
	"encoding/binary"
	"fmt"
	"go.uber.org/zap"
	"hash/crc32"
	"net"
	"sync"
)

func init() {
	// 初始化zap logger
	var err error
	gvaLog, err = zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("初始化logger失败: %v", err))
	}
}

const (
	MagicNumber     uint16 = 0xCAFE
	ProtocolVersion byte   = 0x01
	HeaderSize             = 12
	CRCSize                = 4
	MinPacketSize          = HeaderSize + CRCSize
)

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

var packetBuffers = make(map[byte]*PacketBuffer)
var packetBuffersMutex sync.Mutex

// ListenSerialConnection 监听物理串口连接通道数据（.fa00）
func ListenSerialConnection(conn net.Conn, kvmID string, log *zap.Logger, processCompleteTaskData ProcessCompleteTaskDataFunc) {
	if log == nil {
		// 如果传入的logger为nil，使用默认logger
		log = gvaLog
	}
	defer conn.Close()

	var receivedBuf []byte
	for {
		buf := make([]byte, BufSize)
		n, err := conn.Read(buf)
		if err != nil {
			log.Error("Error reading from socket:", zap.Error(err))
			log.Info(fmt.Sprintf("agent[%s] 连接断开", kvmID))
			// 清理所有未完成的数据包
			packetBuffersMutex.Lock()
			packetBuffers = make(map[byte]*PacketBuffer)
			packetBuffersMutex.Unlock()
			return
		}

		receivedBuf = append(receivedBuf, buf[:n]...)

		// 处理所有完整的数据包
		for len(receivedBuf) >= MinPacketSize {
			// 查找Magic Number
			magicIndex := -1
			for i := 0; i <= len(receivedBuf)-2; i++ {
				if binary.BigEndian.Uint16(receivedBuf[i:i+2]) == MagicNumber {
					magicIndex = i
					break
				}
			}

			if magicIndex == -1 || len(receivedBuf[magicIndex:]) < MinPacketSize {
				break
			}

			// 解析header
			header := PacketHeader{
				MagicNumber: binary.BigEndian.Uint16(receivedBuf[magicIndex:]),
				Version:     receivedBuf[magicIndex+2],
				SeqNum:      binary.BigEndian.Uint32(receivedBuf[magicIndex+3:]),
				PacketType:  receivedBuf[magicIndex+7],
				Status:      receivedBuf[magicIndex+8],
				TaskID:      receivedBuf[magicIndex+9],
				DataLen:     binary.BigEndian.Uint16(receivedBuf[magicIndex+10:]),
			}

			// 验证版本
			if header.Version != ProtocolVersion {
				receivedBuf = receivedBuf[magicIndex+1:]
				continue
			}

			totalLen := HeaderSize + int(header.DataLen) + CRCSize
			if len(receivedBuf[magicIndex:]) < totalLen {
				break
			}

			packet := receivedBuf[magicIndex : magicIndex+totalLen]

			// 验证CRC
			expectedCRC := binary.BigEndian.Uint32(packet[totalLen-CRCSize:])
			actualCRC := crc32.ChecksumIEEE(packet[:totalLen-CRCSize])
			if expectedCRC != actualCRC {
				receivedBuf = receivedBuf[magicIndex+1:]
				continue
			}

			// 处理数据包
			packetBuffersMutex.Lock()
			switch int(header.Status) {
			case DataStart:
				packetBuffers[header.TaskID] = &PacketBuffer{
					Data:    packet[HeaderSize : totalLen-CRCSize],
					LastSeq: header.SeqNum,
				}
			case DataTransfer:
				if buf, exists := packetBuffers[header.TaskID]; exists && header.SeqNum == buf.LastSeq+1 {
					buf.Data = append(buf.Data, packet[HeaderSize:totalLen-CRCSize]...)
					buf.LastSeq = header.SeqNum
				}
			case DataEnd:
				if _, exists := packetBuffers[header.TaskID]; !exists {
					// 单片数据，直接处理
					if err := processCompleteTaskData(int(header.PacketType), packet[HeaderSize:totalLen-CRCSize], kvmID); err != nil {
						log.Error("处理数据失败:", zap.Error(err))
					}
				} else {
					// 多片数据的最后一片
					if buf, exists := packetBuffers[header.TaskID]; exists && header.SeqNum == buf.LastSeq+1 {
						buf.Data = append(buf.Data, packet[HeaderSize:totalLen-CRCSize]...)
						buf.Complete = true

						// 处理完整数据
						if err := processCompleteTaskData(int(header.PacketType), buf.Data, kvmID); err != nil {
							log.Error("处理数据失败:", zap.Error(err))
						}

						// 清理缓存
						delete(packetBuffers, header.TaskID)
					}
				}

			}
			packetBuffersMutex.Unlock()

			// 移除已处理的数据
			receivedBuf = receivedBuf[magicIndex+totalLen:]
		}
	}
}
