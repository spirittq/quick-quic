package shared

const (
	ServerAddr = "localhost"
	PortPub    = "3001"
	PortSub    = "3002"
)

const (
	MaxIdleTimeout  = 5
	KeepAlivePeriod = 5
)

type ContextKey string

const (
	ContextName ContextKey = "ContextName"
)

type ContextValue string

const (
	PubClientReceiveCtx ContextValue = "PubClientReceiveCtx"
	PubClientInputCtx   ContextValue = "PubClientInputCtx"
	PubClientSendCtx    ContextValue = "PubClientSendCtx"
	SubClientSendCtx    ContextValue = "SubClientSendCtx"
)
