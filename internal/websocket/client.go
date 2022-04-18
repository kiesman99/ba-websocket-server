package websocket

import (
	"context"
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID   string
	Conn *websocket.Conn
	Pool *Pool
}

type Message struct {
	Type int    `json:"type"`
	Body string `json:"body"`
}

func (c *Client) Read(ctx context.Context) {
	defer func() {
		c.Pool.Unregister <- c
		c.Conn.Close()
	}()

	for {
		messageType, p, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		message := Message{Type: messageType, Body: string(p)}
		json, err := json.Marshal(&message)
		if err != nil {
			log.Println(err)
			return
		}

		c.Pool.Broadcast <- json
	}
}

func (c *Client) EchoChamber(ctx context.Context) {
	defer func() {
		c.Pool.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		// c.Pool.Broadcast <- message
	}
}
