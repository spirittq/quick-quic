package connections

import "sync"

type Counter struct {
	Count int
	Mu    sync.Mutex
}

var pubCount, subCount Counter
var MessageChan chan []byte
var SubConnectedChan, SubNotConnectedChan chan bool
