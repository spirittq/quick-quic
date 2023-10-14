package core

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"shared"
	"shared/streams"
	"time"

	"github.com/quic-go/quic-go"
)

var acceptStreamChan chan quic.Stream

// Initiates subscriber client. If successfully connected to the server, starts background processes.
// Is blocked until stream accept fails, tries to re-connect to the server once.
// TODO test
var RunSubClient = func(ctx context.Context) {
	logger := log.Default()

	var tlsConfig *tls.Config
	var quicConfig *quic.Config
	acceptStreamChan = make(chan quic.Stream, 1)

	tlsConfig = &tls.Config{InsecureSkipVerify: true}
	quicConfig = &quic.Config{
		MaxIdleTimeout:  time.Second * shared.MaxIdleTimeout,
		KeepAlivePeriod: time.Second * shared.KeepAlivePeriod,
	}

	for {
		logger.Println("Connecting to the server")

		clientCtx, cancel := context.WithCancel(ctx)
		clientReceiveCtx := context.WithValue(clientCtx, shared.ContextName, shared.SubClientSendCtx)
		conn, err := quic.DialAddr(clientCtx, fmt.Sprintf("%v:%v", shared.ServerAddr, shared.PortSub), tlsConfig, quicConfig)
		if err != nil {
			logger.Fatalf("Could not connect to the server: %v", err)
		}

		logger.Println("Connection established")

		go receiveMessage(clientReceiveCtx, conn)
		err = streams.AcceptStream(clientCtx, conn, acceptStreamChan)
		if err != nil {
			logger.Printf("Lost connection to server with %v", err)
		}
		cancel()
	}
}

// Wrapper to receive message stream and run a custom function after message stream is received.
func receiveMessage(ctx context.Context, conn quic.Connection) {
	logger := log.Default()

	var postReceiveMessage = func(messageStream streams.MessageStream) {
		logger.Println(messageStream.Message)
	}

	go streams.ReceiveMessage(ctx, acceptStreamChan, postReceiveMessage)
}
