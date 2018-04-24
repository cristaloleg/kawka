package main

import "github.com/cristaloleg/kawka"

func main() {
	brokers := []string{}
	kawka := kawka.New(brokers, myHandler)
	_ = kawka.Start()
}

func myHandler(data []byte) (topic string, content interface{}, err error) {
	return "", nil, nil
}
