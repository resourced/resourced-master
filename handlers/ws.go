package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"github.com/resourced/resourced-master/dal"
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

func ApiWSAccessToken(w http.ResponseWriter, r *http.Request) {
	accessToken := mux.Vars(r)["id"]

	db := context.Get(r, "db").(*sqlx.DB)

	// Check if access token exists
	accessTokenRow, err := dal.NewAccessToken(db).GetByAccessToken(nil, accessToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if accessTokenRow == nil {
		err = errors.New("Unrecognized access token")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Upgrade connection to full duplex TCP connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	wsTraffickers := context.Get(r, "wsTraffickers").(*wstrafficker.WSTraffickers)

	wsTrafficker, err := wsTraffickers.SaveConnection(accessToken, conn)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()

	for {
		select {
		case message, ok := <-wsTrafficker.Chans.Send:
			if !ok {
				wsTrafficker.Write(websocket.CloseMessage, []byte{})
				return
			}
			if err := wsTrafficker.Write(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := wsTrafficker.Write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}
