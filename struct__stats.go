package ringbuffer

import (
	"fmt"
)

type Stats struct {
	barrier_batch_num uint64
	read_batch_num    uint64
	write_batch_num   uint64
	barrier_seq_num   uint64
	write_seq_num     uint64
	__padding         [3]uint64
}

func (stats *Stats) GetBarrierBatchNum() uint64 {
	return stats.barrier_batch_num
}

func (stats *Stats) GetReadBatchNum() uint64 {
	return stats.read_batch_num
}

func (stats *Stats) GetWriteBatchNum() uint64 {
	return stats.write_batch_num
}

func (stats *Stats) GetBarrierSeqNum() uint64 {
	return stats.barrier_seq_num
}

func (stats *Stats) GetWriteSeqNum() uint64 {
	return stats.write_seq_num
}

func (stats *Stats) String() string {
	return fmt.Sprintf("GO State - Batch [ B %d | R %d | W %d ] Seq [ B %d | W %d ]",
		stats.barrier_batch_num,
		stats.read_batch_num,
		stats.write_batch_num,
		stats.barrier_seq_num,
		stats.write_seq_num)
}
