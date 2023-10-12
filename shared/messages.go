package shared

import (
	"context"
	"log"

	"github.com/quic-go/quic-go"
)

var SendMessage = func(ctx context.Context, conn quic.Connection, writeChan chan MessageStream) {
	logger := log.Default()

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-writeChan:
			marshalledMsg, err := ToJson[MessageStream](msg)
			if err != nil {
				logger.Printf("Failed to marshal message: %v", err)
			}
			err = WriteStream(conn, marshalledMsg)
			if err != nil {
				logger.Printf("Failed to send message: %v", err)
			}
		}
	}
}
