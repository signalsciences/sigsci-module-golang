package sigsci

// NOTE: THIS FILE WAS PRODUCED BY THE
// MSGP CODE GENERATION TOOL (github.com/tinylib/msgp)
// DO NOT EDIT

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *RPCMsgIn) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "AccessKeyID":
			z.AccessKeyID, err = dc.ReadString()
			if err != nil {
				return
			}
		case "ModuleVersion":
			z.ModuleVersion, err = dc.ReadString()
			if err != nil {
				return
			}
		case "ServerVersion":
			z.ServerVersion, err = dc.ReadString()
			if err != nil {
				return
			}
		case "ServerFlavor":
			z.ServerFlavor, err = dc.ReadString()
			if err != nil {
				return
			}
		case "ServerName":
			z.ServerName, err = dc.ReadString()
			if err != nil {
				return
			}
		case "Timestamp":
			z.Timestamp, err = dc.ReadInt64()
			if err != nil {
				return
			}
		case "NowMillis":
			z.NowMillis, err = dc.ReadInt64()
			if err != nil {
				return
			}
		case "RemoteAddr":
			z.RemoteAddr, err = dc.ReadString()
			if err != nil {
				return
			}
		case "Method":
			z.Method, err = dc.ReadString()
			if err != nil {
				return
			}
		case "Scheme":
			z.Scheme, err = dc.ReadString()
			if err != nil {
				return
			}
		case "URI":
			z.URI, err = dc.ReadString()
			if err != nil {
				return
			}
		case "Protocol":
			z.Protocol, err = dc.ReadString()
			if err != nil {
				return
			}
		case "TLSProtocol":
			z.TLSProtocol, err = dc.ReadString()
			if err != nil {
				return
			}
		case "TLSCipher":
			z.TLSCipher, err = dc.ReadString()
			if err != nil {
				return
			}
		case "WAFResponse":
			z.WAFResponse, err = dc.ReadInt32()
			if err != nil {
				return
			}
		case "ResponseCode":
			z.ResponseCode, err = dc.ReadInt32()
			if err != nil {
				return
			}
		case "ResponseMillis":
			z.ResponseMillis, err = dc.ReadInt64()
			if err != nil {
				return
			}
		case "ResponseSize":
			z.ResponseSize, err = dc.ReadInt64()
			if err != nil {
				return
			}
		case "HeadersIn":
			var zb0002 uint32
			zb0002, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.HeadersIn) >= int(zb0002) {
				z.HeadersIn = (z.HeadersIn)[:zb0002]
			} else {
				z.HeadersIn = make([][2]string, zb0002)
			}
			for za0001 := range z.HeadersIn {
				var zb0003 uint32
				zb0003, err = dc.ReadArrayHeader()
				if err != nil {
					return
				}
				if zb0003 != uint32(2) {
					err = msgp.ArrayError{Wanted: uint32(2), Got: zb0003}
					return
				}
				for za0002 := range z.HeadersIn[za0001] {
					z.HeadersIn[za0001][za0002], err = dc.ReadString()
					if err != nil {
						return
					}
				}
			}
		case "HeadersOut":
			var zb0004 uint32
			zb0004, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.HeadersOut) >= int(zb0004) {
				z.HeadersOut = (z.HeadersOut)[:zb0004]
			} else {
				z.HeadersOut = make([][2]string, zb0004)
			}
			for za0003 := range z.HeadersOut {
				var zb0005 uint32
				zb0005, err = dc.ReadArrayHeader()
				if err != nil {
					return
				}
				if zb0005 != uint32(2) {
					err = msgp.ArrayError{Wanted: uint32(2), Got: zb0005}
					return
				}
				for za0004 := range z.HeadersOut[za0003] {
					z.HeadersOut[za0003][za0004], err = dc.ReadString()
					if err != nil {
						return
					}
				}
			}
		case "PostBody":
			z.PostBody, err = dc.ReadString()
			if err != nil {
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *RPCMsgIn) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 21
	// write "AccessKeyID"
	err = en.Append(0xde, 0x0, 0x15, 0xab, 0x41, 0x63, 0x63, 0x65, 0x73, 0x73, 0x4b, 0x65, 0x79, 0x49, 0x44)
	if err != nil {
		return err
	}
	err = en.WriteString(z.AccessKeyID)
	if err != nil {
		return
	}
	// write "ModuleVersion"
	err = en.Append(0xad, 0x4d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e)
	if err != nil {
		return err
	}
	err = en.WriteString(z.ModuleVersion)
	if err != nil {
		return
	}
	// write "ServerVersion"
	err = en.Append(0xad, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e)
	if err != nil {
		return err
	}
	err = en.WriteString(z.ServerVersion)
	if err != nil {
		return
	}
	// write "ServerFlavor"
	err = en.Append(0xac, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x46, 0x6c, 0x61, 0x76, 0x6f, 0x72)
	if err != nil {
		return err
	}
	err = en.WriteString(z.ServerFlavor)
	if err != nil {
		return
	}
	// write "ServerName"
	err = en.Append(0xaa, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x4e, 0x61, 0x6d, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteString(z.ServerName)
	if err != nil {
		return
	}
	// write "Timestamp"
	err = en.Append(0xa9, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70)
	if err != nil {
		return err
	}
	err = en.WriteInt64(z.Timestamp)
	if err != nil {
		return
	}
	// write "NowMillis"
	err = en.Append(0xa9, 0x4e, 0x6f, 0x77, 0x4d, 0x69, 0x6c, 0x6c, 0x69, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteInt64(z.NowMillis)
	if err != nil {
		return
	}
	// write "RemoteAddr"
	err = en.Append(0xaa, 0x52, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x41, 0x64, 0x64, 0x72)
	if err != nil {
		return err
	}
	err = en.WriteString(z.RemoteAddr)
	if err != nil {
		return
	}
	// write "Method"
	err = en.Append(0xa6, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Method)
	if err != nil {
		return
	}
	// write "Scheme"
	err = en.Append(0xa6, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Scheme)
	if err != nil {
		return
	}
	// write "URI"
	err = en.Append(0xa3, 0x55, 0x52, 0x49)
	if err != nil {
		return err
	}
	err = en.WriteString(z.URI)
	if err != nil {
		return
	}
	// write "Protocol"
	err = en.Append(0xa8, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c)
	if err != nil {
		return err
	}
	err = en.WriteString(z.Protocol)
	if err != nil {
		return
	}
	// write "TLSProtocol"
	err = en.Append(0xab, 0x54, 0x4c, 0x53, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c)
	if err != nil {
		return err
	}
	err = en.WriteString(z.TLSProtocol)
	if err != nil {
		return
	}
	// write "TLSCipher"
	err = en.Append(0xa9, 0x54, 0x4c, 0x53, 0x43, 0x69, 0x70, 0x68, 0x65, 0x72)
	if err != nil {
		return err
	}
	err = en.WriteString(z.TLSCipher)
	if err != nil {
		return
	}
	// write "WAFResponse"
	err = en.Append(0xab, 0x57, 0x41, 0x46, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteInt32(z.WAFResponse)
	if err != nil {
		return
	}
	// write "ResponseCode"
	err = en.Append(0xac, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x43, 0x6f, 0x64, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteInt32(z.ResponseCode)
	if err != nil {
		return
	}
	// write "ResponseMillis"
	err = en.Append(0xae, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x4d, 0x69, 0x6c, 0x6c, 0x69, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteInt64(z.ResponseMillis)
	if err != nil {
		return
	}
	// write "ResponseSize"
	err = en.Append(0xac, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x53, 0x69, 0x7a, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteInt64(z.ResponseSize)
	if err != nil {
		return
	}
	// write "HeadersIn"
	err = en.Append(0xa9, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x49, 0x6e)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.HeadersIn)))
	if err != nil {
		return
	}
	for za0001 := range z.HeadersIn {
		err = en.WriteArrayHeader(uint32(2))
		if err != nil {
			return
		}
		for za0002 := range z.HeadersIn[za0001] {
			err = en.WriteString(z.HeadersIn[za0001][za0002])
			if err != nil {
				return
			}
		}
	}
	// write "HeadersOut"
	err = en.Append(0xaa, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x4f, 0x75, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.HeadersOut)))
	if err != nil {
		return
	}
	for za0003 := range z.HeadersOut {
		err = en.WriteArrayHeader(uint32(2))
		if err != nil {
			return
		}
		for za0004 := range z.HeadersOut[za0003] {
			err = en.WriteString(z.HeadersOut[za0003][za0004])
			if err != nil {
				return
			}
		}
	}
	// write "PostBody"
	err = en.Append(0xa8, 0x50, 0x6f, 0x73, 0x74, 0x42, 0x6f, 0x64, 0x79)
	if err != nil {
		return err
	}
	err = en.WriteString(z.PostBody)
	if err != nil {
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *RPCMsgIn) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 21
	// string "AccessKeyID"
	o = append(o, 0xde, 0x0, 0x15, 0xab, 0x41, 0x63, 0x63, 0x65, 0x73, 0x73, 0x4b, 0x65, 0x79, 0x49, 0x44)
	o = msgp.AppendString(o, z.AccessKeyID)
	// string "ModuleVersion"
	o = append(o, 0xad, 0x4d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e)
	o = msgp.AppendString(o, z.ModuleVersion)
	// string "ServerVersion"
	o = append(o, 0xad, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e)
	o = msgp.AppendString(o, z.ServerVersion)
	// string "ServerFlavor"
	o = append(o, 0xac, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x46, 0x6c, 0x61, 0x76, 0x6f, 0x72)
	o = msgp.AppendString(o, z.ServerFlavor)
	// string "ServerName"
	o = append(o, 0xaa, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x4e, 0x61, 0x6d, 0x65)
	o = msgp.AppendString(o, z.ServerName)
	// string "Timestamp"
	o = append(o, 0xa9, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70)
	o = msgp.AppendInt64(o, z.Timestamp)
	// string "NowMillis"
	o = append(o, 0xa9, 0x4e, 0x6f, 0x77, 0x4d, 0x69, 0x6c, 0x6c, 0x69, 0x73)
	o = msgp.AppendInt64(o, z.NowMillis)
	// string "RemoteAddr"
	o = append(o, 0xaa, 0x52, 0x65, 0x6d, 0x6f, 0x74, 0x65, 0x41, 0x64, 0x64, 0x72)
	o = msgp.AppendString(o, z.RemoteAddr)
	// string "Method"
	o = append(o, 0xa6, 0x4d, 0x65, 0x74, 0x68, 0x6f, 0x64)
	o = msgp.AppendString(o, z.Method)
	// string "Scheme"
	o = append(o, 0xa6, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x65)
	o = msgp.AppendString(o, z.Scheme)
	// string "URI"
	o = append(o, 0xa3, 0x55, 0x52, 0x49)
	o = msgp.AppendString(o, z.URI)
	// string "Protocol"
	o = append(o, 0xa8, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c)
	o = msgp.AppendString(o, z.Protocol)
	// string "TLSProtocol"
	o = append(o, 0xab, 0x54, 0x4c, 0x53, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c)
	o = msgp.AppendString(o, z.TLSProtocol)
	// string "TLSCipher"
	o = append(o, 0xa9, 0x54, 0x4c, 0x53, 0x43, 0x69, 0x70, 0x68, 0x65, 0x72)
	o = msgp.AppendString(o, z.TLSCipher)
	// string "WAFResponse"
	o = append(o, 0xab, 0x57, 0x41, 0x46, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65)
	o = msgp.AppendInt32(o, z.WAFResponse)
	// string "ResponseCode"
	o = append(o, 0xac, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x43, 0x6f, 0x64, 0x65)
	o = msgp.AppendInt32(o, z.ResponseCode)
	// string "ResponseMillis"
	o = append(o, 0xae, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x4d, 0x69, 0x6c, 0x6c, 0x69, 0x73)
	o = msgp.AppendInt64(o, z.ResponseMillis)
	// string "ResponseSize"
	o = append(o, 0xac, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x53, 0x69, 0x7a, 0x65)
	o = msgp.AppendInt64(o, z.ResponseSize)
	// string "HeadersIn"
	o = append(o, 0xa9, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x49, 0x6e)
	o = msgp.AppendArrayHeader(o, uint32(len(z.HeadersIn)))
	for za0001 := range z.HeadersIn {
		o = msgp.AppendArrayHeader(o, uint32(2))
		for za0002 := range z.HeadersIn[za0001] {
			o = msgp.AppendString(o, z.HeadersIn[za0001][za0002])
		}
	}
	// string "HeadersOut"
	o = append(o, 0xaa, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x4f, 0x75, 0x74)
	o = msgp.AppendArrayHeader(o, uint32(len(z.HeadersOut)))
	for za0003 := range z.HeadersOut {
		o = msgp.AppendArrayHeader(o, uint32(2))
		for za0004 := range z.HeadersOut[za0003] {
			o = msgp.AppendString(o, z.HeadersOut[za0003][za0004])
		}
	}
	// string "PostBody"
	o = append(o, 0xa8, 0x50, 0x6f, 0x73, 0x74, 0x42, 0x6f, 0x64, 0x79)
	o = msgp.AppendString(o, z.PostBody)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *RPCMsgIn) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "AccessKeyID":
			z.AccessKeyID, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "ModuleVersion":
			z.ModuleVersion, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "ServerVersion":
			z.ServerVersion, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "ServerFlavor":
			z.ServerFlavor, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "ServerName":
			z.ServerName, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "Timestamp":
			z.Timestamp, bts, err = msgp.ReadInt64Bytes(bts)
			if err != nil {
				return
			}
		case "NowMillis":
			z.NowMillis, bts, err = msgp.ReadInt64Bytes(bts)
			if err != nil {
				return
			}
		case "RemoteAddr":
			z.RemoteAddr, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "Method":
			z.Method, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "Scheme":
			z.Scheme, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "URI":
			z.URI, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "Protocol":
			z.Protocol, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "TLSProtocol":
			z.TLSProtocol, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "TLSCipher":
			z.TLSCipher, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "WAFResponse":
			z.WAFResponse, bts, err = msgp.ReadInt32Bytes(bts)
			if err != nil {
				return
			}
		case "ResponseCode":
			z.ResponseCode, bts, err = msgp.ReadInt32Bytes(bts)
			if err != nil {
				return
			}
		case "ResponseMillis":
			z.ResponseMillis, bts, err = msgp.ReadInt64Bytes(bts)
			if err != nil {
				return
			}
		case "ResponseSize":
			z.ResponseSize, bts, err = msgp.ReadInt64Bytes(bts)
			if err != nil {
				return
			}
		case "HeadersIn":
			var zb0002 uint32
			zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.HeadersIn) >= int(zb0002) {
				z.HeadersIn = (z.HeadersIn)[:zb0002]
			} else {
				z.HeadersIn = make([][2]string, zb0002)
			}
			for za0001 := range z.HeadersIn {
				var zb0003 uint32
				zb0003, bts, err = msgp.ReadArrayHeaderBytes(bts)
				if err != nil {
					return
				}
				if zb0003 != uint32(2) {
					err = msgp.ArrayError{Wanted: uint32(2), Got: zb0003}
					return
				}
				for za0002 := range z.HeadersIn[za0001] {
					z.HeadersIn[za0001][za0002], bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
				}
			}
		case "HeadersOut":
			var zb0004 uint32
			zb0004, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.HeadersOut) >= int(zb0004) {
				z.HeadersOut = (z.HeadersOut)[:zb0004]
			} else {
				z.HeadersOut = make([][2]string, zb0004)
			}
			for za0003 := range z.HeadersOut {
				var zb0005 uint32
				zb0005, bts, err = msgp.ReadArrayHeaderBytes(bts)
				if err != nil {
					return
				}
				if zb0005 != uint32(2) {
					err = msgp.ArrayError{Wanted: uint32(2), Got: zb0005}
					return
				}
				for za0004 := range z.HeadersOut[za0003] {
					z.HeadersOut[za0003][za0004], bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
				}
			}
		case "PostBody":
			z.PostBody, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *RPCMsgIn) Msgsize() (s int) {
	s = 3 + 12 + msgp.StringPrefixSize + len(z.AccessKeyID) + 14 + msgp.StringPrefixSize + len(z.ModuleVersion) + 14 + msgp.StringPrefixSize + len(z.ServerVersion) + 13 + msgp.StringPrefixSize + len(z.ServerFlavor) + 11 + msgp.StringPrefixSize + len(z.ServerName) + 10 + msgp.Int64Size + 10 + msgp.Int64Size + 11 + msgp.StringPrefixSize + len(z.RemoteAddr) + 7 + msgp.StringPrefixSize + len(z.Method) + 7 + msgp.StringPrefixSize + len(z.Scheme) + 4 + msgp.StringPrefixSize + len(z.URI) + 9 + msgp.StringPrefixSize + len(z.Protocol) + 12 + msgp.StringPrefixSize + len(z.TLSProtocol) + 10 + msgp.StringPrefixSize + len(z.TLSCipher) + 12 + msgp.Int32Size + 13 + msgp.Int32Size + 15 + msgp.Int64Size + 13 + msgp.Int64Size + 10 + msgp.ArrayHeaderSize
	for za0001 := range z.HeadersIn {
		s += msgp.ArrayHeaderSize
		for za0002 := range z.HeadersIn[za0001] {
			s += msgp.StringPrefixSize + len(z.HeadersIn[za0001][za0002])
		}
	}
	s += 11 + msgp.ArrayHeaderSize
	for za0003 := range z.HeadersOut {
		s += msgp.ArrayHeaderSize
		for za0004 := range z.HeadersOut[za0003] {
			s += msgp.StringPrefixSize + len(z.HeadersOut[za0003][za0004])
		}
	}
	s += 9 + msgp.StringPrefixSize + len(z.PostBody)
	return
}

// DecodeMsg implements msgp.Decodable
func (z *RPCMsgIn2) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "RequestID":
			z.RequestID, err = dc.ReadString()
			if err != nil {
				return
			}
		case "ResponseCode":
			z.ResponseCode, err = dc.ReadInt32()
			if err != nil {
				return
			}
		case "ResponseMillis":
			z.ResponseMillis, err = dc.ReadInt64()
			if err != nil {
				return
			}
		case "ResponseSize":
			z.ResponseSize, err = dc.ReadInt64()
			if err != nil {
				return
			}
		case "HeadersOut":
			var zb0002 uint32
			zb0002, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.HeadersOut) >= int(zb0002) {
				z.HeadersOut = (z.HeadersOut)[:zb0002]
			} else {
				z.HeadersOut = make([][2]string, zb0002)
			}
			for za0001 := range z.HeadersOut {
				var zb0003 uint32
				zb0003, err = dc.ReadArrayHeader()
				if err != nil {
					return
				}
				if zb0003 != uint32(2) {
					err = msgp.ArrayError{Wanted: uint32(2), Got: zb0003}
					return
				}
				for za0002 := range z.HeadersOut[za0001] {
					z.HeadersOut[za0001][za0002], err = dc.ReadString()
					if err != nil {
						return
					}
				}
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *RPCMsgIn2) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 5
	// write "RequestID"
	err = en.Append(0x85, 0xa9, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x49, 0x44)
	if err != nil {
		return err
	}
	err = en.WriteString(z.RequestID)
	if err != nil {
		return
	}
	// write "ResponseCode"
	err = en.Append(0xac, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x43, 0x6f, 0x64, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteInt32(z.ResponseCode)
	if err != nil {
		return
	}
	// write "ResponseMillis"
	err = en.Append(0xae, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x4d, 0x69, 0x6c, 0x6c, 0x69, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteInt64(z.ResponseMillis)
	if err != nil {
		return
	}
	// write "ResponseSize"
	err = en.Append(0xac, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x53, 0x69, 0x7a, 0x65)
	if err != nil {
		return err
	}
	err = en.WriteInt64(z.ResponseSize)
	if err != nil {
		return
	}
	// write "HeadersOut"
	err = en.Append(0xaa, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x4f, 0x75, 0x74)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.HeadersOut)))
	if err != nil {
		return
	}
	for za0001 := range z.HeadersOut {
		err = en.WriteArrayHeader(uint32(2))
		if err != nil {
			return
		}
		for za0002 := range z.HeadersOut[za0001] {
			err = en.WriteString(z.HeadersOut[za0001][za0002])
			if err != nil {
				return
			}
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *RPCMsgIn2) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 5
	// string "RequestID"
	o = append(o, 0x85, 0xa9, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x49, 0x44)
	o = msgp.AppendString(o, z.RequestID)
	// string "ResponseCode"
	o = append(o, 0xac, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x43, 0x6f, 0x64, 0x65)
	o = msgp.AppendInt32(o, z.ResponseCode)
	// string "ResponseMillis"
	o = append(o, 0xae, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x4d, 0x69, 0x6c, 0x6c, 0x69, 0x73)
	o = msgp.AppendInt64(o, z.ResponseMillis)
	// string "ResponseSize"
	o = append(o, 0xac, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x53, 0x69, 0x7a, 0x65)
	o = msgp.AppendInt64(o, z.ResponseSize)
	// string "HeadersOut"
	o = append(o, 0xaa, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73, 0x4f, 0x75, 0x74)
	o = msgp.AppendArrayHeader(o, uint32(len(z.HeadersOut)))
	for za0001 := range z.HeadersOut {
		o = msgp.AppendArrayHeader(o, uint32(2))
		for za0002 := range z.HeadersOut[za0001] {
			o = msgp.AppendString(o, z.HeadersOut[za0001][za0002])
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *RPCMsgIn2) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "RequestID":
			z.RequestID, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "ResponseCode":
			z.ResponseCode, bts, err = msgp.ReadInt32Bytes(bts)
			if err != nil {
				return
			}
		case "ResponseMillis":
			z.ResponseMillis, bts, err = msgp.ReadInt64Bytes(bts)
			if err != nil {
				return
			}
		case "ResponseSize":
			z.ResponseSize, bts, err = msgp.ReadInt64Bytes(bts)
			if err != nil {
				return
			}
		case "HeadersOut":
			var zb0002 uint32
			zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.HeadersOut) >= int(zb0002) {
				z.HeadersOut = (z.HeadersOut)[:zb0002]
			} else {
				z.HeadersOut = make([][2]string, zb0002)
			}
			for za0001 := range z.HeadersOut {
				var zb0003 uint32
				zb0003, bts, err = msgp.ReadArrayHeaderBytes(bts)
				if err != nil {
					return
				}
				if zb0003 != uint32(2) {
					err = msgp.ArrayError{Wanted: uint32(2), Got: zb0003}
					return
				}
				for za0002 := range z.HeadersOut[za0001] {
					z.HeadersOut[za0001][za0002], bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
				}
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *RPCMsgIn2) Msgsize() (s int) {
	s = 1 + 10 + msgp.StringPrefixSize + len(z.RequestID) + 13 + msgp.Int32Size + 15 + msgp.Int64Size + 13 + msgp.Int64Size + 11 + msgp.ArrayHeaderSize
	for za0001 := range z.HeadersOut {
		s += msgp.ArrayHeaderSize
		for za0002 := range z.HeadersOut[za0001] {
			s += msgp.StringPrefixSize + len(z.HeadersOut[za0001][za0002])
		}
	}
	return
}

// DecodeMsg implements msgp.Decodable
func (z *RPCMsgOut) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "WAFResponse":
			err = z.WAFResponse.DecodeMsg(dc)
			if err != nil {
				return
			}
		case "RequestID":
			z.RequestID, err = dc.ReadString()
			if err != nil {
				return
			}
		case "RequestHeaders":
			var zb0002 uint32
			zb0002, err = dc.ReadArrayHeader()
			if err != nil {
				return
			}
			if cap(z.RequestHeaders) >= int(zb0002) {
				z.RequestHeaders = (z.RequestHeaders)[:zb0002]
			} else {
				z.RequestHeaders = make([][2]string, zb0002)
			}
			for za0001 := range z.RequestHeaders {
				var zb0003 uint32
				zb0003, err = dc.ReadArrayHeader()
				if err != nil {
					return
				}
				if zb0003 != uint32(2) {
					err = msgp.ArrayError{Wanted: uint32(2), Got: zb0003}
					return
				}
				for za0002 := range z.RequestHeaders[za0001] {
					z.RequestHeaders[za0001][za0002], err = dc.ReadString()
					if err != nil {
						return
					}
				}
			}
		default:
			err = dc.Skip()
			if err != nil {
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z *RPCMsgOut) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 3
	// write "WAFResponse"
	err = en.Append(0x83, 0xab, 0x57, 0x41, 0x46, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65)
	if err != nil {
		return err
	}
	err = z.WAFResponse.EncodeMsg(en)
	if err != nil {
		return
	}
	// write "RequestID"
	err = en.Append(0xa9, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x49, 0x44)
	if err != nil {
		return err
	}
	err = en.WriteString(z.RequestID)
	if err != nil {
		return
	}
	// write "RequestHeaders"
	err = en.Append(0xae, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73)
	if err != nil {
		return err
	}
	err = en.WriteArrayHeader(uint32(len(z.RequestHeaders)))
	if err != nil {
		return
	}
	for za0001 := range z.RequestHeaders {
		err = en.WriteArrayHeader(uint32(2))
		if err != nil {
			return
		}
		for za0002 := range z.RequestHeaders[za0001] {
			err = en.WriteString(z.RequestHeaders[za0001][za0002])
			if err != nil {
				return
			}
		}
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z *RPCMsgOut) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "WAFResponse"
	o = append(o, 0x83, 0xab, 0x57, 0x41, 0x46, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65)
	o, err = z.WAFResponse.MarshalMsg(o)
	if err != nil {
		return
	}
	// string "RequestID"
	o = append(o, 0xa9, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x49, 0x44)
	o = msgp.AppendString(o, z.RequestID)
	// string "RequestHeaders"
	o = append(o, 0xae, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x48, 0x65, 0x61, 0x64, 0x65, 0x72, 0x73)
	o = msgp.AppendArrayHeader(o, uint32(len(z.RequestHeaders)))
	for za0001 := range z.RequestHeaders {
		o = msgp.AppendArrayHeader(o, uint32(2))
		for za0002 := range z.RequestHeaders[za0001] {
			o = msgp.AppendString(o, z.RequestHeaders[za0001][za0002])
		}
	}
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *RPCMsgOut) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			return
		}
		switch msgp.UnsafeString(field) {
		case "WAFResponse":
			bts, err = z.WAFResponse.UnmarshalMsg(bts)
			if err != nil {
				return
			}
		case "RequestID":
			z.RequestID, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				return
			}
		case "RequestHeaders":
			var zb0002 uint32
			zb0002, bts, err = msgp.ReadArrayHeaderBytes(bts)
			if err != nil {
				return
			}
			if cap(z.RequestHeaders) >= int(zb0002) {
				z.RequestHeaders = (z.RequestHeaders)[:zb0002]
			} else {
				z.RequestHeaders = make([][2]string, zb0002)
			}
			for za0001 := range z.RequestHeaders {
				var zb0003 uint32
				zb0003, bts, err = msgp.ReadArrayHeaderBytes(bts)
				if err != nil {
					return
				}
				if zb0003 != uint32(2) {
					err = msgp.ArrayError{Wanted: uint32(2), Got: zb0003}
					return
				}
				for za0002 := range z.RequestHeaders[za0001] {
					z.RequestHeaders[za0001][za0002], bts, err = msgp.ReadStringBytes(bts)
					if err != nil {
						return
					}
				}
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *RPCMsgOut) Msgsize() (s int) {
	s = 1 + 12 + z.WAFResponse.Msgsize() + 10 + msgp.StringPrefixSize + len(z.RequestID) + 15 + msgp.ArrayHeaderSize
	for za0001 := range z.RequestHeaders {
		s += msgp.ArrayHeaderSize
		for za0002 := range z.RequestHeaders[za0001] {
			s += msgp.StringPrefixSize + len(z.RequestHeaders[za0001][za0002])
		}
	}
	return
}
