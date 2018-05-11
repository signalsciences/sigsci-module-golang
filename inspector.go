package sigsci

// Inspector is an interface to implement how the
// module communicates with the inspection engine.
type Inspector interface {
	InitModule(*RPCMsgIn, *RPCMsgOut) error
	PreRequest(*RPCMsgIn, *RPCMsgOut) error
	PostRequest(*RPCMsgIn, *RPCMsgOut) error
	UpdateRequest(*RPCMsgIn2, *RPCMsgOut) error
}
