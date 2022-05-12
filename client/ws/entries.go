package ws

import (
	"context"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

const (
	RequestCommand = iota
	RequestHeartBeat
	ResponseHeartBeat
)

type Request struct {
	RequestType int    `json:"request_type"`
	Message     string `json:"message"`
}

type Response struct {
	ResponseType int         `json:"response_type"`
	Success      bool        `json:"success"`
	Error        string      `json:"error"`
	ID           string      `json:"id"`
	Topic        string      `json:"topic"`
	Data         interface{} `json:"data"`
	Message      string      `json:"meesage"`
}

func RenderErrorResponse(c *websocket.Conn, ctx context.Context, err string) {
	resp := &Response{
		ResponseType: RequestCommand,
		Success:      false,
		Error:        err,
	}
	wsjson.Write(ctx, c, resp)
	c.Close(websocket.StatusUnsupportedData, err)
}
