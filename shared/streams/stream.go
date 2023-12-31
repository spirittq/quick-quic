package streams

import (
	"context"
	"io"

	"github.com/quic-go/quic-go"
)

var quicAcceptStream = (quic.Connection).AcceptStream

// Reads message from the quic stream
var ReadStream = func(stream quic.Stream) ([]byte, error) {

	var err error
	var n int
	var receivedData []byte

	buffer := make([]byte, 1024)
	for err != io.EOF {
		n, err = stream.Read(buffer)
		if err != nil && err != io.EOF {
			return []byte{}, err
		}
		receivedData = append(receivedData, buffer[:n]...)
		if err == io.EOF {
			break
		}
	}
	return receivedData, nil
}

// Writes message to the quic stream
var WriteStream = func(conn quic.Connection, msg []byte) error {

	stream, err := conn.OpenStream()
	if err != nil {
		return err
	}
	_, err = stream.Write(msg)
	closeErr := stream.Close()

	if err != nil {
		return err
	}
	return closeErr
}

// Waits to accept stream. When it is available, sends stream through the channel.
var AcceptStream = func(ctx context.Context, conn quic.Connection, acceptStreamChan chan quic.Stream) error {
	for {
		stream, err := quicAcceptStream(conn, ctx)
		if err != nil {
			return err
		}
		acceptStreamChan <- stream
	}
}
