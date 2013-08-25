package ringbuffer

import ()

type SingleEntryBatch struct {
	Buffer *RingBuffer
	SeqNum uint64
}

type SingleEntry struct {
	Data []byte
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

func (batch *SingleEntryBatch) GetEntryAt(index uint16) RingBufferEntryWriter {
	if index != 1 {
		panic("Invalid index into single entry batch")
	}
	return &SingleEntry{
		Data: nil, //batch.Buffer.GetEntryAt(batch.SeqNum),
	}
}

func (batch *SingleEntryBatch) Publish() (*PublishToken, error) {
	err := batch.Buffer.Publish(batch)
	return nil, err
}

func (batch *SingleEntryBatch) Cancel() error {
	return batch.Buffer.Cancel(batch)
}

func (entry *SingleEntry) GetBuffer() []byte {
	return nil
}

func (entry *SingleEntry) CopyFrom(source []byte) {
}
