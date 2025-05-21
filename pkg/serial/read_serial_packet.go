package serial

import (
	"encoding/binary"
	"fmt"
	"github.com/xuchao-ovo/agent-sdk-go/global"
	"github.com/xuchao-ovo/agent-sdk-go/pkg/protocol"
	"go.uber.org/zap"
	"hash/crc32"
	"net"
	"sync"
)

const BufSize = 316

// 用于缓存每个业务的数据包
var taskDataMap = make(map[int][]byte)
var taskDataMapMutex sync.Mutex

// 用于缓存每个采集任务的数据包
var collectTaskDataMap = make(map[int][]byte)
var collectTaskDataMapMutex sync.Mutex

// 用于缓存物理串口的的数据包
var packetBuffers = make(map[byte]*protocol.PacketBuffer)
var packetBuffersMutex sync.Mutex

// 定义日志
var gvaLog *zap.Logger

const (
	MagicNumber     uint16 = 0xCAFE
	ProtocolVersion byte   = 0x01
	HeaderSize             = 12
	CRCSize                = 4
	MinPacketSize          = HeaderSize + CRCSize
)

// ProcessCompleteDataFunc 定义外部传入的 ProcessCompleteDataFunc 函数签名
type ProcessCompleteDataFunc func(packetType int, data []byte, kvmID string) error

// ListenConnection 监听数据采集连接通道数据（.fa2）
func ListenConnection(conn net.Conn, kvmID string, agentMap map[string]interface{}, agentMapMutex sync.Mutex, log *zap.Logger, processCompleteTaskData ProcessCompleteDataFunc) {
	defer conn.Close()
	gvaLog = log

	var receivedBuf []byte
	for {
		buf := make([]byte, BufSize)
		n, err := conn.Read(buf)
		if err != nil {
			gvaLog.Error("Error reading from socket:", zap.Error(err))
			gvaLog.Info(fmt.Sprintf("agent[%s].fa2 连接断开", kvmID))
			agentMapMutex.Lock()
			delete(agentMap, kvmID)
			agentMapMutex.Unlock()
			return
		}

		// 过滤无效数据包
		if int(buf[0]) != global.TaskCollect && int(buf[0]) != global.MetricCollect {
			continue
		}

		receivedBuf = append(receivedBuf, buf[:n]...)
		// 兼容旧版本心跳采集数据
		for len(receivedBuf) >= 3 && len(receivedBuf) < BufSize {
			dataStatus := int(receivedBuf[2])
			taskID := int(receivedBuf[1])
			// 过滤无效数据包
			if isValidPacket(taskID, dataStatus) {
				continue
			}
			packetLen := int(receivedBuf[1])<<8 | int(receivedBuf[2])
			totalLen := 3 + packetLen

			if len(receivedBuf) < totalLen {
				break // 数据不完整，继续等待接收
			}
			packetType := int(receivedBuf[0])
			data := receivedBuf[3:totalLen]
			// 移除已处理的数据
			receivedBuf = receivedBuf[totalLen:]
			// 解析完整数据
			err = processCompleteTaskData(packetType, data, kvmID)
			if err != nil {
				gvaLog.Error("解析数据错误:", zap.Error(err))
				continue
			}
		}

		// 处理接收的数据包
		for len(receivedBuf) >= BufSize {
			// 解析数据包头部信息
			packet := receivedBuf[:BufSize]
			receivedBuf = receivedBuf[BufSize:]
			dataLen := int(packet[3])<<16 | int(packet[4])<<8 | int(packet[5])
			totalLen := 6 + dataLen
			dataStatus := int(packet[2])
			taskID := int(packet[1])
			// 过滤无效数据包
			if !isValidPacket(taskID, dataStatus) {
				continue
			}
			// 开始接收数据，添加数据到缓存中
			if dataStatus == protocol.DataStart {
				collectTaskDataMapMutex.Lock()
				if _, ok := collectTaskDataMap[taskID]; ok {
					delete(collectTaskDataMap, taskID)
				}
				collectTaskDataMap[taskID] = append(collectTaskDataMap[taskID], packet...)
				collectTaskDataMapMutex.Unlock()
			}
			// 跳过不完整数据包
			collectTaskDataMapMutex.Lock()
			if _, ok := collectTaskDataMap[taskID]; !ok && totalLen >= BufSize {
				collectTaskDataMapMutex.Unlock()
				continue
			}
			collectTaskDataMapMutex.Unlock()
			// 处理传输中数据，追加数据到缓存中
			if dataStatus == protocol.DataTransfer {
				collectTaskDataMapMutex.Lock()
				collectTaskDataMap[taskID] = append(collectTaskDataMap[taskID], packet[6:]...)
				collectTaskDataMapMutex.Unlock()
			}
			// 处理传输结束数据
			if dataStatus == protocol.DataEnd {
				err = handleEndPacket(packet, taskID, totalLen, kvmID, processCompleteTaskData)
				if err != nil {
					gvaLog.Error("解析数据错误:", zap.Error(err))
					continue
				}
			}
		}
	}
}

// ListenTaskConnection 监听任务连接通道数据（.fa）
func ListenTaskConnection(conn net.Conn, kvmID string, log *zap.Logger, processCompleteTaskData ProcessCompleteDataFunc) {
	defer conn.Close()
	gvaLog = log

	var receivedBuf []byte
	for {
		buf := make([]byte, BufSize)
		n, err := conn.Read(buf)
		if err != nil {
			gvaLog.Error("Error reading from socket:", zap.Error(err))
			gvaLog.Info(fmt.Sprintf("agent[%s].fa 连接断开", kvmID))
			return
		}

		// 过滤无效数据包
		if int(buf[0]) != global.OldAgentBackCollect && int(buf[0]) != global.TaskCallBackCollect {
			continue
		}

		receivedBuf = append(receivedBuf, buf[:n]...)
		// 兼容旧版本任务数据
		for len(receivedBuf) >= 3 && len(receivedBuf) == BufSize && int(receivedBuf[0]) == global.OldAgentBackCollect {
			packetLen := int(receivedBuf[1])<<8 | int(receivedBuf[2])
			totalLen := 3 + packetLen

			packetType := int(receivedBuf[0])
			data := receivedBuf[3:totalLen]
			// 移除已处理的数据
			receivedBuf = receivedBuf[4096:]
			// 解析完整数据
			err = processCompleteTaskData(packetType, data, kvmID)
			if err != nil {
				gvaLog.Error("解析数据错误:", zap.Error(err))
				continue
			}
		}

		// 处理接收的数据包
		for len(receivedBuf) >= BufSize {
			// 解析数据包头部信息
			packet := receivedBuf[:BufSize]
			receivedBuf = receivedBuf[BufSize:]
			dataLen := int(packet[3])<<16 | int(packet[4])<<8 | int(packet[5])
			totalLen := 6 + dataLen
			dataStatus := int(packet[2])
			taskID := int(packet[1])
			// 过滤无效数据包
			if !isValidPacket(taskID, dataStatus) {
				continue
			}
			// 开始接收数据，添加数据到缓存中
			if dataStatus == protocol.DataStart {
				taskDataMapMutex.Lock()
				if _, ok := taskDataMap[taskID]; ok {
					delete(taskDataMap, taskID)
				}
				taskDataMap[taskID] = append(taskDataMap[taskID], packet...)
				taskDataMapMutex.Unlock()
			}
			// 跳过不完整数据包
			taskDataMapMutex.Lock()
			if _, ok := taskDataMap[taskID]; !ok && totalLen >= BufSize {
				taskDataMapMutex.Unlock()
				continue
			}
			taskDataMapMutex.Unlock()
			// 处理传输中数据，追加数据到缓存中
			if dataStatus == protocol.DataTransfer {
				taskDataMapMutex.Lock()
				taskDataMap[taskID] = append(taskDataMap[taskID], packet[6:]...)
				taskDataMapMutex.Unlock()
			}
			// 处理传输结束数据
			if dataStatus == protocol.DataEnd {
				err = handleEndTaskPacket(packet, taskID, totalLen, kvmID, processCompleteTaskData)
				if err != nil {
					gvaLog.Error("解析数据错误:", zap.Error(err))
					continue
				}
			}
		}
	}
}

// ListenSerialConnection 监听物理串口连接通道数据（.fa00）
func ListenSerialConnection(conn net.Conn, kvmID string, log *zap.Logger, processCompleteTaskData ProcessCompleteDataFunc) {
	defer conn.Close()
	gvaLog = log

	var receivedBuf []byte
	for {
		buf := make([]byte, BufSize)
		n, err := conn.Read(buf)
		if err != nil {
			log.Error("Error reading from socket:", zap.Error(err))
			log.Info(fmt.Sprintf("agent[%s].fa00 连接断开", kvmID))
			// 清理所有未完成的数据包
			packetBuffersMutex.Lock()
			packetBuffers = make(map[byte]*protocol.PacketBuffer)
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
			header := protocol.PacketHeader{
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
			case protocol.DataStart:
				packetBuffers[header.TaskID] = &protocol.PacketBuffer{
					Data:    packet[HeaderSize : totalLen-CRCSize],
					LastSeq: header.SeqNum,
				}
			case protocol.DataTransfer:
				if buf, exists := packetBuffers[header.TaskID]; exists && header.SeqNum == buf.LastSeq+1 {
					buf.Data = append(buf.Data, packet[HeaderSize:totalLen-CRCSize]...)
					buf.LastSeq = header.SeqNum
				}
			case protocol.DataEnd:
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

// 校验数据包格式
func isValidPacket(taskID int, dataStatus int) bool {
	if taskID < 255 && taskID > 0 && (dataStatus == protocol.DataStart || dataStatus == protocol.DataTransfer || dataStatus == protocol.DataEnd) {
		return true
	}
	return false
}

// handleEndPacket 处理数据结束包
func handleEndPacket(packet []byte, taskID int, totalLen int, kvmID string, processCompleteAgentDataData ProcessCompleteDataFunc) error {
	actualDataEnd := totalLen - len(collectTaskDataMap[taskID]) + 6
	if actualDataEnd > BufSize {
		actualDataEnd = BufSize
	}
	collectTaskDataMapMutex.Lock()
	if totalLen <= BufSize {
		collectTaskDataMap[taskID] = append(collectTaskDataMap[taskID], packet[:totalLen]...)
	} else {
		collectTaskDataMap[taskID] = append(collectTaskDataMap[taskID], packet[6:actualDataEnd]...)
	}
	collectTaskData := collectTaskDataMap[taskID]
	packetType := int(collectTaskData[0])
	data := collectTaskData[6:]
	// 清除已处理的数据包
	delete(collectTaskDataMap, taskID)
	collectTaskDataMapMutex.Unlock()
	// 解析完整数据
	err := processCompleteAgentDataData(packetType, data, kvmID)
	if err != nil {
		return err
	}
	return nil
}

// handleEndPacket 处理数据结束包
func handleEndTaskPacket(packet []byte, taskID int, totalLen int, kvmID string, processCompleteTaskData ProcessCompleteDataFunc) error {
	actualDataEnd := totalLen - len(taskDataMap[taskID]) + 6
	if actualDataEnd > BufSize {
		actualDataEnd = BufSize
	}
	taskDataMapMutex.Lock()
	if totalLen <= BufSize {
		taskDataMap[taskID] = append(taskDataMap[taskID], packet[:totalLen]...)
	} else {
		taskDataMap[taskID] = append(taskDataMap[taskID], packet[6:actualDataEnd]...)
	}
	taskData := taskDataMap[taskID]
	packetType := int(taskData[0])
	data := taskData[6:]
	// 清除已处理的数据包
	delete(taskDataMap, taskID)
	taskDataMapMutex.Unlock()
	// 解析完整数据
	err := processCompleteTaskData(packetType, data, kvmID)
	if err != nil {
		return err
	}
	return nil
}
