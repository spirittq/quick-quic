package connections

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"shared"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
)

var InitPubServer = func(ctx context.Context, tlsConfig *tls.Config, quicConfig *quic.Config) {

	logger := log.Default()

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

	go func() {
		for {
			if subCount.Count == 0 {
				for i := 0; i < pubCount.Count; i++ {
					SubNotConnectedChan <- true

				}
				time.Sleep(time.Second * 10)
			}
		}
	}()
}

var handlePubClient = func(ctx context.Context, conn quic.Connection) {

	logger := log.Default()
	logger.Println("New Pub connected")

	pubCount.Mu.Lock()
	pubCount.Count++
	pubCount.Mu.Unlock()

	pubClientCtx, cancel := context.WithCancel(ctx)

	defer func() {
		pubCount.Mu.Lock()
		pubCount.Count--
		pubCount.Mu.Unlock()
		cancel()
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go sendMessagePub(pubClientCtx, conn, &wg)
	wg.Add(1)
	go receiveMessagePub(pubClientCtx, conn, &wg)
	wg.Wait()
}

var sendMessagePub = func(ctx context.Context, conn quic.Connection, wg *sync.WaitGroup) {
	defer wg.Done()
	logger := log.Default()

	go func() {
		msgToPub := shared.MessageStream{
			Message: "New subscriber has connected",
		}
		msg, err := shared.ToJson[shared.MessageStream](msgToPub)
		if err != nil {
			logger.Printf("Unable to marshal message: %v", err)
		}
		for {
			select {
			case <-ctx.Done():
				return
			case <-SubConnectedChan:
				err := shared.WriteStream(conn, msg)
				if err != nil {
					logger.Printf("Failed to send message: %v", err)
				}
			}
		}
	}()

	go func() {
		msgToPub := shared.MessageStream{
			Message: "No subscribers are connected",
		}
		msg, err := shared.ToJson[shared.MessageStream](msgToPub)
		if err != nil {
			logger.Printf("Unable to marshal message: %v", err)
		}
		for {
			select {
			case <-ctx.Done():
				return
			case <-SubNotConnectedChan:
				err := shared.WriteStream(conn, msg)
				if err != nil {
					logger.Printf("Failed to send message: %v", err)
				}
			}
		}
	}()
}

var receiveMessagePub = func(ctx context.Context, conn quic.Connection, wg *sync.WaitGroup) {

	defer wg.Done()

	logger := log.Default()

	for {
		stream, err := conn.AcceptStream(ctx)
		if err != nil {
			logger.Printf("Pub closed with %v", err)
			return
		}

		go func(stream quic.Stream) {
			receivedData, err := shared.ReadStream(stream)
			if err != nil {
				logger.Printf("Failed to read message: %v", err)
				return
			}
			logger.Println("Pub message received")
			for i := 0; i < subCount.Count; i++ {
				MessageChan <- receivedData
			}
		}(stream)
	}
}
