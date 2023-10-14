package connections

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"shared"
	"shared/streams"

	"github.com/quic-go/quic-go"
)

// Initiates subscriber server and starts listening to the clients.
var InitSubServer = func(ctx context.Context, tlsConfig *tls.Config, quicConfig *quic.Config) {
	logger := log.Default()

	acceptStreamSubChan = make(chan quic.Stream, 1)

	ln, err := quic.ListenAddr(fmt.Sprintf("%v:%v", shared.ServerAddr, shared.PortSub), tlsConfig, quicConfig)
	if err != nil {
		log.Fatalf("Error during sub server initialization: %v", err)
	}

	go func() {
		for {
			conn, err := ln.Accept(ctx)
			if err != nil {
				logger.Printf("Failed to accept sub client connection: %v", err)
				continue
			}
			go handleSubClient(ctx, conn)
		}
	}()
}

// Increases subscriber count upon start and decreases it upon end. Starts background processes.
// Is blocked until stream accept fails.
var handleSubClient = func(ctx context.Context, conn quic.Connection) {
	logger := log.Default()
	logger.Println("New Sub connected")

	for i := 0; i < PubCount.Count; i++ {
		NewSubChan <- NewSubMessage
	}

	SubCount.Mu.Lock()
	SubCount.Count++
	SubCount.Mu.Unlock()

	subClientCtx, cancel := context.WithCancel(ctx)
	subClientSendCtx := context.WithValue(subClientCtx, shared.ContextName, shared.SubClientSendCtx)

	defer func() {
		SubCount.Mu.Lock()
		SubCount.Count--
		SubCount.Mu.Unlock()
		cancel()
	}()

	go sendMessageToSub(subClientSendCtx, conn)
	err := streams.AcceptStream(subClientCtx, conn, acceptStreamSubChan)
	if err != nil {
		logger.Printf("Sub disconnected with: %v", err)
	}
}

// Wrapper to send message stream to subscriber client.
var sendMessageToSub = func(ctx context.Context, conn quic.Connection) {
	go streams.SendMessage(ctx, conn, MessageChan)
}
