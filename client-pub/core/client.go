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

var connectingToServer chan bool
var jugglingMessage chan string

func RunClient(ctx context.Context) {
	logger := log.Default()

	var tlsConfig *tls.Config
	var quicConfig *quic.Config
	connectingToServer = make(chan bool, 1)
	jugglingMessage = make(chan string, 1)
	jugglingMessage <- ""

	tlsConfig = &tls.Config{InsecureSkipVerify: true}
	quicConfig = &quic.Config{
		MaxIdleTimeout: 5 * time.Second,
	}

	go func() {
		for {
			logger.Println("Connecting to the server")
			clientCtx, cancel := context.WithCancel(ctx)
			conn, err := quic.DialAddr(clientCtx, "localhost:3001", tlsConfig, quicConfig)
			if err != nil {
				logger.Fatalf("Could not connect to the server: %v", err)
			}
			fmt.Println("Connection established")

			go sendMessage(clientCtx, conn)
			go receiveMessage(clientCtx, conn)
			<-connectingToServer
			cancel()
		}
	}()
}

func sendMessage(ctx context.Context, conn quic.Connection) {

	logger := log.Default()
	initMsg := <-jugglingMessage

	for {
		var msg string

		switch initMsg {
		case "":
			reader := bufio.NewReader(os.Stdin)
			msg, _ = reader.ReadString('\n')
			msg = strings.TrimSpace(msg)
		default:
			msg = initMsg
			initMsg = ""
		}

		select {
		case <-ctx.Done():
			jugglingMessage <- msg
			return
		default:
			err := shared.WriteStream(conn, msg)
			if err != nil {
				logger.Printf("Failed to send message: %v", err)
			}
		}
	}
}

func receiveMessage(ctx context.Context, conn quic.Connection) {

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
			fmt.Println(string(receivedData))
		}(stream)
	}
}
