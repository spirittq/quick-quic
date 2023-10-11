package core

import (
	"crypto/tls"
	"sync"

	"github.com/quic-go/quic-go"
)

type Counter struct {
	Count int
	Mu    sync.Mutex
}

var pubCount, subCount Counter
var tlsConfig *tls.Config
var quicConfig *quic.Config
var messageChan chan string
var subConnectedChan, subNotConnectedChan chan bool
