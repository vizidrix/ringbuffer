package ringbuffer

/*
#include "ringbuffer.h"
*/
import "C"
import (
	"errors"
	"fmt"
	"log"
	//"os"
	"reflect"
	"strings"
	"unsafe"
)

type BUFFER_TYPES uint16
type BATCHING_MODE uint8

const (
	L0  BUFFER_TYPES = iota //         1 * 64 	=         64 (iota has been reset)
	L1                      //         8 * 64 	=        512
	L2                      //        16 * 64 	=      1,024
	L3                      //        32 * 64 	=      2,048
	L4                      //        64 * 64 	=      4,096
	L5                      //       8 * 4096 	=     32,768
	L6                      //      16 * 4096 	=     65,536
	L7                      //      32 * 4096 	=    131,072
	L8                      // 	    64 * 4096 	=    262,144
	L9                      //  8 * 64 * 4096 	=  2,097,152
	L10                     // 16 * 64 * 4096 	=  4,194,304
	L11                     // 32 * 64 * 4096 	=  8,388,608
	L12                     // 64 * 64 * 4096 	= 16,777,216
)

const (
	NONE        BATCHING_MODE = iota // Disables tracking of batches, enables use of full entry space for data
	SMALL_BATCH                      // Uses a single byte to track batch size for a heade size of 5
	LARGE_BATCH                      // Uses two bytes to track batch size for a header size of 6
)

type RingBuffer struct {
	buffer_ptr *[0]byte
}

type PublishToken struct {
	Published chan struct{}
	Failed    chan struct{}
}

func NewRingBuffer(buffer_type BUFFER_TYPES, batching_mode BATCHING_MODE, data_size uint64) (*RingBuffer, error) {
	//DebugPrint("Making ring buffer [ Mode: %d / Size: %d / Blocks: %d ]", writer_mode, buffer_size, entry_size)
	buffer := &RingBuffer{}
	result := C.rb_init_buffer(&buffer.buffer_ptr, C.uint8_t(buffer_type), C.rb_batching_mode_t(batching_mode), C.uint64_t(data_size))
	if result != 0 {
		return nil, errors.New(fmt.Sprintf("Error initializing ring buffer [%d]", result))
	}
	return buffer, nil
}

func (buffer *RingBuffer) Close() error {
	result := C.rb_release_buffer(buffer.buffer_ptr)
	if result != 0 {
		return errors.New(fmt.Sprintf("Error closing ring buffer [%d]", result))
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

func (buffer *RingBuffer) PublishRaw(data []byte) (*PublishToken, error) {
	batch, err := buffer.Claim(1) //(1)
	if err != nil {
		return nil, err
	}
	batch.Entry(0).CopyFrom(data)
	//batch.Entries[0].CopyFrom(data)
	//token, err := batch.Publish()
	token, err := batch.Publish()
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (buffer *RingBuffer) Claim(count uint16) (RingBufferBatchWriter, error) {
	//var batch_ptr unsafe.Pointer //unsafe.Pointer // = unsafe.Pointer(&count) //**[0]byte
	//ptr := &struct{}{}
	//batch_ptr := unsafe.Pointer(ptr)
	//batch := &Batch{}
	//var batch *Batch = &Batch{}
	//batch_ptr := unsafe.Pointer(batch) //*new(unsafe.Pointer)
	//log.Printf("batch_ptr: %s", batch_ptr)
	//var batch_ptr *[0]byte
	//result := 0
	//batch_ptr := unsafe.Pointer(batch_addr)
	//var batch_ptr *C.uint64_t
	seq_num, err := C.rb_claim(buffer.buffer_ptr, C.uint16_t(count))
	//log.Printf("err: %s", err)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error: %d", seq_num))
	}

	switch buffer.GetInfo().GetBatchingMode() {
	case NONE:
		{
			//log.Printf("Batching mode NONE")
			return buffer.NewSingleEntryBatch(uint64(seq_num)), nil
		}
	case SMALL_BATCH:
		{
			//log.Printf("Batching mode SMALL BATCH")
			return buffer.NewMultiEntryBatch(SMALL_BATCH, uint64(seq_num)), nil
		}
	case LARGE_BATCH:
		{
			//log.Printf("Batching mode LARGE BATCH")
			return buffer.NewMultiEntryBatch(LARGE_BATCH, uint64(seq_num)), nil
		}
	}

	return nil, errors.New("Unable to find batch type")
}

func (buffer *RingBuffer) Entry(seq_num uint64) []byte {
	ptr := C.rb_get_entry(buffer.buffer_ptr, C.uint64_t(seq_num))
	//log.Printf("ptr: %v - seq: %d", ptr, seq_num)
	//size := C.int(buffer.GetInfo().GetEntrySize())
	size := int(buffer.GetInfo().GetEntrySize())
	data := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(ptr)),
		Len:  size,
		Cap:  size,
	}))
	return data
	//return C.GoBytes(ptr, size)
}

func (buffer *RingBuffer) Cancel(batch RingBufferBatchWriter) error {
	//C.rb_cancel_batch(buffer.buffer_ptr, (*[0]byte)(unsafe.Pointer(batch)), C.uint16_t(len(batch.Entries)))
	C.rb_cancel(buffer.buffer_ptr, C.uint64_t(batch.GetSeqNum()), C.uint16_t(batch.GetBatchSize()))

	return nil
}

func (buffer *RingBuffer) Publish(batch RingBufferBatchWriter) error {
	//C.rb_publish_batch(buffer.buffer_ptr, (*[0]byte)(unsafe.Pointer(batch)), C.uint16_t(len(batch.Entries)))
	C.rb_publish(buffer.buffer_ptr, C.uint64_t(batch.GetSeqNum()), C.uint16_t(batch.GetBatchSize()))

	return nil
}

/* This section works!! */
/*
	batch := *(*[]*byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(batch_ptr),
		Len:  int(count),
		Cap:  int(count),
	}))
	log.Printf("Batch: %s", ((*Batch)(unsafe.Pointer(&batch))))

	//log.Printf("Len: %d", len(batch))
	//log.Printf("Entry size: %d", int(buffer.GetInfo().entry_size))
	//log.Printf("Batch: %s", batch[:])
	for i := 0; i < int(count); i++ {
		//log.Printf("Batch[%d]: %s", i, batch[i])
		entry := *(*Entry)(unsafe.Pointer(&reflect.SliceHeader{
			Data: uintptr(unsafe.Pointer(batch[i])),
			Len:  int(buffer.GetInfo().entry_size),
			Cap:  int(buffer.GetInfo().entry_size),
		}))
		log.Printf("Entry[%d]: %v", i, entry.Data[:])
	}
*/

//var temp *Batch = (*Batch)(unsafe.Pointer(&reflect.SliceHeader{
/*
	batch = (*Batch)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(batch_ptr),
		Len:  int(count),
		Cap:  int(count),
	}))
	for i := 0; i < int(count); i++ {
		batch.Entries[i] = (*BatchEntry)(unsafe.Pointer(&reflect.SliceHeader{
			Data: uintptr(unsafe.Pointer(batch.Entries[i])),
			Len:  int(buffer.GetInfo().entry_size),
			Cap:  int(buffer.GetInfo().entry_size),
		}))
		//log.Printf("Batch Entry[%d]: %v", i, batch.Entries[i].Data[:])
	}
*/
//log.Printf("Batch: %s", batch)
//return batch, nil

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
	//log.Printf(C.GoString(format))
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
