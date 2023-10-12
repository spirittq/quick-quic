package connections

import (
	"shared"
	"sync"

	"github.com/quic-go/quic-go"
)

type Counter struct {
	Count int
	Mu    sync.Mutex
}

var PubCount, SubCount Counter
var MessageChan, NewSubChan, NoSubsChan chan shared.MessageStream
var acceptStreamPubChan, acceptStreamSubChan chan quic.Stream

var NewSubMessage = shared.MessageStream{
	Message: "New subscriber has connected",
}

var NoSubsMessage = shared.MessageStream{
	Message: "No subscribers are connected",
}
