package core

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"shared/streams"
	"testing"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/stretchr/testify/assert"
)

var InputMessageTestCases = []struct {
	Description        string
	ReadStringResponse string
	ReadStringError    error
	ExpectedChannelLen int
	Success            bool
}{
	{
		Description:        "Successfully reads input and puts message into channel",
		ReadStringResponse: "test message",
		ReadStringError:    nil,
		ExpectedChannelLen: 1,
		Success:            true,
	},
	{
		Description:        "Fails to read input",
		ReadStringResponse: "",
		ReadStringError:    errors.New("error"),
		ExpectedChannelLen: 0,
		Success:            false,
	},
}

func TestInputMessage(t *testing.T) {
	for _, testCase := range InputMessageTestCases {
		t.Run(testCase.Description, func(t *testing.T) {

			jugglingMessage = make(chan streams.MessageStream, 1)
			readStringChan := make(chan bool, 1)

			ReadString = func(b *bufio.Reader, delim byte) (string, error) {
				<-readStringChan
				return testCase.ReadStringResponse, testCase.ReadStringError
			}

			ctx := context.TODO()
			go inputMessage(ctx)
			readStringChan <- true
			time.Sleep(1 * time.Second)

			assert.Equal(t, testCase.ExpectedChannelLen, len(jugglingMessage))
			if testCase.Success {
				actualMsg := <-jugglingMessage
				assert.Equal(t, testCase.ReadStringResponse, actualMsg.Message)
			}

		})
	}
}

var RunPubClientTestCases = []struct {
	Description        string
	AcceptStreamError  error
	QuicDialAddrError  error
	GoroutinesExecuted int
	QuicDialAddrCount  int
}{
	{
		Description:        "Successfully runs pub client",
		AcceptStreamError:  nil,
		QuicDialAddrError:  nil,
		GoroutinesExecuted: 3,
		QuicDialAddrCount:  1,
	},
	{
		Description:        "Lost connection to the server and retries",
		AcceptStreamError:  errors.New("error"),
		QuicDialAddrError:  nil,
		GoroutinesExecuted: 3,
		QuicDialAddrCount:  2,
	},
	{
		Description:        "Unable to connect to the server and repeats",
		AcceptStreamError:  nil,
		QuicDialAddrError:  errors.New("error"),
		GoroutinesExecuted: 0,
		QuicDialAddrCount:  2,
	},
}

func TestRunPubClient(t *testing.T) {

	_inputMessage := inputMessage
	_sendMessage := sendMessage
	_receiveMessage := receiveMessage

	defer func() {
		inputMessage = _inputMessage
		sendMessage = _sendMessage
		receiveMessage = _receiveMessage
	}()

	for _, testCase := range RunPubClientTestCases {
		t.Run(testCase.Description, func(t *testing.T) {

			var goroutines int
			var quicDialAddrCount int
			var quicDialChan = make(chan bool, 1)

			quicDialAddr = func(ctx context.Context, addr string, tlsConf *tls.Config, conf *quic.Config) (quic.Connection, error) {
				quicDialAddrCount++
				<-quicDialChan
				var conn quic.Connection
				return conn, testCase.QuicDialAddrError
			}

			inputMessage = func(ctx context.Context) {
				goroutines++
			}
			sendMessage = func(ctx context.Context, conn quic.Connection) {
				goroutines++
			}
			receiveMessage = func(ctx context.Context, conn quic.Connection) {
				goroutines++
			}
			streams.AcceptStream = func(ctx context.Context, conn quic.Connection, acceptStreamChan chan quic.Stream) error {
				for {
					if testCase.AcceptStreamError != nil {
						break
					}
				}
				return testCase.AcceptStreamError
			}

			go RunPubClient(context.TODO())
			quicDialChan <- true
			time.Sleep(1 * time.Second)

			assert.Equal(t, testCase.GoroutinesExecuted, goroutines)
		})
	}
}
