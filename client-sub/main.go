package main

import (
	"client-sub/core"
	"context"
	"log"
	"os"
	"os/signal"
)

func main() {
	logger := log.Default()
	ctx := context.Background()
	defer ctx.Done()

	go core.RunSubClient(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	logger.Printf("Client shutdown")
}
