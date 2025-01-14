package metrics

import (
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"net"
	"sync"
)

const BufSize = 4096

var taskDataMap = make(map[int][]byte) // 用于缓存每个任务的数据包
var taskDataMapMutex sync.Mutex

var gvaLog *zap.Logger

// 数据状态
var (
	DataStart    = 0
	DataTransfer = 1
	DataEnd      = 2
)

// ListenConnection 监听单个连接的数据
func ListenConnection(conn net.Conn, kvmID string, agentMap map[string]interface{}, agentMapMutex sync.Mutex, agentData chan map[string]interface{}, log *zap.Logger) {
	defer conn.Close()
	gvaLog = log

	var receivedBuf []byte
	for {
		buf := make([]byte, BufSize)
		n, err := conn.Read(buf)
		if err != nil {
			gvaLog.Error("Error reading from socket:", zap.Error(err))
			gvaLog.Info(fmt.Sprintf("agent[%s] 连接断开", kvmID))
			agentMapMutex.Lock()
			delete(agentMap, kvmID)
			agentMapMutex.Unlock()
			return
		}
		// 过滤无效数据包
		if int(buf[0]) != 2 {
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
			err = processCompleteTaskData(packetType, data, kvmID, agentData)
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
			if dataStatus == DataStart {
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
			if dataStatus == DataTransfer {
				taskDataMapMutex.Lock()
				taskDataMap[taskID] = append(taskDataMap[taskID], packet[6:]...)
				taskDataMapMutex.Unlock()
			}
			// 处理传输结束数据
			if dataStatus == DataEnd {
				err = handleEndPacket(packet, taskID, totalLen, kvmID, agentData)
				if err != nil {
					gvaLog.Error("解析数据错误:", zap.Error(err))
					continue
				}
			}
		}
	}
}

// 校验数据包格式
func isValidPacket(taskID int, dataStatus int) bool {
	if taskID < 255 && taskID > 0 && (dataStatus == DataStart || dataStatus == DataTransfer || dataStatus == DataEnd) {
		return true
	}
	return false
}

// handleEndPacket 处理数据结束包
func handleEndPacket(packet []byte, taskID int, totalLen int, kvmID string, agentData chan map[string]interface{}) error {
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
	err := processCompleteTaskData(packetType, data, kvmID, agentData)
	if err != nil {
		return err
	}
	return nil
}

// processCompleteTaskData 解析完整数据
func processCompleteTaskData(packetType int, data []byte, kvmID string, agentData chan map[string]interface{}) error {
	switch packetType {
	case 2:
		var metricsHostInfo MetricsHostInfo
		err := json.Unmarshal(data, &metricsHostInfo)
		if err != nil {
			return err
		}
		metricsHostInfo.KvmID = kvmID
		agentData <- metricsHostInfo.ToMap()
		gvaLog.Info(fmt.Sprintf("收到[ %s ]的[ %s ]采集数据", kvmID, metricsHostInfo.MetricsName))

	default:
		return errors.New("未知数据类型")
	}
	return nil
}
