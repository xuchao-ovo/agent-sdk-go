package task

import (
	"context"
	"encoding/json"
	"time"
)

type TaskBackJson struct {
	Worker   string `json:"worker"`
	TaskUUID string `json:"taskUUID"`
	Data     string `json:"data"`
	Message  string `json:"message"`
	Progress int    `json:"progress"`
	Success  int    `json:"success"`
}

func failOnError(err error, msg string) {
	if err != nil {
		global.GvaLog.Error(msg, zap.Error(err))
	}
}
