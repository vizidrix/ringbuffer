package ringbuffer

import (
	"fmt"
)

type Batch struct {
	//Info     *RingBufferInfo
	//BatchNum uint64
	//SeqNum   uint64
	//Size    uint8
	//PubMask uint8

	Entries []*Entry
}

type Entry struct {
	Data []byte
}

func (batch *Batch) GetBatchNum() uint64 {
	return batch.Entries[0].GetBatchNum()
}

func (batch *Batch) GetBatchSize() uint16 {
	return uint16(len(batch.Entries))
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
	//return fmt.Sprintf("%s [#%d_seq%d_sz%d]", batch.Info, batch.BatchNum, batch.SeqNum, batch.Size)
	//return fmt.Sprintf("[#%d_seq%d_sz%d]", batch.GetBatchNum(), -1, len(batch.Entries))
	return fmt.Sprintf("%s", batch.Entries)
}

func (entry *Entry) GetBatchNum() uint64 {
	/*
		if len(entry.Data) < 6 {
			panic(fmt.Sprintf("Invalid data length: %d", len(entry.Data)))
		}
		batch_num := uint64(entry.Data[0]) << 24 &
			uint64(entry.Data[1]) << 16 &
			uint64(entry.Data[2]) << 8 &
			uint64(entry.Data[3]) << 0
		return batch_num
	*/
	return 10
}

// TODO: Add GetBatchIndex

func (entry *Entry) CopyFrom(data []byte) error {
	//DebugPrint("Copied to entry: % x", data)

	return nil
}
