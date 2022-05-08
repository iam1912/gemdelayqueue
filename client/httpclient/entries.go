package httpclient

type AddRequest struct {
	Topic string `json:"topic"`
	ID    string `json:"id"`
	Delay int64    `json:"delay"`
	TTR   int64    `json:"ttr"`
	Body  string `json:"body"`
}

type Response struct {
	Success bool        `json:"success"`
	Error   string      `json:"error"`
	ID      string      `json:"id"`
	Topic   string      `json:"topic"`
	Data    interface{} `json:"data"`
}

var (
	InvalidParam = "invalid params"
)
