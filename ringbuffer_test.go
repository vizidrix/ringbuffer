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

func GetData(size uint16) []byte {
	isize := int(size)
	data := make([]byte, isize)
	for i := 0; i < isize; i++ {
		data[i] = byte((isize - i) & 0xFF)
	}
	return data
}

func Test(t *testing.T) { TestingT(t) }

type Given_nothing struct{}

type Given_a_size_64_writer_buffer struct {
	buffer *RingBuffer
}

type Given_batching_mode_none struct {
	buffer *RingBuffer
}

var _ = Suite(&Given_nothing{})
var _ = Suite(&Given_batching_mode_none{})
var _ = Suite(&Given_a_size_64_writer_buffer{})

func (g *Given_batching_mode_none) SetUpTest(c *C) {
	g.buffer, _ = NewRingBuffer(L0, NONE, 6)
}

func (g *Given_batching_mode_none) TearDownTest(c *C) {
	g.buffer.Close()
}

func (g *Given_a_size_64_writer_buffer) SetUpTest(c *C) {
	g.buffer, _ = NewRingBuffer(L0, LARGE_BATCH, 6)
}

func (g *Given_a_size_64_writer_buffer) TearDownTest(c *C) {
	g.buffer.Close()
}

func (g *Given_nothing) Test_Should_be_able_to_create_a_ring_buffer(c *C) {
	buffer, _ := NewRingBuffer(L6, LARGE_BATCH, 2)
	info := buffer.GetInfo()

	c.Assert(info.GetBufferType(), Equals, L6)
	c.Assert(info.GetBufferSize(), Equals, uint64(65536))
	c.Assert(info.GetDataSize(), Equals, uint64(2))
	c.Assert(info.GetEntrySize(), Equals, uint64(8))
	buffer.Close()
}

func (g *Given_batching_mode_none) Test_Should_return_error_if_claim_batch_called_with_multiple(c *C) {
	_, err := g.buffer.Claim(2)

	if err == nil {
		c.Fatalf("Batch mode none shouldn't allow multiple claim: %s", err)
	}
}

func (g *Given_batching_mode_none) Test_Should_return_error_if_claim_batch_called_with_zero(c *C) {
	_, err := g.buffer.Claim(0)

	if err == nil {
		c.Fatalf("Should not allow claim size zero: %s", err)
	}
}

func (g *Given_batching_mode_none) Test_Should_not_return_error_if_claim_batch_called_with_one(c *C) {
	_, err := g.buffer.Claim(1)

	if err != nil {
		c.Fatalf("Should allow claim size one for batch mode none: %s", err)
	}
}

func (g *Given_batching_mode_none) Test_Should_return_a_batch_instance(c *C) {
	batch, _ := g.buffer.Claim(1)

	if batch == nil {
		c.Fatalf("Should have returned a valid batch")
	}
	c.Assert(batch.GetBatchNum(), Equals, uint64(0))
	c.Assert(batch.GetSeqNum(), Equals, uint64(0))
	c.Assert(batch.GetBatchSize(), Equals, uint16(1))
}

func (g *Given_batching_mode_none) Test_Should_be_able_to_write_to_and_verify_single_entry_batch(c *C) {
	batch, _ := g.buffer.Claim(1)
	batch.Entry(0).CopyFrom(GetData(10))
	buffer := batch.Entry(0).GetBuffer()

	c.Assert(buffer[0], Equals, uint8(10))
	c.Assert(buffer[4], Equals, uint8(6))
}

func (g *Given_batching_mode_none) Test_Should_not_overlap_multiple_buffers(c *C) {
	batch1, _ := g.buffer.Claim(1)
	batch1.Entry(0).CopyFrom(GetData(10))
	batch2, _ := g.buffer.Claim(1)
	batch2.Entry(0).CopyFrom(GetData(20))

	buffer1 := batch1.Entry(0).GetBuffer()
	buffer2 := batch2.Entry(0).GetBuffer()

	for i := 0; i < len(buffer1); i++ {
		if buffer1[i] == buffer2[i] {
			c.Fail()
		}
	}
}

func (g *Given_a_size_64_writer_buffer) Test_Should_populate_RingBufferInfo(c *C) {
	info := g.buffer.GetInfo()

	c.Assert(info.GetBufferType(), Equals, L0)
	c.Assert(info.GetBufferSize(), Equals, uint64(64))
	c.Assert(info.GetDataSize(), Equals, uint64(6))
	c.Assert(info.GetEntrySize(), Equals, uint64(12))
}

func (g *Given_a_size_64_writer_buffer) Test_Should_not_allow_allocation_beyond_max_size(c *C) {
	_, err := g.buffer.Claim(255)

	if err == nil {
		c.Fatalf("Claim requested is larger than buffer")
	}
}

func (g *Given_a_size_64_writer_buffer) Test_Should_return_error_if_full(c *C) {
	g.buffer.Claim(63) // Claim nearly all the slots

	_, err := g.buffer.Claim(2)

	if err == nil {
		c.Fatalf("Buffer should have been full if claim is larger than remaining: %s", err)
	}
}

func (g *Given_a_size_64_writer_buffer) Test_Should_return_error_if_buffer_is_too_full_on_last_slot(c *C) {
	g.buffer.Claim(64) // Claim all the slots

	_, err := g.buffer.Claim(1)

	if err == nil {
		c.Fatalf("Buffer should have been full if no slots are available: %s", err)
	}
}

func (g *Given_a_size_64_writer_buffer) Test_Should_increment_batch_num(c *C) {
	c.Assert(g.buffer.GetStats().GetBatchNum(), Equals, uint64(0))
	g.buffer.Claim(1)

	c.Assert(g.buffer.GetStats().GetBatchNum(), Equals, uint64(1))
}

func (g *Given_a_size_64_writer_buffer) Test_Should_return_seq_num(c *C) {
	batch, _ := g.buffer.Claim(1)

	if batch == nil {
		c.Fatalf("Should have returned a valid batch")
	}
	//c.Assert(seq_num, Equals, 0)
}

/*
func (g *Given_a_size_64_writer_buffer) Test_Should_match_position_and_info_when_a_batch_is_claimed(c *C) {
	//EnableDebug()
	//defer DisableDebug()

	batch1, _ := g.buffer.Claim(1)
	batch2, _ := g.buffer.Claim(2)
	batch3, _ := g.buffer.Claim(3)

	//log.Printf("Batch1: %s [ % v ]", batch1, batch1)
	//log.Printf("Batch2: %s [ % v ]", batch2, batch2)
	//log.Printf("Batch3: %s [ % v ]", batch3, batch3)

	c.Assert(batch1.GetBatchNum(), Equals, uint64(1))
	c.Assert(batch2.GetBatchNum(), Equals, uint64(2))
	c.Assert(batch3.GetBatchNum(), Equals, uint64(3))

	c.Assert(batch1.GetBatchSize(), Equals, uint16(1))
	c.Assert(batch2.GetBatchSize(), Equals, uint16(2))
	c.Assert(batch3.GetBatchSize(), Equals, uint16(3))
}
*/

/*
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
*/

// 0.28 ns/op empty
// 32 ns/op with noop just method calls
func Benchmark_Logic(b *testing.B) {
	//data := make([]byte, 10)
	//var batch *Batch
	//buffer, _ := NewRingBuffer(SINGLE_WRITER, 16, 10)
	//defer buffer.Close()
	log.Printf("Bench: %d", b.N)
	for i := 0; i < b.N; i++ {
		buffer, _ := NewRingBuffer(L0, LARGE_BATCH, 2)

		batch, _ := buffer.Claim(1)
		batch.Cancel()
		//batch, _ = buffer.Claim(1)
		//batch.Entries[0].CopyFrom(data)
		//batch.Publish()

		buffer.Close()
	}
}

func Benchmark_Claim_and_cancel(b *testing.B) {
	log.Printf("Bench: %d", b.N)
	buffer, _ := NewRingBuffer(L12, LARGE_BATCH, 2)
	defer buffer.Close()
	for i := 0; i < b.N; i++ {
		batch, _ := buffer.Claim(1)
		batch.Cancel()
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
	log.Printf("", reflect.SliceHeader{}, errors.New("stuff"), strings.HasPrefix("s", "q"), unsafe.Pointer(nil), time.Millisecond)
}
