package ringbuffer

import ()

type MultiEntryBatch struct {
	Buffer *RingBuffer
	Mode   BATCHING_MODE
	SeqNum uint64
}

/*
type MultiEntry struct {
	Data []byte
}
*/

func (buffer *RingBuffer) NewMultiEntryBatch(mode BATCHING_MODE, seq_num uint64) RingBufferBatchWriter {
	return &MultiEntryBatch{
		Buffer: buffer,
		Mode:   mode,
		SeqNum: seq_num,
	}
}

func (batch *MultiEntryBatch) GetBatchNum() uint64 {
	return batch.SeqNum
}

func (batch *MultiEntryBatch) GetSeqNum() uint64 {
	return batch.SeqNum
}

func (batch *MultiEntryBatch) GetBatchSize() uint16 {
	return 1
}

func (batch *MultiEntryBatch) Entry(index uint64) []byte {
	if index != 1 {
		panic("Invalid index into single entry batch")
	}
	return batch.Buffer.Entry(batch.SeqNum + index)
	//return &MultiEntry{
	//	Data: nil, //batch.Buffer.GetEntryAt(batch.SeqNum),
	//}
}

func (batch *MultiEntryBatch) CopyTo(index uint64, data []byte) {
	copy(batch.Buffer.Entry(batch.SeqNum + index)[0:], data[:])
}

func (batch *MultiEntryBatch) Publish() (*PublishToken, error) {
	err := batch.Buffer.Publish(batch)
	return nil, err
}

func (batch *MultiEntryBatch) Cancel() error {
	return batch.Buffer.Cancel(batch)
}

/*
func (entry *MultiEntry) GetBuffer() []byte {
	return nil
}

func (entry *MultiEntry) CopyFrom(source []byte) {
}
*/
