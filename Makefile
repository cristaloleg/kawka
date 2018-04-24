build:
	go build ./...

run:
	go run cmd/main.go

test:
	go test -v -race -count=1 ./...