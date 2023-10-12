package shared

import (
	"context"
	"log"

	"github.com/quic-go/quic-go"
)

var SendMessage = func(ctx context.Context, conn quic.Connection, messageChan chan MessageStream) {
	logger := log.Default()

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-messageChan:
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

var ReceiveMessage = func(ctx context.Context, acceptStreamChan chan quic.Stream, postReceiveMessage PostReceiveMessage) {
	logger := log.Default()

	for {
		stream := <-acceptStreamChan

		go func(stream quic.Stream) {
			receivedData, err := ReadStream(stream)
			if err != nil {
				logger.Printf("Failed to read message: %v", err)
				return
			}
			unmarshalledData, err := FromJson[MessageStream](receivedData)
			if err != nil {
				logger.Printf("Failed to unmarshall message: %v", err)
				return
			}
			postReceiveMessage(unmarshalledData)
		}(stream)
	}
}
