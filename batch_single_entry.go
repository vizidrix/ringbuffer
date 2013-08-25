package ringbuffer

import (
	"log"
)

type SingleEntryBatch struct {
	Buffer *RingBuffer
	SeqNum uint64
}

type SingleEntry struct {
	data []byte
}

func (buffer *RingBuffer) NewSingleEntryBatch(seq_num uint64) RingBufferBatchWriter {
	//log.Printf("Batch seq_num: %d", seq_num)
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

func (batch *SingleEntryBatch) Entry(index uint64) RingBufferEntryWriter {
	//if seq_num&batch.Buffer.GetInfo().GetSizeMask() != 0 {
	if index != 0 {
		panic("Invalid index into single entry batch")
	}
	seq_num := batch.SeqNum + index
	buffer := batch.Buffer.Entry(seq_num)
	//DebugPrint("buffer: %v", buffer)
	return &SingleEntry{
		data: buffer,
	}
}

func (batch *SingleEntryBatch) Publish() (*PublishToken, error) {
	err := batch.Buffer.Publish(batch)
	// TODO: figure out how to do pub token
	return nil, err
}

func (batch *SingleEntryBatch) Cancel() error {
	return batch.Buffer.Cancel(batch)
}

func (entry *SingleEntry) GetBuffer() []byte {
	return entry.data
}

func (entry *SingleEntry) CopyFrom(source []byte) {
	//log.Printf("CopyFrom: %v", source)
	//log.Printf("b-Data: %v", entry.data[0:])
	copy(entry.data[0:], source[:])
	//log.Printf("a-Data: %v", entry.data[0:])
}

func batch_single_entry_ignore() {
	log.Printf("")
}
