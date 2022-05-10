package ws

import (
	"context"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type Response struct {
	Success bool        `json:"success"`
	Error   string      `json:"error"`
	ID      string      `json:"id"`
	Topic   string      `json:"topic"`
	Data    interface{} `json:"data"`
}

func RenderErrorResponse(c *websocket.Conn, ctx context.Context, err string) {
	resp := &Response{
		Success: false,
		Error:   err,
	}
	wsjson.Write(ctx, c, resp)
	c.Close(websocket.StatusUnsupportedData, err)
}
