package router

import (
	"log"
	"net/http"

	"github.com/iam1912/gemseries/gemdelayqueue/client/dqclient"
	"github.com/iam1912/gemseries/gemdelayqueue/client/httpclient"
	"github.com/iam1912/gemseries/gemdelayqueue/client/ws"
	"github.com/iam1912/gemseries/gemdelayqueue/config"
)

func Serve() {
	c := config.MustGetConfig()
	client, err := dqclient.New(c)
	if err != nil {
		panic(err)
	}
	wsHandler := ws.New(client)
	go ws.Broadcaster.Run()
	http.HandleFunc("/dq/ws", wsHandler.Ws)

	dqHandler := httpclient.New(client)
	http.HandleFunc("/dq/add", dqHandler.Add)
	http.HandleFunc("/dq/pop", dqHandler.Pop)
	http.HandleFunc("/dq/finish", dqHandler.Finish)
	http.HandleFunc("/dq/delete", dqHandler.Delete)
	http.HandleFunc("/dq/info", dqHandler.GetInfo)

	log.Printf("localhost%s\n", c.Port)
	http.ListenAndServe(c.Port, nil)
}
