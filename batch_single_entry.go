package ringbuffer

import (
	"log"
	//"unsafe"
)

type SingleEntryBatch struct {
	Buffer *RingBuffer
	SeqNum uint64
}

type SingleEntry struct {
	data []byte
}

func (buffer *RingBuffer) NewSingleEntryBatch(seq_num uint64) RingBufferBatchWriter {
	return &SingleEntryBatch{
		Buffer: buffer,
		SeqNum: seq_num,
	}
}

func (batch *SingleEntryBatch) GetBatchNum() uint64 {
	return batch.SeqNum
}

func (batch *SingleEntryBatch) GetSeqNum() uint64 {
	return batch.SeqNum
}

func (batch *SingleEntryBatch) GetBatchSize() uint16 {
	return 1
}

func (batch *SingleEntryBatch) Entry(index uint64) []byte {
	if index != 0 {
		panic("Invalid index into single entry batch")
	}
	return batch.Buffer.Entry(batch.SeqNum + index)
	//return (*SingleEntry)()
	//buffer := batch.Buffer.Entry(batch.SeqNum + index)
	//return &SingleEntry{
	//	data: buffer,
	//}
}

func (batch *SingleEntryBatch) CopyTo(index uint64, data []byte) {
	if index != 0 {
		panic("Invalid index into single entry batch")
	}
	copy(batch.Buffer.Entry(batch.SeqNum + index)[0:], data[:])
}

func (batch *SingleEntryBatch) Publish() (*PublishToken, error) {
	err := batch.Buffer.Publish(batch)
	// TODO: figure out how to do pub token
	return nil, err
}

func (batch *SingleEntryBatch) Cancel() error {
	return batch.Buffer.Cancel(batch)
}

/*
func (entry *SingleEntry) GetBuffer() []byte {
	return entry.data
}

func (entry *SingleEntry) CopyFrom(source []byte) {
	copy(entry.data[0:], source[:])
}
*/

func batch_single_entry_ignore() {
	log.Printf("")
}
