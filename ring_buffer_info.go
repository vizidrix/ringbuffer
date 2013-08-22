package ringbuffer

import (
	"fmt"
)

type RingBufferInfo struct {
	buffer_type uint8
	buffer_size uint64
	chunk_count uint8
	data_size   uint64
}

func (info *RingBufferInfo) GetBufferType() BUFFER_TYPES {
	return (BUFFER_TYPES)(info.buffer_type)
}

func (info *RingBufferInfo) GetBufferSize() uint64 {
	return info.buffer_size
}

func (info *RingBufferInfo) GetChunkCount() uint8 {
	return info.chunk_count
}

func (info *RingBufferInfo) GetDataSize() uint64 {
	return info.data_size
}

func (info *RingBufferInfo) String() string {
	return fmt.Sprintf("[%d_%dslots_%dx32-%db]", info.buffer_type, info.buffer_size, info.chunk_count, info.data_size)
}
