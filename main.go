package main

import (
	"fmt"
	"log"
	"net/http"

	ws "github.com/kiesman99/ba-websocket-server/pkg/websocket"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	fmt.Println("Starting BA Websocket Server...")

	telegrafPool := ws.NewPool("telegraf")
	go telegrafPool.Start()
	http.HandleFunc("/telegraf", func(rw http.ResponseWriter, r *http.Request) {
		telegrafHandler(telegrafPool, rw, r)
	})

	displayPool := ws.NewPool("display")
	go displayPool.Start()
	http.HandleFunc("/display", func(rw http.ResponseWriter, r *http.Request) {
		displayHandler(displayPool, rw, r)
	})

	go func() {
		fmt.Println("Starting Websocket Severs...")
		log.Fatal(http.ListenAndServe(":3210", nil))
	}()

	go func() {
		fmt.Println("Starting Prometheus Severs...")
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe(":2112", nil))
	}()

	for {
	}
}
