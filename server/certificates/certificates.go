package certificates

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"log"
	"math/big"
)

// Basic TLS config for for representation, taken from https://github.com/quic-go/quic-go/blob/master/example/echo/echo.go
var GenerateTLSConfig = func() *tls.Config {
	logger := log.Default()

	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		logger.Panicln(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		logger.Panicln(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		logger.Panicln(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
	}
}
