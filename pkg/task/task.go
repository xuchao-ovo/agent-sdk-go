package task

type TaskBackJson struct {
	Worker   string `json:"worker"`
	TaskUUID string `json:"taskUUID"`
	Data     string `json:"data"`
	Message  string `json:"message"`
	Progress int    `json:"progress"`
	Success  int    `json:"success"`
}
