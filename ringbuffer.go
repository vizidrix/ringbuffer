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
	//log.Printf("Creating ring buffer")
	_, err := C.rb_init_buffer(&buffer.buffer_ptr, C.uint64_t(batch_size), C.uint64_t(buffer_size), C.uint64_t(data_size))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error initializing ring buffer [%d]", err))
	}
	//log.Printf("Created ring buffer")
	return buffer, nil
}

func (buffer *RingBuffer) Close() error {
	log.Printf("Closing: %s", buffer)
	_, err := C.rb_free_buffer(&buffer.buffer_ptr)
	log.Printf("Freed")
	if err != nil {
		return errors.New(fmt.Sprintf("Error closing ring buffer [%d]", err))
	}
	log.Printf("Return from Close")
	return nil
}

func Temp(buffer *RingBuffer, count uint64) {
	C.temp(buffer.buffer_ptr, C.uint64_t(count))
}

/*
func (buffer *RingBuffer) Claim(count uint16) (batch *Batch, cancelToken *struct{}, err error) {
	cancelToken = &struct{}{}
	c_batch, err := C.rb_claim(buffer.buffer_ptr, C.uint16_t(count), unsafe.Pointer(cancelToken))
	batch = (*Batch)(unsafe.Pointer(c_batch))
	return
}
*/

func (buffer *RingBuffer) Claim(count uint16) *ClaimResult {
	result := NewClaimResult()
	cancelToken := &struct{}{}
	go func(token *struct{}) {
		<-result.CancelChan
		token = nil // Cancel token
		log.Printf("Finished cancel goroutine")
	}(cancelToken)
	log.Printf("Cancel goroutine launched")
	go func(token *struct{}) {
		c_batch, err := C.rb_claim(buffer.buffer_ptr, C.uint16_t(count), unsafe.Pointer(token))
		log.Printf("Claimed: %s", c_batch)
		if err != nil {
			log.Printf("Got err")
			result.ErrorChan <- err
		} else {
			log.Printf("No err")
			batch_ptr := (*Batch)(unsafe.Pointer(c_batch))
			// TODO: Handle potential race where result chan is closed when result comes back
			log.Printf("Converted batch")
			result.ResultChan <- batch_ptr
			log.Printf("Return after converting")
		}
		log.Printf("Finished value goroutine")
	}(cancelToken)
	log.Printf("Value goroutine launched")
	return result
}

/*
func (buffer *RingBuffer) Claim(count uint16) *ClaimResult {
	result := &ClaimResult{
		ResultChan: make(chan *Batch),
	}
	// Setup cancel for C func
	var cancelToken *struct{} = &struct{}{}
	go func() {
		defer result.Tomb.Done()
		batch, err := C.rb_claim(buffer.buffer_ptr, C.uint16_t(count), unsafe.Pointer(cancelToken))

		if err != nil {
			result.Tomb.Kill(err)
			return
		}

		result.ResultChan <- (*Batch)(unsafe.Pointer(batch))
	}()
	go func() {
		select {
		case <-result.Tomb.Dying():
			{
				// Flipping this to a non-zero will exit the C function
				if cancelToken != nil {
					cancelToken = nil
				}
			}
		}

	}()
	return result
}
*/

func (buffer *RingBuffer) Batch(batch_num uint64) *Batch {
	return (*Batch)(unsafe.Pointer(C.rb_get_batch(buffer.buffer_ptr, C.uint64_t(batch_num))))
}

func (buffer *RingBuffer) Entry(seq_num uint64) []byte {
	return *(*[]byte)(unsafe.Pointer((*reflect.SliceHeader)(C.rb_get_entry_slice(buffer.buffer_ptr, C.uint64_t(seq_num)))))
}

func (buffer *RingBuffer) Publish(batch *Batch) error {
	_, err := C.rb_publish(buffer.buffer_ptr, (*C.rb_batch)(unsafe.Pointer(batch)))
	return err
}

func (buffer *RingBuffer) Release(batch *Batch) error {
	_, err := C.rb_release(buffer.buffer_ptr, (*C.rb_batch)(unsafe.Pointer(batch)))
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

func (buffer *RingBuffer) BatchStateString() string {
	result := ""
	for i := 0; i < int(buffer.GetInfo().GetBatchBufferSize()); i++ {
		batch := buffer.Batch(uint64(i))
		key := batch.GetState().String()[0:1]
		result += key + "_"
	}
	return result
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
