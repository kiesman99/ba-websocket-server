package websocket

import (
	"context"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/otel"
)

var (
	connGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "websocket_connections",
			Help: "Number of current connections to the websocket instances",
		},
		[]string{"endpoint"},
	)

	messageCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "websocket_messages",
		Help: "Number of messages that were received by the websocket instances",
	},
		[]string{"endpoint"},
	)
)

type Pool struct {
	Name       string
	Register   chan *Client
	Unregister chan *Client
	Clients    map[*Client]bool
	Broadcast  chan []byte
}

func NewPool(name string) *Pool {
	return &Pool{
		Name:       name,
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte),
	}
}

func (pool *Pool) Start(name string, ctx context.Context) {
	for {
		select {
		case client := <-pool.Register:
			// if the Register chan receives a new element
			// add the user to the pools clients list
			fmt.Println("New User Connected")
			pool.Clients[client] = true
			connGauge.WithLabelValues(pool.Name).Inc()
		case client := <-pool.Unregister:
			// if the Unregister chan receives a new element
			// remove the client from the pools client list
			delete(pool.Clients, client)
			connGauge.WithLabelValues(pool.Name).Dec()
		case message := <-pool.Broadcast:
			broadCastCtx, broadCastSpan := otel.Tracer(name).Start(ctx, "Broadcast")
			messageCount.WithLabelValues(pool.Name).Inc()
			for client := range pool.Clients {
				func() {
					_, writeSpan := otel.Tracer(name).Start(broadCastCtx, "Write")
					defer writeSpan.End()
					if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
						fmt.Println(err)
						writeSpan.End()
						return
					}
				}()
			}
			broadCastSpan.End()
		}
	}
}
