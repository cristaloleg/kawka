package kawka

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	kafka "github.com/Shopify/sarama"
	"github.com/gorilla/websocket"
)

// MessageHandler ...
type MessageHandler func(data []byte) (topic string, content interface{}, err error)

type dataHandler func(data []byte) error

// Kawka ...
type Kawka struct {
	producer   kafka.SyncProducer
	aproducer  kafka.AsyncProducer
	hub        *wsHub
	wsUpgrader websocket.Upgrader
	handler    MessageHandler
	isAsync    bool
}

// Message ...
type Message struct {
	ID   string          `json:"id"`
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// New ...
func New(brokers []string, handler MessageHandler, opts ...interface{}) *Kawka {
	wk := &Kawka{
		hub:        newHub(),
		wsUpgrader: websocket.Upgrader{},
		handler:    handler,
	}

	if wk.isAsync {
		p, err := newAsyncProducer(brokers)
		if err != nil {
			panic(err)
		}
		wk.aproducer = p
	} else {
		p, err := newSyncProducer(brokers)
		if err != nil {
			panic(err)
		}
		wk.producer = p
	}

	if wk.handler == nil {
		wk.handler = defaultMessageHandler
	}
	return wk
}

// Start ...
func (wk *Kawka) Start() error {
	http.HandleFunc("/ws", wk.wsHandler)

	return http.ListenAndServe(":5987", nil)
}

// Stop will stop Kawka processing data from websockets.
func (wk *Kawka) Stop() error {
	return nil
}

func (wk *Kawka) process(data []byte) error {
	topic, content, err := wk.handler(data)
	if err != nil {
		log.Printf("error on SendMessage: %s\n", err.Error())
		return err
	}

	msg := &kafka.ProducerMessage{
		Topic: topic,
		Value: kafka.StringEncoder(content.(string)),
	}

	if wk.isAsync {
		wk.aproducer.Input() <- msg
		<-wk.aproducer.Successes()
	} else {
		_, _, err = wk.producer.SendMessage(msg)
	}

	if err != nil {
		log.Printf("error on SendMessage: %s\n", err.Error())
		return err
	}
	return nil
}

func newSyncProducer(brokers []string) (kafka.SyncProducer, error) {
	config := kafka.NewConfig()
	config.ChannelBufferSize = 1
	config.Version = kafka.V0_10_0_1
	config.Producer.Return.Successes = true

	producer, err := kafka.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}
	return producer, nil
}

func newAsyncProducer(brokers []string) (kafka.AsyncProducer, error) {
	config := kafka.NewConfig()
	config.ChannelBufferSize = 1
	config.Version = kafka.V0_10_0_1
	config.Producer.Return.Successes = true

	producer, err := kafka.NewAsyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}
	return producer, nil
}

func defaultMessageHandler(data []byte) (topic string, content interface{}, err error) {
	var msg *Message
	if err := json.Unmarshal(data, msg); err != nil {
		return "", nil, err
	}
	return msg.Type, msg, nil
}

func (wk *Kawka) wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := wk.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		w.Write([]byte(err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	client := newWsClient(conn, wk.hub, wk.process)

	wk.hub.Connect(client)

	go client.readPump()
}

const (
	pongWait       = 60 * time.Second
	maxMessageSize = 512
)

type wsClient struct {
	hub     *wsHub
	conn    *websocket.Conn
	process dataHandler
}

func newWsClient(conn *websocket.Conn, hub *wsHub, process dataHandler) *wsClient {
	client := &wsClient{
		conn:    conn,
		hub:     hub,
		process: process,
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
			c.process(msg)

		case websocket.BinaryMessage:
		case websocket.CloseMessage:
		case websocket.PingMessage:
		case websocket.PongMessage:
		default:
		}
	}
}

type wsHub struct {
	sync.RWMutex
	clients map[*wsClient]struct{}
}

func newHub() *wsHub {
	h := &wsHub{
		clients: make(map[*wsClient]struct{}),
	}
	return h
}

func (h *wsHub) Connect(client *wsClient) {
	h.Lock()
	h.clients[client] = struct{}{}
	h.Unlock()
}

func (h *wsHub) Disconnect(client *wsClient) {
	h.Lock()
	delete(h.clients, client)
	h.Unlock()
}
