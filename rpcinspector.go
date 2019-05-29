package sigsci

import (
	"fmt"
	"net"
	"net/rpc"
	"time"
)

// RPCInspector is an Inspector implemented as RPC calls to the agent
type RPCInspector struct {
	Network           string
	Address           string
	Timeout           time.Duration
	Debug             bool
	InitRPCClientFunc func() (*rpc.Client, error)
	FiniRPCClientFunc func(*rpc.Client, error)
}

// ModuleInit sends a RPC.ModuleInit message to the agent
func (ri *RPCInspector) ModuleInit(in *RPCMsgIn, out *RPCMsgOut) error {
	client, err := ri.GetRPCClient()
	if err == nil {
		err = client.Call("RPC.ModuleInit", in, out)
		ri.CloseRPCClient(client, err)
	}

	if err != nil {
		return fmt.Errorf("RPC.ModuleInit call failed: %s", err)
	}

	return nil
}

// PreRequest sends a RPC.PreRequest message to the agent
func (ri *RPCInspector) PreRequest(in *RPCMsgIn, out *RPCMsgOut) error {
	client, err := ri.GetRPCClient()
	if err == nil {
		err = client.Call("RPC.PreRequest", in, out)
		ri.CloseRPCClient(client, err)
	}

	if err != nil {
		return fmt.Errorf("RPC.PreRequest call failed: %s", err)
	}

	return nil
}

// PostRequest sends a RPC.PostRequest message to the agent
func (ri *RPCInspector) PostRequest(in *RPCMsgIn, out *RPCMsgOut) error {
	client, err := ri.GetRPCClient()
	if err == nil {
		var rpcout int
		err = client.Call("RPC.PostRequest", in, &rpcout)
		ri.CloseRPCClient(client, err)

		// Always success as the rpcout is not currently used
		out.WAFResponse = 200
		out.RequestID = ""
		out.RequestHeaders = nil
	}

	if err != nil {
		return fmt.Errorf("RPC.PostRequest call failed: %s", err)
	}

	return nil
}

// UpdateRequest sends a RPC.UpdateRequest message to the agent
func (ri *RPCInspector) UpdateRequest(in *RPCMsgIn2, out *RPCMsgOut) error {
	client, err := ri.GetRPCClient()
	if err == nil {

		var rpcout int
		err = client.Call("RPC.UpdateRequest", in, &rpcout)
		ri.CloseRPCClient(client, err)

		// Always success as the rpcout is not currently used
		out.WAFResponse = 200
		out.RequestID = ""
		out.RequestHeaders = nil
	}

	return err
}

// GetRPCClient gets a RPC client
func (ri *RPCInspector) GetRPCClient() (*rpc.Client, error) {
	if ri.InitRPCClientFunc != nil {
		return ri.InitRPCClientFunc()
	}

	conn, err := ri.getConnection()
	if err != nil {
		return nil, err
	}
	rpcCodec := NewMsgpClientCodec(conn)
	return rpc.NewClientWithCodec(rpcCodec), nil
}

// CloseRPCClient closes a RPC client
func (ri *RPCInspector) CloseRPCClient(client *rpc.Client, err error) {
	if ri.FiniRPCClientFunc != nil {
		ri.FiniRPCClientFunc(client, err)
		return
	}
	client.Close()
}

func (ri *RPCInspector) makeConnection() (net.Conn, error) {
	deadline := time.Now().Add(ri.Timeout)
	conn, err := net.DialTimeout(ri.Network, ri.Address, ri.Timeout)
	if err != nil {
		return nil, err
	}
	conn.SetDeadline(deadline)
	return conn, nil
}

func (ri *RPCInspector) getConnection() (net.Conn, error) {
	// here for future expansion to use pools, etc.
	return ri.makeConnection()
}
