package core

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

var RunServer = func(ctx context.Context) {
	logger := log.Default()
	messageChan = make(chan string)
	subNotConnectedChan = make(chan bool)
	subConnectedChan = make(chan bool)

	logger.Println("Loading temp certificate")

	cert, err := tls.LoadX509KeyPair(fmt.Sprintf("%v%v", shared.CertPath, shared.CertTemp), fmt.Sprintf("%v%v", shared.CertPath, shared.CertKeyTemp))
	if err != nil {
		log.Fatal("Could not load certificates")
	}

	logger.Println("Certificates loaded")

	tlsConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	quicConfig = &quic.Config{
		MaxIdleTimeout: time.Second * 5,
	}

	logger.Println("Initializing pub server")

	go initPubServer(ctx)

	logger.Printf("Pub server initialization complete, listening on: %v", shared.PortPub)
	logger.Println("Initializing pub server")

	_, err = quic.ListenAddr(fmt.Sprintf("%v:%v", shared.ServerAddr, shared.PortSub), tlsConfig, quicConfig)
	if err != nil {
		log.Fatalf("Error during sub server initialization: %v", err)
	}

	logger.Printf("Sub server initialization complete, listening on: %v", shared.PortSub)

}

var initPubServer = func(ctx context.Context) {

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
			go handlePubClient(conn, ctx)
		}
	}()

	go func() {
		for {
			if subCount.Count == 0 {
				for i := 0; i < pubCount.Count; i++ {
					subNotConnectedChan <- true

				}
				time.Sleep(time.Second * 3)
			}
		}
	}()
}

var handlePubClient = func(conn quic.Connection, ctx context.Context) {

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
		msg := "New subscriber has connected"
		for {
			select {
			case <-ctx.Done():
				return
			case <-subConnectedChan:
				err := shared.WriteStream(conn, msg)
				if err != nil {
					logger.Printf("Failed to send message: %v", err)
				}
			}
		}
	}()

	go func() {
		msg := "No subscribers are connected"
		for {
			select {
			case <-ctx.Done():
				return
			case <-subNotConnectedChan:
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
			messageChan <- string(receivedData)
		}(stream)
	}
}
