package websocket

import (
	"fmt"

	"github.com/gorilla/websocket"
)

type Pool struct {
	Register   chan *Client
	Unregister chan *Client
	Clients    map[*Client]bool
	Broadcast  chan []byte
}

func NewPool() *Pool {
	return &Pool{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte),
	}
}

func (pool *Pool) Start() {
	for {
		select {
		case client := <-pool.Register:
			// if the Register chan receives a new element
			// add the user to the pools clients list
			fmt.Println("New User Connected")
			pool.Clients[client] = true
		case client := <-pool.Unregister:
			// if the Unregister chan receives a new element
			// remove the client from the pools client list
			delete(pool.Clients, client)
		case message := <-pool.Broadcast:
			fmt.Println("Sending message to all clients in pool")
			for client := range pool.Clients {
				if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
					fmt.Println(err)
					return
				}

				// if err := client.Conn.WriteJSON(message); err != nil {
				// 	fmt.Println(err)
				// 	return
				// }
			}
		}
	}
}
