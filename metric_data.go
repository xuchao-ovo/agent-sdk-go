package metrics

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

// SystemData 系统信息采集信息 => PC1
type SystemData struct {
	Hostname           string `json:"hostname"`           // 主机名
	CPUModel           string `json:"cpuModel"`           // CPU型号
	LogicalCores       int    `json:"logicalCores"`       // 逻辑处理器数量
	SystemArchitecture string `json:"systemArchitecture"` // 系统架构
	Manufacture        string `json:"manufacture"`        // 系统厂商
	SystemDescription  string `json:"systemDescription"`  // 系统描述
	ManufactureDate    string `json:"manufactureDate"`    // 生产时间
	InstallDate        string `json:"installDate"`        // 系统安装时间
	Uptime             string `json:"uptime"`             // 系统运行时间                                                                                                                                                                              version:2022-05-23 15:08
}

// NetInfo 网卡采集信息 => PC2
type NetInfo struct {
	Hostname string `json:"hostname"` // 主机名
	Name     string `json:"name"`     // 网卡名称
	IPv4     string `json:"ipv4"`     // IPv4地址
	IPv4Mask string `json:"ipv4Mask"` // IPv4掩码
	IPv6     string `json:"ipv6"`     // IPv6地址
	IPv6Mask string `json:"ipv6Mask"` // IPv6掩码
	Gateway  string `json:"gateway"`  // 网关
	DNS      string `json:"dns"`      // DNS地址
	MAC      string `json:"mac"`      // MAC地址
	Status   string `json:"status"`   // 状态
}

// ProcessInfo 进程采集信息 => PC3
type ProcessInfo struct {
	PId              int32   `json:"pId"`              // 进程ID;应记录进程ID。
	PpId             int32   `json:"ppId"`             // 父进程ID;应记录启动进程的父进程ID。
	Account          string  `json:"account"`          // 帐户名称;应记录进程执行的账户名称。
	ProcessName      string  `json:"processName"`      // 进程名称;应记录进程的进程名称。
	MemoryUseBytes   string  `json:"memoryUseBytes"`   // 内存使用大小;应记录进程占用的内存大小。
	MemoryUseRate    float64 `json:"memoryUseRate"`    // 内存使用率;应记录进程占用的内存利用率。
	CpuUseRate       float64 `json:"cpuUseRate"`       // CPU利用率;应记录进程占用的CPU大小。
	IoReadBytes      uint64  `json:"ioReadBytes"`      // 读取字节数;应记录进程读取的字节数。
	IoWriteBytes     uint64  `json:"ioWriteBytes"`     // 写入字节数;应记录进程写入的字节数。
	IoReadRate       float64 `json:"ioReadRate"`       // IO 读速率
	IoWriteRate      float64 `json:"ioWriteRate"`      // IO 写速率
	ProcessStartDate string  `json:"processStartDate"` // 进程创建时间;应记录进程启动的时间。
	DynamicLib       string  `json:"dynamicLib"`       // 动态库;应记录进程依赖的所有动态库。
	Cmd              string  `json:"cmd"`              // 进程执行命令
	ProcessStatus    int     `json:"processStatus"`    // 进程状态;应填写进程的状态，1=正在运行，2=处于休眠状态，3= 停止或被追踪，4=僵尸进程，5=进入内存交换，6=死掉的 进程。
	ProcessPath      string  `json:"processPath"`      // 进程路径;应记录进程启动的路径。
	CollectedAt      uint64  `json:"collectedAt"`      // 采集时间(秒级时间戳)
}

// PortInfo 端口采集信息 => PC4
type PortInfo struct {
	ListenAddr      string `json:"listenAddr"`      // 监听地址
	Port            uint32 `json:"port"`            // 端口
	Protocol        string `json:"protocol"`        // 协议
	ConnectionCount uint32 `json:"connectionCount"` // 连接数
	ProcessPid      int32  `json:"processPid"`      // 进程PID
	Process         string `json:"process"`         // 进程名称
	ProcessPath     string `json:"processPath"`     // 进程路径
	ProcessCreate   string `json:"processCreate"`   // 进程创建时间
	Cmd             string `json:"cmd"`             // 进程命令
}

// ArpInfo 网络互连采集信息 => PC5
type ArpInfo struct {
	CacheIp   string `json:"cacheIp"` // 缓存IP;缓存IP
	NetworkIp string `json:"networkIp"`
	CacheMac  string `json:"cacheMac"`  // 缓存MAC;缓存MAC
	IsGateway bool   `json:"isGateway"` // 是否网关;是否网关
	CacheType string `json:"cacheType"` // 缓存类型;缓存类型
}

// UserInfo 所有用户信息采集  => PC6
type UserInfo struct {
	Name     string `json:"name"`      // 用户名
	FullName string `json:"full_name"` // 用户全名
	Domain   string `json:"domain"`    // 用户所属域
	SID      string `json:"sid"`       // 用户SID
	Disabled bool   `json:"disabled"`  // 是否禁用;是否禁用
}

// FileModifyData 文件变动采集信息 => PC7
type FileModifyData struct {
	FileName         string `json:"fileName"`         // 文件名;应记录发生变化的文件的文件名。
	FilePath         string `json:"filePath"`         // 文件路径;应记录发生变化的文件的路径。
	Operate          string `json:"operate"`          // 操作;应记录发生变化的文件的操作。
	UpdateTime       int64  `json:"updateTime"`       // 修改时间;应记录发生变化的文件的修改时间。
	IsAllowedCreate  bool   `json:"isAllowedCreate"`  // 是否允许新建;应记录文件是否可以新建，1=是，0=否。
	OriginalFileHash string `json:"originalFileHash"` // 原始文件hash;应记录发生变化的文件的原始hash。
	UpdatedFileHash  string `json:"updatedFileHash"`  // 修改后文件hash;应记录发生变化后的文件的hash。
	ThreatLevel      string `json:"threatLevel"`      // 威胁级别;应记录该类变更的威胁级别。
}

// CommandModifyData 系统命令采集信息 => PC8
type CommandModifyData struct {
	Command     string `json:"command"`     // 系统命令
	CollectTime int64  `json:"collectTime"` // 采集时间
}

// CronTaskData 定时任务采集信息 => PC9
type CronTaskData struct {
	HostName        string `json:"hostName"`        // 主机名
	TaskName        string `json:"taskName"`        // 任务名
	NextRunTime     string `json:"nextRunTime"`     // 下次运行时间
	Mode            string `json:"mode"`            // 模式
	LoginType       string `json:"loginType"`       // 登录状态
	LastRunTime     string `json:"lastRunTime"`     // 上次运行时间
	LastRunResult   string `json:"lastRunResult"`   // 上次结果
	CreateBy        string `json:"createBy"`        // 创建者
	Command         string `json:"command"`         // 要运行的任务
	Description     string `json:"description"`     // 注释
	TaskState       string `json:"taskState"`       // 计划任务状态
	FreeTime        string `json:"freeTime"`        // 空闲时间
	PowerManagement string `json:"powerManagement"` // 电源管理
	RunAsUser       string `json:"runAsUser"`       // 作为用户运行
	Key1            string `json:"key1"`            // 删除没有计划的任务
	Key2            string `json:"key2"`            // 如果运行了 X 小时 X 分钟，停止任务
	Schedule        string `json:"schedule"`        // 计划
	TaskType        string `json:"taskType"`        // 计划类型
}

// LoginInfo 用户登录信息采集 => PC10
type LoginInfo struct {
	Name                  string    // 用户名
	Domain                string    // 域
	StartTime             time.Time // 登录时间
	AuthenticationPackage string    // 认证包名称
	LogType               uint32    //
}

// HeartBeatInfo 心跳采集信息 => PC11
type HeartBeatInfo struct {
	HostName         string `json:"hostName"`         // 主机名称
	Platform         string `json:"platform"`         // 平台
	Type             uint   `json:"type"`             // 类型
	Config           Config `json:"config"`           // 配置信息
	CollectionStatus bool   `json:"collectionStatus"` // 采集状态
}

type Config struct {
	Version       string         `json:"version"`
	MetricConfig  []MetricConfig `json:"metricConfig"`
	SoftwareTools []Software     `json:"softwareTools"`
}

// CpuInfo CPU采集信息 => PC12
type CpuInfo struct {
	CpuUseRate float64 `json:"cpuUseRate"`
}

// DiskData 磁盘采集信息 => PC13
type DiskData struct {
	Total       string  `json:"total"`       // 磁盘总量
	Used        string  `json:"used"`        // 磁盘已用
	Free        string  `json:"free"`        // 磁盘剩余
	UsedPercent float64 `json:"usedPercent"` // 磁盘使用率
	Disks       []Disk  `json:"disks"`       // 磁盘信息
}
type Disk struct {
	Name        string  `json:"name"`        // 磁盘名称
	Total       string  `json:"total"`       // 磁盘总量
	Used        string  `json:"used"`        // 磁盘已用
	Free        string  `json:"free"`        // 磁盘剩余
	UsedPercent float64 `json:"usedPercent"` // 磁盘使用率
}

// MemInfo 内存采集信息 => PC14
type MemInfo struct {
	MemoryUseRate    float64 `json:"memoryUseRate"`
	MemoryUseBytes   string  `json:"memoryUseBytes"`
	MemoryTotalBytes string  `json:"memoryTotalBytes"`
	MemoryFreeBytes  string  `json:"memoryFreeBytes"`
}

// NetSendInfo 网卡发包速率采集信息 => PC15
type NetSendInfo struct {
	Name          string `json:"name"`          // 网卡名称
	PacketsSent   uint64 `json:"packetsSent"`   // 发包数量
	BytesSentRate uint64 `json:"bytesSentRate"` // 发包速率
}

// NetRecvInfo 网卡收包速率采集信息 => PC16
type NetRecvInfo struct {
	Name          string `json:"name"`          // 网卡名称
	PacketsRecv   uint64 `json:"packetsRecv"`   // 收包数量
	BytesRecvRate uint64 `json:"bytesRecvRate"` // 收包速率
}

// SoftwareData 已安装应用采集信息 => PC18
type SoftwareData struct {
	DisplayName     string `json:"displayName"`     //`json:"software"`         // 应用名称
	DisplayVersion  string `json:"displayVersion"`  //`json:"version"`          // 应用版本
	InstallLocation string `json:"installLocation"` //`json:"install_location"` // 安装位置
	Publisher       string `json:"publisher"`       //`json:"vendor"`           // 发布者
	InstallDate     string `json:"installDate"`     //`json:"install_date"`     // 安装时间
}

// FirewallStatus 防火墙状态采集 => PC19
type FirewallStatus struct {
	FirewallName string `json:"firewallName"`
	Status       bool   `json:"status"`
}

// HttpPacketData HTTP状态采集 => PC20
type HttpPacketData struct {
	Request  HTTPPacket `json:"request"`
	Response HTTPPacket `json:"response"`
}

// SSHInfo SSH连接信息采集 => PC21
type SSHInfo struct {
	User      string `json:"user"`       // 用户名
	TTY       string `json:"tty"`        // 终端
	LoginTime string `json:"login_time"` // 登录时间
	ClientIP  string `json:"client_ip"`  // 客户端IP地址
}

// RDPLog RDP连接信息采集 => PC22
type RDPLog struct {
	Server string `json:"server"` // 服务器地址
	User   string `json:"user"`   // 用户名
}

// EventLogInfo 事件日志采集 => PC23
type EventLogInfo struct {
	TimeGenerated string `json:"timeGenerated"` // 事件生成时间
	EventID       int    `json:"eventId"`       // 事件ID
	EventType     string `json:"eventType"`     // 事件类型
	Source        string `json:"source"`        // 事件来源
	Message       string `json:"message"`       // 事件消息
	LogName       string `json:"logName"`       // 日志名称
}

// RequestData 用于存储HTTP请求头信息
type RequestData struct {
	Method               string `json:"method"`
	URL                  string `json:"url"`
	XForwardedFor        string `json:"x_forwarded_for"`
	Connection           string `json:"connection"`
	Host                 string `json:"host"`
	UserAgent            string `json:"user_agent"`
	Accept               string `json:"accept"`
	AcceptLanguage       string `json:"accept_language"`
	AcceptEncoding       string `json:"accept_encoding"`
	AccessControlMethod  string `json:"access_control_request_method"`
	AccessControlHeaders string `json:"access_control_request_headers"`
	Referer              string `json:"referer"`
	Origin               string `json:"origin"`
	XToken               string `json:"xtoken"`
}

// HTTPPacket HTTP请求和响应信息
type HTTPPacket struct {
	ReqType string
	Method  string
	Body    string
	URL     string
	Host    string
	Payload string
}

type MetricConfig struct {
	Name     string `json:"name"`
	Describe string `json:"describe"`
	Interval uint   `json:"interval"`
	Level    uint   `json:"level"`
	Enabled  bool   `json:"enabled"`
}

// IOCacheInfo 进程 IO 缓存信息
type IOCacheInfo struct {
	IoReadBytes  uint64 `json:"ioReadBytes"`  // 读取字节数;应记录进程读取的字节数。
	IoWriteBytes uint64 `json:"ioWriteBytes"` // 写入字节数;应记录进程写入的字节数。
	CollectedAt  uint64 `json:"collectedAt"`  // 采集时间;记录进程采集时间。
}

// Software 软件工具库信息
type Software struct {
	ID          uint      `json:"id"` // 主键ID
	CreatedAt   time.Time // 创建时间
	UpdatedAt   time.Time // 更新时间
	UUID        uuid.UUID `json:"uuid"`    //UUID
	Name        string    `json:"name"`    //名称
	Type        uint      `json:"type"`    //类型
	Path        string    `json:"path"`    //具体路径
	Size        string    `json:"size"`    //大小
	Comment     string    `json:"comment"` //描述
	Status      bool      `json:"status"`
	SizeDefault string    `json:"size_default" gorm:"-"` //大小
	Version     string    `json:"version"`               //版本
	OS          uint      `json:"os"`                    //适用系统
	Enabled     bool      `json:"enabled"`               //是否公开
	DownloadUrl string    `json:"download_url"`          //下载地址
}
