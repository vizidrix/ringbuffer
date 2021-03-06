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

var _ = Suite(&Given_nothing{})
var _ = Suite(&Given_a_size_64_writer_buffer{})

func (g *Given_a_size_64_writer_buffer) SetUpTest(c *C) {
	g.buffer, _ = NewRingBuffer(64, 6)
}

func (g *Given_a_size_64_writer_buffer) TearDownTest(c *C) {
	g.buffer.Close()
}

func (g *Given_nothing) Test_Should_be_able_to_create_a_ring_buffer(c *C) {
	buffer, _ := NewRingBuffer(65000, 2)
	info := buffer.GetInfo()

	c.Assert(info.GetBufferSize(), Equals, uint64(65536))
	c.Assert(info.GetEntrySize(), Equals, uint64(2))
	buffer.Close()
}

func (g *Given_a_size_64_writer_buffer) Test_Should_return_error_if_claim_batch_called_with_zero(c *C) {
	_, err := g.buffer.Claim(0)

	if err == nil {
		c.Fatalf("Should not allow claim size zero: %s", err)
	}
}

func (g *Given_a_size_64_writer_buffer) Test_Should_not_return_error_if_claim_batch_called_with_one(c *C) {
	_, err := g.buffer.Claim(1)

	if err != nil {
		c.Fatalf("Should allow claim size one for batch mode none: %s", err)
	}
}

func (g *Given_a_size_64_writer_buffer) Test_Should_return_a_batch_instance(c *C) {
	batch, _ := g.buffer.Claim(1)

	if batch == nil {
		c.Fatalf("Should have returned a valid batch")
	}
	c.Assert(batch.BatchNum, Equals, uint64(0))
	c.Assert(batch.SeqNum, Equals, uint64(0))
	c.Assert(batch.BatchSize, Equals, uint64(1))
}

func (g *Given_a_size_64_writer_buffer) Test_Should_be_able_to_write_to_and_verify_single_entry_batch(c *C) {
	batch, _ := g.buffer.Claim(1)
	copy(g.buffer.Entry(batch.SeqNum), GetData(10))
	buffer := g.buffer.Entry(batch.SeqNum)

	c.Assert(buffer[0], Equals, uint8(10))
	c.Assert(buffer[4], Equals, uint8(6))
}

func (g *Given_a_size_64_writer_buffer) Test_Should_not_overlap_multiple_buffers(c *C) {
	batch1, _ := g.buffer.Claim(1)
	batch2, _ := g.buffer.Claim(1)

	copy(g.buffer.Entry(batch1.SeqNum), GetData(10))
	copy(g.buffer.Entry(batch2.SeqNum), GetData(20))

	buffer1 := g.buffer.Entry(batch1.SeqNum)
	buffer2 := g.buffer.Entry(batch2.SeqNum)

	//log.Printf("Buffer overlap 1: %v == %v", buffer1, buffer2)
	for i := 0; i < len(buffer1); i++ {
		if buffer1[i] == buffer2[i] {
			c.Fail()
		}
	}
}

func (g *Given_a_size_64_writer_buffer) Test_Should_not_overlap_multiple_buffers_interleaved(c *C) {
	batch1, _ := g.buffer.Claim(1)
	buffer1 := g.buffer.Entry(batch1.SeqNum)
	copy(g.buffer.Entry(batch1.SeqNum), GetData(10))

	batch2, _ := g.buffer.Claim(1)
	buffer2 := g.buffer.Entry(batch2.SeqNum)
	copy(g.buffer.Entry(batch2.SeqNum), GetData(20))

	//log.Printf("Buffer overlap 2: %v == %v", buffer1, buffer2)
	for i := 0; i < len(buffer1); i++ {
		if buffer1[i] == buffer2[i] {
			c.Fail()
		}
	}
	c.Assert(buffer1[0], Equals, uint8(10))
	c.Assert(buffer2[0], Equals, uint8(20))
	c.Assert(g.buffer.Entry(batch1.SeqNum)[0], Equals, uint8(10))
	c.Assert(g.buffer.Entry(batch2.SeqNum)[0], Equals, uint8(20))
}

func (g *Given_a_size_64_writer_buffer) Test_Should_populate_RingBufferInfo(c *C) {
	info := g.buffer.GetInfo()

	c.Assert(info.GetBufferSize(), Equals, uint64(64))
	c.Assert(info.GetEntrySize(), Equals, uint64(6))
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

// 0.28 ns/op empty
// 32 ns/op with noop just method calls
// Each interop call costs ~30-45ns
func Benchmark_Logic(b *testing.B) {
	//data := make([]byte, 10)
	//var batch *Batch
	//buffer, _ := NewRingBuffer(SINGLE_WRITER, 16, 10)
	//defer buffer.Close()
	log.Printf("Bench: %d", b.N)
	for i := 0; i < b.N; i++ {
		buffer, _ := NewRingBuffer(64, 2)

		batch, _ := buffer.Claim(1)
		//batch.Entry(0).CopyFrom(GetData(10))
		//batch.CopyTo(0, )
		data := buffer.Entry(batch.SeqNum)
		//data := batch.Entry(0)
		if data[0] != uint8(10) {
			b.Fail()
		}
		//batch.Publish()
		buffer.Publish(batch)
		//batch, _ = buffer.Claim(1)
		//batch.Entries[0].CopyFrom(data)
		//batch.Publish()

		buffer.Close()
	}
}

// 64 -> 20 -> 16 -> 15
// 44 / 1
//  4 / 10
//  1 / 100

func Benchmark_ClaimAndPublish_1(b *testing.B) {
	b.StopTimer()
	buffer, err := NewRingBuffer(64, 2)
	if err != nil {
		log.Printf("Err in buffer create: %s", err)
	}
	defer buffer.Close()
	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		buffer.ClaimAndPublish(1)
	}
}

func Benchmark_ClaimAndPublish_10(b *testing.B) {
	b.StopTimer()
	buffer, err := NewRingBuffer(64, 2)
	if err != nil {
		log.Printf("Err in buffer create: %s", err)
	}
	defer buffer.Close()
	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		buffer.ClaimAndPublish(10)
	}
}

func Benchmark_ClaimAndPublish_100(b *testing.B) {
	b.StopTimer()
	buffer, err := NewRingBuffer(64, 2)
	if err != nil {
		log.Printf("Err in buffer create: %s", err)
	}
	defer buffer.Close()
	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		buffer.ClaimAndPublish(100)
	}
}

func Benchmark_ClaimAndPublishB_1(b *testing.B) {
	b.StopTimer()
	buffer, err := NewRingBuffer(64, 2)
	if err != nil {
		log.Printf("Err in buffer create: %s", err)
	}
	defer buffer.Close()
	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		batch, _ := buffer.Claim(1)
		buffer.Publish(batch)
	}
}

func Benchmark_ClaimAndPublishB_10(b *testing.B) {
	b.StopTimer()
	buffer, err := NewRingBuffer(64, 2)
	if err != nil {
		log.Printf("Err in buffer create: %s", err)
	}
	defer buffer.Close()
	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 10; j++ {
			batch, _ := buffer.Claim(1)
			buffer.Publish(batch)
		}
	}
}

func Benchmark_ClaimAndPublishB_100(b *testing.B) {
	b.StopTimer()
	buffer, err := NewRingBuffer(64, 2)
	if err != nil {
		log.Printf("Err in buffer create: %s", err)
	}
	defer buffer.Close()
	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 100; j++ {
			batch, _ := buffer.Claim(1)
			buffer.Publish(batch)
		}
	}
}

func _Benchmark_Claim_and_cancel(b *testing.B) {
	b.StopTimer()
	buffer, err := NewRingBuffer(64, 2)
	if err != nil {
		log.Printf("Err in buffer create: %s", err)
	}
	//data := GetData(2)
	//batch, _ := buffer.Claim(1)
	defer buffer.Close()
	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {

		buffer.ClaimAndPublish(10)

		/*
			batch, _ := buffer.Claim(1) // 344 ns/op - down to 72 ns/op
			buffer.Publish(batch)
		*/

		/*
			batch, err := buffer.Claim(1)
			if err != nil {
				//log.Printf("Err: %s", err)
			} else {
				//entry := batch.Entry(0)
				//batch.Entry(0)
				//data := batch.Entry(0)

				//batch.CopyTo(0, data)
				//batch.CopyTo(0, data)
				//batch.CopyTo(0, data)
				//entry[0] = 99

				//data := batch.Entry(0).GetBuffer()
				//if data[0] != uint8(10) {
				//	b.Fail()
				//}
				batch.Cancel()
			}
		*/
	}
}

func ring_buffer_test_ignore() {
	log.Println(fmt.Sprintf("", 10))
	log.Printf("", reflect.SliceHeader{}, errors.New("stuff"), strings.HasPrefix("s", "q"), unsafe.Pointer(nil), time.Millisecond)
}
