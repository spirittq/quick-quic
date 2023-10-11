package shared

import (
	"io"

	"github.com/quic-go/quic-go"
)

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
