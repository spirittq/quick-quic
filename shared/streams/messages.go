package streams

import (
	"context"
	"log"
	"shared/utils"

	"github.com/quic-go/quic-go"
)

// Waits for the channel to receive message stream, encodes it to json and sends it to the peers. Finishes if context is closed.
var SendMessage = func(ctx context.Context, conn quic.Connection, messageChan chan MessageStream) {
	logger := log.Default()

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-messageChan:
			marshalledMsg, err := utils.ToJson[MessageStream](msg)
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

// Waits for the channel to receive a quic stream, decodes json and executes custom function from the wrapper.
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
			unmarshalledData, err := utils.FromJson[MessageStream](receivedData)
			if err != nil {
				logger.Printf("Failed to unmarshall message: %v", err)
				return
			}
			postReceiveMessage(unmarshalledData)
		}(stream)
	}
}
