package core

import (
	"bufio"
	"context"
	"errors"
	"shared/streams"
	"testing"
	"time"

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
		Description: "Successfully reads input and puts message into channel",
		ReadStringResponse: "test message",
		ReadStringError: nil,
		ExpectedChannelLen: 1,
		Success: true,
	},
	{
		Description: "Fails to read input",
		ReadStringResponse: "",
		ReadStringError: errors.New("error"),
		ExpectedChannelLen: 0,
		Success: false,
	},
}

func TestInputMessage(t *testing.T) {

	for _, testCase := range InputMessageTestCases {
		t.Run(testCase.Description, func(t *testing.T) {

			jugglingMessage = make(chan streams.MessageStream, 1)
			readStringChan := make(chan bool, 1)

			ReadString = func(b *bufio.Reader, delim byte) (string, error) {
				<- readStringChan
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
