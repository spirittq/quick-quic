package connections

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"shared"
	"sync"

	"github.com/quic-go/quic-go"
)

var InitSubServer = func(ctx context.Context, tlsConfig *tls.Config, quicConfig *quic.Config) {
	logger := log.Default()

	ln, err := quic.ListenAddr(fmt.Sprintf("%v:%v", shared.ServerAddr, shared.PortSub), tlsConfig, quicConfig)
	if err != nil {
		log.Fatalf("Error during sub server initialization: %v", err)
	}

	go func() {
		for {
			conn, err := ln.Accept(ctx)
			if err != nil {
				logger.Printf("Failed to accept sub client connection: %v", err)
				continue
			}
			go handleSubClient(ctx, conn)
		}
	}()
}

var handleSubClient = func(ctx context.Context, conn quic.Connection) {
	logger := log.Default()
	logger.Println("New Sub connected")

	for i := 0; i < pubCount.Count; i++ {
		SubConnectedChan <- true
	}

	subCount.Mu.Lock()
	subCount.Count++
	subCount.Mu.Unlock()

	subClientCtx, cancel := context.WithCancel(ctx)

	defer func() {
		subCount.Mu.Lock()
		subCount.Count--
		subCount.Mu.Unlock()
		cancel()
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go sendMessageSub(subClientCtx, conn, &wg)
	wg.Add(1)
	go acceptStreamSub(subClientCtx, conn, &wg)
	wg.Wait()
}

var sendMessageSub = func(ctx context.Context, conn quic.Connection, wg *sync.WaitGroup) {
	defer wg.Done()
	logger := log.Default()

	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("sub canceled")
				return
			case msg := <-MessageChan:
				err := shared.WriteStream(conn, msg)
				if err != nil {
					logger.Printf("Failed to send message: %v", err)
				}
			}
		}
	}()
}

var acceptStreamSub = func(ctx context.Context, conn quic.Connection, wg *sync.WaitGroup) {
	defer wg.Done()

	logger := log.Default()

	for {
		_, err := conn.AcceptStream(ctx)
		if err != nil {
			logger.Printf("Sub closed with %v", err)
			return
		}
	}
}
