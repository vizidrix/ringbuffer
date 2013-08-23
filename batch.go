package ringbuffer

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"unsafe"
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
	if len(entry.Data) < 6 {
		panic(fmt.Sprintf("Invalid data length: %d", len(entry.Data)))
	}
	//log.Printf("entry.Data[0]: %s", uint64(entry.Data[0])<<24)
	//batch_num := (uint64(entry.Data[0]) << 24) |
	//	(uint64(entry.Data[1]) << 16)
	batch_num := (uint64(entry.Data[0])<<24 |
		uint64(entry.Data[1])<<16 |
		uint64(entry.Data[2])<<8 |
		uint64(entry.Data[3])<<0)
	//log.Printf("batch_num: %d", batch_num)
	return batch_num
}

func (entry *Entry) GetIndex() uint16 {
	if len(entry.Data) < 6 {
		panic(fmt.Sprintf("Invalid data length: %d", len(entry.Data)))
	}
	index := (uint16(entry.Data[0])<<8 |
		uint16(entry.Data[1])<<0)
	return index
}

// TODO: Add GetBatchIndex

func (entry *Entry) CopyFrom(data []byte) error {
	//DebugPrint("Copied to entry: % x", data)

	return nil
}

func batch_ignore() {
	log.Println(fmt.Sprintf("", 10))
	log.Printf("", reflect.SliceHeader{}, errors.New("stuff"), strings.HasPrefix("s", "q"), unsafe.Pointer(nil))
}
