package httpclient

import (
	"encoding/json"
	"net/http"
)

func CheckCommand(r *http.Request) string {
	cmd := r.FormValue("command")
	if cmd == "add" || cmd == "pop" || cmd == "finish" || cmd == "delete" || cmd == "info" {
		return cmd
	}
	return ""
}

func RenderSuccessJSON(w http.ResponseWriter, id, topic string, data interface{}) {
	result, _ := json.Marshal(Response{
		Success: true,
		ID:      id,
		Topic:   topic,
		Data:    data,
	})
	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}

func RenderFailureJSON(w http.ResponseWriter, err string) {
	result, _ := json.Marshal(Response{
		Success: false,
		Error:   err,
	})
	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}
