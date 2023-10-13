package streams

import (
	"context"
	"io"

	"github.com/quic-go/quic-go"
)

// Reads message from the quic stream
func ReadStream(stream quic.Stream) ([]byte, error) {

	receivedData := []byte{}

	buffer := make([]byte, 1024)
	n, err := stream.Read(buffer)
	if err != nil && err != io.EOF {
		return receivedData, err
	}
	receivedData = buffer[:n]
	return receivedData, nil
}

// Writes message to the quic stream
func WriteStream(conn quic.Connection, msg []byte) error {
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

func AcceptStream(ctx context.Context, conn quic.Connection, acceptStreamChan chan quic.Stream) (error) {
	defer ctx.Done()

	for {
		stream, err := conn.AcceptStream(ctx)
		if err != nil {
			return err
		}
		acceptStreamChan <- stream
	}
}
