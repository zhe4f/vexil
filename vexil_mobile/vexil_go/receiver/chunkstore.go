package receiver

import (
	"crypto/sha256"
	"fmt"
	"sort"
)

// chunkEntry 记录一个已接收数据块的元信息
type chunkEntry struct {
	offset int64
	size   int64
	hash   [32]byte
}

// WriteError 表示 Worker 中的写入错误
type WriteError struct {
	Offset  int64
	Size    int64
	FileIdx int
	Err     error
}

func (e *WriteError) Error() string {
	return fmt.Sprintf("写入失败 [offset=%d, size=%d]: %v", e.Offset, e.Size, e.Err)
}

// resetCollected 清空已收集的块信息
func (r *Receiver) resetCollected() {
	r.collectMu.Lock()
	r.collected = r.collected[:0]
	r.collectMu.Unlock()

	r.receivedSet.Range(func(key, value interface{}) bool {
		r.receivedSet.Delete(key)
		return true
	})
}

// addChunk 记录一个已接收的块
func (r *Receiver) addChunk(offset int64, size int64, hash [32]byte) {
	r.collectMu.Lock()
	r.collected = append(r.collected, chunkEntry{offset: offset, size: size, hash: hash})
	r.collectMu.Unlock()
}

// computeFinalHash 对所有已接收块的哈希做排序去重后串联，计算最终 SHA-256
func (r *Receiver) computeFinalHash() string {
	r.collectMu.RLock()
	entries := make([]chunkEntry, len(r.collected))
	copy(entries, r.collected)
	r.collectMu.RUnlock()

	seen := make(map[int64]bool)
	unique := make([]chunkEntry, 0, len(entries))
	for _, e := range entries {
		if !seen[e.offset] {
			unique = append(unique, e)
			seen[e.offset] = true
		}
	}

	sort.Slice(unique, func(i, j int) bool {
		return unique[i].offset < unique[j].offset
	})

	hasher := sha256.New()
	for _, e := range unique {
		hasher.Write(e.hash[:])
	}
	return fmt.Sprintf("%x", hasher.Sum(nil))
}