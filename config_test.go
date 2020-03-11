package sigsci

import (
	"net/http"
	"testing"
	"time"
)

func TestModuleConfig(t *testing.T) {
	// Everything default
	c, err := NewModuleConfig()
	if err != nil {
		t.Fatalf("Failed to create module config: %s", err)
	}
	if c.AllowUnknownContentLength() != DefaultAllowUnknownContentLength {
		t.Errorf("Unexpected AllowUnknownContentLength: %v", c.AllowUnknownContentLength())
	}
	if c.AltResponseCodes() != nil {
		t.Errorf("Unexpected AltResponseCodes: %v", c.AltResponseCodes())
	}
	if c.AnomalyDuration() != DefaultAnomalyDuration {
		t.Errorf("Unexpected AnomalyDuration: %v", c.AnomalyDuration())
	}
	if c.AnomalySize() != DefaultAnomalySize {
		t.Errorf("Unexpected AnomalySize: %v", c.AnomalySize())
	}
	if c.Debug() != DefaultDebug {
		t.Errorf("Unexpected Debug: %v", c.Debug())
	}
	if c.HeaderExtractor() != nil {
		t.Errorf("Unexpected HeaderExtractor: %p", c.HeaderExtractor())
	}
	if c.Inspector() != DefaultInspector {
		t.Errorf("Unexpected Inspector: %v", c.Inspector())
	}
	if c.InspectorInit() != nil {
		t.Errorf("Unexpected InspectorInit: %p", c.InspectorInit())
	}
	if c.InspectorFini() != nil {
		t.Errorf("Unexpected InspectorFini: %p", c.InspectorFini())
	}
	if c.MaxContentLength() != DefaultMaxContentLength {
		t.Errorf("Unexpected MaxContentLength: %v", c.MaxContentLength())
	}
	if c.ModuleIdentifier() != DefaultModuleIdentifier {
		t.Errorf("Unexpected ModuleIdentifier: %v", c.ModuleIdentifier())
	}
	if c.RPCAddress() != DefaultRPCAddress {
		t.Errorf("Unexpected RPCAddress: %v", c.RPCAddress())
	}
	if c.RPCNetwork() != DefaultRPCNetwork {
		t.Errorf("Unexpected RPCNetwork: %v", c.RPCNetwork())
	}
	if c.RPCAddressString() != DefaultRPCNetwork+":"+DefaultRPCAddress {
		t.Errorf("Unexpected RPCAddressString: %v", c.RPCAddressString())
	}
	if c.ServerIdentifier() != DefaultServerIdentifier {
		t.Errorf("Unexpected ServerIdentifier: %v", c.ServerIdentifier())
	}
	if c.Timeout() != DefaultTimeout {
		t.Errorf("Unexpected Timeout: %v", c.Timeout())
	}
	if c.IsAllowCode(406) {
		t.Errorf("Unexpected IsAllowCode(406): %v", c.IsAllowCode(406))
	}
	if !c.IsAllowCode(200) {
		t.Errorf("Unexpected IsAllowCode(200): %v", c.IsAllowCode(200))
	}
	if !c.IsBlockCode(406) {
		t.Errorf("Unexpected IsBlockCode(406): %v", c.IsBlockCode(406))
	}
	if c.IsBlockCode(403) {
		t.Errorf("Unexpected IsBlockCode(403): %v", c.IsBlockCode(403))
	}
	if c.IsBlockCode(200) {
		t.Errorf("Unexpected IsBlockCode(200): %v", c.IsBlockCode(200))
	}

	// Everything configured non-default
	c, err = NewModuleConfig(
		AllowUnknownContentLength(true),
		AltResponseCodes(403),
		AnomalyDuration(10*time.Second),
		AnomalySize(8192),
		CustomInspector(&RPCInspector{}, func(_ *http.Request) bool { return true }, func(_ *http.Request) {}),
		CustomHeaderExtractor(func(_ *http.Request) (http.Header, error) { return nil, nil }),
		Debug(true),
		MaxContentLength(500000),
		Socket("tcp", "0.0.0.0:1234"),
		Timeout(10*time.Millisecond),
	)
	if err != nil {
		t.Fatalf("Failed to create module config: %s", err)
	}
	if c.AllowUnknownContentLength() != true {
		t.Errorf("Unexpected AllowUnknownContentLength: %v", c.AllowUnknownContentLength())
	}
	altResponseCodes := c.AltResponseCodes()
	if len(altResponseCodes) != 1 || altResponseCodes[0] != 403 {
		t.Errorf("Unexpected AltResponseCodes: %v", altResponseCodes)
	}
	if c.AnomalyDuration() != 10*time.Second {
		t.Errorf("Unexpected AnomalyDuration: %v", c.AnomalyDuration())
	}
	if c.AnomalySize() != 8192 {
		t.Errorf("Unexpected AnomalySize: %v", c.AnomalySize())
	}
	if c.Debug() != true {
		t.Errorf("Unexpected Debug: %v", c.Debug())
	}
	if c.HeaderExtractor() == nil {
		t.Errorf("Unexpected HeaderExtractor: %p", c.HeaderExtractor())
	}
	if c.Inspector() == DefaultInspector {
		t.Errorf("Unexpected Inspector: %v", c.Inspector())
	}
	if c.InspectorInit() == nil {
		t.Errorf("Unexpected InspectorInit: %p", c.InspectorInit())
	}
	if c.InspectorFini() == nil {
		t.Errorf("Unexpected InspectorFini: %p", c.InspectorFini())
	}
	if c.MaxContentLength() != 500000 {
		t.Errorf("Unexpected MaxContentLength: %v", c.MaxContentLength())
	}
	if c.ModuleIdentifier() != DefaultModuleIdentifier {
		t.Errorf("Unexpected ModuleIdentifier: %v", c.ModuleIdentifier())
	}
	if c.RPCAddress() != "0.0.0.0:1234" {
		t.Errorf("Unexpected RPCAddress: %v", c.RPCAddress())
	}
	if c.RPCNetwork() != "tcp" {
		t.Errorf("Unexpected RPCNetwork: %v", c.RPCNetwork())
	}
	if c.RPCAddressString() != "tcp:0.0.0.0:1234" {
		t.Errorf("Unexpected RPCAddressString: %v", c.RPCAddressString())
	}
	if c.ServerIdentifier() != DefaultServerIdentifier {
		t.Errorf("Unexpected ServerIdentifier: %v", c.ServerIdentifier())
	}
	if c.Timeout() != 10*time.Millisecond {
		t.Errorf("Unexpected Timeout: %v", c.Timeout())
	}
	if c.IsAllowCode(406) {
		t.Errorf("Unexpected IsAllowCode(406): %v", c.IsAllowCode(406))
	}
	if !c.IsAllowCode(200) {
		t.Errorf("Unexpected IsAllowCode(200): %v", c.IsAllowCode(200))
	}
	if !c.IsBlockCode(406) {
		t.Errorf("Unexpected IsBlockCode(406): %v", c.IsBlockCode(406))
	}
	if !c.IsBlockCode(403) {
		t.Errorf("Unexpected IsBlockCode(403): %v", c.IsBlockCode(403))
	}
	if c.IsBlockCode(200) {
		t.Errorf("Unexpected IsBlockCode(200): %v", c.IsBlockCode(200))
	}
}
