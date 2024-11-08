package sigsci

//go:generate go run github.com/tinylib/msgp@v1.2.4 -unexported -tests=false

//
// This is for messages to and from the agent
//

// RPCMsgIn is the primary message from the webserver module to the agent
type RPCMsgIn struct {
	AccessKeyID    string      // AccessKeyID optional, what Site does this belong too (deprecated)
	ModuleVersion  string      // The module build version
	ServerVersion  string      // Main server identifier "apache 2.0.46..."
	ServerFlavor   string      // Any other webserver configuration info  (optional)
	ServerName     string      // As in request website URL
	Timestamp      int64       // Start of request in the number of seconds elapsed since January 1, 1970 UTC.
	NowMillis      int64       // Current time, the number of milliseconds elapsed since January 1, 1970 UTC.
	RemoteAddr     string      // Remote IP Address, from request socket
	Method         string      // GET/POST, etc...
	Scheme         string      // http/https
	URI            string      // /path?query
	Protocol       string      // HTTP protocol
	TLSProtocol    string      // e.g. TLSv1.2
	TLSCipher      string      // e.g. ECDHE-RSA-AES128-GCM-SHA256
	WAFResponse    int32       // Optional
	ResponseCode   int32       // HTTP Response Status Code, -1 if unknown
	ResponseMillis int64       // HTTP Milliseconds - How many milliseconds did the full request take, -1 if unknown
	ResponseSize   int64       // HTTP Response size, -1 if unknown
	HeadersIn      [][2]string // HTTP Request headers (slice of name/value pairs); nil ok
	HeadersOut     [][2]string // HTTP Response headers (slice of name/value pairs); nil ok
	PostBody       string      // HTTP Request body; empty string if none
}

// RPCMsgOut is sent back to the webserver
type RPCMsgOut struct {
	WAFResponse    int32
	RequestID      string      `json:",omitempty"`                  // Set if the server expects an UpdateRequest with this ID (UUID)
	RequestHeaders [][2]string `json:",omitempty"`                  // Any additional information in the form of additional request headers
	RespActions    []Action    `json:",omitempty" msg:",omitempty"` // Add or Delete application response headers
}

const (
	AddHdr int8 = iota + 1
	SetHdr
	SetNEHdr
	DelHdr
)

//msgp:tuple Action
type Action struct {
	Code int8
	Args []string
}

// RPCMsgIn2 is a follow-up message from the webserver to the Agent
// Note there is no formal response to this message
type RPCMsgIn2 struct {
	RequestID      string // The request id (UUID)
	ResponseCode   int32  // HTTP status code did the webserver send back
	ResponseMillis int64  // How many milliseconds did the full request take
	ResponseSize   int64  // how many bytes did the webserver send back
	HeadersOut     [][2]string
}
