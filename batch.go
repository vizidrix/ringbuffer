package ringbuffer

import (
	"fmt"
)

type Batch struct {
	Info     *RingBufferInfo
	BatchNum uint64
	SeqNum   uint64
	Size     uint8
	PubMask  uint8

	Entries []Entry
}

func (batch *Batch) Publish() (*PublishToken, error) {
	//DebugPrint("Published: %d", len(batch))
	// Listen for success event in goroutine

	return &PublishToken{
		Published: make(chan struct{}),
		Failed:    make(chan struct{}),
	}, nil
}

func (batch *Batch) String() string {
	return fmt.Sprintf("%s [#%d_seq%d_sz%d]", batch.Info, batch.BatchNum, batch.SeqNum, batch.Size)
}
