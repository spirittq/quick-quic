package streams

const (
	ServerAddr  = "localhost"
	PortPub     = "3001"
	PortSub     = "3002"
	CertPath    = "../certificates/"
	CertTemp    = "ca.crt"
	CertKeyTemp = "ca.key"
)

const (
	MaxIdleTimeout  = 5
	KeepAlivePeriod = 5
)

type PostReceiveMessage func(MessageStream)
