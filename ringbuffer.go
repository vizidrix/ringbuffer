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

//export DebugPrintf
func DebugPrintf(format *C.char) {
	log.Printf(C.GoString(format))
}

func ring_buffer_ignore() {
	log.Println(fmt.Sprintf("", 10))
	log.Printf("", reflect.SliceHeader{}, errors.New("stuff"), strings.HasPrefix("s", "q"), unsafe.Pointer(nil))
}

func NewRingBuffer() {
	var temp *C.rb_buffer

	C.rb_init_buffer(&temp, 10, 10)
	C.rb_release_buffer(temp)
	log.Printf("Buffer: % v", temp)
}

//#cgo CFLAGS: -I. -fPIC
//#cgo LDFLAGS: -lstdc++ -w -hostobj -L. ringbuffer.c

//#include "ringbuffer.h"
//#include <stdlib.h>

//#cgo CFLAGS: -I/go/vizidrix/src/github.com/vizidrix/ringbuffer/
//#cgo LDFLAGS: /go/vizidrix/src/github.com/vizidrix/ringbuffer/ringbuffer.c

//#cgo LDFLAGS: -lringbuffer.h
//#include <ringbuffer.h>
//#cgo LDFLAGS: -lccv
//#cgo pkg-config: ringbufer
//#cgo LDFLAGS: -L/ringbuffer.cgo2.o
