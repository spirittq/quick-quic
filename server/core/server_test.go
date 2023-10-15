package core

import (
	"context"
	"crypto/tls"
	"errors"
	"server/certificates"
	"server/connections"
	"testing"

	"github.com/quic-go/quic-go"
	"github.com/stretchr/testify/assert"
)

var RunServerTestCases = []struct {
	Description               string
	GenerateTLSConfigResponse *tls.Config
	GenerateTLSConfigError    error
	InitPubServerError        error
	InitSubServerError        error
	Success                   bool
}{
	{
		Description:               "Successfully runs server",
		GenerateTLSConfigResponse: &tls.Config{},
		GenerateTLSConfigError:    nil,
		InitPubServerError:        nil,
		InitSubServerError:        nil,
		Success:                   true,
	},
	{
		Description:               "Fails to run server due to TLS config error",
		GenerateTLSConfigResponse: nil,
		GenerateTLSConfigError:    errors.New("error"),
		InitPubServerError:        nil,
		InitSubServerError:        nil,
		Success:                   false,
	},
	{
		Description:               "Fails to run server due to pub server init error",
		GenerateTLSConfigResponse: nil,
		GenerateTLSConfigError:    nil,
		InitPubServerError:        errors.New("error"),
		InitSubServerError:        nil,
		Success:                   false,
	},
	{
		Description:               "Fails to run server due to sub server init error",
		GenerateTLSConfigResponse: nil,
		GenerateTLSConfigError:    nil,
		InitPubServerError:        nil,
		InitSubServerError:        errors.New("error"),
		Success:                   false,
	},
}

func TestRunServer(t *testing.T) {
	for _, testCase := range RunServerTestCases {
		t.Run(testCase.Description, func(t *testing.T) {

			certificates.GenerateTLSConfig = func() (*tls.Config, error) {
				return testCase.GenerateTLSConfigResponse, testCase.GenerateTLSConfigError
			}

			connections.InitPubServer = func(ctx context.Context, tlsConfig *tls.Config, quicConfig *quic.Config) error {
				return testCase.InitPubServerError
			}

			connections.InitSubServer = func(ctx context.Context, tlsConfig *tls.Config, quicConfig *quic.Config) error {
				return testCase.InitSubServerError
			}

			err := RunServer(context.TODO())

			if testCase.Success {
				assert.Nil(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
