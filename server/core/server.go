package core

import (
	"context"
	"log"
	"server/certificates"
	"server/connections"
	"shared/streams"
	"time"

	"github.com/quic-go/quic-go"
)

var RunServer = func(ctx context.Context) {
	logger := log.Default()
	connections.MessageChan = make(chan streams.MessageStream, 1)
	connections.NoSubsChan = make(chan streams.MessageStream, 1)
	connections.NewSubChan = make(chan streams.MessageStream, 1)

	logger.Println("Generating TLS config")

	tlsConfig := certificates.GenerateTLSConfig()

	logger.Println("TLS config generated")

	quicConfig := &quic.Config{
		MaxIdleTimeout:  time.Second * streams.MaxIdleTimeout,
		KeepAlivePeriod: time.Second * streams.KeepAlivePeriod,
	}

	logger.Println("Initializing pub server")

	go connections.InitPubServer(ctx, tlsConfig, quicConfig)

	logger.Printf("Pub server initialization complete, listening on: %v", streams.PortPub)
	logger.Println("Initializing sub server")

	go connections.InitSubServer(ctx, tlsConfig, quicConfig)

	logger.Printf("Sub server initialization complete, listening on: %v", streams.PortSub)
}
