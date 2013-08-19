package ringbuffer

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"testing"
	"unsafe"
)

func ring_buffer_test_ignore() {
	log.Println(fmt.Sprintf("", 10))
	log.Printf("", reflect.SliceHeader{}, errors.New("stuff"), strings.HasPrefix("s", "q"), unsafe.Pointer(nil))
}

func Test_Should_init_ring_buffer(t *testing.T) {
	NewRingBuffer()

	t.Fail()
}
