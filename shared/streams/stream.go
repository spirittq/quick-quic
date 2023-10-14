package streams

import (
	"context"
	"io"

	"github.com/quic-go/quic-go"
)

// Reads message from the quic stream
// TODO test
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
// TODO test
var WriteStream = func(conn quic.Connection, msg []byte) error {
	var closeErr error

	stream, err := conn.OpenStream()
	if err != nil {
		return err
	}
	_, err = stream.Write(msg)
	stream.Close()

	if err != nil {
		return err
	}
	return closeErr
}

// Waits to accept stream. When it is available, sends stream through the channel.
// TODO test
var AcceptStream = func(ctx context.Context, conn quic.Connection, acceptStreamChan chan quic.Stream) error {
	for {
		stream, err := conn.AcceptStream(ctx)
		if err != nil {
			return err
		}
		acceptStreamChan <- stream
	}
}
