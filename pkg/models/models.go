package models

const (
	StatusPending   = "pending"
	StatusCompleted = "completed"
	StatusFailed    = "failed"

	StatusNeedToSend = "need to send"
)

// структура для состояния выражения
type Expression struct {
	ID     string  `json:"id"`
	Status string  `json:"status"`
	Result float64 `json:"result"`
}

// структура задачи
type Task struct {
	ID     string `json:"id"`
	Status string `json:"status"`

	Arg1           float64 `json:"arg1"`
	Arg2           float64 `json:"arg2"`
	Operation      string  `json:"operation"`
	Operation_time int     `json:"operation_time"`

	Result float64 `json:"result,omitempty"`
	Error  string  `json:"error,omitempty"`
}
