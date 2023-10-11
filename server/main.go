package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"server/core"
)

func main() {

	logger := log.Default()
	ctx := context.Background()
	defer ctx.Done()

	go core.RunServer(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	logger.Println("Server shut down")
}
