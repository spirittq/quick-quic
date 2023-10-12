package connections

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"shared"
	"time"

	"github.com/quic-go/quic-go"
)

var InitPubServer = func(ctx context.Context, tlsConfig *tls.Config, quicConfig *quic.Config) {
	logger := log.Default()

	acceptStreamPubChan = make(chan quic.Stream, 1)

	ln, err := quic.ListenAddr(fmt.Sprintf("%v:%v", shared.ServerAddr, shared.PortPub), tlsConfig, quicConfig)
	if err != nil {
		log.Fatalf("Error during pub server initialization: %v", err)
	}

	go monitorSubs()

	go func() {
		for {
			conn, err := ln.Accept(ctx)
			if err != nil {
				logger.Printf("Failed to accept pub client connection: %v", err)
				continue
			}
			go handlePubClient(ctx, conn)
		}
	}()

}

var handlePubClient = func(ctx context.Context, conn quic.Connection) {
	logger := log.Default()
	logger.Println("New Pub connected")

	PubCount.Mu.Lock()
	PubCount.Count++
	PubCount.Mu.Unlock()

	pubClientCtx, cancel := context.WithCancel(ctx)

	defer func() {
		PubCount.Mu.Lock()
		PubCount.Count--
		PubCount.Mu.Unlock()
		cancel()
	}()

	go sendMessageToPub(pubClientCtx, conn)
	go receiveMessageFromPub(pubClientCtx, conn)
	err := shared.AcceptStream(pubClientCtx, conn, acceptStreamPubChan)
	if err != nil {
		logger.Printf("Pub disconnected with: %v", err)
	}
}

var sendMessageToPub = func(ctx context.Context, conn quic.Connection) {
	go shared.SendMessage(ctx, conn, NewSubChan)
	go shared.SendMessage(ctx, conn, NoSubsChan)
}

var receiveMessageFromPub = func(ctx context.Context, conn quic.Connection) {
	logger := log.Default()

	var postReceiveMessage = func(messageStream shared.MessageStream) {
		logger.Println("Pub message received")
		for i := 0; i < SubCount.Count; i++ {
			MessageChan <- messageStream
		}
	}

	go shared.ReceiveMessage(ctx, acceptStreamPubChan, postReceiveMessage)
}

var monitorSubs = func() {
	for {
		if SubCount.Count == 0 {
			for i := 0; i < PubCount.Count; i++ {
				NoSubsChan <- NoSubsMessage
			}
			time.Sleep(time.Second * 10)
		}
	}
}
