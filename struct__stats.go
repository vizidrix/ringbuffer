package ringbuffer

import (
	"fmt"
)

type Stats struct {
	barrier_batch_num [8]uint64
	read_batch_num    [8]uint64
	write_batch_num   [8]uint64
	barrier_seq_num   [8]uint64
	write_seq_num     [8]uint64
}

func (stats *Stats) GetBarrierBatchNum() uint64 {
	return stats.barrier_batch_num[0]
}

func (stats *Stats) GetReadBatchNum() uint64 {
	return stats.read_batch_num[0]
}

func (stats *Stats) GetWriteBatchNum() uint64 {
	return stats.write_batch_num[0]
}

func (stats *Stats) GetBarrierSeqNum() uint64 {
	return stats.barrier_seq_num[0]
}

func (stats *Stats) GetWriteSeqNum() uint64 {
	return stats.write_seq_num[0]
}

func (stats *Stats) String() string {
	return fmt.Sprintf("GO State - Batch [ B %d | R %d | W %d ] Seq [ B %d | W %d ]",
		stats.barrier_batch_num[0],
		stats.read_batch_num[0],
		stats.write_batch_num[0],
		stats.barrier_seq_num[0],
		stats.write_seq_num[0])
}
