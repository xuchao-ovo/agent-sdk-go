// metrics/metrics.go

package metrics

import (
	"encoding/json"
	"fmt"
	"utils"
)

// 定义 MetricsHostInfo 结构体
type MetricsHostInfo struct {
	KvmID       string `json:"kvm_id"`
	MetricsCode string `json:"metrics_code"`
	MetricsName string `json:"metrics_name"`
	Summary     string `json:"summary"`
}

// addSummary 函数用于根据采集数据生成简洁摘要
func AddSummary(agentData *MetricsHostInfo, data json.RawMessage) {
	// 根据 MetricsCode 判断数据类型并处理
	switch agentData.MetricsCode {
	case "PC1": // 系统信息
		var systemData map[string]interface{}
		if err := json.Unmarshal(data, &systemData); err != nil {
			utils.LogError("Failed to parse system data: %v", err)
			return
		}
		// 生成摘要信息
		agentData.Summary = fmt.Sprintf("OS: %s, Version: %s", systemData["os"], systemData["version"])

	case "PC2": // 网络信息
		var netInfo []map[string]interface{}
		if err := json.Unmarshal(data, &netInfo); err != nil {
			utils.LogError("Failed to parse network info: %v", err)
			return
		}
		// 生成摘要信息
		agentData.Summary = fmt.Sprintf("Network interfaces: %d", len(netInfo))

	case "PC3": // 进程信息
		var processInfo []map[string]interface{}
		if err := json.Unmarshal(data, &processInfo); err != nil {
			utils.LogError("Failed to parse process info: %v", err)
			return
		}
		// 生成摘要信息
		agentData.Summary = fmt.Sprintf("Processes count: %d", len(processInfo))

	default:
		utils.LogError("Unknown MetricsCode: %s", agentData.MetricsCode)
	}
}
