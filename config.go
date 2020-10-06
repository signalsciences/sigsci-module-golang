package sigsci

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"runtime"
	"time"
)

var (
	// DefaultAllowUnknownContentLength is the default value
	DefaultAllowUnknownContentLength = false
	// DefaultAnomalyDuration is the default value
	DefaultAnomalyDuration = 1 * time.Second
	// DefaultAnomalySize is the default value
	DefaultAnomalySize = int64(512 * 1024)
	// DefaultDebug is the default value
	DefaultDebug = false
	// DefaultInspector is the default value
	DefaultInspector = Inspector(nil)
	// DefaultMaxContentLength is the default value
	DefaultMaxContentLength = int64(100000)
	// DefaultModuleIdentifier is the default value
	DefaultModuleIdentifier = "sigsci-module-golang " + version
	// DefaultRPCAddress is the default value
	DefaultRPCAddress = "/var/run/sigsci.sock"
	// DefaultRPCNetwork is the default value
	DefaultRPCNetwork = "unix"
	// DefaultTimeout is the default value
	DefaultTimeout = 100 * time.Millisecond
	// DefaultServerIdentifier is the default value
	DefaultServerIdentifier = runtime.Version()
	// DefaultServerFlavor is the default value
	DefaultServerFlavor = ""
)

// HeaderExtractorFunc is a header extraction function
type HeaderExtractorFunc func(*http.Request) (http.Header, error)

// ModuleConfig is a configuration object for a Module
type ModuleConfig struct {
	allowUnknownContentLength bool
	anomalyDuration           time.Duration
	anomalySize               int64
	debug                     bool
	headerExtractor           HeaderExtractorFunc
	inspector                 Inspector
	inspInit                  InspectorInitFunc
	inspFini                  InspectorFiniFunc
	maxContentLength          int64
	moduleIdentifier          string
	rpcAddress                string
	rpcNetwork                string
	serverIdentifier          string
	serverFlavor              string
	timeout                   time.Duration
}

// NewModuleConfig returns an object with any options set
func NewModuleConfig(options ...ModuleConfigOption) (*ModuleConfig, error) {
	c := &ModuleConfig{
		allowUnknownContentLength: DefaultAllowUnknownContentLength,
		anomalyDuration:           DefaultAnomalyDuration,
		anomalySize:               DefaultAnomalySize,
		debug:                     DefaultDebug,
		headerExtractor:           nil,
		inspector:                 DefaultInspector,
		inspInit:                  nil,
		inspFini:                  nil,
		maxContentLength:          DefaultMaxContentLength,
		moduleIdentifier:          DefaultModuleIdentifier,
		rpcAddress:                DefaultRPCAddress,
		rpcNetwork:                DefaultRPCNetwork,
		serverIdentifier:          DefaultServerIdentifier,
		serverFlavor:              DefaultServerFlavor,
		timeout:                   DefaultTimeout,
	}
	if err := c.SetOptions(options...); err != nil {
		return nil, err
	}
	return c, nil
}

// SetOptions sets options specified as functional arguments
func (c *ModuleConfig) SetOptions(options ...ModuleConfigOption) error {
	for _, opt := range options {
		if opt == nil {
			continue
		}
		err := opt(c)
		if err != nil {
			return err
		}
	}
	return nil
}

// IsBlockCode returns true if the code is a configured block code
func (c *ModuleConfig) IsBlockCode(code int) bool {
	return code >= 300 && code <= 599
}

// IsAllowCode returns true if the code is an allow code
func (c *ModuleConfig) IsAllowCode(code int) bool {
	return code == 200
}

// AllowUnknownContentLength returns the configuration value
func (c *ModuleConfig) AllowUnknownContentLength() bool {
	return c.allowUnknownContentLength
}

// AltResponseCodes returns the configuration value
//
// Deprecated: The `AltResponseCodes` concept has
// been deprecated in favor of treating all response
// codes 300-599 as blocking codes. Due to
// this, this method will always return nil. It is left
// here to avoid breakage, but will eventually be removed.
func (c *ModuleConfig) AltResponseCodes() []int {
	return nil
}

// AnomalyDuration returns the configuration value
func (c *ModuleConfig) AnomalyDuration() time.Duration {
	return c.anomalyDuration
}

// AnomalySize returns the configuration value
func (c *ModuleConfig) AnomalySize() int64 {
	return c.anomalySize
}

// Debug returns the configuration value
func (c *ModuleConfig) Debug() bool {
	return c.debug
}

// HeaderExtractor returns the configuration value
func (c *ModuleConfig) HeaderExtractor() func(r *http.Request) (http.Header, error) {
	return c.headerExtractor
}

// Inspector returns the inspector
func (c *ModuleConfig) Inspector() Inspector {
	return c.inspector
}

// InspectorInit returns the inspector init function
func (c *ModuleConfig) InspectorInit() InspectorInitFunc {
	return c.inspInit
}

// InspectorFini returns the inspector fini function
func (c *ModuleConfig) InspectorFini() InspectorFiniFunc {
	return c.inspFini
}

// MaxContentLength returns the configuration value
func (c *ModuleConfig) MaxContentLength() int64 {
	return c.maxContentLength
}

// ModuleIdentifier returns the configuration value
func (c *ModuleConfig) ModuleIdentifier() string {
	return c.moduleIdentifier
}

// RPCAddress returns the configuration value
func (c *ModuleConfig) RPCAddress() string {
	return c.rpcAddress
}

// RPCNetwork returns the configuration value
func (c *ModuleConfig) RPCNetwork() string {
	return c.rpcNetwork
}

// RPCAddressString returns the RPCNetwork/RPCAddress as a combined string
func (c *ModuleConfig) RPCAddressString() string {
	return c.rpcNetwork + ":" + c.rpcAddress
}

// ServerIdentifier returns the configuration value
func (c *ModuleConfig) ServerIdentifier() string {
	return c.serverIdentifier
}

// ServerFlavor returns the configuration value
func (c *ModuleConfig) ServerFlavor() string {
	return c.serverFlavor
}

// Timeout returns the configuration value
func (c *ModuleConfig) Timeout() time.Duration {
	return c.timeout
}

// Functional Config Options

// ModuleConfigOption is a functional config option for configuring the module
// See: https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
type ModuleConfigOption func(*ModuleConfig) error

// AllowUnknownContentLength is a function argument to set the ability
// to read the body when the content length is not specified.
//
// NOTE: This can be dangerous (fill RAM) if set when the max content
//       length is not limited by the server itself. This is intended
//       for use with gRPC where the max message receive length is limited.
//       Do NOT enable this if there is no limit set on the request
//       content length!
func AllowUnknownContentLength(allow bool) ModuleConfigOption {
	return func(c *ModuleConfig) error {
		c.allowUnknownContentLength = allow
		return nil
	}
}

// AltResponseCodes is a function argument to set the alternative
// response codes from the agent that are considered "blocking"
//
// Deprecated: The `AltResponseCodes` concept has
// been deprecated in favor of treating all response
// codes 300-599 as blocking codes. Due to
// this, this method will always return nil. It is left
// here to avoid breakage, but will eventually be removed.
func AltResponseCodes(codes ...int) ModuleConfigOption {
	return nil
}

// AnomalyDuration is a function argument to indicate when to send data
// to the inspector if the response was abnormally slow
func AnomalyDuration(dur time.Duration) ModuleConfigOption {
	return func(c *ModuleConfig) error {
		c.anomalyDuration = dur
		return nil
	}
}

// AnomalySize is a function argument to indicate when to send data to
// the inspector if the response was abnormally large
func AnomalySize(size int64) ModuleConfigOption {
	return func(c *ModuleConfig) error {
		c.anomalySize = size
		return nil
	}
}

// CustomInspector is a function argument that sets a custom inspector,
// an optional inspector initializer to decide if inspection should occur, and
// an optional inspector finalizer that can perform any post-inspection steps
func CustomInspector(insp Inspector, init InspectorInitFunc, fini InspectorFiniFunc) ModuleConfigOption {
	return func(c *ModuleConfig) error {
		c.inspector = insp
		c.inspInit = init
		c.inspFini = fini
		return nil
	}
}

// CustomHeaderExtractor is a function argument that sets a function to extract
// an alternative header object from the request. It is primarily intended only
// for internal use.
func CustomHeaderExtractor(fn func(r *http.Request) (http.Header, error)) ModuleConfigOption {
	return func(c *ModuleConfig) error {
		c.headerExtractor = fn
		return nil
	}
}

// Debug turns on debug logging
func Debug(enable bool) ModuleConfigOption {
	return func(c *ModuleConfig) error {
		c.debug = enable
		return nil
	}
}

// MaxContentLength is a function argument to set the maximum post
// body length that will be processed
func MaxContentLength(size int64) ModuleConfigOption {
	return func(c *ModuleConfig) error {
		c.maxContentLength = size
		return nil
	}
}

// Socket is a function argument to set where to send data to the
// Signal Sciences Agent. The network argument should be `unix`
// or `tcp` and the corresponding address should be either an absolute
// path or an `address:port`, respectively.
func Socket(network, address string) ModuleConfigOption {
	return func(c *ModuleConfig) error {
		switch network {
		case "unix":
			if !filepath.IsAbs(address) {
				return errors.New(`address must be an absolute path for network="unix"`)
			}
		case "tcp":
			if _, _, err := net.SplitHostPort(address); err != nil {
				return fmt.Errorf(`address must be in "address:port" form for network="tcp": %s`, err)
			}
		default:
			return errors.New(`network must be "tcp" or "unix"`)
		}

		c.rpcNetwork = network
		c.rpcAddress = address

		return nil
	}
}

// Timeout is a function argument that sets the maximum time to wait until
// receiving a reply from the inspector. Once this timeout is reached, the
// module will fail open.
func Timeout(t time.Duration) ModuleConfigOption {
	return func(c *ModuleConfig) error {
		c.timeout = t
		return nil
	}
}

// ModuleIdentifier is a function argument that sets the module name
// and version for custom setups.
// The version should be a sem-version (e.g., "1.2.3")
func ModuleIdentifier(name, version string) ModuleConfigOption {
	return func(c *ModuleConfig) error {
		c.moduleIdentifier = name + " " + version
		return nil
	}
}

// ServerIdentifier is a function argument that sets the server
// identifier for custom setups
func ServerIdentifier(id string) ModuleConfigOption {
	return func(c *ModuleConfig) error {
		c.serverIdentifier = id
		return nil
	}
}

// ServerFlavor is a function argument that sets the server
// flavor for custom setups using revproxy.
func ServerFlavor(serverModule string) ModuleConfigOption {
	return func(c *ModuleConfig) error {
		c.serverFlavor = serverModule
		return nil
	}
}

// FromModuleConfig allow cloning the config
func FromModuleConfig(mcfg *ModuleConfig) ModuleConfigOption {
	return func(c *ModuleConfig) error {
		*c = *mcfg
		return nil
	}
}
