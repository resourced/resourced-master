package wstrafficker

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

func NewWSTrafficker(client *websocket.Conn) *WSTrafficker {
	ws := &WSTrafficker{}
	ws.Chans.Send = make(chan []byte)
	ws.Chans.Receive = make(chan []byte)

	ws.Client = client

	return ws
}

type WSTrafficker struct {
	Chans struct {
		Send    chan []byte
		Receive chan []byte
	}
	Client *websocket.Conn
}

func NewWSTraffickers() *WSTraffickers {
	wss := &WSTraffickers{}
	wss.ByAgentID = make(map[string]*WSTrafficker)
	wss.ByAccessToken = make(map[string]*WSTrafficker)

	return wss
}

type WSTraffickers struct {
	ByAgentID     map[string]*WSTrafficker
	ByAccessToken map[string]*WSTrafficker
}

func (wss *WSTraffickers) SaveConnection(conn *websocket.Conn) error {
	_, payloadJson, err := conn.ReadMessage()
	if err != nil {
		return err
	}

	var payload map[string]interface{}

	err = json.Unmarshal(payloadJson, &payload)
	if err != nil {
		return err
	}

	agentIdInterface, agentIdExists := payload["AgentID"]
	if agentIdExists {
		agentId := agentIdInterface.(string)
		wss.ByAgentID[agentId] = NewWSTrafficker(conn)
	}

	accessTokenInterface, accessTokenExists := payload["AccessToken"]
	if accessTokenExists {
		accessToken := accessTokenInterface.(string)
		wss.ByAccessToken[accessToken] = NewWSTrafficker(conn)
	}

	return nil
}
