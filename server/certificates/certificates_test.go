package certificates

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

var GenerateTLSConfigTestCases = []struct {
	Description                string
	RsaGenerateKeyError        error
	x509CreateCertificateError error
	tlsX509KeyPairError        error
	Success                    bool
}{
	{
		Description:                "Successfully generate tls config",
		RsaGenerateKeyError:        nil,
		x509CreateCertificateError: nil,
		tlsX509KeyPairError:        nil,
		Success:                    true,
	},
	{
		Description:                "Returns error if rsaGenerateKey fails",
		RsaGenerateKeyError:        errors.New("error"),
		x509CreateCertificateError: nil,
		tlsX509KeyPairError:        nil,
		Success:                    false,
	},
	{
		Description:                "Returns error if x509CreateCertificate fails",
		RsaGenerateKeyError:        nil,
		x509CreateCertificateError: errors.New("error"),
		tlsX509KeyPairError:        nil,
		Success:                    false,
	},
	{
		Description:                "Returns error if tlsX509KeyPair fails",
		RsaGenerateKeyError:        nil,
		x509CreateCertificateError: nil,
		tlsX509KeyPairError:        errors.New("error"),
		Success:                    false,
	},
}

func TestGenerateTLSConfig(t *testing.T) {
	for _, testCase := range GenerateTLSConfigTestCases {
		t.Run(testCase.Description, func(t *testing.T) {

			_rsaGenerateKey := rsaGenerateKey
			_x509CreateCertificate := x509CreateCertificate
			_tlsX509KeyPair := tlsX509KeyPair

			defer func() {
				rsaGenerateKey = _rsaGenerateKey
				x509CreateCertificate = _x509CreateCertificate
				tlsX509KeyPair = _tlsX509KeyPair
			}()

			if testCase.RsaGenerateKeyError != nil {
				rsaGenerateKey = func(random io.Reader, bits int) (*rsa.PrivateKey, error) {
					return nil, testCase.RsaGenerateKeyError
				}
			}

			if testCase.x509CreateCertificateError != nil {
				x509CreateCertificate = func(rand io.Reader, template, parent *x509.Certificate, pub, priv any) ([]byte, error) {
					return nil, testCase.x509CreateCertificateError
				}
			}
			if testCase.tlsX509KeyPairError != nil {
				tlsX509KeyPair = func(certPEMBlock, keyPEMBlock []byte) (tls.Certificate, error) {
					return tls.Certificate{}, testCase.tlsX509KeyPairError
				}
			}

			tlsConfig, err := GenerateTLSConfig()

			if testCase.Success {
				assert.NotEmpty(t, tlsConfig)
				assert.Nil(t, err)
			} else {
				assert.Nil(t, tlsConfig)
				assert.Error(t, err)
			}
		})
	}
}
