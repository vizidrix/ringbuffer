package ringbuffer

import (
	"fmt"
)

type Batch struct {
	BatchNum        uint64
	BatchSize       uint64
	SeqNum          uint64
	GroupFlags      uint16
	ReaderFlags     [16]uint8
	ReleaseCallback uint64
	Data            [22]uint8
}

func (batch *Batch) String() string {
	return fmt.Sprintf("Batch# %d - Seq# % x [%d + %d] - Group: % x\n\tReader Flags: %s\n\tReleased Callback: %s\n\tData: %s",
		batch.BatchNum,
		batch.SeqNum,
		batch.SeqNum,
		batch.BatchSize,
		batch.GroupFlags,
		batch.ReaderFlags,
		batch.ReleaseCallback,
		batch.Data)
}
