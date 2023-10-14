package core

import (
	"context"
	"crypto/tls"
	"errors"
	"server/certificates"
	"server/connections"
	"testing"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/stretchr/testify/assert"
)

var RunServerTestCases = []struct {
	Description               string
	GenerateTLSConfigResponse *tls.Config
	GenerateTLSConfigError    error
	Success                   bool
	ExpectedExecutedCount     int
}{
	{
		Description:               "Successfully runs server",
		GenerateTLSConfigResponse: &tls.Config{},
		GenerateTLSConfigError:    nil,
		Success:                   true,
		ExpectedExecutedCount:     2,
	},
	{
		Description:               "Fails to run server due to TLS config error",
		GenerateTLSConfigResponse: nil,
		GenerateTLSConfigError:    errors.New("error"),
		Success:                   false,
		ExpectedExecutedCount:     0,
	},
}

func TestRunServer(t *testing.T) {
	for _, testCase := range RunServerTestCases {
		t.Run(testCase.Description, func(t *testing.T) {

			var executed int

			certificates.GenerateTLSConfig = func() (*tls.Config, error) {
				return testCase.GenerateTLSConfigResponse, testCase.GenerateTLSConfigError
			}

			connections.InitPubServer = func(ctx context.Context, tlsConfig *tls.Config, quicConfig *quic.Config) {
				executed++
			}

			connections.InitSubServer = func(ctx context.Context, tlsConfig *tls.Config, quicConfig *quic.Config) {
				executed++
			}

			err := RunServer(context.TODO())
			time.Sleep(2 * time.Second)

			if testCase.Success {
				assert.Nil(t, err)
			} else {
				assert.Error(t, err)
			}
			assert.Equal(t, testCase.ExpectedExecutedCount, executed)
		})
	}
}
