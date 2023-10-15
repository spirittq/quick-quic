package core

import (
	"context"
	"crypto/tls"
	"errors"
	"shared/streams"
	"testing"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/stretchr/testify/assert"
)

var RunSubClientTestCases = []struct {
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
		GoroutinesExecuted: 1,
		QuicDialAddrCount:  1,
	},
	{
		Description:        "Lost connection to the server and retries",
		AcceptStreamError:  errors.New("error"),
		QuicDialAddrError:  nil,
		GoroutinesExecuted: 1,
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

func TestRunSubClient(t *testing.T) {

	_receiveMessage := receiveMessage

	defer func() {
		receiveMessage = _receiveMessage
	}()

	for _, testCase := range RunSubClientTestCases {
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

			go RunSubClient(context.TODO())
			quicDialChan <- true
			time.Sleep(1 * time.Second)

			assert.Equal(t, testCase.GoroutinesExecuted, goroutines)
		})
	}
}
