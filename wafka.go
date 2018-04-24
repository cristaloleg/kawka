package kawka

import (
	"encoding/json"
	"log"
	"net/http"

	kafka "github.com/Shopify/sarama"
	"github.com/gorilla/websocket"
)

// MessageHandler ...
type MessageHandler func(data []byte) (topic string, content interface{}, err error)

// Kawka ...
type Kawka struct {
	producer   kafka.SyncProducer
	wsUpgrader websocket.Upgrader
	handler    MessageHandler
	stream     chan []byte
}

// Message ...
type Message struct {
	ID   string          `json:"id"`
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// New ...
func New(brokers []string, handler MessageHandler, opts ...interface{}) *Kawka {
	p, err := newSyncProducer(brokers)
	if err != nil {
		panic(err)
	}
	wk := &Kawka{
		producer:   p,
		wsUpgrader: websocket.Upgrader{},
		handler:    handler,
		stream:     make(chan []byte),
	}

	if wk.handler == nil {
		wk.handler = defaultMessageHandler
	}
	return wk
}

// Start ...
func (wk *Kawka) Start() error {
	go func() {
		for {
			data := <-wk.stream

			topic, content, err := wk.handler(data)
			if err != nil {
				log.Printf("error on SendMessage: %s\n", err.Error())
				continue
			}

			msg := &kafka.ProducerMessage{
				Topic: topic,
				Value: kafka.StringEncoder(content.(string)),
			}

			_, _, err = wk.producer.SendMessage(msg)
			if err != nil {
				log.Printf("error on SendMessage: %s\n", err.Error())
				continue
			}
		}
	}()

	http.HandleFunc("/ws", wk.wsHandler)

	return http.ListenAndServe(":5987", nil)
}

// Stop will stop Kawka processing data from websockets.
func (wk *Kawka) Stop() error {
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

	client := newWsClient(conn, wk.stream)

	go client.readPump()
}
