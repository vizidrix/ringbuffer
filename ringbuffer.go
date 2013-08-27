package ringbuffer

/*
#include "ringbuffer.h"
*/
import "C"
import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"unsafe"
)

const (
	BATCH_POOL_SIZE = 64 * 4096
)

type RingBuffer struct {
	slices      [BATCH_POOL_SIZE]reflect.SliceHeader
	batch_index uint64
	buffer_ptr  *[0]byte
}

type Batch struct {
	SeqNum    uint64
	BatchNum  uint64
	BatchSize uint64
}

type PublishToken struct {
	Published chan struct{}
	Failed    chan struct{}
}

func NewRingBuffer(buffer_size uint64, data_size uint64) (*RingBuffer, error) {
	//DebugPrint("Making ring buffer [ Mode: %d / Size: %d / Blocks: %d ]", batching_mode, buffer_type, data_size)
	buffer := &RingBuffer{}
	_, err := C.rb_init_buffer(&buffer.buffer_ptr, C.uint64_t(buffer_size), C.uint64_t(data_size))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error initializing ring buffer [%d]", err))
	}
	//size := int(buffer.GetInfo().GetEntrySize())

	// TODO: Fix pool impl
	//buffer.batch_index = 0
	//for i := 0; i < BATCH_POOL_SIZE; i++ {
	//	buffer.slices[i] = reflect.SliceHeader{Data: uintptr(0), Len: size, Cap: size}
	//buffer.batches[i] = &SingleEntryBatch{Buffer: buffer, SeqNum: 0}
	//}
	return buffer, nil
}

func (buffer *RingBuffer) Close() error {
	_, err := C.rb_release_buffer(buffer.buffer_ptr)
	if err != nil {
		return errors.New(fmt.Sprintf("Error closing ring buffer [%d]", err))
	}
	return nil
}

func (buffer *RingBuffer) GetInfo() *RingBufferInfo {
	info_ptr := C.rb_get_info(buffer.buffer_ptr)
	return (*RingBufferInfo)(unsafe.Pointer(info_ptr))
}

func (buffer *RingBuffer) GetStats() *RingBufferStats {
	stats_ptr := C.rb_get_stats(buffer.buffer_ptr)
	return (*RingBufferStats)(unsafe.Pointer(stats_ptr))
}

func (buffer *RingBuffer) Claim(count uint16) (*Batch, error) {
	batch, err := C.rb_claim(buffer.buffer_ptr, C.uint16_t(count))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to claim [%d]: %s", count, err))
	}
	return (*Batch)(unsafe.Pointer(batch)), nil
}

func (buffer *RingBuffer) Entry(seq_num uint64) []byte {
	return *(*[]byte)(unsafe.Pointer((*reflect.SliceHeader)(C.rb_get_entry_slice(buffer.buffer_ptr, C.uint64_t(seq_num)))))
}

func (buffer *RingBuffer) Publish(batch *Batch) error {
	_, err := C.rb_publish(buffer.buffer_ptr, (*C.rb_batch)(unsafe.Pointer(batch)))
	return err
}

func (buffer *RingBuffer) ClaimAndPublish(count int) {
	C.rb_claim_and_publish(buffer.buffer_ptr, C.int(count)) //, C.uint16_t(count))
}

var debugEnabled bool = true

func EnableDebug() {
	debugEnabled = true
}
func DisableDebug() {
	debugEnabled = false
}

//export DebugPrintf
func DebugPrintf(format *C.char) {
	DebugPrint(C.GoString(format))
}
func DebugPrint(format string, args ...interface{}) {
	if !debugEnabled {
		return
	}
	log.Printf(format, args...)
}

func ring_buffer_ignore() {
	log.Println(fmt.Sprintf("", 10))
	log.Printf("", reflect.SliceHeader{}, errors.New("stuff"), strings.HasPrefix("s", "q"), unsafe.Pointer(nil))
}
