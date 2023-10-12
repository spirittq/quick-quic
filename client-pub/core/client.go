package core

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"shared"
	"strings"
	"time"

	"github.com/quic-go/quic-go"
)

var jugglingMessage chan shared.MessageStream
var acceptStreamChan chan quic.Stream

func RunPubClient(ctx context.Context) {
	logger := log.Default()

	var tlsConfig *tls.Config
	var quicConfig *quic.Config
	acceptStreamChan = make(chan quic.Stream, 1)
	jugglingMessage = make(chan shared.MessageStream, 1)

	tlsConfig = &tls.Config{InsecureSkipVerify: true}
	quicConfig = &quic.Config{
		// MaxIdleTimeout:  time.Second * shared.MaxIdleTimeout,
		// KeepAlivePeriod: time.Second * shared.KeepAlivePeriod,
		MaxIdleTimeout: 5 * time.Second,
	}

	for {
		logger.Println("Connecting to the server")
		clientCtx, cancel := context.WithCancel(ctx)
		conn, err := quic.DialAddr(clientCtx, fmt.Sprintf("%v:%v", shared.ServerAddr, shared.PortPub), tlsConfig, quicConfig)
		if err != nil {
			logger.Fatalf("Could not connect to the server: %v", err)
		}
		fmt.Println("Connection established")

		go InputMessage()
		go sendMessage(clientCtx, conn)
		go receiveMessage(clientCtx, conn, acceptStreamChan)
		err = shared.AcceptStream(clientCtx, conn, acceptStreamChan)
		if err != nil {
			logger.Printf("Lost connection to server with %v", err)
		}
		cancel()
	}
}

func InputMessage() {
	for {
		var msg shared.MessageStream
		reader := bufio.NewReader(os.Stdin)
		msg.Message, _ = reader.ReadString('\n')
		msg.Message = strings.TrimSpace(msg.Message)
		jugglingMessage <- msg
	}
}

func sendMessage(ctx context.Context, conn quic.Connection) {
	go shared.SendMessage(ctx, conn, jugglingMessage)
}

func receiveMessage(ctx context.Context, conn quic.Connection, acceptStreamChan chan quic.Stream) {
	logger := log.Default()
	defer ctx.Done()

	for {
		stream := <-acceptStreamChan

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
			logger.Println(unmarshalledMsg.Message)
		}(stream)
	}
}
