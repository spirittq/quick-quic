package connections

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

var receiveMessageFromPubTestCases = []struct {
	Description string
	Subcount    int
}{
	{
		Description: "Successfully receives message from pub client and sends it to the channel if sub is connected",
		Subcount:    1,
	},
	{
		Description: "Successfully receive message from pub client but doesn't sent it because no subs are connected",
		Subcount:    0,
	},
	{
		Description: "Successfully receives message from pub client and sends it to the channel x amount of subs connected",
		Subcount:    5,
	},
}

func TestReceiveMessageFromPub(t *testing.T) {

	var testConnection quic.Connection
	var testStream quic.Stream

	for _, testCase := range receiveMessageFromPubTestCases {
		t.Run(testCase.Description, func(t *testing.T) {
			MessageChan = make(chan streams.MessageStream, 1)
			acceptStreamPubChan = make(chan quic.Stream, 1)
			var messageSentTimes int

			expectedMessage := streams.MessageStream{
				Message: "testMessage",
			}

			SubCount.Count = testCase.Subcount

			streams.ReadStream = func(stream quic.Stream) ([]byte, error) {
				return []byte(`{"message":"testMessage"}`), nil
			}

			acceptStreamPubChan <- testStream
			go receiveMessageFromPub(context.TODO(), testConnection)
			time.Sleep(1 * time.Second)

			for i := 0; i < SubCount.Count; i++ {
				actualMessage := <-MessageChan
				assert.Equal(t, expectedMessage, actualMessage)
				messageSentTimes++
			}
			assert.Equal(t, testCase.Subcount, messageSentTimes)
		})
	}
}

var MonitorSubTestCases = []struct {
	Description           string
	SubCount              int
	PubCount              int
	expectedChannelLength int
}{
	{
		Description:           "If no subscribers are connected, message is sent to the channel (1 publisher is connected)",
		SubCount:              0,
		PubCount:              1,
		expectedChannelLength: 1,
	},
	{
		Description:           "If no subscribers are connected, message is sent to the channel (multiple publishers are connected)",
		SubCount:              0,
		PubCount:              5,
		expectedChannelLength: 1,
	},
	{
		Description:           "If no subscribers are connected, message is not sent to the channel (if no publishers are connected)",
		SubCount:              0,
		PubCount:              0,
		expectedChannelLength: 0,
	},
	{
		Description:           "If 1 subscriber is connected, message is not sent to the channel",
		SubCount:              1,
		PubCount:              1,
		expectedChannelLength: 0,
	},
	{
		Description:           "if more than 1 subscriber is connected, message is not sent to the channel",
		SubCount:              5,
		PubCount:              5,
		expectedChannelLength: 0,
	},
}

func TestMonitorsSub(t *testing.T) {
	for _, testCase := range MonitorSubTestCases {
		t.Run(testCase.Description, func(t *testing.T) {
			NoSubsChan = make(chan streams.MessageStream, 1)
			SubCount.Count = testCase.SubCount
			PubCount.Count = testCase.PubCount
			go monitorSubs()
			time.Sleep(1 * time.Second)
			assert.Equal(t, testCase.expectedChannelLength, len(NoSubsChan))
		})
	}
}

var InitPubServerTestCases = []struct {
	Description            string
	QuicListenAddrResposne *quic.Listener
	QuicListenAddrError    error
	LnAcceptError          error
	LnAcceptExecutionCount int
	Success                bool
}{
	{
		Description:            "Successfully initiates pub server",
		QuicListenAddrResposne: &quic.Listener{},
		QuicListenAddrError:    nil,
		LnAcceptError:          nil,
		LnAcceptExecutionCount: 2,
		Success:                true,
	},
	{
		Description:            "Fails to initiate pub server",
		QuicListenAddrResposne: nil,
		QuicListenAddrError:    errors.New("error"),
		LnAcceptError:          nil,
		LnAcceptExecutionCount: 0,
		Success:                false,
	},
	{
		Description: "If accepting connection fails, then continues to accept another one",
		QuicListenAddrResposne: &quic.Listener{},
		QuicListenAddrError:    nil,
		LnAcceptError:          errors.New("error"),
		LnAcceptExecutionCount: 2,
		Success:                true,
	},
}

func TestInitPubServer(t *testing.T) {

	_monitorSubs := monitorSubs
	_handlePubClient := handlePubClient

	defer func() {
		monitorSubs = _monitorSubs
		handlePubClient = _handlePubClient
	}()

	for _, testCase := range InitPubServerTestCases {
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

			monitorSubs = func() {}

			handlePubClient = func(ctx context.Context, conn quic.Connection) {}

			err := InitPubServer(context.TODO(), &tls.Config{}, &quic.Config{})
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
