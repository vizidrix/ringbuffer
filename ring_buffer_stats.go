package ringbuffer

import (
	"fmt"
)

type RingBufferStats struct {
	read_seq_num  uint64
	read_barrier  uint64
	write_seq_num uint64
	write_barrier uint64
	batch_num     uint64
}

func (stats *RingBufferStats) GetReadSeqNum() uint64 {
	return stats.read_seq_num
}

func (stats *RingBufferStats) GetReadBarrier() uint64 {
	return stats.read_barrier
}

func (stats *RingBufferStats) GetWriteSeqNum() uint64 {
	return stats.write_seq_num
}

func (stats *RingBufferStats) GetWriterBarrier() uint64 {
	return stats.write_barrier
}

func (stats *RingBufferStats) GetBatchNum() uint64 {
	return stats.batch_num
}

func (stats *RingBufferStats) String() string {
	return fmt.Sprintf("R[%d | %d] W[%d | %d] { %d }",
		stats.read_seq_num,
		stats.read_barrier,
		stats.write_seq_num,
		stats.write_barrier,
		stats.batch_num)
}
