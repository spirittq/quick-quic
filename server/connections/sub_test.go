package connections

import (
	"context"
	"crypto/tls"
	"errors"
	"testing"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/stretchr/testify/assert"
)

var InitSubServerTestCases = []struct {
	Description            string
	QuicListenAddrResposne *quic.Listener
	QuicListenAddrError    error
	LnAcceptError          error
	LnAcceptExecutionCount int
	Success                bool
}{
	{
		Description:            "Successfully initiates sub server",
		QuicListenAddrResposne: &quic.Listener{},
		QuicListenAddrError:    nil,
		LnAcceptError:          nil,
		LnAcceptExecutionCount: 2,
		Success:                true,
	},
	{
		Description:            "Fails to initiate sub server",
		QuicListenAddrResposne: nil,
		QuicListenAddrError:    errors.New("error"),
		LnAcceptError:          nil,
		LnAcceptExecutionCount: 0,
		Success:                false,
	},
	{
		Description:            "If accepting connection fails, then continues to accept another one",
		QuicListenAddrResposne: &quic.Listener{},
		QuicListenAddrError:    nil,
		LnAcceptError:          errors.New("error"),
		LnAcceptExecutionCount: 2,
		Success:                true,
	},
}

func TestInitSubServer(t *testing.T) {

	_handleSubClient := handleSubClient

	defer func() {
		handleSubClient = _handleSubClient
	}()

	for _, testCase := range InitSubServerTestCases {
		t.Run(testCase.Description, func(t *testing.T) {

			var actualLnAcceptExecutionCount int
			acceptedChan := make(chan bool, 1)

			quicListenAddr = func(addr string, tlsConf *tls.Config, config *quic.Config) (*quic.Listener, error) {
				return testCase.QuicListenAddrResposne, testCase.QuicListenAddrError
			}

			lnAccept = func(l *quic.Listener, ctx context.Context) (quic.Connection, error) {
				actualLnAcceptExecutionCount++
				<-acceptedChan
				var conn quic.Connection
				return conn, testCase.LnAcceptError
			}

			handleSubClient = func(ctx context.Context, conn quic.Connection) {}

			err := InitSubServer(context.TODO(), &tls.Config{}, &quic.Config{})
			acceptedChan <- true
			if testCase.Success {
				assert.Nil(t, err)
			} else {
				assert.Error(t, err)
			}
			time.Sleep(1 * time.Second)
			assert.Equal(t, testCase.LnAcceptExecutionCount, actualLnAcceptExecutionCount)
		})
	}
}
