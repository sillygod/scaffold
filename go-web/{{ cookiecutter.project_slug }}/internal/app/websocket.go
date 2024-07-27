package app

import (
	"{{ cookiecutter.project_name }}/events"
	"bytes"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lerenn/asyncapi-codegen/pkg/extensions"
	"github.com/lerenn/asyncapi-codegen/pkg/extensions/brokers"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var _ extensions.BrokerController = (*Client)(nil)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	context context.Context
	hub     *Hub

	// The websocket connection.
	conn *websocket.Conn

	submap map[string]*extensions.BrokerChannelSubscription

	// Buffered channel of outbound messages.
	send chan []byte

	//
	ctrl *events.AppController
}

func NewWSClient(hub *Hub, conn *websocket.Conn, bufferSize int) *Client {
	client := &Client{
		hub:     hub,
		context: context.Background(),
		conn:    conn,
		send:    make(chan []byte, bufferSize),
		submap:  make(map[string]*extensions.BrokerChannelSubscription),
	}

	client.conn.SetReadLimit(maxMessageSize)
	client.conn.SetReadDeadline(time.Now().Add(pongWait))
	client.conn.SetPongHandler(func(string) error { client.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	client.hub.register <- client

	return client
}

func (c *Client) BindAppController(ctrl *events.AppController) {
	c.ctrl = ctrl
}

// TODO: verify the AppController implementation

// Publish push a message to the broker
func (c *Client) Publish(_ context.Context, channel string, bm extensions.BrokerMessage) error {

	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
	}()

	select {
	case _, ok := <-c.send:
		c.conn.SetWriteDeadline(time.Now().Add(writeWait))
		if !ok {
			// The hub closed the channel.
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return nil
		}

		w, err := c.conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return nil
		}

		w.Write(bm.Payload)
		w.Write(newline)

		if err := w.Close(); err != nil {
			return nil
		}
	case <-ticker.C:
		c.conn.SetWriteDeadline(time.Now().Add(writeWait))
		if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			return nil
		}
	}
	return nil
}

func (c *Client) readPump() {

	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		payload := map[string]interface{}{}
		_ = json.Unmarshal(message, &payload)

		// TODO: error handling?
		sub := c.submap[payload["event"].(string)]

		sub.TransmitReceivedMessage(extensions.NewAcknowledgeableBrokerMessage(
			extensions.BrokerMessage{
				Payload: message,
			},
			NoopAcknowledgementHandler{},
		))
		c.hub.broadcast <- message
	}
}

// Subscribe gets a message from the broker
func (c *Client) Subscribe(ctx context.Context, channel string) (extensions.BrokerChannelSubscription, error) {
	// In this case, it may work like a frontend client that receives messages
	sub := extensions.NewBrokerChannelSubscription(
		make(chan extensions.AcknowledgeableBrokerMessage, brokers.BrokerMessagesQueueSize),
		make(chan any, 1),
	)

	// TODO: fix bugs -- distinct dispatch message for each channel

	c.submap[channel] = &sub
	go c.readPump()

	sub.WaitForCancellationAsync(func() {
		c.Close()
	})

	return sub, nil

}

func (c *Client) Context() context.Context {
	return c.context
}

func (c *Client) Close() error {
	c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))

	c.ctrl.Close(c.context)
	return nil
}

type Hub struct {
	// Registered clients.
	clients map[*Client]bool
	// Inbound messages from the clients.
	broadcast chan []byte
	// Register requests from the clients.
	register chan *Client
	// Unregister requests from clients.
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// NOTE: do I need a ack mechanism?
var _ extensions.BrokerAcknowledgment = (*NoopAcknowledgementHandler)(nil)

type NoopAcknowledgementHandler struct {
}

// AckMessage acknowledges the message.
func (k NoopAcknowledgementHandler) AckMessage() {

}

// NakMessage negatively acknowledges the message.
func (k NoopAcknowledgementHandler) NakMessage() {

}
