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

type BUFFER_TYPES uint16

const (
	L0  BUFFER_TYPES = iota //           1 * 64 	=            64 (iota has been reset)
	L1                      //           8 * 64 	=           512
	L2                      //          16 * 64 	=         1,024
	L3                      //          32 * 64 	=         2,048
	L4                      //          64 * 64 	=         4,096
	L5                      //         8 * 4096 	=        32,768
	L6                      //        16 * 4096 	=        65,536
	L7                      //        32 * 4096 	=       131,072
	L8                      // 	      64 * 4096 	=       262,144
	L9                      //    8 * 64 * 4096 	=     2,097,152
	L10                     //   16 * 64 * 4096 	=     4,194,304
	L11                     //   32 * 64 * 4096 	=     8,388,608
	L12                     //   64 * 64 * 4096 	=    16,777,216
	L13                     //  8 * 4096 * 4086  	=   134,217,728
	L14                     // 16 * 4096 * 4096  	=   268,435,456
	L15                     // 32 * 4096 * 4096  	=   536,870,912
	L16                     // 64 * 4096 * 4096  	= 1,073,741,824
)

const (
	BATCH_POOL_SIZE = 64 * 4096
)

type RingBuffer struct {
	slices      [BATCH_POOL_SIZE]reflect.SliceHeader
	batch_index uint64
	buffer_ptr  *[0]byte
}

type ClaimResult struct {
	SeqNum    uint64
	BatchNum  uint64
	BatchSize uint64
}

type PublishToken struct {
	Published chan struct{}
	Failed    chan struct{}
}

func NewRingBuffer(buffer_type BUFFER_TYPES, data_size uint64) (*RingBuffer, error) {
	//DebugPrint("Making ring buffer [ Mode: %d / Size: %d / Blocks: %d ]", batching_mode, buffer_type, data_size)
	buffer := &RingBuffer{}
	_, err := C.rb_init_buffer(&buffer.buffer_ptr, C.uint8_t(buffer_type), C.uint64_t(data_size))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error initializing ring buffer [%d]", err))
	}
	size := int(buffer.GetInfo().GetEntrySize())
	//buffer.temp = reflect.SliceHeader{Data: uintptr(0), Len: size, Cap: size}
	buffer.batch_index = 0
	for i := 0; i < BATCH_POOL_SIZE; i++ {
		buffer.slices[i] = reflect.SliceHeader{Data: uintptr(0), Len: size, Cap: size}
		//buffer.batches[i] = &SingleEntryBatch{Buffer: buffer, SeqNum: 0}
	}
	//DebugPrint("Made ring buffer [ Mode: %d / Size: %d / Blocks: %d ]", batching_mode, buffer_type, data_size)
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

/*
func (buffer *RingBuffer) PublishRaw(data []byte) (*PublishToken, error) {
	batch, err := buffer.Claim(1)
	if err != nil {
		return nil, err
	}
	copy(batch.Entry(0)[0:], data[:])
	token, err := batch.Publish()
	if err != nil {
		return nil, err
	}
	return token, nil
}
*/

func (buffer *RingBuffer) Claim(count uint16) (*ClaimResult, error) {
	//seq_num, err := C.rb_claim(buffer.buffer_ptr, C.uint16_t(count))
	claim_result, err := C.rb_claim(buffer.buffer_ptr, C.uint16_t(count))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error: %d", claim_result.seq_num))
	}

	// TODO: Map claim_result into go struct
	// TODO: Switch buffer operations to use claim_result instead of batch
	//return buffer.NewSingleEntryBatch(uint64(claim_result.seq_num)), nil
	//return (*ClaimResult)(unsafe.Pointer(&claim_result)), nil
	return (*ClaimResult)(unsafe.Pointer(claim_result)), nil
}

func (buffer *RingBuffer) Entry(seq_num uint64) []byte {
	buffer.batch_index++
	buffer.slices[buffer.batch_index&(BATCH_POOL_SIZE-1)].Data = uintptr(C.rb_get_entry(buffer.buffer_ptr, C.uint64_t(seq_num)))
	return *(*[]byte)(unsafe.Pointer(&buffer.slices[buffer.batch_index&(BATCH_POOL_SIZE-1)]))
}

func (buffer *RingBuffer) Publish(batch *ClaimResult) error {
	//_, err := C.rb_publish(buffer.buffer_ptr, C.uint64_t(batch.SeqNum), C.uint16_t(batch.BatchSize))
	_, err := C.rb_publish(buffer.buffer_ptr, (*C.rb_claim_result)(unsafe.Pointer(batch)))

	return err
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
