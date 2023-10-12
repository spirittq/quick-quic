package core

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"server/connections"
	"shared"
	"time"

	"github.com/quic-go/quic-go"
)

var RunServer = func(ctx context.Context) {
	logger := log.Default()
	connections.MessageChan = make(chan shared.MessageStream, 1)
	connections.NoSubsChan = make(chan shared.MessageStream, 1)
	connections.NewSubChan = make(chan shared.MessageStream, 1)

	logger.Println("Loading temp certificate")

	cert, err := tls.LoadX509KeyPair(fmt.Sprintf("%v%v", shared.CertPath, shared.CertTemp), fmt.Sprintf("%v%v", shared.CertPath, shared.CertKeyTemp))
	if err != nil {
		log.Fatal("Could not load certificates")
	}

	logger.Println("Certificates loaded")

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	quicConfig := &quic.Config{
		// MaxIdleTimeout:  time.Second * shared.MaxIdleTimeout,
		// KeepAlivePeriod: time.Second * shared.KeepAlivePeriod,
		MaxIdleTimeout: 5 * time.Second,
	}

	logger.Println("Initializing pub server")

	go connections.InitPubServer(ctx, tlsConfig, quicConfig)

	logger.Printf("Pub server initialization complete, listening on: %v", shared.PortPub)
	logger.Println("Initializing sub server")

	go connections.InitSubServer(ctx, tlsConfig, quicConfig)

	logger.Printf("Sub server initialization complete, listening on: %v", shared.PortSub)
	logger.Println("Initializing sub monitoring")

	go monitorSubs()

	logger.Println("Sub monitoring initialized")
}

var monitorSubs = func() {
	for {
		if connections.SubCount.Count == 0 {
			for i := 0; i < connections.PubCount.Count; i++ {
				connections.NoSubsChan <- connections.NoSubsMessage
			}
			time.Sleep(time.Second * 10)
		}
	}
}
