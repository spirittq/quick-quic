package connections

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"shared"

	"github.com/quic-go/quic-go"
)

var InitSubServer = func(ctx context.Context, tlsConfig *tls.Config, quicConfig *quic.Config) {
	logger := log.Default()

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

	defer func() {
		SubCount.Mu.Lock()
		SubCount.Count--
		SubCount.Mu.Unlock()
		cancel()
	}()

	acceptStreamChan := make(chan quic.Stream, 1)

	go sendMessageToSub(subClientCtx, conn)
	err := shared.AcceptStream(subClientCtx, conn, acceptStreamChan)
	if err != nil {
		logger.Printf("Sub disconnected with: %v", err)
	}
}

var sendMessageToSub = func(ctx context.Context, conn quic.Connection) {
	go sendMessageToClient(ctx, conn, MessageChan)
}
