package core

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"shared/streams"
	"strings"
	"time"

	"github.com/quic-go/quic-go"
)

var jugglingMessage chan streams.MessageStream
var acceptStreamChan chan quic.Stream

// Initiates publisher client. If successfully connected to the server, starts background processes.
// Is blocked until stream accept fails, tries to re-connect to the server once.
func RunPubClient(ctx context.Context) {
	logger := log.Default()

	var tlsConfig *tls.Config
	var quicConfig *quic.Config
	acceptStreamChan = make(chan quic.Stream, 1)
	jugglingMessage = make(chan streams.MessageStream, 1)

	tlsConfig = &tls.Config{InsecureSkipVerify: true}
	quicConfig = &quic.Config{
		MaxIdleTimeout:  time.Second * streams.MaxIdleTimeout,
		KeepAlivePeriod: time.Second * streams.KeepAlivePeriod,
	}

	for {
		logger.Println("Connecting to the server")

		clientCtx, cancel := context.WithCancel(ctx)
		conn, err := quic.DialAddr(clientCtx, fmt.Sprintf("%v:%v", streams.ServerAddr, streams.PortPub), tlsConfig, quicConfig)
		if err != nil {
			logger.Fatalf("Could not connect to the server: %v", err)
		}

		logger.Println("Connection established")

		go inputMessage()
		go sendMessage(clientCtx, conn)
		go receiveMessage(clientCtx, conn)
		err = streams.AcceptStream(clientCtx, conn, acceptStreamChan)
		if err != nil {
			logger.Printf("Lost connection to server with %v", err)
		}
		cancel()
	}
}

// Allows client to read publisher's input and put it into channel for sending message streams.
func inputMessage() {

	logger := log.Default()
	var err error

	for {

		var msg streams.MessageStream
		reader := bufio.NewReader(os.Stdin)
		msg.Message, err = reader.ReadString('\n')
		if err != nil {
			logger.Printf("Failed to read input: %v", err)
		}
		msg.Message = strings.TrimSpace(msg.Message)
		jugglingMessage <- msg
	}
}

// Wrapper to send message stream.
func sendMessage(ctx context.Context, conn quic.Connection) {
	go streams.SendMessage(ctx, conn, jugglingMessage)
}

// Wrapper to receive message stream and run a custom function after message stream is received.
func receiveMessage(ctx context.Context, conn quic.Connection) {
	logger := log.Default()

	var postReceiveMessage = func(messageStream streams.MessageStream) {
		logger.Println(messageStream.Message)
	}

	go streams.ReceiveMessage(ctx, acceptStreamChan, postReceiveMessage)
}
