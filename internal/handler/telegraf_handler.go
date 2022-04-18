package handler

import (
	"fmt"
	"log"
	"net/http"

	ws "github.com/kiesman99/ba-websocket-server/internal/websocket"
)

func (handler Handler) TelegrafHandler(pool *ws.Pool, w http.ResponseWriter, r *http.Request) {
	// TODO: Implement real endpoint for telegraf
	// currently this is only a echo endpoint
	conn, err := ws.Upgrade(w, r)
	if err != nil {
		fmt.Fprintf(w, "%+V\n", err)
	}

	client := &ws.Client{
		Conn: conn,
		Pool: pool,
	}

	pool.Register <- client

	defer func() {
		pool.Unregister <- client
		conn.Close()
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
	}
}
