package core

import (
	"context"
	"log"
	"server/certificates"
	"server/connections"
	"shared"
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

	tlsConfig, err := certificates.GenerateTLSConfig()
	if err != nil {
		logger.Fatalf("Failed to generate TLS config: %v", err)
	}

	logger.Println("TLS config generated")

	quicConfig := &quic.Config{
		MaxIdleTimeout:  time.Second * shared.MaxIdleTimeout,
		KeepAlivePeriod: time.Second * shared.KeepAlivePeriod,
	}

	logger.Println("Initializing pub server")

	go connections.InitPubServer(ctx, tlsConfig, quicConfig)

	logger.Printf("Pub server initialization complete, listening on: %v", shared.PortPub)
	logger.Println("Initializing sub server")

	go connections.InitSubServer(ctx, tlsConfig, quicConfig)

	logger.Printf("Sub server initialization complete, listening on: %v", shared.PortSub)
}
