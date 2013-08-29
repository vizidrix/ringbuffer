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

type RingBuffer struct {
	buffer_ptr *[0]byte
}

type PublishToken struct {
	Published chan struct{}
	Failed    chan struct{}
}

func NewRingBuffer(batch_size uint64, buffer_size uint64, data_size uint64) (*RingBuffer, error) {
	buffer := &RingBuffer{}
	_, err := C.rb_init_buffer(&buffer.buffer_ptr, C.uint64_t(batch_size), C.uint64_t(buffer_size), C.uint64_t(data_size))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error initializing ring buffer [%d]", err))
	}
	return buffer, nil
}

func (buffer *RingBuffer) Close() error {
	_, err := C.rb_free_buffer(&buffer.buffer_ptr)
	if err != nil {
		return errors.New(fmt.Sprintf("Error closing ring buffer [%d]", err))
	}
	return nil
}

func (buffer *RingBuffer) Claim(count uint16) *ClaimResult {
	result := &ClaimResult{
		ResultChan: make(chan *Batch),
	}
	// Setup cancel for C func
	var cancelToken byte = 0
	go func() {
		select {
		case <-result.Tomb.Dying():
			{
			}
		}
		// Flipping this to a non-zero will exit the C function
		cancelToken = 1
	}()

	go func() {
		defer result.Tomb.Done()
		batch, err := C.rb_claim(buffer.buffer_ptr, C.uint16_t(count), unsafe.Pointer(&cancelToken))

		if err != nil {
			result.Tomb.Kill(err)
		}

		result.ResultChan <- (*Batch)(unsafe.Pointer(batch))
	}()
	return result
}

func (buffer *RingBuffer) Entry(seq_num uint64) []byte {
	return *(*[]byte)(unsafe.Pointer((*reflect.SliceHeader)(C.rb_get_entry_slice(buffer.buffer_ptr, C.uint64_t(seq_num)))))
}

func (buffer *RingBuffer) Publish(batch *Batch) error {
	//_, err := C.rb_publish(buffer.buffer_ptr, (*C.rb_batch)(unsafe.Pointer(batch)))
	_, err := C.rb_publish((*C.rb_batch)(unsafe.Pointer(batch)))
	return err
}

/*
func (buffer *RingBuffer) ClaimAndPublish(count int) {
	C.rb_claim_and_publish(buffer.buffer_ptr, C.int(count)) //, C.uint16_t(count))
}
*/

//

//

//

//

func (buffer *RingBuffer) GetInfo() *Info {
	info_ptr := C.rb_get_info(buffer.buffer_ptr)
	return (*Info)(unsafe.Pointer(info_ptr))
}

func (buffer *RingBuffer) GetStats() *Stats {
	stats_ptr := C.rb_get_stats(buffer.buffer_ptr)
	return (*Stats)(unsafe.Pointer(stats_ptr))
}

//

//

//

//

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
