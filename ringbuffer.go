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

type RingBuffer struct {
	buffer_ptr *[0]byte
}

type PublishToken struct {
	Published chan struct{}
	Failed    chan struct{}
}

func NewRingBuffer(buffer_type BUFFER_TYPES, data_size uint64) (*RingBuffer, error) {
	//DebugPrint("Making ring buffer [ Mode: %d / Size: %d / Blocks: %d ]", writer_mode, buffer_size, entry_size)
	buffer := &RingBuffer{}
	result := C.rb_init_buffer(&buffer.buffer_ptr, C.uint8_t(buffer_type), C.uint64_t(data_size))
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

func (buffer *RingBuffer) Publish(data []byte) (*PublishToken, error) {
	batch, err := buffer.Claim(1)
	if err != nil {
		return nil, err
	}
	batch.Entries[0].CopyFrom(data)
	token, err := batch.Publish()
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (buffer *RingBuffer) Claim(count uint16) (*Batch, error) {
	//var batch_ptr unsafe.Pointer //unsafe.Pointer // = unsafe.Pointer(&count) //**[0]byte
	//ptr := &struct{}{}
	//batch_ptr := unsafe.Pointer(ptr)
	batch := &Batch{}
	batch_ptr := unsafe.Pointer(batch) //*new(unsafe.Pointer)
	//log.Printf("batch_ptr: %s", batch_ptr)
	//var batch_ptr *[0]byte
	//result := 0
	//batch_ptr := unsafe.Pointer(batch_addr)
	result := C.rb_claim(buffer.buffer_ptr, &batch_ptr, C.uint16_t(count))
	if result != 0 {
		return nil, errors.New(fmt.Sprintf("Error: %d", result))
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
	batch = (*Batch)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(batch_ptr),
		Len:  int(count),
		Cap:  int(count),
	}))
	for i := 0; i < int(count); i++ {
		batch.Entries[i] = (*Entry)(unsafe.Pointer(&reflect.SliceHeader{
			Data: uintptr(unsafe.Pointer(batch.Entries[i])),
			Len:  int(buffer.GetInfo().entry_size),
			Cap:  int(buffer.GetInfo().entry_size),
		}))
		//log.Printf("Batch Entry[%d]: %v", i, batch.Entries[i].Data[:])
	}
	//log.Printf("Batch: %s", batch)
	return batch, nil

	//entry := *(*Entry)(unsafe.Pointer(&reflect.SliceHeader{

	//batch := *(*[]*byte)(unsafe.Pointer(&reflect.SliceHeader{
	/*
		batch := *(*[]*Entry)(unsafe.Pointer(&reflect.SliceHeader{
			Data: uintptr(batch_ptr),
			Len:  int(count),
			Cap:  int(count),
		}))
		log.Printf("Len: %d", len(batch))
		log.Printf("Entry size: %d", int(buffer.GetInfo().entry_size))
		log.Printf("Batch: %s", batch[:])
		for i := 0; i < int(count); i++ {
			log.Printf("Batch[%d]: %s", i, batch[i])
			entry := *(*Entry)(unsafe.Pointer(&reflect.SliceHeader{
				Data: uintptr(unsafe.Pointer(&batch[i])),
				Len:  int(buffer.GetInfo().entry_size),
				Cap:  int(buffer.GetInfo().entry_size),
			}))
			//entry.Data[0] = 10
			log.Printf("Entry[%d]: %v", 0, entry.Data[:])

		}
	*/
	/*
		log.Printf("Batch: %s [ % v ]", batch_ptr, batch_ptr)
		log.Printf("Batch: %s [ % v ]", &batch_ptr, &batch_ptr)
		var slice [][]byte
		ptr := uintptr(unsafe.Pointer(batch_ptr)) //unsafe.Pointer(batch_ptr))
		head := (*reflect.SliceHeader)((unsafe.Pointer(&slice)))
		head.Len = int(count)
		head.Cap = int(count)
		head.Data = ptr

		log.Printf("slice: %s [ %s ]", slice, ptr)
		log.Printf("Len: %d", len(slice))
	*/

	//return nil, nil
	/*
		var slice []Entry
		ptr := uintptr(unsafe.Pointer(array))
		//log.Printf("Batch: %s", (*C.char)(*batch_ptr))
		head := (*reflect.SliceHeader)((unsafe.Pointer(&slice)))
		head.Data = ptr
		head.Len = int(count)
		head.Cap = int(count)

		return
	*/
	//slice := *(*[]byte)(unsafe.Pointer(&head))
	//batch.Entries = *(*[]Entry)(unsafe.Pointer(&head))
	//temp := (*[]byte)(batch_ptr)

	//log.Printf("Array: % v", array) //(*temp)[0:10])
	//log.Printf("Slice: % v", slice)
	//log.Printf("Slice: % v", slice[0])

	/*
		for i := 0; i < 1; i++ { //int(count); i++ {
			//data_array := slice[i].Data
			//var data_slice []byte
			//data_ptr := uintptr(unsafe.Pointer(data_array))
			data_ptr := uintptr(unsafe.Pointer(&slice[i].Data))
			data_head := (*reflect.SliceHeader)((unsafe.Pointer(&slice[i].Data)))
			data_head.Data = data_ptr
			size := buffer.GetInfo().data_size
			//size := 8
			data_head.Len = int(size)
			data_head.Cap = int(size)
		}
		log.Printf("Slice: % v", slice[0])
	*/
	/*
		head := reflect.SliceHeader{
			Data: uintptr(*batch_ptr),
			Len:  int(count),
			Cap:  int(count),
		}
		entries := (*(*[]byte)(unsafe.Pointer(&head)))
		one := reflect.SliceHeader{
			Data: uintptr(entries[0]),
			Len:  8,
			Cap:  8,
		}
		log.Printf("Batch[%d]: %s", count, one)
		if batch_ptr != nil {
			return nil, errors.New(fmt.Sprintf("Err: %d", batch_ptr))
		}
	*/

	//batch := &Batch{
	//	Entries: make([]Entry, 10),
	//Entries: *(*[]Entry)(unsafe.Pointer(batch_ptr)),
	//}

	//log.Printf("Batch: %s", batch)
	//return nil, nil
	//return batch, nil
	//batch := make([]Entry, count)
	//DebugPrint("Producer claimed: %d", count)

	//return &Batch{
	//	Entries: make([]Entry, count),
	//}, nil
}

func (buffer *RingBuffer) CancelBatch(batch *Batch) error {
	C.rb_cancel(buffer.buffer_ptr, (*[0]byte)(unsafe.Pointer(batch)), C.uint16_t(len(batch.Entries)))

	return nil
}

func (buffer *RingBuffer) PublishBatch(batch *Batch) error {
	C.rb_publish(buffer.buffer_ptr, (*[0]byte)(unsafe.Pointer(batch)), C.uint16_t(len(batch.Entries)))

	return nil
}

/*
















*/

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
