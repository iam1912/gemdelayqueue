package ws

import "nhooyr.io/websocket"

type broadcast struct {
	users    map[string]*WsConn
	register chan *WsConn
	leave    chan *WsConn
}

var Broadcaster = &broadcast{
	users:    make(map[string]*WsConn),
	register: make(chan *WsConn),
	leave:    make(chan *WsConn),
}

func (b *broadcast) Run() {
	for {
		select {
		case wsConn := <-b.register:
			b.users[wsConn.ID] = wsConn
		case wsConn := <-b.leave:
			delete(b.users, wsConn.ID)
			wsConn.conn.Close(websocket.StatusAbnormalClosure, "connection closed")
			wsConn.close()
		}
	}
}

func (b *broadcast) login(user *WsConn) {
	b.register <- user
}

func (b *broadcast) logout(user *WsConn) {
	b.leave <- user
}
