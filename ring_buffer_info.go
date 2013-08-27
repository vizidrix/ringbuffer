package ringbuffer

import (
	"fmt"
)

type RingBufferInfo struct {
	buffer_size uint64
	size_mask   uint64
	entry_size  uint64
	total_size  uint64
}

func (info *RingBufferInfo) GetBufferSize() uint64 {
	return info.buffer_size
}

func (info *RingBufferInfo) GetSizeMask() uint64 {
	return info.size_mask
}

func (info *RingBufferInfo) GetEntrySize() uint64 {
	return info.entry_size
}

func (info *RingBufferInfo) GetTotalSize() uint64 {
	return info.total_size
}

func (info *RingBufferInfo) String() string {
	return fmt.Sprintf("[%dslots_%d]", info.buffer_size, info.total_size)
}
