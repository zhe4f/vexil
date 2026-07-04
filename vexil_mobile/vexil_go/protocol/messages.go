package protocol

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
)

// ============ ConnRole ============

type ConnRole struct {
	Role string `json:"role"`
}

func EncodeConnRole(r ConnRole) (*Frame, error) {
	payload, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("encode conn role: %w", err)
	}
	return &Frame{Type: MSG_CONN_ROLE, Payload: payload}, nil
}

func DecodeConnRole(f *Frame) (*ConnRole, error) {
	var r ConnRole
	if err := json.Unmarshal(f.Payload, &r); err != nil {
		return nil, fmt.Errorf("decode conn role: %w", err)
	}
	return &r, nil
}

// ============ Hello ============

type Hello struct {
	NumConns int `json:"c"`
}

func EncodeHello(h Hello) (*Frame, error) {
	payload, err := json.Marshal(h)
	if err != nil {
		return nil, fmt.Errorf("encode hello: %w", err)
	}
	return &Frame{Type: MSG_HELLO, Payload: payload}, nil
}

func DecodeHello(f *Frame) (*Hello, error) {
	var h Hello
	if err := json.Unmarshal(f.Payload, &h); err != nil {
		return nil, fmt.Errorf("decode hello: %w", err)
	}
	return &h, nil
}

// ============ Ready ============

func EncodeReady() (*Frame, error) {
	return &Frame{Type: MSG_READY, Payload: nil}, nil
}

// ============ Go ============

func EncodeGo() (*Frame, error) {
	return &Frame{Type: MSG_GO, Payload: nil}, nil
}

// ============ FileRequest ============

type FileRequest struct {
	FileName string `json:"n"`
	FileSize int64  `json:"s"`
	NumConns int    `json:"c"`
	SenderName string `json:"h"`
}

func EncodeFileRequest(req FileRequest) (*Frame, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("encode file request: %w", err)
	}
	return &Frame{Type: MSG_FILE_REQUEST, Payload: payload}, nil
}

func DecodeFileRequest(f *Frame) (*FileRequest, error) {
	var req FileRequest
	if err := json.Unmarshal(f.Payload, &req); err != nil {
		return nil, fmt.Errorf("decode file request: %w", err)
	}
	return &req, nil
}

// ============ Complete ============

func EncodeComplete(checksum []byte) (*Frame, error) {
	return &Frame{Type: MSG_COMPLETE, Payload: checksum}, nil
}

// ============ ACK Range ============

type AckRange struct {
	Offset int64
	Length int64
}

func EncodeAckRange(a AckRange) (*Frame, error) {
	payload := make([]byte, 16)
	binary.BigEndian.PutUint64(payload[0:8], uint64(a.Offset))
	binary.BigEndian.PutUint64(payload[8:16], uint64(a.Length))
	return &Frame{Type: MSG_ACK_RANGE, Payload: payload}, nil
}

func DecodeAckRange(f *Frame) AckRange {
	if len(f.Payload) < 16 {
		return AckRange{}
	}
	return AckRange{
		Offset: int64(binary.BigEndian.Uint64(f.Payload[0:8])),
		Length: int64(binary.BigEndian.Uint64(f.Payload[8:16])),
	}
}

// ============ Resume Range ============

func EncodeResumeRange(missing []AckRange) (*Frame, error) {
	payload := make([]byte, 4+len(missing)*16)
	binary.BigEndian.PutUint32(payload[0:4], uint32(len(missing)))
	for i, m := range missing {
		off := 4 + i*16
		binary.BigEndian.PutUint64(payload[off:off+8], uint64(m.Offset))
		binary.BigEndian.PutUint64(payload[off+8:off+16], uint64(m.Length))
	}
	return &Frame{Type: MSG_RESUME_RANGE, Payload: payload}, nil
}

func DecodeResumeRange(f *Frame) []AckRange {
	if len(f.Payload) < 4 {
		return nil
	}
	count := binary.BigEndian.Uint32(f.Payload[0:4])
	ranges := make([]AckRange, 0, count)
	for i := uint32(0); i < count; i++ {
		off := 4 + int(i)*16
		if off+16 > len(f.Payload) {
			break
		}
		ranges = append(ranges, AckRange{
			Offset: int64(binary.BigEndian.Uint64(f.Payload[off : off+8])),
			Length: int64(binary.BigEndian.Uint64(f.Payload[off+8 : off+16])),
		})
	}
	return ranges
}

// ============ BatchStart ============

type BatchStart struct {
	TotalUnits uint32         `json:"tu"`
	Files      []FileManifest `json:"fs"`
}

type FileManifest struct {
	Name string `json:"n"`
	Size int64  `json:"s"`
}

func EncodeBatchStart(b BatchStart) (*Frame, error) {
	payload, err := json.Marshal(b)
	if err != nil {
		return nil, fmt.Errorf("encode batch start: %w", err)
	}
	return &Frame{Type: MSG_BATCH_START, Payload: payload}, nil
}

func DecodeBatchStart(f *Frame) (*BatchStart, error) {
	var b BatchStart
	if err := json.Unmarshal(f.Payload, &b); err != nil {
		return nil, fmt.Errorf("decode batch start: %w", err)
	}
	return &b, nil
}

// ============ BatchEnd ============

func EncodeBatchEnd() (*Frame, error) {
	return &Frame{Type: MSG_BATCH_END, Payload: nil}, nil
}

// ============ BlockHash ============

type BlockHash struct {
	Offset int64
	Hash   [32]byte
}

func EncodeResumeRangeEx(missing []AckRange, existing []BlockHash) (*Frame, error) {
	payload := make([]byte, 4+len(missing)*16+4+len(existing)*40)

	binary.BigEndian.PutUint32(payload[0:4], uint32(len(missing)))
	for i, m := range missing {
		off := 4 + i*16
		binary.BigEndian.PutUint64(payload[off:off+8], uint64(m.Offset))
		binary.BigEndian.PutUint64(payload[off+8:off+16], uint64(m.Length))
	}

	existingOff := 4 + len(missing)*16
	binary.BigEndian.PutUint32(payload[existingOff:existingOff+4], uint32(len(existing)))
	for i, b := range existing {
		off := existingOff + 4 + i*40
		binary.BigEndian.PutUint64(payload[off:off+8], uint64(b.Offset))
		copy(payload[off+8:off+40], b.Hash[:])
	}

	return &Frame{Type: MSG_RESUME_RANGE, Payload: payload}, nil
}

func DecodeResumeRangeEx(f *Frame) ([]AckRange, []BlockHash) {
	if len(f.Payload) < 4 {
		return nil, nil
	}

	missingCount := binary.BigEndian.Uint32(f.Payload[0:4])
	missing := make([]AckRange, missingCount)
	for i := uint32(0); i < missingCount; i++ {
		off := 4 + int(i)*16
		if off+16 > len(f.Payload) {
			break
		}
		missing[i] = AckRange{
			Offset: int64(binary.BigEndian.Uint64(f.Payload[off : off+8])),
			Length: int64(binary.BigEndian.Uint64(f.Payload[off+8 : off+16])),
		}
	}

	existingOff := 4 + int(missingCount)*16
	if existingOff+4 > len(f.Payload) {
		return missing, nil
	}

	existingCount := binary.BigEndian.Uint32(f.Payload[existingOff : existingOff+4])
	existing := make([]BlockHash, existingCount)
	for i := uint32(0); i < existingCount; i++ {
		off := existingOff + 4 + int(i)*40
		if off+40 > len(f.Payload) {
			break
		}
		existing[i] = BlockHash{
			Offset: int64(binary.BigEndian.Uint64(f.Payload[off : off+8])),
		}
		copy(existing[i].Hash[:], f.Payload[off+8:off+40])
	}

	return missing, existing
}