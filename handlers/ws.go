package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/context"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func setWSConnection(w http.ResponseWriter, r *http.Request, conn *websocket.Conn) {
	wsConnections := context.Get(r, "wsConnections").(map[string]*websocket.Conn)

	_, payloadJson, err := conn.ReadMessage()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var payload map[string]interface{}

	err = json.Unmarshal(payloadJson, &payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	agentIdInterface := payload["AgentID"]
	if agentIdInterface == nil {
		return
	}
	agentId := agentIdInterface.(string)

	wsConnections[agentId] = conn
}

func ApiWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	setWSConnection(w, r, conn)

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
