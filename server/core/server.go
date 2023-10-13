package core

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
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

	logger.Println("Loading temp certificate")

	cert, err := tls.LoadX509KeyPair(fmt.Sprintf("%v%v", streams.CertPath, streams.CertTemp), fmt.Sprintf("%v%v", streams.CertPath, streams.CertKeyTemp))
	if err != nil {
		log.Fatal("Could not load certificates")
	}

	logger.Println("Certificates loaded")

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
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
