package ringbuffer

import (
	"fmt"
)

type Batch struct {
	batch_num       uint64
	batch_size      uint64
	seq_num         uint64
	data            [5]uint64
	state           BatchStates
	__state_padding [7]uint64
	group_flags     [8]uint64
	reader_flags    [8]uint64
}

func (batch *Batch) GetBatchNum() uint64 {
	return batch.batch_num
}

func (batch *Batch) GetBatchSize() uint64 {
	return batch.batch_size
}

func (batch *Batch) GetSeqNum() uint64 {
	return batch.seq_num
}

func (batch *Batch) GetState() BatchStates {
	return batch.state
}

func (batch *Batch) GetData() [5]uint64 {
	return batch.data
}

func (batch *Batch) GetGroupFlags() uint64 {
	return batch.group_flags[0]
}

func (batch *Batch) String() string {
	return fmt.Sprintf("Batch# %d - State: %s - Seq# % x [%d + %d] - Group: % x\n\tReader Flags: %s\n\tData: %s",
		batch.batch_num,
		batch.state,
		batch.seq_num,
		batch.seq_num,
		batch.batch_size,
		batch.group_flags[0],
		batch.reader_flags,
		batch.data)
}
