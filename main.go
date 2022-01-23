package main

import (
	"fmt"
	"log"
	"net/http"

	ws "github.com/kiesman99/ba-websocket-server/pkg/websocket"
	// "github.com/kiesman99/ba-websocket-server/pkg/websocket"
)

func telegrafHandler(pool *ws.Pool, w http.ResponseWriter, r *http.Request) {
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
	client.EchoChamber()
}

func displayHandler(pool *ws.Pool, w http.ResponseWriter, r *http.Request) {
	conn, err := ws.Upgrade(w, r)
	if err != nil {
		fmt.Fprintf(w, "%+V\n", err)
	}

	client := &ws.Client{
		Conn: conn,
		Pool: pool,
	}

	pool.Register <- client
	client.Read()
}

func main() {

	fmt.Println("BA Websocket Server")

	telegrafPool := ws.NewPool()
	go telegrafPool.Start()
	http.HandleFunc("/telegraf", func(rw http.ResponseWriter, r *http.Request) {
		telegrafHandler(telegrafPool, rw, r)
	})

	displayPool := ws.NewPool()
	go displayPool.Start()
	http.HandleFunc("/display", func(rw http.ResponseWriter, r *http.Request) {
		displayHandler(displayPool, rw, r)
	})

	log.Fatal(http.ListenAndServe(":3210", nil))
}
