package kawka

import (
	"encoding/json"
	"io"
	"log"
	"net"

	kafka "github.com/Shopify/sarama"
	"github.com/gobwas/ws"
)

// MessageHandler ...
type MessageHandler func(data []byte) (topic string, content interface{}, err error)

// Kawka ...
type Kawka struct {
	producer kafka.SyncProducer

	handler MessageHandler
	stream  chan []byte
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
		producer: p,
		handler:  handler,
		stream:   make(chan []byte),
	}

	if wk.handler == nil {
		wk.handler = defaultMessageHandler
	}
	return wk
}

// Start ...
func (wk *Kawka) Start() error {
	ln, err := net.Listen("tcp", "localhost:5985")
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}
		_, err = ws.Upgrade(conn)
		if err != nil {
			// handle error
		}

		go func() {
			defer conn.Close()

			for {
				header, err := ws.ReadHeader(conn)
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
					ws.Cipher(payload, header.Mask, 0)
				}

				topic, content, err := wk.handler(payload)
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
	}
	return nil
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
