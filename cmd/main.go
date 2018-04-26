package main

import (
	"flag"
	"os"
	"strings"

	"github.com/cristaloleg/kawka"
)

var (
	addr      = flag.String("addr", ":8080", "The address to bind to")
	brokers   = flag.String("brokers", os.Getenv("KAFKA_PEERS"), "The Kafka brokers to connect to, as a comma separated list")
	verbose   = flag.Bool("verbose", false, "Turn on Sarama logging")
	topic     = flag.String("topic", "test", "topic name")
	partition = flag.Int64("partition", 0, "partition")
	// certFile  = flag.String("certificate", "", "The optional certificate file for client authentication")
	// keyFile   = flag.String("key", "", "The optional key file for client authentication")
	// caFile    = flag.String("ca", "", "The optional certificate authority file for TLS client authentication")
	// verifySsl = flag.Bool("verify", false, "Optional verify ssl certificates chain")
)

func init() {
	flag.Parse()

	if *brokers == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func main() {
	brokerList := strings.Split(*brokers, ",")
	if len(brokerList) == 0 {
		panic("brokers are unavailable")
	}

	kawka := kawka.New(
		kawka.WithBrokers(brokerList),
		kawka.WithPort(5986),
		kawka.WithHandler(func(data []byte) (string, []byte, error) {
			return *topic, data, nil
		}),
	)
	_ = kawka.Start()
}
