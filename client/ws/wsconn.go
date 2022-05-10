package ws

import (
	"context"
	"errors"
	"io"

	"github.com/iam1912/gemseries/gemdelayqueue/client/dqclient"
	"github.com/iam1912/gemseries/gemdelayqueue/utils"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type WsConn struct {
	ID             string
	commandChannel chan *Response
	conn           *websocket.Conn
}

func NewWsConn(id string, conn *websocket.Conn) *WsConn {
	return &WsConn{
		ID:             id,
		commandChannel: make(chan *Response),
		conn:           conn,
	}
}

func (ws *WsConn) Write(ctx context.Context) {
	for cmd := range ws.commandChannel {
		wsjson.Write(ctx, ws.conn, cmd)
	}
}

func (ws *WsConn) Read(client *dqclient.Client, ctx context.Context) error {
	var (
		command map[string]string
		err     error
	)
	for {
		err = wsjson.Read(ctx, ws.conn, &command)
		if err != nil {
			var closeErr websocket.CloseError
			if errors.As(err, &closeErr) {
				return nil
			} else if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		ws.HandleCommand(client, command, ctx)
	}
}

func (ws *WsConn) HandleCommand(client *dqclient.Client, command map[string]string, ctx context.Context) error {
	var invalidErr = errors.New("invalid param")

	switch command["command"] {
	case "add":
		id, topic, delay, ttr, maxTries, body := command["id"], command["topic"], command["delay"], command["ttr"], command["max_tries"], command["body"]
		if id == "" || topic == "" || delay == "" || ttr == "" || body == "" {
			RenderErrorResponse(ws.conn, ctx, "invalid param")
			return invalidErr
		}
		err := client.Add(ctx, id, topic, utils.StringToInt64(delay), utils.StringToInt64(ttr), utils.StringToInt(maxTries), body)
		if err != nil {
			RenderErrorResponse(ws.conn, ctx, err.Error())
			return err
		}
		ws.sendChannel("", "", "success add")
	case "pop":
		topic := command["topic"]
		if topic == "" {
			RenderErrorResponse(ws.conn, ctx, "invalid param")
			return invalidErr
		}
		job, err := client.RPop(ctx, topic)
		if err != nil {
			RenderErrorResponse(ws.conn, ctx, err.Error())
			return err
		}
		ws.sendChannel(job.ID, job.Topic, job.Body)
	case "delete":
		id := command["id"]
		if id == "" {
			RenderErrorResponse(ws.conn, ctx, "invalid param")
			return invalidErr
		}
		err := client.Deleted(ctx, id)
		if err != nil {
			RenderErrorResponse(ws.conn, ctx, err.Error())
			return err
		}
		ws.sendChannel("", "", "success delete")
	case "finish":
		id := command["id"]
		if id == "" {
			RenderErrorResponse(ws.conn, ctx, "invalid param")
			return invalidErr
		}
		err := client.Finish(ctx, id)
		if err != nil {
			RenderErrorResponse(ws.conn, ctx, err.Error())
			return err
		}
		ws.sendChannel("", "", "success finish")
	}
	return nil
}

func (ws *WsConn) sendChannel(id string, topic string, data interface{}) {
	resp := &Response{
		Success: true,
		ID:      id,
		Topic:   topic,
		Data:    data,
	}
	ws.commandChannel <- resp
}

func (ws *WsConn) close() {
	close(ws.commandChannel)
}
