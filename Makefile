build:
	go build ./...

run:
	go run cmd/main.go -brokers=127.0.0.1:9092 -topic=test

test:
	go test -v -race -count=1 ./...

install:
	go get github.com/gobwas/ws
	go get github.com/Shopify/sarama
