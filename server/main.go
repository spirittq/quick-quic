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

	err := core.RunServer(ctx)
	if err != nil {
		log.Fatalf("Could not start server: %v", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	logger.Println("Server shut down")
}
