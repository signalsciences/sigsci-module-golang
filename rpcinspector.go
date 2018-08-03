package sigsci

import (
	"fmt"
	"net"
	"net/rpc"
	"time"
)

// RPCInspector is an Inspector implemented as RPC calls to the agent
type RPCInspector struct {
	Network string
	Address string
	Timeout time.Duration
	Debug   bool
}

// ModuleInit sends a RPC.ModuleInit message to the agent
func (ri *RPCInspector) ModuleInit(in *RPCMsgIn, out *RPCMsgOut) error {
	conn, err := ri.getConnection()
	if err == nil {
		rpcCodec := newMsgpClientCodec(conn)
		client := rpc.NewClientWithCodec(rpcCodec)
		err = client.Call("RPC.ModuleInit", in, out)
		client.Close() // TBD: reuse conn
	}

	// TBD: wrap error instead of prefixing
	if err != nil {
		return fmt.Errorf("RPC.ModuleInit call failed: %s", err)
	}

	return nil
}

// PreRequest sends a RPC.PreRequest message to the agent
func (ri *RPCInspector) PreRequest(in *RPCMsgIn, out *RPCMsgOut) error {
	conn, err := ri.getConnection()
	if err == nil {
		rpcCodec := newMsgpClientCodec(conn)
		client := rpc.NewClientWithCodec(rpcCodec)
		err = client.Call("RPC.PreRequest", in, out)
		client.Close() // TBD: reuse conn
	}

	// TBD: wrap error instead of prefixing
	if err != nil {
		return fmt.Errorf("RPC.PreRequest call failed: %s", err)
	}

	return nil
}

// PostRequest sends a RPC.PostRequest message to the agent
func (ri *RPCInspector) PostRequest(in *RPCMsgIn, out *RPCMsgOut) error {
	conn, err := ri.getConnection()
	if err == nil {
		rpcCodec := newMsgpClientCodec(conn)
		client := rpc.NewClientWithCodec(rpcCodec)

		var rpcout int
		err = client.Call("RPC.PostRequest", in, &rpcout)
		client.Close()

		// Fake the output until RPC call is updated
		out.WAFResponse = 200
		out.RequestID = ""
		out.RequestHeaders = nil
	}

	// TBD: wrap error instead of prefixing
	if err != nil {
		return fmt.Errorf("RPC.PostRequest call failed: %s", err)
	}

	return nil
}

// UpdateRequest sends a RPC.UpdateRequest message to the agent
func (ri *RPCInspector) UpdateRequest(in *RPCMsgIn2, out *RPCMsgOut) error {
	conn, err := ri.getConnection()
	if err == nil {
		rpcCodec := newMsgpClientCodec(conn)
		client := rpc.NewClientWithCodec(rpcCodec)

		var rpcout int
		err = client.Call("RPC.UpdateRequest", in, &rpcout)
		client.Close()

		// Fake the output until RPC call is updated
		out.WAFResponse = 200
		out.RequestID = ""
		out.RequestHeaders = nil
	}

	return err
}

func (ri *RPCInspector) makeConnection() (net.Conn, error) {
	conn, err := net.DialTimeout(ri.Network, ri.Address, ri.Timeout)
	if err != nil {
		return nil, err
	}
	conn.SetDeadline(time.Now().Add(ri.Timeout))
	return conn, nil
}

func (ri *RPCInspector) getConnection() (net.Conn, error) {
	// here for future expansion to use pools, etc.
	return ri.makeConnection()
}
