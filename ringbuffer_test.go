package ringbuffer_test

import (
	"errors"
	"fmt"
	. "github.com/vizidrix/ringbuffer"
	. "launchpad.net/gocheck"
	"log"
	"reflect"
	"strings"
	"testing"
	"time"
	"unsafe"
)

func Test(t *testing.T) { TestingT(t) }

type Given_a_single_writer_buffer struct {
	buffer *RingBuffer
}

var _ = Suite(&Given_a_single_writer_buffer{})

func (g *Given_a_single_writer_buffer) SetUpTest(c *C) {
	g.buffer, _ = NewRingBuffer(SINGLE_WRITER, L0, 4)
}

func (g *Given_a_single_writer_buffer) TearDownTest(c *C) {
	g.buffer.Close()
}

func (g *Given_a_single_writer_buffer) Test_When_a_batch_is_claimed(c *C) {
	batch, _ := g.buffer.Claim(1)

	c.Assert(batch.Number, Equals, 1)
	c.Assert(batch.Size, Equals, 1)
}

func (g *Given_a_single_writer_buffer) Test_When_a_single_entry_is_published(c *C) {
	data := make([]byte, 10)
	token, _ := g.buffer.Publish(data)

	select {
	case count := <-token.Published:
		{
			c.Errorf("Batch published: %d", count)
		}
	case err := <-token.Failed:
		{
			c.Errorf("Error in publish token: %s", err)
		}
	case <-time.After(10 * time.Millisecond):
		{
			c.Errorf("Publish timed out")
		}
	}
}

func (g *Given_a_single_writer_buffer) Test_When_a_sequence_is_published(c *C) { // Test_Should_upate_seq_num_when_entry_is_published(c *C) {
	//func Test_Should_upate_seq_num_when_entry_is_published(t *testing.T) {
	data := make([]byte, 10)
	batch, _ := g.buffer.Claim(1)
	batch.Entries[0].CopyFrom(data)
	token, _ := batch.Publish()

	select {
	case count := <-token.Published:
		{
			c.Errorf("Batch published: %d", count)
		}
	case err := <-token.Failed:
		{
			c.Errorf("Error in publish token: %s", err)
		}
	case <-time.After(10 * time.Millisecond):
		{
			c.Errorf("Publish timed out")
		}
	}

	//c.Assert("a", Equals, "b")
}

func (g *Given_a_single_writer_buffer) Test_Should_init_ring_buffer(c *C) {
	//func Test_Should_init_ring_buffer(t *testing.T) {
	//EnableDebug()
	//defer DisableDebug()

	data := make([]byte, 10)
	batch, err := g.buffer.Claim(1)
	if err != nil {
		c.Errorf("Error claiming batch: %s", err)
	}

	err = batch.Entries[0].CopyFrom(data)
	if err != nil {
		c.Errorf("Error copying to entry: %s", err)
	}

	token, err := batch.Publish()
	if err != nil {
		c.Errorf("Error publishing batch: %s", err)
	}

	select {
	case count := <-token.Published:
		{
			c.Errorf("Batch published: %d", count)
		}
	case err := <-token.Failed:
		{
			c.Errorf("Error in publish token: %s", err)
		}
	case <-time.After(10 * time.Millisecond):
		{
			c.Errorf("Publish timed out")
		}
	}

	//c.Fail()
}

// 0.28 ns/op empty
// 32 ns/op with noop just method calls
func Benchmark_Logic(b *testing.B) {
	data := make([]byte, 10)
	var batch *Batch
	buffer, _ := NewRingBuffer(SINGLE_WRITER, 16, 10)
	defer buffer.Close()
	for i := 0; i < b.N; i++ {
		batch, _ = buffer.Claim(1)
		batch.Entries[0].CopyFrom(data)
		batch.Publish()
	}
}

// go test -c
// ./eventstore.test -test.bench=.benchname -test.cpuprofile=cpu.out
// ./eventstore.test -test.v -test.fun xxx -test.cpuprofile=cpu.out -test.memprofile=mem.out -test.bench=.Bench_name
// go tool pprof eventstore.test cpu.out

// go test github.com/vizidrix/eventstore -bench .Trim -benchmem

// go test github.com/vizidrx/ringbuffer -gocheck.v

func ring_buffer_test_ignore() {
	log.Println(fmt.Sprintf("", 10))
	log.Printf("", reflect.SliceHeader{}, errors.New("stuff"), strings.HasPrefix("s", "q"), unsafe.Pointer(nil))
}
