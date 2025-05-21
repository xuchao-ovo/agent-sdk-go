package serial

import (
	"encoding/binary"
	"errors"
	"github.com/xuchao-ovo/agent-sdk-go/pkg/utils"
	"hash/crc32"
	"log"
	"net"
	"sync"
	"time"
)

var mu sync.Mutex
var seqNum uint32 // 新增序列号计数器

type PacketHeader struct {
	MagicNumber uint16
	Version     byte
	SeqNum      uint32
	PacketType  byte
	Status      byte
	TaskID      byte
	DataLen     uint16
}

func Write(con net.Conn, writeType int, byteData []byte, pool utils.TaskIDPool) error {
	mu.Lock()
	defer mu.Unlock()

	const dataSize = 300
	dataLen := len(byteData)

	taskID, err := pool.GetTaskID()
	if err != nil {
		return err
	}

	for offset := 0; offset < dataLen; offset += dataSize {
		// 准备数据块
		end := offset + dataSize
		if end > dataLen {
			end = dataLen
		}
		chunk := byteData[offset:end]

		// 确定数据状态
		var status byte
		if offset == 0 && dataLen > dataSize {
			status = 0 // DataStart
		} else if offset+dataSize >= dataLen {
			status = 2 // DataEnd
		} else {
			status = 1 // DataTransfer
		}

		// 创建 header
		header := PacketHeader{
			MagicNumber: MagicNumber,
			Version:     ProtocolVersion,
			SeqNum:      seqNum,
			PacketType:  byte(writeType),
			Status:      status,
			TaskID:      byte(taskID),
			DataLen:     uint16(len(chunk)),
		}

		// 准备完整数据包（header 12 字节 + 数据 + CRC32 4 字节）
		packet := make([]byte, 12+len(chunk)+4)

		// 写入 header
		binary.BigEndian.PutUint16(packet[0:], header.MagicNumber)
		packet[2] = header.Version
		binary.BigEndian.PutUint32(packet[3:], header.SeqNum)
		packet[7] = header.PacketType
		packet[8] = header.Status
		packet[9] = header.TaskID
		binary.BigEndian.PutUint16(packet[10:], header.DataLen)

		// 写入数据
		copy(packet[12:], chunk)

		// 写入 CRC32 校验码
		crc := crc32.ChecksumIEEE(packet[:12+len(chunk)])
		binary.BigEndian.PutUint32(packet[12+len(chunk):], crc)

		// 写入串口
		_, err = con.Write(packet)
		if err != nil {
			log.Println("Error writing to serial port, taskID:", taskID, "error:", err)
			pool.RecycleTaskID(taskID)
			return errors.New("写入数据失败")
		}
		time.Sleep(20 * time.Millisecond)

		seqNum++ // 增加序列号
	}

	pool.RecycleTaskID(taskID)
	return nil
}
