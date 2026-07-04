package receiver

import (
	"sort"

	"vexil/internal/protocol"
)

func (r *Receiver) findFirstMissing(offsets []int64, sizes []int64, totalSize int64, maxChunk int64) int64 {
	if len(offsets) == 0 {
		return 0
	}

	type block struct {
		offset int64
		size   int64
	}
	blocks := make([]block, len(offsets))
	for i, off := range offsets {
		sz := maxChunk
		if i < len(sizes) && sizes[i] > 0 {
			sz = sizes[i]
		}
		blocks[i] = block{offset: off, size: sz}
	}
	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i].offset < blocks[j].offset
	})

	expected := int64(0)
	for _, b := range blocks {
		if b.offset > expected {
			return expected
		}
		if b.offset+b.size > expected {
			expected = b.offset + b.size
		}
	}

	if expected < totalSize {
		return expected
	}
	return totalSize
}

func (r *Receiver) isComplete(totalSize int64, maxChunk int64) bool {
	chunkSize := maxChunk
	expectedChunks := (totalSize + chunkSize - 1) / chunkSize

	count := 0
	r.receivedSet.Range(func(key, value interface{}) bool {
		count++
		return true
	})

	return int64(count) >= expectedChunks
}

func (r *Receiver) findMissingRanges(offsets []int64, sizes []int64, totalSize int64, maxChunk int64) []protocol.AckRange {
	if len(offsets) == 0 {
		return []protocol.AckRange{{Offset: 0, Length: totalSize}}
	}

	type block struct {
		offset int64
		size   int64
	}
	blocks := make([]block, len(offsets))
	for i, off := range offsets {
		sz := maxChunk
		if i < len(sizes) && sizes[i] > 0 {
			sz = sizes[i]
		}
		blocks[i] = block{offset: off, size: sz}
	}
	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i].offset < blocks[j].offset
	})

	var missing []protocol.AckRange
	expected := int64(0)

	for _, b := range blocks {
		if b.offset > expected {
			missing = append(missing, protocol.AckRange{
				Offset: expected,
				Length: b.offset - expected,
			})
		}
		if b.offset+b.size > expected {
			expected = b.offset + b.size
		}
	}

	if expected < totalSize {
		missing = append(missing, protocol.AckRange{
			Offset: expected,
			Length: totalSize - expected,
		})
	}

	return missing
}