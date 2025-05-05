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
	Error  string  `json:"error"`
}

// структура задачи
type Task struct {
	ID string `json:"id"`

	Arg1      float64 `json:"arg1"`
	Arg2      float64 `json:"arg2"`
	Operation string  `json:"operation"`

	Result float64 `json:"result,omitempty"`
	Error  string  `json:"error,omitempty"`
}
