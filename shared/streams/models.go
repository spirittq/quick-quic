package streams

// Struct contains only message string now, but can have more information stored upon need. Previously had 'Empty' bool to decide whether to
// prompt publisher to input message or message is received from goroutine of previous disrupted connection.
// See commint 4b9aa3e
type MessageStream struct {
	Message string `json:"message"`
	// Empty   bool
}
