package handler

import (
	"fmt"
	"net/http"

	ws "github.com/kiesman99/ba-websocket-server/internal/websocket"
)

func (handler Handler) DisplayHandler(pool *ws.Pool, w http.ResponseWriter, r *http.Request) {
	conn, err := ws.Upgrade(w, r)
	if err != nil {
		fmt.Fprintf(w, "%+V\n", err)
	}

	client := &ws.Client{
		Conn: conn,
		Pool: pool,
	}

	pool.Register <- client
	// client.Read(context.Background())
}
