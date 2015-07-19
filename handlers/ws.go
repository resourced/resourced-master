package handlers

import (
	"net/http"
	"time"

	"github.com/gorilla/context"
	"github.com/gorilla/websocket"
	"github.com/resourced/resourced-master/wstrafficker"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// CheckOrigin == true allows CORS permission.
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func ApiWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	wsTraffickers := context.Get(r, "wsTraffickers").(*wstrafficker.WSTraffickers)
	wsTraffickers.SaveConnection(conn)

	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()

	// for {
	// 	select {
	// 	case message, ok := <-c.Send:
	// 		if !ok {
	// 			conn.SetWriteDeadline(time.Now().Add(writeWait))

	// 			c.Write(websocket.CloseMessage, []byte{})
	// 			return
	// 		}
	// 		if err := c.Write(websocket.TextMessage, message); err != nil {
	// 			return
	// 		}
	// 	case <-ticker.C:
	// 		if err := c.Write(websocket.PingMessage, []byte{}); err != nil {
	// 			return
	// 		}
	// 	}
	// }

	// for {
	// 	messageType, p, err := conn.ReadMessage()
	// 	if err != nil {
	// 		return
	// 	}

	// 	println("Received on Server Side: " + string(p))

	// 	err = conn.WriteMessage(messageType, p)
	// 	if err != nil {
	// 		return
	// 	}
	// }
}
