package sdk

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Kvm struct {
	ID string
}

type MetricsHostInfo struct {
	AgentID     string      `json:"agentId"`     //  探针ID
	KvmID       string      `json:"kvmId"`       // 虚拟机ID
	MetricsCode string      `json:"metricsCode"` // 指标编号
	MetricsName string      `json:"metricsName"` // 指标名
	MetricsType string      `json:"metricsType"` // 轮询方式
	Summary     string      `json:"summary"`     // 摘要
	MetricsData interface{} `json:"metricsData"` // 指标数据
	Level       uint        `json:"level"`       // 指标等级
	Interval    uint        `json:"interval"`    // 采集周期
}

func (t *MetricsHostInfo) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"KvmID":       t.KvmID,
		"MetricsCode": t.MetricsCode,
		"MetricsName": t.MetricsName,
		"MetricsType": t.MetricsType,
		"Summary":     t.Summary,
		"MetricsData": t.MetricsData,
		"Level":       t.Level,
		"Interval":    t.Interval,
	}
}

const BufSize = 4096

var agentData chan map[string]interface{}

// 数据状态
var (
	DataStart    = 0
	DataTransfer = 1
	DataEnd      = 2
)

// ProcessReceivedData 处理接收到的数据包
func ProcessReceivedData(buf []byte, receivedBuf []byte, kvmID string) ([]byte, error) {
	// 过滤无效数据包
	if int(buf[0]) != 2 {
		return receivedBuf, nil // 返回未修改的缓存
	}

	receivedBuf = append(receivedBuf, buf...)

	// 兼容旧版本心跳采集数据
	for len(receivedBuf) >= 3 && len(receivedBuf) < BufSize {
		dataStatus := int(receivedBuf[2])
		taskID := int(receivedBuf[1])

		// 校验数据包格式
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

		// 调用处理函数解析数据
		err := ProcessPacket([]byte{byte(packetType)}, kvmID, data)
		if err != nil {
			return receivedBuf, fmt.Errorf("解析数据错误: %v", err)
		}
	}

	return receivedBuf, nil
}

func ProcessPacket(packetType []byte, kvmID string, data []byte) error {
	// 处理不同类型的数据包
	switch packetType[0] {
	case 2: // 假设 2 是合法的数据包类型
		var metricsHostInfo MetricsHostInfo
		err := json.Unmarshal(data, &metricsHostInfo)
		if err != nil {
			return err
		}
		metricsHostInfo.KvmID = kvmID
		agentData <- metricsHostInfo.ToMap()

		GvaLog.Info(fmt.Sprintf("收到[ %s ]的[ %s ]采集数据", kvmID, metricsHostInfo.MetricsName))
	default:
		return errors.New("未知数据类型")
	}

	return nil
}

// 校验数据包格式
func isValidPacket(taskID int, dataStatus int) bool {
	if taskID < 255 && taskID > 0 && (dataStatus == DataStart || dataStatus == DataTransfer || dataStatus == DataEnd) {
		return true
	}
	return false
}
