package ringbuffer

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"unsafe"
)

type RingBufferBatchWriter interface {
	GetBatchNum() uint64
	GetSeqNum() uint64
	GetBatchSize() uint16
	Entry(index uint64) []byte
	CopyTo(index uint64, data []byte)
	Publish() (*PublishToken, error)
	Cancel() error
}

type RingBufferEntryWriter interface {
	GetBuffer() []byte      // Zero copy option
	CopyFrom(source []byte) // Standard mem copy option
}

func batch_ignore() {
	log.Println(fmt.Sprintf("", 10))
	log.Printf("", reflect.SliceHeader{}, errors.New("stuff"), strings.HasPrefix("s", "q"), unsafe.Pointer(nil))
}
