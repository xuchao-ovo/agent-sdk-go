package metrics

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xuchao-ovo/agent-sdk-go/pkg/metrics/types"
)

// GetSummary 获取摘要
func GetSummary(agentData types.MetricsHostInfo) (string, error) {
	// 将[]interface{}转换为json
	metricsDataJson, err := json.Marshal(agentData.MetricsData)
	if err != nil {
		return "", err
	}
	// 获取探针信息
	switch agentData.MetricsCode {
	case "PC1":
		var systemInfo types.SystemData
		if err = json.Unmarshal(metricsDataJson, &systemInfo); err != nil {
			return "", err
		}
		// 设置摘要信息
		agentData.Summary = fmt.Sprintf("操作系统:%s，版本: %s", systemInfo.Manufacture, systemInfo.SystemDescription)
	case "PC2":
		var netInfo []types.NetInfo
		if err = json.Unmarshal(metricsDataJson, &netInfo); err != nil {
			return "", err
		}
		if len(netInfo) == 0 {
			agentData.Summary = "未查询到网卡信息"
			return agentData.Summary, nil
		}
		var IPv4, Mac []string
		for _, net := range netInfo {
			if net.IPv4 != "" && net.IPv4 != "127.0.0.1" {
				IPv4 = append(IPv4, net.IPv4)
			}
			if net.MAC != "" {
				Mac = append(Mac, net.MAC)
			}
		}
		agentData.Summary = fmt.Sprintf("共%d个网卡，IP分别为%v", len(netInfo), IPv4)
	case "PC3":
		var processInfo []types.ProcessInfo
		if err = json.Unmarshal(metricsDataJson, &processInfo); err != nil {
			return "", err
		}
		if len(processInfo) == 0 {
			agentData.Summary = "未查询到进程信息"
			return agentData.Summary, nil
		}
		var totalCpuUseRate, totalMemoryUseRate float64
		for _, proc := range processInfo {
			totalCpuUseRate += proc.CpuUseRate
			totalMemoryUseRate += proc.MemoryUseRate
		}
		agentData.Summary = fmt.Sprintf("共%d个进程，共占用%.2f%% CPU、%.2f MB 内存", len(processInfo), totalCpuUseRate, totalMemoryUseRate)
	case "PC4":
		var portInfo []types.PortInfo
		if err = json.Unmarshal(metricsDataJson, &portInfo); err != nil {
			return "", err
		}
		if len(portInfo) == 0 {
			agentData.Summary = "未开放端口"
			return agentData.Summary, nil
		}
		var portList []string
		for i, port := range portInfo {
			if i < 3 { // 只显示前三个端口
				portList = append(portList, fmt.Sprintf("%s:%d", port.ListenAddr, port.Port))
			} else {
				break
			}
		}
		agentData.Summary = fmt.Sprintf("共开放%d个端口，包括：%s 等", len(portInfo), strings.Join(portList, "、"))
	case "PC5":
		var arpInfo []types.ArpInfo
		if err = json.Unmarshal(metricsDataJson, &arpInfo); err != nil {
			return "", err
		}
		ipList := make([]string, 0)
		if len(arpInfo) == 0 {
			agentData.Summary = "有过网络连接的IP：无"
			return agentData.Summary, nil
		}
		for i, v := range arpInfo {
			if i < 3 { // 只显示前三个ip
				ipList = append(ipList, v.CacheIp)
			} else {
				break
			}
		}
		agentData.Summary = fmt.Sprintf("有过网络连接的IP：%s等", strings.Join(ipList, "、"))
	case "PC6":
		var userInfo []types.UserInfo
		if err = json.Unmarshal(metricsDataJson, &userInfo); err != nil {
			return "", err
		}
		if len(userInfo) == 0 {
			agentData.Summary = "未查询到用户信息"
			return agentData.Summary, nil
		}
		var userList []string
		for i, user := range userInfo {
			if i < 3 && user.Name != "" { // 只显示前三个用户
				userList = append(userList, user.Name)
			} else {
				break
			}
		}
		agentData.Summary = fmt.Sprintf("共%d个用户，包括：%s等", len(userInfo), strings.Join(userList, "、"))
	case "PC7":
		var fileModifyInfo types.FileModifyData
		if err = json.Unmarshal(metricsDataJson, &fileModifyInfo); err != nil {
			return "", err
		}
		var operate string
		if fileModifyInfo.Operate == "create" {
			operate = "创建"
		} else if fileModifyInfo.Operate == "write" {
			operate = "写入"
		} else if fileModifyInfo.Operate == "remove" {
			operate = "删除"
		} else if fileModifyInfo.Operate == "rename" {
			operate = "重命名"
		} else if fileModifyInfo.Operate == "chmod" {
			operate = "修改权限"
		}
		agentData.Summary = fmt.Sprintf("文件 [%s] 被%s", fileModifyInfo.FileName, operate)
	case "PC9":
		var cronTaskData []types.CronTaskData
		if err = json.Unmarshal(metricsDataJson, &cronTaskData); err != nil {
			return "", err
		}
		if len(cronTaskData) == 0 {
			agentData.Summary = "未查询到定时任务"
			return agentData.Summary, nil
		}
		var taskList []string
		for i, task := range cronTaskData {
			if i < 3 && task.TaskName != "" {
				taskList = append(taskList, task.TaskName)
			} else {
				break
			}
		}
		agentData.Summary = fmt.Sprintf("%d个定时任务，包括：%s等", len(cronTaskData), strings.Join(taskList, "、"))
	case "PC10":
		var loginInfo []types.LoginInfo
		if err = json.Unmarshal(metricsDataJson, &loginInfo); err != nil {
			return "", err
		}
		// 设置摘要信息
		loginCount := len(loginInfo)
		loginTypes := make(map[uint32]int)
		for _, info := range loginInfo {
			loginTypes[info.LogType]++
		}
		loginTypesSummary := ""
		for logType, count := range loginTypes {
			loginTypesSummary += fmt.Sprintf("%d种登录方式（类型 %d: %d 次），", count, logType, count)
		}
		agentData.Summary = fmt.Sprintf("%d个用户登录，%s", loginCount, loginTypesSummary)
	case "PC11":
		agentData.Summary = "探针心跳"
	case "PC12":
		var cpuInfo types.CpuInfo
		if err = json.Unmarshal(metricsDataJson, &cpuInfo); err != nil {
			return "", err
		}
		agentData.Summary = fmt.Sprintf("CPU使用率：%.2f%%", cpuInfo.CpuUseRate)
	default:
		agentData.Summary = fmt.Sprintf("未知指标类型: %s", agentData.MetricsCode)
	}

	return agentData.Summary, nil
}
