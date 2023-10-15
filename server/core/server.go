package core

import (
	"context"
	"errors"
	"log"
	"server/certificates"
	"server/connections"
	"shared"
	"shared/streams"
	"time"

	"github.com/quic-go/quic-go"
)

// Server initialization with configuration.
var RunServer = func(ctx context.Context) error {
	logger := log.Default()
	connections.MessageChan = make(chan streams.MessageStream, 1)
	connections.NoSubsChan = make(chan streams.MessageStream, 1)
	connections.NewSubChan = make(chan streams.MessageStream, 1)

	logger.Println("Generating TLS config")

	tlsConfig, err := certificates.GenerateTLSConfig()
	if err != nil {
		return errors.Join(errors.New("error during tls config generation"), err)
	}

	logger.Println("TLS config generated")

	quicConfig := &quic.Config{
		MaxIdleTimeout:  time.Second * shared.MaxIdleTimeout,
		KeepAlivePeriod: time.Second * shared.KeepAlivePeriod,
	}

	logger.Println("Initializing pub server")

	err = connections.InitPubServer(ctx, tlsConfig, quicConfig)
	if err != nil {
		return errors.Join(errors.New("error during pub server initialization"), err)
	}

	logger.Printf("Pub server initialization complete, listening on: %v", shared.PortPub)
	logger.Println("Initializing sub server")

	err = connections.InitSubServer(ctx, tlsConfig, quicConfig)
	if err != nil {
		return errors.Join(errors.New("error during sub server initialization"), err)
	}

	logger.Printf("Sub server initialization complete, listening on: %v", shared.PortSub)

	return nil
}
