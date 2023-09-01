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
	// ModuleInit can be called when the module starts up. This allows the module
	// data (e.g., `ModuleVersion`, `ServerVersion`, `ServerFlavor`, etc.) to be
	// sent to the collector so that the agent shows up initialized without having
	// to wait for data to be sent through the inspector. This should only be called
	// once when the app/module starts.
	ModuleInit(*RPCMsgIn, *RPCMsgOut) error
	// PreRequest is called before the request is processed by the app. The results
	// should be analyzed for any anomalies or blocking conditions. In addition, any
	// `RequestID` returned in the response should be recorded for future use.
	PreRequest(*RPCMsgIn, *RPCMsgOut) error
	// PostRequest is called after the request has been processed by the app and the
	// response data (e.g., status code, headers, etc.) is available. This should be
	// called if there was NOT a `RequestID` in the response to a previous `PreRequest`
	// call for the same transaction (if a `RequestID` was in the response, then it
	// should be used in an `UpdateRequest` call instead).
	PostRequest(*RPCMsgIn, *RPCMsgOut) error
	// UpdateRequest is called after the request has been processed by the app and the
	// response data (e.g., status code, headers, etc.) is available. This should be used
	// instead of a `PostRequest` call when a prior `PreRequest` call for the same
	// transaction included a `RequestID`. In this case, this call is updating the data
	// collected in the `PreRequest` with the given response data (e.g., status code,
	// headers, etc.).
	UpdateRequest(*RPCMsgIn2, *RPCMsgOut) error
	// LogRequest can be optionally called if the waf-data-log-all config option in the agent
	// is set to true.  This is used to log to the waf-data-log files specified in the config option
	// and this RPC call is required because in order to log once per HTTP request we need to know
	// upfront which of the following PreRequest, UpdateRequest or PostRequest calls will be made
	// and this information is only available in the context of the module
	LogRequest(*RPCMsgIn, *RPCMsgOut) error
}
