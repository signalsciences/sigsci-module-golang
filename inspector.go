package sigsci

import "net/http"

// InspectorInitFunc is called to decide if the request should be inspected
// Return true if inspection should occur for the request or false if
// inspection should be bypassed.
type InspectorInitFunc func(*http.Request) bool

// InspectorFiniFunc is called after any inspection on the request is completed
type InspectorFiniFunc func(*http.Request)

// Inspector is an interface to implement how the
// module communicates with the inspection engine.
type Inspector interface {
	ModuleInit(*RPCMsgIn, *RPCMsgOut) error
	PreRequest(*RPCMsgIn, *RPCMsgOut) error
	PostRequest(*RPCMsgIn, *RPCMsgOut) error
	UpdateRequest(*RPCMsgIn2, *RPCMsgOut) error
}
