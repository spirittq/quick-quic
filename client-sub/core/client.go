package core

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"shared"
	"time"

	"github.com/quic-go/quic-go"
)

var acceptStreamChan chan quic.Stream

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
		conn, err := quic.DialAddr(clientCtx, fmt.Sprintf("%v:%v", shared.ServerAddr, shared.PortSub), tlsConfig, quicConfig)
		if err != nil {
			logger.Fatalf("Could not connect to the server: %v", err)
		}

		logger.Println("Connection established")

		go receiveMessage(clientCtx, conn)
		err = shared.AcceptStream(clientCtx, conn, acceptStreamChan)
		if err != nil {
			logger.Printf("Lost connection to server with %v", err)
		}
		cancel()
	}
}

func receiveMessage(ctx context.Context, conn quic.Connection) {
	logger := log.Default()

	var postReceiveMessage = func(messageStream shared.MessageStream) {
		logger.Println(messageStream.Message)
	}

	go shared.ReceiveMessage(ctx, acceptStreamChan, postReceiveMessage)
}
