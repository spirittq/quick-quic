package connections

import (
	"shared/streams"
	"sync"

	"github.com/quic-go/quic-go"
)

type Counter struct {
	Count int
	Mu    sync.Mutex
}

var PubCount, SubCount Counter
var MessageChan, NewSubChan, NoSubsChan chan streams.MessageStream
var acceptStreamPubChan, acceptStreamSubChan chan quic.Stream

var NewSubMessage = streams.MessageStream{
	Message: "New subscriber has connected",
}

var NoSubsMessage = streams.MessageStream{
	Message: "No subscribers are connected",
}
