package sigsci

import (
	"net/http"
	"testing"
	"time"
)

func TestDefaultModuleConfig(t *testing.T) {
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
	for code := 300; code < 600; code++ {
		if c.IsAllowCode(code) {
			t.Errorf("Unexpected IsAllowCode(%d): %v", code, c.IsAllowCode(code))
		}
	}
	if !c.IsAllowCode(200) {
		t.Errorf("Unexpected IsAllowCode(200): %v", c.IsAllowCode(200))
	}
	for code := 300; code < 600; code++ {
		if !c.IsBlockCode(code) {
			t.Errorf("Unexpected IsBlockCode(%d): %v", code, c.IsBlockCode(code))
		}
	}
	if c.IsBlockCode(600) {
		t.Errorf("Unexpected IsBlockCode(600): %v", c.IsBlockCode(600))
	}
	if c.IsBlockCode(200) {
		t.Errorf("Unexpected IsBlockCode(200): %v", c.IsBlockCode(200))
	}
}

func TestConfiguredModuleConfig(t *testing.T) {
	c, err := NewModuleConfig(
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
	if c.AltResponseCodes() != nil {
		t.Errorf("Unexpected AltResponseCodes from deprecated option (should be nil): %v", c.AltResponseCodes())
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
	for code := 300; code < 600; code++ {
		if c.IsAllowCode(code) {
			t.Errorf("Unexpected IsAllowCode(%d): %v", code, c.IsAllowCode(code))
		}
	}
	if !c.IsAllowCode(200) {
		t.Errorf("Unexpected IsAllowCode(200): %v", c.IsAllowCode(200))
	}
	for code := 300; code < 600; code++ {
		if !c.IsBlockCode(code) {
			t.Errorf("Unexpected IsBlockCode(%d): %v", code, c.IsBlockCode(code))
		}
	}
	if c.IsBlockCode(600) {
		t.Errorf("Unexpected IsBlockCode(600): %v", c.IsBlockCode(600))
	}
	if c.IsBlockCode(200) {
		t.Errorf("Unexpected IsBlockCode(200): %v", c.IsBlockCode(200))
	}
}

func TestFromModuleConfig(t *testing.T) {
	c0, err := NewModuleConfig(
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

	c, err := NewModuleConfig(
		FromModuleConfig(c0),
	)
	if err != nil {
		t.Fatalf("Failed to create module config: %s", err)
	}
	if c.AllowUnknownContentLength() != true {
		t.Errorf("Unexpected AllowUnknownContentLength: %v", c.AllowUnknownContentLength())
	}
	if c.AltResponseCodes() != nil {
		t.Errorf("Unexpected AltResponseCodes from deprecated option (should be nil): %v", c.AltResponseCodes())
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
	for code := 300; code < 600; code++ {
		if c.IsAllowCode(code) {
			t.Errorf("Unexpected IsAllowCode(%d): %v", code, c.IsAllowCode(code))
		}
	}
	if !c.IsAllowCode(200) {
		t.Errorf("Unexpected IsAllowCode(200): %v", c.IsAllowCode(200))
	}
	for code := 300; code < 600; code++ {
		if !c.IsBlockCode(code) {
			t.Errorf("Unexpected IsBlockCode(%d): %v", code, c.IsBlockCode(code))
		}
	}
	if c.IsBlockCode(600) {
		t.Errorf("Unexpected IsBlockCode(600): %v", c.IsBlockCode(600))
	}
	if c.IsBlockCode(200) {
		t.Errorf("Unexpected IsBlockCode(200): %v", c.IsBlockCode(200))
	}
}
