# Agent SDK Go

这是一个用于与Agent探针进行通信的Go语言SDK库。通过该SDK，外部应用可以方便地采集系统指标、执行远程任务，并处理来自探针的数据。

## 目录结构

agent-sdk-go/
├── api/ # 公共API接口定义
│ ├── client.go # 定义客户端接口
│ ├── client_impl.go # 客户端接口实现
│ └── options.go # 定义配置选项
│
├── pkg/ # 可重用的内部包
│ ├── metrics/ # 指标相关定义
│ │ ├── types/ # 指标数据类型
│ │ │ ├── system.go # 系统相关指标类型
│ │ │ ├── network.go # 网络相关指标类型
│ │ │ ├── process.go # 进程相关指标类型
│ │ │ └── ...
│ │ ├── processor.go # 指标数据处理
│ │ └── collector.go # 指标采集功能
│ │
│ ├── serial/ # 串口通信相关
│ │ ├── sendPacket.go # 发送包实现
│ │ └── socket.go # 串口连接实现
│ │
│ ├── task/ # 任务相关定义
│ │ └── task.go # 任务数据模型
│ │
│ ├── protocol/ # 通信协议相关
│ │ └── packet.go # 数据包定义
│ │
│ └── utils/ # 通用工具函数
│ └── task.go # 任务工具函数
│
├── internal/ # 内部实现
│ └── processor/ # 数据处理器实现
│ └── data_processor.go # 数据处理实现
│
├── examples/ # 使用示例
│ └── basic/ # 基础使用示例
│ └── main.go # 示例代码
│
├── global/ # 全局定义
│ └── model.go # 全局模型定义
│
├── go.mod
├── go.sum
└── README.md


## 快速开始

### 安装

```bash
go get github.com/xuchao-ovo/agent-sdk-go
```

### 使用示例

```go
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/xuchao-ovo/agent-sdk-go/api"
)

func main() {
	// 初始化SDK客户端
	client, err := api.NewClient(api.Options{
		SerialPort: "/dev/ttyS0",
		LogLevel:   "info",
	})
	if err != nil {
		log.Fatalf("初始化SDK失败: %v", err)
	}
	defer client.Close()

	// 注册指标数据处理回调
	client.RegisterMetricHandler(func(metricCode string, data interface{}) {
		fmt.Printf("收到指标数据: %s\n", metricCode)
	})

	// 开始监听数据
	go func() {
		if err := client.StartListening(); err != nil {
			log.Fatalf("监听失败: %v", err)
		}
	}()

	// 采集系统信息
	systemInfo, err := client.CollectMetric("PC1")
	if err != nil {
		log.Printf("采集系统信息失败: %v", err)
	} else {
		fmt.Printf("系统信息: %v\n", systemInfo)
	}

	// 阻塞主程序，持续运行一段时间
	time.Sleep(30 * time.Second)
}
```

更多使用示例可以查看 `examples` 目录。

## 文档

- [API文档](./docs/api.md)
- [指标文档](./docs/metrics.md)
- [使用指南](./docs/getting-started.md)

## 贡献

欢迎提交Issues或Pull Requests！

## 许可证

本项目采用 [LICENSE](./LICENSE) 许可证。
