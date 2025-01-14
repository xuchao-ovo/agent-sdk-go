package metrics

import (
	"encoding/json"
	"fmt"
	"strings"
)

type MetricsHostInfo struct {
	KvmID       string      `json:"KvmID"`
	MetricsCode string      `json:"metricsCode"` // 指标编号
	MetricsName string      `json:"metricsName"` // 指标名
	MetricsType string      `json:"metricsType"` // 轮询方式
	Summary     string      `json:"summary"`     // 摘要
	MetricsData interface{} `json:"metricsData"` // 指标数据
	Level       uint        `json:"level"`       // 指标等级
	Interval    uint        `json:"interval"`    // 采集周期
}

// GetSummary 获取摘要
func GetSummary(agentData MetricsHostInfo) (string, error) {
	// 将[]interface{}转换为json
	metricsDataJson, err := json.Marshal(agentData.MetricsData)
	if err != nil {
		return "", err
	}
	// 获取探针信息
	switch agentData.MetricsCode {
	case "PC1":
		var systemInfo SystemData
		if err = json.Unmarshal(metricsDataJson, &systemInfo); err != nil {
			return "", err
		}
		// 设置摘要信息
		agentData.Summary = fmt.Sprintf("操作系统:%s，版本: %s", systemInfo.Manufacture, systemInfo.SystemDescription)
	case "PC2":
		var netInfo []NetInfo
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
		var processInfo []ProcessInfo
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
		var portInfo []PortInfo
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
		var arpInfo []ArpInfo
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
		var userInfo []UserInfo
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
		var fileModifyInfo FileModifyData
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
		var cronTaskData []CronTaskData
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
		var loginInfo []LoginInfo
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
		var cpuInfo CpuInfo
		if err = json.Unmarshal(metricsDataJson, &cpuInfo); err != nil {
			return "", err
		}
		agentData.Summary = fmt.Sprintf("CPU使用率：%.2f%%", cpuInfo.CpuUseRate)
	case "PC13":
		var diskInfo DiskData
		if err = json.Unmarshal(metricsDataJson, &diskInfo); err != nil {
			return "", err
		}
		agentData.Summary = fmt.Sprintf("磁盘使用率：%.2f%%", diskInfo.UsedPercent)
	case "PC14":
		var memoryInfo MemInfo
		if err = json.Unmarshal(metricsDataJson, &memoryInfo); err != nil {
			return "", err
		}
		agentData.Summary = fmt.Sprintf("内存使用率：%.2f%%", memoryInfo.MemoryUseRate)

	case "PC15":
		var netSendInfo []NetSendInfo
		if err = json.Unmarshal(metricsDataJson, &netSendInfo); err != nil {
			return "", err
		}
		netSendList := []string{}
		for _, netSend := range netSendInfo {
			if netSend.PacketsSent != 0 || netSend.BytesSentRate != 0 {
				summary := fmt.Sprintf("网卡[ %s ]发包速率: %.2f KB/s, 发包数: %d", netSend.Name, float64(netSend.BytesSentRate), netSend.PacketsSent)
				netSendList = append(netSendList, summary)
			}
		}
		if len(netSendList) == 0 {
			agentData.Summary = fmt.Sprintf("网卡发包速率为 0KB/s")
			return agentData.Summary, nil
		}
		agentData.Summary = fmt.Sprintf(strings.Join(netSendList, "、"))
	case "PC16":
		var netRecvInfo []NetRecvInfo
		if err = json.Unmarshal(metricsDataJson, &netRecvInfo); err != nil {
			return "", err
		}
		netRecvList := []string{}
		for _, netRecv := range netRecvInfo {
			if netRecv.PacketsRecv != 0 || netRecv.BytesRecvRate != 0 {
				summary := fmt.Sprintf("网卡[ %s ]收包速率: %.2f KB/s, 收包数: %d", netRecv.Name, float64(netRecv.BytesRecvRate), netRecv.PacketsRecv)
				netRecvList = append(netRecvList, summary)
			}
		}
		if len(netRecvList) == 0 {
			agentData.Summary = fmt.Sprintf("网卡收包速率为 0KB/s")
			return agentData.Summary, nil
		}
		agentData.Summary = fmt.Sprintf(strings.Join(netRecvList, "、"))
	case "PC18":
		var softwareInfo []SoftwareData
		if err = json.Unmarshal(metricsDataJson, &softwareInfo); err != nil {
			return "", err
		}
		// 设置摘要信息
		softwareCount := len(softwareInfo)
		softwareList := []string{}
		for i, software := range softwareInfo {
			if i < 3 { // 只显示前三个软件
				softwareList = append(softwareList, software.DisplayName)
			} else {
				break
			}
		}
		if len(softwareList) == 0 {
			agentData.Summary = fmt.Sprintf("未安装三方软件")
		} else {
			agentData.Summary = fmt.Sprintf("共%d个软件，包括：%s等", softwareCount, strings.Join(softwareList, "、"))
		}
	case "PC19":
		var firewallInfo []FirewallStatus
		if err = json.Unmarshal(metricsDataJson, &firewallInfo); err != nil {
			return "", err
		}
		var summary string
		for _, info := range firewallInfo {
			switch info.FirewallName {
			case "domainProfile":
				summary += fmt.Sprintf("%s: %s", "域防火墙", firewallStatus(info.Status))
				break
			case "privateProfile":
				summary += fmt.Sprintf("%s: %s", "专用防火墙", firewallStatus(info.Status))
				break
			case "publicProfile":
				summary += fmt.Sprintf("%s: %s", "公用防火墙", firewallStatus(info.Status))
				break
			default:
				summary += fmt.Sprintf("%s: %s", info.FirewallName, firewallStatus(info.Status))
				break
			}

		}
		agentData.Summary = summary
	case "PC21":
		var sshInfo []SSHInfo
		if err = json.Unmarshal(metricsDataJson, &sshInfo); err != nil {
			return "", err
		}
		ipList := make([]string, 0)
		for i, v := range sshInfo {
			if i < 3 { // 只显示前三个ip
				ipList = append(ipList, v.ClientIP)
			} else {
				break
			}
		}
		agentData.Summary = fmt.Sprintf("有过SSH连接的IP：%s等", strings.Join(ipList, "、"))
	case "PC22":
		var rdpLog []RDPLog
		if err = json.Unmarshal(metricsDataJson, &rdpLog); err != nil {
			return "", err
		}
		ipList := make([]string, 0)
		for i, v := range rdpLog {
			if i < 3 { // 只显示前三个ip
				ipList = append(ipList, v.Server)
			} else {
				break
			}
		}
		agentData.Summary = fmt.Sprintf("有过RDP连接的IP：%s等", strings.Join(ipList, "、"))
	case "PC23":
		var eventLogs []EventLogInfo
		if err = json.Unmarshal(metricsDataJson, &eventLogs); err != nil {
			return "", err
		}
		if len(eventLogs) == 0 {
			agentData.Summary = "无日志信息"
			return agentData.Summary, nil
		}
		agentData.Summary = fmt.Sprintf("共采集到 %d 条日志信息，最近一条日志：%s", len(eventLogs), eventLogs[0].Message)

	}
	return agentData.Summary, nil
}

func firewallStatus(status bool) string {
	switch status {
	case true:
		return "开启"
	case false:
		return "关闭"
	default:
		return "未知"
	}
}
