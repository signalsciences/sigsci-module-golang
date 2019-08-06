package sigsci

import (
	"fmt"
	"io"
	"net"
	"net/rpc"

	"github.com/tinylib/msgp/msgp"
)

// Adaptors for golang's RPC mechanism using MSGPACK
//
// * http://msgpack.org
// * https://golang.org/pkg/net/rpc/
//

// defines the MSGPACK RPC format
type msgpClientCodec struct {
	dec *msgp.Reader
	enc *msgp.Writer
	c   io.Closer
}

// NewMsgpClientCodec creates a new rpc.ClientCodec from an existing connection
func NewMsgpClientCodec(conn io.ReadWriteCloser) rpc.ClientCodec {
	return msgpClientCodec{
		dec: msgp.NewReader(conn),
		enc: msgp.NewWriter(conn),
		c:   conn,
	}
}

func (c msgpClientCodec) Close() error {
	return c.c.Close()
}

func (c msgpClientCodec) WriteRequest(r *rpc.Request, x interface{}) error {
	if err := c.enc.WriteArrayHeader(4); err != nil {
		return fmt.Errorf("WriteRequest failed in writing array header: %s", err)
	}

	if err := c.enc.WriteInt(0); err != nil {
		return fmt.Errorf("WriteRequest failed in requiting rpc msg type 0: %s", err)
	}

	if err := c.enc.WriteUint32(uint32(r.Seq)); err != nil {
		return fmt.Errorf("WriteRequest failed in writing sequence id: %s", err)
	}

	if err := c.enc.WriteString(r.ServiceMethod); err != nil {
		return fmt.Errorf("WriteRequest failed in writing service method: %s", err)
	}

	if err := c.enc.WriteArrayHeader(1); err != nil {
		return fmt.Errorf("WriteRequest failed in writing arg array header: %s", err)
	}

	if err := c.enc.WriteIntf(x); err != nil {
		return fmt.Errorf("WriteRequest failed in writing %T: %s", x, err)
	}

	if err := c.enc.Flush(); err != nil {
		return fmt.Errorf("WriteRequest failed in flushing: %s", err)
	}

	return nil
}

func (c msgpClientCodec) ReadResponseHeader(r *rpc.Response) error {
	sz, err := c.dec.ReadArrayHeader()
	if err != nil || sz != 4 {
		if cerr := knownError(err); cerr != nil {
			return cerr
		}
		if err == nil && sz != 4 {
			err = fmt.Errorf("invalid array size %d", sz)
		}
		return fmt.Errorf("ReadResponseHeader failed in initial array: %s", err)
	}

	msgtype, err := c.dec.ReadUint()
	if err != nil || msgtype != 1 {
		if cerr := knownError(err); cerr != nil {
			return cerr
		}
		if err == nil && msgtype != 1 {
			err = fmt.Errorf("invalid message type %d", msgtype)
		}
		return fmt.Errorf("ReadResponseHeader failed in message type: %s", err)
	}

	seqID, err := c.dec.ReadUint()
	if err != nil {
		if cerr := knownError(err); cerr != nil {
			return cerr
		}
		return fmt.Errorf("ReadResponseHeader failed in error type: %s", err)
	}
	r.Seq = uint64(seqID)

	err = c.dec.ReadNil()
	if err != nil {
		if cerr := knownError(err); cerr != nil {
			return cerr
		}
		// if there is an error maybe its not nil
		//  try to read string.  if still an error
		//  then assume its bad response
		rawerr, err := c.dec.ReadString()
		if err != nil {
			if cerr := knownError(err); cerr != nil {
				return cerr
			}
			return fmt.Errorf("ReadResponseHeader failed in error message: %s", err)
		}
		return fmt.Errorf("remote error: %s", rawerr)
	}
	return nil
}

func (c msgpClientCodec) ReadResponseBody(x interface{}) error {
	if x == nil {
		return nil
	}

	// if its a decode-able object, then sort it out.
	if obj, ok := x.(msgp.Decodable); ok {
		if err := obj.DecodeMsg(c.dec); err != nil {
			if cerr := knownError(err); cerr != nil {
				return cerr
			}
			return fmt.Errorf("ReadResponseBody failed in obj decode: %s", err)
		}
		return nil
	}

	// we use a plain "int" for response codes just hardwired this
	// case. in future use an object to simplify this.
	//
	if xint, ok := x.(*int); ok {
		val, err := c.dec.ReadInt()
		if err != nil {
			if cerr := knownError(err); cerr != nil {
				return cerr
			}
			return fmt.Errorf("ReadResponseBody failed in int decode: %s", err)
		}
		*xint = val
		return nil
	}

	return fmt.Errorf("ReadResponseBody failed: unable to decode")
}

// knownError checks the error against known errors, does any fixups and returns an error to indicate it is a known error that should not be handled further as well as the replaced error. A nil error is returned if the original error should be handled further.
func knownError(err error) error {
	if err == nil {
		return nil
	}
	if err == io.EOF {
		return err
	}
	if nerr, ok := err.(net.Error); ok {
		if nerr.Timeout() {
			return nerr
		}
	}
	return nil
}
