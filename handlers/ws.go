package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/resourced/resourced-master/libhttp"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func setWSConnection(w http.ResponseWriter, r *http.Request, conn *websocket.Conn) {
	cookieStore := context.Get(r, "cookieStore").(*sessions.CookieStore)

	session, _ := cookieStore.Get(r, "resourcedmaster-session")

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

	wsConnections := session.Values["wsConnections"].(map[string]*websocket.Conn)
	wsConnections[agentId] = conn

	err = session.Save(r, w)
	if err != nil {
		libhttp.HandleErrorJson(w, err)
		return
	}
}

func ApiWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	setWSConnection(w, r, conn)

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			return
		}

		println("Received on Server Side: " + string(p))

		err = conn.WriteMessage(messageType, p)
		if err != nil {
			return
		}
	}
}
