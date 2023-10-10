package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"shared"
	"time"

	"github.com/quic-go/quic-go"
)

var pubCount int

func main() {

	logger := log.Default()
	ctx := context.Background()
	defer ctx.Done()

	var tlsConfig *tls.Config
	var quicConfig *quic.Config

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
		MaxIdleTimeout:  time.Second * 5,
		KeepAlivePeriod: time.Second * 30,
	}

	logger.Println("Initializing pub server")

	pubLn, err := quic.ListenAddr(fmt.Sprintf("%v:%v", shared.ServerAddr, shared.PortPub), tlsConfig, quicConfig)
	if err != nil {
		log.Fatalf("Error during pub server initialization: %v", err)
	}

	go listenForPubClient(pubLn, ctx)

	logger.Printf("Pub server initialization complete, listening on: %v", shared.PortPub)
	logger.Println("Initializing pub server")

	_, err = quic.ListenAddr(fmt.Sprintf("%v:%v", shared.ServerAddr, shared.PortSub), tlsConfig, quicConfig)
	if err != nil {
		log.Fatalf("Error during sub server initialization: %v", err)
	}

	logger.Printf("Sub server initialization complete, listening on: %v", shared.PortSub)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	logger.Printf("Server shut down")
}

var listenForPubClient = func(ln *quic.Listener, ctx context.Context) {

	logger := log.Default()

	pubServerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	for {
		conn, err := ln.Accept(pubServerCtx)
		if err != nil {
			logger.Printf("Failed to accept pub client connection: %v", err)
			continue
		}
		go handlePubClient(conn, ctx)
	}
}

var handlePubClient = func(conn quic.Connection, ctx context.Context) {

	logger := log.Default()
	logger.Println("New Pub connected")

	pubCount++
	pubClientCtx, cancel := context.WithCancel(ctx)

	defer func() {
		cancel()
		pubCount--
	}()

	// TODO separare shared func
	for {
		stream, err := conn.AcceptStream(pubClientCtx)
		if err != nil {
			logger.Printf("Pub closed with %v", err)
			break
		}

		go func(stream quic.Stream) {
			buffer := make([]byte, 1024)
			n, err := stream.Read(buffer)
			if err != nil && err != io.EOF {
				logger.Printf("Failed to read message: %v", err)
				return
			}
			receivedData := buffer[:n]
			logger.Printf("Pub sent message: %v", string(receivedData))
		}(stream)
	}
}
