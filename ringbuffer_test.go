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
	// *******

	//temp := make([]byte, 10, 20)
	temp := reflect.SliceHeader{
		Data: uintptr(10),
		Len:  10,
		Cap:  20,
	}
	temp.Data = 30

	log.Printf("temp: %x - %v - %s", temp, temp, temp)
	//c.Fail()
}

func (g *Given_batching_mode_none) Test_Should_be_able_to_write_to_and_verify_single_entry_batch(c *C) {
	batch, _ := g.buffer.Claim(1)
	batch.CopyTo(0, GetData(10))
	//batch.Entry(0).CopyFrom(GetData(10))
	buffer := batch.Entry(0)

	c.Assert(buffer[0], Equals, uint8(10))
	c.Assert(buffer[4], Equals, uint8(6))
}

func (g *Given_batching_mode_none) Test_Should_not_overlap_multiple_buffers(c *C) {
	batch1, _ := g.buffer.Claim(1)
	batch1.CopyTo(0, GetData(10))
	batch2, _ := g.buffer.Claim(1)
	batch2.CopyTo(0, GetData(20))

	buffer1 := batch1.Entry(0)
	buffer2 := batch2.Entry(0)

	for i := 0; i < len(buffer1); i++ {
		if buffer1[i] == buffer2[i] {
			c.Fail()
		}
	}
}

func (g *Given_batching_mode_none) Test_Should_not_overlap_multiple_buffers_interleaved(c *C) {
	batch1, _ := g.buffer.Claim(1)
	buffer1 := batch1.Entry(0)
	batch1.CopyTo(0, GetData(10))

	batch2, _ := g.buffer.Claim(1)
	buffer2 := batch2.Entry(0)
	batch2.CopyTo(0, GetData(20))

	for i := 0; i < len(buffer1); i++ {
		if buffer1[i] == buffer2[i] {
			c.Fail()
		}
	}
	c.Assert(buffer1[0], Equals, uint8(10))
	c.Assert(buffer2[0], Equals, uint8(20))
	c.Assert(batch1.Entry(0)[0], Equals, uint8(10))
	c.Assert(batch2.Entry(0)[0], Equals, uint8(20))
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
	//c.Fail()
}

/*

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
		//batch.Entry(0).CopyFrom(GetData(10))
		//batch.CopyTo(0, )
		data := batch.Entry(0)
		if data[0] != uint8(10) {
			b.Fail()
		}
		batch.Cancel()
		//batch, _ = buffer.Claim(1)
		//batch.Entries[0].CopyFrom(data)
		//batch.Publish()

		buffer.Close()
	}
}

func Benchmark_Claim_and_cancel(b *testing.B) {
	log.Printf("Bench: %d", b.N)
	//buffer, _ := NewRingBuffer(L12, LARGE_BATCH, 2)
	buffer, _ := NewRingBuffer(L10, NONE, 64)
	data := GetData(10)
	//batch, _ := buffer.Claim(1)
	defer buffer.Close()
	for i := 0; i < b.N; i++ {
		//buffer.Claim(1) // 344 ns/op

		batch, _ := buffer.Claim(1)
		entry := batch.Entry(0)
		//batch.Entry(0)
		//data := batch.Entry(0)

		batch.CopyTo(0, data)
		entry[0] = 99

		//data := batch.Entry(0).GetBuffer()
		//if data[0] != uint8(10) {
		//	b.Fail()
		//}
		//batch.Cancel()
	}
}

func ring_buffer_test_ignore() {
	log.Println(fmt.Sprintf("", 10))
	log.Printf("", reflect.SliceHeader{}, errors.New("stuff"), strings.HasPrefix("s", "q"), unsafe.Pointer(nil), time.Millisecond)
}
