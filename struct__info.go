package ringbuffer

import (
	"fmt"
)

type Info struct {
	batch_buffer_size uint64
	batch_size_mask   uint64
	data_buffer_size  uint64
	data_size_mask    uint64
	entry_size        uint64
	total_data_size   uint64
}

func (info *Info) GetBatchBufferSize() uint64 {
	return info.batch_buffer_size
}

func (info *Info) GetBatchSizeMask() uint64 {
	return info.batch_size_mask
}

func (info *Info) GetDataBufferSize() uint64 {
	return info.data_buffer_size
}

func (info *Info) GetDataSizeMask() uint64 {
	return info.data_size_mask
}

func (info *Info) GetEntrySize() uint64 {
	return info.entry_size
}

func (info *Info) GetTotalDataSize() uint64 {
	return info.total_data_size
}

func (info *Info) String() string {
	return fmt.Sprintf("[Batches: %d - Entries: %d - Entry Size(bytes): %d - Size(bytes): %d]",
		info.batch_buffer_size,
		info.data_buffer_size,
		info.entry_size,
		info.total_data_size)
}
