package utils

import "fmt"

// LogError 用于输出错误日志
func LogError(format string, args ...interface{}) {
	fmt.Printf("ERROR: "+format+"\n", args...)
}
