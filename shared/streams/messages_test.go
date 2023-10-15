package streams

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/stretchr/testify/assert"
)

var SendMessageTestCases = []struct {
	Description      string
	WriteStreamError error
}{
	{
		Description:      "Successfully sends message",
		WriteStreamError: nil,
	},
	{
		Description:      "Fails to send message",
		WriteStreamError: errors.New("testError"),
	},
}

func TestSendMessage(t *testing.T) {

	_WriteStream := WriteStream
	defer func() {
		WriteStream = _WriteStream
	}()

	testMessageChan := make(chan MessageStream, 1)
	var testConnection quic.Connection

	for _, testCase := range SendMessageTestCases {
		t.Run(testCase.Description, func(t *testing.T) {
			var executed int

			ctx := context.TODO()

			WriteStream = func(conn quic.Connection, msg []byte) error {
				executed++
				return testCase.WriteStreamError
			}

			go SendMessage(ctx, testConnection, testMessageChan)
			testMessageChan <- MessageStream{}
			time.Sleep(2 * time.Second)
			assert.Equal(t, 1, executed)
		})
	}
}

var ReceiveMessageTestCases = []struct {
	Description           string
	ExpectedMessage       MessageStream
	ReadStreamResponse    []byte
	ReadStreamError       error
	ExpectedExecutedCount int
}{
	{
		Description:        "Succesfully reads message and executes postReceiveMessage function",
		ExpectedMessage:    MessageStream{Message: "test"},
		ReadStreamResponse: []byte(`{"message":"test"}`),
		ReadStreamError: nil,
		ExpectedExecutedCount: 2,
	},
	{
		Description: "Fails to read message and doesn't execute postReceiveMessage function",
		ExpectedMessage: MessageStream{},
		ReadStreamResponse: []byte{},
		ReadStreamError: errors.New("testError"),
		ExpectedExecutedCount: 1,
	},
}

func TestReceiveMessage(t *testing.T) {

	_ReadStream := ReadStream
	defer func() {
		ReadStream = _ReadStream
	}()

	testAcceptStreamChan := make(chan quic.Stream, 1)
	var testStream quic.Stream

	for _, testCase := range ReceiveMessageTestCases {
		t.Run(testCase.Description, func(t *testing.T) {
			var executed int

			ctx := context.TODO()

			ReadStream = func(stream quic.Stream) ([]byte, error) {
				executed++
				return testCase.ReadStreamResponse, testCase.ReadStreamError
			}

			postReceiveMessage := func(actualMessage MessageStream) {
				executed++
				assert.Equal(t, testCase.ExpectedMessage, actualMessage)
			}

			go ReceiveMessage(ctx, testAcceptStreamChan, postReceiveMessage)
			testAcceptStreamChan <- testStream
			time.Sleep(1 * time.Second)
			assert.Equal(t, testCase.ExpectedExecutedCount, executed)
		})
	}
}
