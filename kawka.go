package kawka

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"strconv"

	kafka "github.com/Shopify/sarama"
	websocket "github.com/gobwas/ws"
)

// MessageHandler ...
type MessageHandler func(data []byte) (topic string, content []byte, err error)

// Kawka ...
type Kawka struct {
	port     int
	producer kafka.SyncProducer
	brokers  []string
	handler  MessageHandler
	stream   chan []byte
}

// Message ...
type Message struct {
	ID   string          `json:"id,omitempty"`
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// New ...
func New(opts ...Option) *Kawka {
	wk := &Kawka{
		handler: defaultMessageHandler,
		stream:  make(chan []byte),
	}

	for _, op := range opts {
		if err := op(wk); err != nil {
			panic(err)
		}
	}

	if err := wk.initProducer(wk.brokers); err != nil {
		panic(err)
	}

	if wk.handler == nil {
		wk.handler = defaultMessageHandler
	}
	return wk
}

// Start ...
func (wk *Kawka) Start() error {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(wk.port))
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}
		_, err = websocket.Upgrade(conn)
		if err != nil {
			// handle error
		}

		go func() {
			defer conn.Close()

			for {
				header, err := websocket.ReadHeader(conn)
				if err != nil {
					// handle error
				}

				// TODO: use pool
				payload := make([]byte, header.Length)
				_, err = io.ReadFull(conn, payload)
				if err != nil {
					// handle error
				}
				if header.Masked {
					websocket.Cipher(payload, header.Mask, 0)
				}

				topic, content, err := wk.handler(payload)
				if err != nil {
					log.Printf("error in handler: %s\n", err.Error())
					continue
				}

				msg := &kafka.ProducerMessage{
					Topic: topic,
					Value: kafka.ByteEncoder(content),
				}

				_, _, err = wk.producer.SendMessage(msg)
				if err != nil {
					log.Printf("error on SendMessage: %s\n", err.Error())
					continue
				}
			}
		}()
	}
}

// Stop will stop Kawka processing data from websockets.
func (wk *Kawka) Stop() error {
	if err := wk.producer.Close(); err != nil {
		return err
	}
	return nil
}

func (wk *Kawka) initProducer(brokers []string) error {
	config := kafka.NewConfig()
	config.Version = kafka.V0_10_0_1
	config.Producer.Partitioner = kafka.NewManualPartitioner
	config.Producer.Return.Successes = true

	var err error
	wk.producer, err = kafka.NewSyncProducer(brokers, config)
	if err != nil {
		return err
	}
	return nil
}

func defaultMessageHandler(data []byte) (topic string, content []byte, err error) {
	var msg *Message
	if err := json.Unmarshal(data, msg); err != nil {
		return "", nil, err
	}
	return msg.Type, data, nil
}
