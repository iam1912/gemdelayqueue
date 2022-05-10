package ws

import (
	"net/http"

	"github.com/iam1912/gemseries/gemdelayqueue/client/dqclient"
	"github.com/iam1912/gemseries/gemdelayqueue/log"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type Handler struct {
	Client *dqclient.Client
}

func New(client *dqclient.Client) Handler {
	return Handler{Client: client}
}

func (h Handler) Ws(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id := r.FormValue("id")
	if id == "" {
		wsjson.Write(r.Context(), conn, "invalid param")
		conn.Close(websocket.StatusUnsupportedData, "invalid param")
		return
	}

	wsConn := NewWsConn(id, conn)
	Broadcaster.login(wsConn)
	log.Infof("%s is register delayqueue\n", id)

	go wsConn.Write(r.Context())
	err = wsConn.Read(h.Client, r.Context())

	Broadcaster.logout(wsConn)
	log.Infof("%s is leave delayqueue\n", id)

	if err == nil {
		conn.Close(websocket.StatusNormalClosure, "")
	} else {
		log.Infof("read from client error:", err)
		conn.Close(websocket.StatusInternalError, "Read from client error")
	}
}
