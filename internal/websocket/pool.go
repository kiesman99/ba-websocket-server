package websocket

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	"github.com/gorilla/websocket"
	"github.com/kiesman99/ba-websocket-server/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
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
	Logger     *zap.SugaredLogger
}

func NewPool(name string, logger *zap.SugaredLogger) *Pool {
	return &Pool{
		Name:       name,
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte),
		Logger:     logger,
	}
}

func (pool *Pool) Start(ctx context.Context) {
	for {
		select {
		case client := <-pool.Register:
			registerClient(pool, client)
		case client := <-pool.Unregister:
			unRegisterClient(pool, client)
		case message := <-pool.Broadcast:
			go broadcastMessage(pool, message, ctx)
		}
	}
}

func registerClient(pool *Pool, client *Client) {
	pool.Logger.Infow("New User Connected", "pool", pool.Name)
	// if the Register chan receives a new element
	// add the user to the pools clients list
	pool.Clients[client] = true
	connGauge.WithLabelValues(pool.Name).Inc()
}

func unRegisterClient(pool *Pool, client *Client) {
	pool.Logger.Infow("User Unregistered", "pool", pool.Name)
	// if the Unregister chan receives a new element
	// remove the client from the pools client list
	delete(pool.Clients, client)
	connGauge.WithLabelValues(pool.Name).Dec()
}

func broadcastMessage(pool *Pool, message []byte, ctx context.Context) {
	pool.Logger.Infow("Broadcasting message",
		// "message", string(message),
		"pool", pool.Name,
		"clientCount", len(pool.Clients),
		"payload", utils.Truncate(string(message), 10),
	)

	broadCastCtx, broadCastSpan := otel.Tracer("ba-ws-server").Start(ctx, fmt.Sprintf("Broadcast %s", pool.Name))
	messageCount.WithLabelValues(pool.Name).Inc()
	for client := range pool.Clients {
		func() {
			_, writeSpan := otel.Tracer("ba-ws-server").Start(broadCastCtx, fmt.Sprintf("Write %s", pool.Name))
			defer writeSpan.End()
			if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				pool.Logger.Error(err)
				writeSpan.End()
				return
			}
		}()
	}
	broadCastSpan.End()
}
