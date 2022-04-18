package handler

import (
	"encoding/json"
	"net/http"

	"github.com/kiesman99/ba-websocket-server/pkg/utils"

	ws "github.com/kiesman99/ba-websocket-server/internal/websocket"
)

type MessageType int

const (
	DisplayMessage MessageType = iota
	TelegrafMessage
)

type Message struct {
	Type    MessageType `json:"type"`
	Message string      `json:"message"`
}

func (handler Handler) DistributionHandler(displayPool *ws.Pool, telegrafPool *ws.Pool, w http.ResponseWriter, r *http.Request) {
	suggarLogger := handler.Logger.Sugar()
	conn, err := ws.Upgrade(w, r)
	if err != nil {
		suggarLogger.Errorw("Error Upgrading distributionHandler", "err", err)
	}

	defer func() {
		suggarLogger.Info("Disconnect")
		conn.Close()
	}()

	suggarLogger.Info("New connection")

	for {
		_, payload, err := conn.ReadMessage()
		if err != nil {
			suggarLogger.Errorf("Error reading message", err)
			return
		}

		var msg Message
		if err := json.Unmarshal(payload, &msg); err != nil {
			suggarLogger.Warnw("Message invalid format", "payload", string(payload))
			suggarLogger.Errorf("Error Unmarshalling", err)
			continue
		}

		suggarLogger.Infow("Distributer received message", "type", msg.Type, "payload", utils.Truncate(msg.Message, 10))

		switch msg.Type {
		case DisplayMessage:
			displayPool.Broadcast <- payload
		case TelegrafMessage:
			telegrafPool.Broadcast <- []byte(msg.Message)
		}
	}
}
