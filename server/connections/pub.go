package connections

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"shared/streams"
	"time"

	"github.com/quic-go/quic-go"
)

// Initiates publisher server and starts listening to the clients.
var InitPubServer = func(ctx context.Context, tlsConfig *tls.Config, quicConfig *quic.Config) {
	logger := log.Default()

	acceptStreamPubChan = make(chan quic.Stream, 1)

	ln, err := quic.ListenAddr(fmt.Sprintf("%v:%v", streams.ServerAddr, streams.PortPub), tlsConfig, quicConfig)
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

// Increases publisher count upon start and decreases it upon end. Starts background processes.
// Is blocked until stream accept fails.
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
	err := streams.AcceptStream(pubClientCtx, conn, acceptStreamPubChan)
	if err != nil {
		logger.Printf("Pub disconnected with: %v", err)
	}
}

// Wrapper to send message stream to publisher client.
var sendMessageToPub = func(ctx context.Context, conn quic.Connection) {
	go streams.SendMessage(ctx, conn, NewSubChan)
	go streams.SendMessage(ctx, conn, NoSubsChan)
}

// Wrapper to receive message stream from publisher client and run a custom function after message stream is received.
var receiveMessageFromPub = func(ctx context.Context, conn quic.Connection) {
	logger := log.Default()

	var postReceiveMessage = func(messageStream streams.MessageStream) {
		logger.Println("Pub message received")
		for i := 0; i < SubCount.Count; i++ {
			MessageChan <- messageStream
		}
	}

	go streams.ReceiveMessage(ctx, acceptStreamPubChan, postReceiveMessage)
}

// Monitors if no subscribers are connected, puts a message stream in a channel if no subscribers are connected.
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
