package connections

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"shared"

	"github.com/quic-go/quic-go"
)

var InitPubServer = func(ctx context.Context, tlsConfig *tls.Config, quicConfig *quic.Config) {
	logger := log.Default()

	acceptStreamPubChan = make(chan quic.Stream, 1)

	ln, err := quic.ListenAddr(fmt.Sprintf("%v:%v", shared.ServerAddr, shared.PortPub), tlsConfig, quicConfig)
	if err != nil {
		log.Fatalf("Error during pub server initialization: %v", err)
	}

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

	for {
		stream := <-acceptStreamPubChan

		go func(stream quic.Stream) {
			receivedData, err := shared.ReadStream(stream)
			if err != nil {
				logger.Printf("Failed to read message: %v", err)
				return
			}
			unmarshalledData, err := shared.FromJson[shared.MessageStream](receivedData)
			if err != nil {
				logger.Printf("Failed to unmarshall message: %v", err)
				return
			}
			logger.Println("Pub message received")
			for i := 0; i < SubCount.Count; i++ {
				MessageChan <- unmarshalledData
			}
		}(stream)
	}
}
