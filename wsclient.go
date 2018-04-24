package kawka

import (
	"bytes"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	pongWait       = 60 * time.Second
	maxMessageSize = 512
)

type wsClient struct {
	hub    *wsHub
	conn   *websocket.Conn
	stream chan<- []byte
}

func newWsClient(conn *websocket.Conn, hub *wsHub, stream chan<- []byte) *wsClient {
	client := &wsClient{
		conn:   conn,
		hub:    hub,
		stream: stream,
	}
	return client
}

func (c *wsClient) readPump() {
	defer func() {
		c.hub.Disconnect(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		msgType, msg, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("error: %v", err)
			break
		}

		switch msgType {
		case websocket.TextMessage:
			msg = bytes.TrimSpace(bytes.Replace(msg, []byte{'\n'}, []byte{' '}, -1))
			c.stream <- msg

		case websocket.BinaryMessage:
		case websocket.CloseMessage:
		case websocket.PingMessage:
		case websocket.PongMessage:
		default:
		}
	}
}
