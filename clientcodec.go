package sigsci

import (
	"errors"
	"fmt"
	"io"
	"net/rpc"

	"github.com/tinylib/msgp/msgp"
)

// defines the MSGPACK RPC format
type msgpClientCodec struct {
	dec *msgp.Reader
	enc *msgp.Writer
	c   io.Closer
}

func newMsgpClientCodec(conn io.ReadWriteCloser) rpc.ClientCodec {
	return &msgpClientCodec{
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
		return fmt.Errorf("WriteRequest flush failed: %s", err)
	}

	return nil
}

func (c msgpClientCodec) ReadResponseHeader(r *rpc.Response) error {
	sz, err := c.dec.ReadArrayHeader()
	if err != nil || sz != 4 {
		return errors.New("Failed ReadResponseHeader on initial array")
	}
	msgtype, err := c.dec.ReadUint()
	if err != nil || msgtype != 1 {
		return errors.New("Failed ReadResponseHeader in mesage type")
	}

	// Sequence ID
	_, err = c.dec.ReadUint()
	if err != nil {
		return errors.New("Failed ReadResponseHeader in error type")
	}
	err = c.dec.ReadNil()
	if err != nil {
		// if there is an error maybe its not nil
		//  try to read string.  if still an error
		//  then assume its bad response
		rawerr, err := c.dec.ReadString()
		if err != nil {
			return errors.New("Failed ReadResponseHeader: Unable to read arg3: " + err.Error())
		}
		return errors.New("Remote Error: " + string(rawerr))
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
			return fmt.Errorf("failed ReadResponseBody in obj decode: %s", err)
		}
		return nil
	}

	// we use a plain "int" for response codes just hardwired this
	// case. in future use an object to simplify this.
	//
	if xint, ok := x.(*int); ok {
		val, err := c.dec.ReadInt()
		if err != nil {
			return fmt.Errorf("failed ReadResponseBody in int decode: %s", err)
		}
		*xint = val
		return nil
	}

	return fmt.Errorf("Unable to decode ReadResponseBody")
}
