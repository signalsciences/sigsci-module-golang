package sigsci

//go:generate msgp -unexported -tests=false

//
// This is for messages to and from the agent
//

// RPCMsgIn is the primary message from the webserver module to the agent
//
type RPCMsgIn struct {
	AccessKeyID string /* AccessKeyID optional, what Site does this belong too */

	ModuleVersion string /* The module build version */
	ServerVersion string /* Main server identifier "apache 2.0.46..." */
	ServerFlavor  string /* Any other webserver configuration info  (optional) */
	ServerName    string /* as in request website URL */
	Timestamp     int64  /* Start of request in the number of seconds elapsed since January 1, 1970 UTC. */
	NowMillis     int64  /* Current time, the number of milliseconds elapsed since January 1, 1970 UTC. */
	RemoteAddr    string /* Remote IP Address, from request socket */
	Method        string /* GET/POST, etc... */
	Scheme        string /* http/https */
	URI           string /* /path?query */
	Protocol      string /* HTTP protocol */
	TLSProtocol   string // e.g. TLSv1.2
	TLSCipher     string // e.g. ECDHE-RSA-AES128-GCM-SHA256

	WAFResponse    int32       /* optional */
	ResponseCode   int32       /* HTTP Response Status Code, -1 if unknown */
	ResponseMillis int64       /* HTTP Milliseconds - How many milliseconds did the full request take, -1 if unknown */
	ResponseSize   int64       /* HTTP Response size, -1 if unknown */
	HeadersIn      [][2]string /* nil ok */
	HeadersOut     [][2]string
	PostBody       string /* empty string if none */
}

// RPCMsgOut is sent back to the webserver
// it contains a a HTTP response code, and an optional UUID
type RPCMsgOut struct {
	WAFResponse    int32
	RequestID      string      `json:",omitempty"` /* optional */
	RequestHeaders [][2]string `json:",omitempty"` /* optional, to set additional request headers */
}

// RPCMsgIn2 is a follow-up message from the webserver to the Agent
// Note there is no formal response to this message
type RPCMsgIn2 struct {
	RequestID      string /* the request id */
	ResponseCode   int32  /* what http status code did the webserver send back */
	ResponseMillis int64  /* How many milliseconds did the full request take */
	ResponseSize   int64  /* how many bytes did the webserver send back */
	HeadersOut     [][2]string
}
