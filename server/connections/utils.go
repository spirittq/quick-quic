package connections

import (
	"context"
	"log"
	"shared"

	"github.com/quic-go/quic-go"
)

var sendMessageToClient = func(ctx context.Context, conn quic.Connection, writeChan chan shared.MessageStream) {
	logger := log.Default()

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-writeChan:
			marshalledMsg, err := shared.ToJson[shared.MessageStream](msg)
			if err != nil {
				logger.Printf("Failed to marshal message: %v", err)
			}
			err = shared.WriteStream(conn, marshalledMsg)
			if err != nil {
				logger.Printf("Failed to send message: %v", err)
			}
		}
	}
}
