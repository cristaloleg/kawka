build:
	go build ./...

run:
	go run cmd/main.go

test:
	go test -v -race -count=1 ./...

install:
	go get github.com/gorilla/websocket
	go get github.com/Shopify/sarama
