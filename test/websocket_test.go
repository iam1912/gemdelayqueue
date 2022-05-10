package test

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/iam1912/gemseries/gemdelayqueue/client/router"
	"github.com/iam1912/gemseries/gemdelayqueue/client/ws"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func TestWebSocket(t *testing.T) {
	var wg sync.WaitGroup
	go func() {
		router.Serve()
	}()
	time.Sleep(time.Second)
	wg.Add(2)
	url := "ws://localhost:9091//dq/ws?id=129"
	conn, _, err := websocket.Dial(context.Background(), url, nil)
	if err != nil {
		log.Fatal(err)
	}
	body := map[string]string{
		"command":   "add",
		"id":        "129",
		"topic":     "delay:order",
		"delay":     "30",
		"ttr":       "30",
		"body":      "test3",
		"max_tries": "5",
	}
	err = wsjson.Write(context.Background(), conn, body)
	resp := &ws.Response{}
	err = wsjson.Read(context.Background(), conn, resp)
	if err != nil {
		t.Error(err)
	} else {
		fmt.Println(resp)
		wg.Done()
	}
	body = map[string]string{
		"command": "info",
		"id":      "129",
	}
	err = wsjson.Write(context.Background(), conn, body)
	err = wsjson.Read(context.Background(), conn, resp)
	if err != nil {
		t.Error(err)
	} else {
		fmt.Println(resp)
		wg.Done()
	}
	wg.Wait()
}
