package protocol

import (
	"encoding/binary"
	"fmt"
)

const (
	MSG_CONN_ROLE    = 0x01
	MSG_HELLO        = 0x02
	MSG_READY        = 0x03
	MSG_FILE_REQUEST = 0x04
	MSG_DATA         = 0x05
	MSG_GO           = 0x06
	MSG_COMPLETE     = 0x07
	MSG_RESUME_RANGE = 0x0D
	MSG_BATCH_START  = 0x0A
	MSG_BATCH_END    = 0x0B
	MSG_ACK_RANGE    = 0x0C
)

const (
	HeaderSize    = 20
	MaxPayloadLen = 16*1024*1024 + 1024
)

type Frame struct {
	Type    uint8
	FileID  uint32
	Offset  uint64
	Payload []byte
}

func (f *Frame) Encode() []byte {
	buf := make([]byte, HeaderSize+len(f.Payload))
	buf[0] = f.Type
	buf[1] = byte(f.FileID >> 16)
	buf[2] = byte(f.FileID >> 8)
	buf[3] = byte(f.FileID)
	binary.BigEndian.PutUint64(buf[8:16], f.Offset)
	binary.BigEndian.PutUint32(buf[16:20], uint32(len(f.Payload)))
	copy(buf[HeaderSize:], f.Payload)
	return buf
}

func DecodeHeader(data []byte) (*Frame, error) {
	if len(data) < HeaderSize {
		return nil, fmt.Errorf("frame too short: %d", len(data))
	}
	payloadLen := binary.BigEndian.Uint32(data[16:20])
	if payloadLen > MaxPayloadLen {
		return nil, fmt.Errorf("payload too large: %d", payloadLen)
	}
	f := &Frame{
		Type:   data[0],
		FileID: uint32(data[1])<<16 | uint32(data[2])<<8 | uint32(data[3]),
		Offset: binary.BigEndian.Uint64(data[8:16]),
	}
	if len(data) > HeaderSize {
		f.Payload = data[HeaderSize:]
	}
	return f, nil
}

func MessageName(t uint8) string {
	switch t {
	case MSG_CONN_ROLE:
		return "CONN_ROLE"
	case MSG_HELLO:
		return "HELLO"
	case MSG_READY:
		return "READY"
	case MSG_FILE_REQUEST:
		return "FILE_REQUEST"
	case MSG_DATA:
		return "DATA"
	case MSG_GO:
		return "GO"
	case MSG_COMPLETE:
		return "COMPLETE"
	case MSG_RESUME_RANGE:
		return "RESUME_RANGE"
	case MSG_BATCH_START:
		return "BATCH_START"
	case MSG_BATCH_END:
		return "BATCH_END"
	case MSG_ACK_RANGE:
		return "ACK_RANGE"
	default:
		return fmt.Sprintf("UNKNOWN(0x%02X)", t)
	}
}