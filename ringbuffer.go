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
	L0  BUFFER_TYPES = 0x01 << iota //         1 * 64 	=         64 (iota has been reset)
	L1                              //         8 * 64 	=        512
	L2                              //        16 * 64 	=      1,024
	L3                              //        32 * 64 	=      2,048
	L4                              //        64 * 64 	=      4,096
	L5                              //       8 * 4096 	=     32,768
	L6                              //      16 * 4096 	=     65,536
	L7                              //      32 * 4096 	=    131,072
	L8                              // 	    64 * 4096 	=    262,144
	L9                              //  8 * 64 * 4096 	=  2,097,152
	L10                             // 16 * 64 * 4096 	=  4,194,304
	L11                             // 32 * 64 * 4096 	=  8,388,608
	L12                             // 64 * 64 * 4096 	= 16,777,216
)

type RingBuffer struct {
	buffer_ptr *[0]byte
}

type Batch struct {
	Number  uint64
	Size    byte
	Entries []Entry
}

type PublishToken struct {
	Published chan struct{}
	Failed    chan struct{}
}

type Entry struct {
	Data []byte
}

func NewRingBuffer(buffer_size BUFFER_TYPES, entry_size uint64) (*RingBuffer, error) {
	//DebugPrint("Making ring buffer [ Mode: %d / Size: %d / Blocks: %d ]", writer_mode, buffer_size, entry_size)
	buffer := &RingBuffer{}
	result, err := C.rb_init_buffer(&buffer.buffer_ptr, C.uint8_t(buffer_size), C.uint8_t(entry_size))
	if result != 0 {
		return nil, errors.New(fmt.Sprintf("Error initializing ring buffer [%d]: %s", result, err))
	}
	return buffer, nil
}

func (buffer *RingBuffer) Close() error {
	result, err := C.rb_release_buffer(buffer.buffer_ptr)
	if result != 0 {
		return errors.New(fmt.Sprintf("Error closing ring buffer [%d]: %s", result, err))
	}
	return nil
}

func (buffer *RingBuffer) GetInfo() *RingBufferInfo {
	info := C.rb_get_info(buffer.buffer_ptr)
	return (*RingBufferInfo)(unsafe.Pointer(info))
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

func (buffer *RingBuffer) Claim(count uint8) (*Batch, error) {
	batch := C.rb_claim(buffer.buffer_ptr, C.uint8_t(count))

	log.Printf("Batch: %s", batch)
	//batch := make([]Entry, count)
	//DebugPrint("Producer claimed: %d", count)

	return &Batch{
		Entries: make([]Entry, count),
	}, nil
}

func (batch *Batch) Publish() (*PublishToken, error) {
	//DebugPrint("Published: %d", len(batch))
	// Listen for success event in goroutine

	return &PublishToken{
		Published: make(chan struct{}),
		Failed:    make(chan struct{}),
	}, nil
}

func (entry *Entry) CopyFrom(data []byte) error {
	//DebugPrint("Copied to entry: % x", data)

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
