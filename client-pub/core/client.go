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
var jugglingMessage chan shared.MessageStream

func RunPubClient(ctx context.Context) {
	logger := log.Default()

	var tlsConfig *tls.Config
	var quicConfig *quic.Config
	connectingToServer = make(chan bool, 1)
	jugglingMessage = make(chan shared.MessageStream, 1)
	jugglingMessage <- shared.MessageStream{Empty: true}

	tlsConfig = &tls.Config{InsecureSkipVerify: true}
	quicConfig = &quic.Config{
		MaxIdleTimeout:  time.Second * 5,
		KeepAlivePeriod: time.Second * 2,
	}

	for {
		logger.Println("Connecting to the server")
		clientCtx, cancel := context.WithCancel(ctx)
		conn, err := quic.DialAddr(clientCtx, fmt.Sprintf("%v:%v", shared.ServerAddr, shared.PortPub), tlsConfig, quicConfig)
		if err != nil {
			logger.Fatalf("Could not connect to the server: %v", err)
		}
		fmt.Println("Connection established")

		go sendMessage(clientCtx, conn)
		go receiveMessage(clientCtx, conn)
		<-connectingToServer
		cancel()
	}
}

func sendMessage(ctx context.Context, conn quic.Connection) {

	logger := log.Default()
	initMsg := <-jugglingMessage

	for {
		var msg shared.MessageStream

		switch {
		case initMsg.Empty:
			reader := bufio.NewReader(os.Stdin)
			msg.Message, _ = reader.ReadString('\n')
			msg.Message = strings.TrimSpace(msg.Message)
		default:
			msg = initMsg
			initMsg.Empty = true
		}

		select {
		case <-ctx.Done():
			jugglingMessage <- msg
			return
		default:
			marshalledMsg, err := shared.ToJson[shared.MessageStream](msg)
			if err != nil {
				logger.Printf("Unable to marshall message: %v", err)
			}
			err = shared.WriteStream(conn, marshalledMsg)
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
			unmarshalledMsg, err := shared.FromJson[shared.MessageStream](receivedData)
			if err != nil {
				logger.Printf("Unable to unmarshall message: %v", err)
				return
			}
			fmt.Println(unmarshalledMsg.Message)
		}(stream)
	}
}
