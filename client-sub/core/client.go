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

var connectingToServer chan bool

var RunSubClient = func(ctx context.Context) {
	logger := log.Default()

	var tlsConfig *tls.Config
	var quicConfig *quic.Config
	connectingToServer = make(chan bool, 1)

	tlsConfig = &tls.Config{InsecureSkipVerify: true}
	quicConfig = &quic.Config{
		MaxIdleTimeout: time.Second * 5,
		KeepAlivePeriod: time.Second * 2,
	}

	for {
		logger.Println("Connecting to the server")
		clientCtx, cancel := context.WithCancel(ctx)
		conn, err := quic.DialAddr(clientCtx, fmt.Sprintf("%v:%v", shared.ServerAddr, shared.PortSub), tlsConfig, quicConfig)
		if err != nil {
			logger.Fatalf("Could not connect to the server: %v", err)
		}
		fmt.Println("Connection established")

		go receiveMessage(clientCtx, conn)
		<-connectingToServer
		cancel()
	}
}

var receiveMessage = func(ctx context.Context, conn quic.Connection) {
	logger := log.Default()
	defer ctx.Done()

	for {
		stream, err := conn.AcceptStream(ctx)
		if err != nil {
			logger.Printf("Lost connection to server with %v", err)
			connectingToServer <- true
			return
		}

		go func(stream quic.Stream) {
			receivedData, err := shared.ReadStream(stream)
			if err != nil {
				logger.Printf("Failed to read message: %v", err)
				return
			}
			unmarshalledMsg, err := shared.FromJson[shared.MessageStream](receivedData)
			if err != nil {
				logger.Printf("Unable to unmarshall message: %v", err)
				return
			}
			fmt.Println(unmarshalledMsg.Message)
		}(stream)
	}
}
