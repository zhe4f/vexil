package receiver

import (
	"encoding/binary"
	"fmt"
	"os"
)

const vexilMagic = 0x56455849
const vexilVersion = 5

type ResumeState struct {
	FileSize        int64
	TotalChunks     uint32
	ContiguousBytes int64
	ChunkHashes     [][32]byte
	ChunkOffsets    []int64
	ChunkSizes      []int64
}

func SaveResume(path string, state *ResumeState) error {
	numChunks := int(state.TotalChunks)
	headerSize := 4 + 4 + 8 + 4 + 8
	chunkSize := numChunks * 48
	totalSize := headerSize + chunkSize

	data := make([]byte, totalSize)
	binary.BigEndian.PutUint32(data[0:4], vexilMagic)
	binary.BigEndian.PutUint32(data[4:8], vexilVersion)
	binary.BigEndian.PutUint64(data[8:16], uint64(state.FileSize))
	binary.BigEndian.PutUint32(data[16:20], state.TotalChunks)
	binary.BigEndian.PutUint64(data[20:28], uint64(state.ContiguousBytes))

	offset := headerSize
	for i := 0; i < numChunks; i++ {
		if i < len(state.ChunkOffsets) {
			binary.BigEndian.PutUint64(data[offset:offset+8], uint64(state.ChunkOffsets[i]))
		}
		offset += 8
		if i < len(state.ChunkSizes) {
			binary.BigEndian.PutUint64(data[offset:offset+8], uint64(state.ChunkSizes[i]))
		}
		offset += 8
		if i < len(state.ChunkHashes) {
			copy(data[offset:offset+32], state.ChunkHashes[i][:])
		}
		offset += 32
	}

	// 优化 3：先写临时文件，再原子 rename
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("写入临时续传文件失败: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("替换续传文件失败: %w", err)
	}
	return nil
}

func LoadResume(path string) (*ResumeState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data) < 28 {
		return nil, fmt.Errorf("invalid resume file")
	}
	if binary.BigEndian.Uint32(data[0:4]) != vexilMagic {
		return nil, fmt.Errorf("bad magic")
	}

	version := binary.BigEndian.Uint32(data[4:8])
	fileSize := int64(binary.BigEndian.Uint64(data[8:16]))
	totalChunks := binary.BigEndian.Uint32(data[16:20])
	contiguousBytes := int64(binary.BigEndian.Uint64(data[20:28]))

	state := &ResumeState{
		FileSize:        fileSize,
		TotalChunks:     totalChunks,
		ContiguousBytes: contiguousBytes,
	}

	headerSize := 4 + 4 + 8 + 4 + 8

	// 清理残留的临时文件
	tmpPath := path + ".tmp"
	os.Remove(tmpPath)

	if version >= 5 {
		chunkSize := 48
		expectedEnd := headerSize + int(totalChunks)*chunkSize
		if len(data) >= expectedEnd {
			state.ChunkOffsets = make([]int64, totalChunks)
			state.ChunkSizes = make([]int64, totalChunks)
			state.ChunkHashes = make([][32]byte, totalChunks)
			off := headerSize
			for i := uint32(0); i < totalChunks; i++ {
				state.ChunkOffsets[i] = int64(binary.BigEndian.Uint64(data[off : off+8]))
				off += 8
				state.ChunkSizes[i] = int64(binary.BigEndian.Uint64(data[off : off+8]))
				off += 8
				copy(state.ChunkHashes[i][:], data[off:off+32])
				off += 32
			}
		}
	} else if version == 4 {
		chunkSize := 40
		expectedEnd := headerSize + int(totalChunks)*chunkSize
		if len(data) >= expectedEnd {
			state.ChunkOffsets = make([]int64, totalChunks)
			state.ChunkHashes = make([][32]byte, totalChunks)
			off := headerSize
			for i := uint32(0); i < totalChunks; i++ {
				state.ChunkOffsets[i] = int64(binary.BigEndian.Uint64(data[off : off+8]))
				off += 8
				copy(state.ChunkHashes[i][:], data[off:off+32])
				off += 32
			}
		}
	} else {
		chunkSize := 32
		expectedEnd := headerSize + int(totalChunks)*chunkSize
		if len(data) >= expectedEnd {
			state.ChunkHashes = make([][32]byte, totalChunks)
			off := headerSize
			for i := uint32(0); i < totalChunks; i++ {
				copy(state.ChunkHashes[i][:], data[off:off+32])
				off += 32
			}
		}
	}

	return state, nil
}