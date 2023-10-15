package streams

import (
	"context"
	"crypto/tls"
	"errors"
	"server/certificates"
	"testing"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadAndWriteStream(t *testing.T) {
	clientAcceptChan := make(chan quic.Stream, 1)
	serverAcceptChan := make(chan quic.Stream, 1)

	serverSentData := []byte("server wrote this message")
	clientSentData := []byte("client wrote this message")
	ctx := context.TODO()
	serverTLSConfig, _ := certificates.GenerateTLSConfig()
	receiverTLSConfig := &tls.Config{InsecureSkipVerify: true}
	quicConfig := &quic.Config{
		MaxIdleTimeout:  time.Second * 5,
		KeepAlivePeriod: time.Second * 5,
	}
	t.Run("Server and client are connected, server sends message and client reads it, then client sends message and server reads it", func(t *testing.T) {
		serverLn, err := quic.ListenAddr("localhost:3000", serverTLSConfig, quicConfig)
		require.Nil(t, err)
		clientConn, err := quic.DialAddr(ctx, "localhost:3000", receiverTLSConfig, quicConfig)
		require.Nil(t, err)
		go AcceptStream(ctx, clientConn, clientAcceptChan)
		serverConn, err := serverLn.Accept(ctx)
		require.Nil(t, err)
		go AcceptStream(ctx, serverConn, serverAcceptChan)
		err = WriteStream(serverConn, serverSentData)
		require.Nil(t, err)
		clientStream := <-clientAcceptChan
		clientReceivedData, err := ReadStream(clientStream)
		require.Nil(t, err)
		require.Equal(t, serverSentData, clientReceivedData)
		err = WriteStream(clientConn, clientSentData)
		require.Nil(t, err)
		serverStream := <-serverAcceptChan
		serverReceivedData, err := ReadStream(serverStream)
		require.Nil(t, err)
		require.Equal(t, clientSentData, serverReceivedData)
		require.Equal(t, 0, len(serverAcceptChan))
		require.Equal(t, 0, len(clientAcceptChan))
	})
}

var AcceptStreamTestCases = []struct {
	Description            string
	AcceptStreamError      error
	AcceptStreamChanLength int
	Success                bool
}{
	{
		Description:            "Successfully accepts stream and puts it into channel",
		AcceptStreamError:      nil,
		AcceptStreamChanLength: 1,
		Success:                true,
	},
	{
		Description:            "Fails to accept stream and returns an error",
		AcceptStreamError:      errors.New("error"),
		AcceptStreamChanLength: 0,
		Success:                false,
	},
}

func TestAcceptStream(t *testing.T) {
	var testStream quic.Stream
	var testConnection quic.Connection
	for _, testCase := range AcceptStreamTestCases {
		t.Run(testCase.Description, func(t *testing.T) {

			acceptStreamChan := make(chan quic.Stream, 1)
			testChan := make(chan bool, 1)

			quicAcceptStream = func(c quic.Connection, ctx context.Context) (quic.Stream, error) {
				<-testChan
				return testStream, testCase.AcceptStreamError
			}

			go AcceptStream(context.TODO(), testConnection, acceptStreamChan)
			testChan <- true
			time.Sleep(1 * time.Second)

			assert.Equal(t, testCase.AcceptStreamChanLength, len(acceptStreamChan))
		})
	}
}
