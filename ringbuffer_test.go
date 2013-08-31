package ringbuffer_test

import (
	"errors"
	"fmt"
	. "github.com/vizidrix/ringbuffer"
	. "launchpad.net/gocheck"
	//. "launchpad.net/tomb"
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

type Given_a_size_4_buffer struct {
	buffer *RingBuffer
}

var _ = Suite(&Given_nothing{})
var _ = Suite(&Given_a_size_4_buffer{})

func (g *Given_a_size_4_buffer) SetUpTest(c *C) {
	g.buffer, _ = NewRingBuffer(4, 4, 4)
}

func (g *Given_a_size_4_buffer) TearDownTest(c *C) {
	g.buffer.Close()
}

//Causes error - Don't delete this, needs to be fixed!!
func (g *Given_nothing) Test_Should_not_segfault_on_claimasync_that_is_not_waited_for(c *C) {
	buffer, _ := NewRingBuffer(1, 1, 1)
	defer buffer.Close()

	//buffer.ClaimAsync(1, 1*time.Millisecond)
}

func (g *Given_nothing) Test_Should_be_able_to_create_a_ring_buffer(c *C) {
	buffer, _ := NewRingBuffer(60, 65000, 2)
	defer buffer.Close()
	info := buffer.GetInfo()

	c.Assert(info.GetBatchBufferSize(), Equals, uint64(64))
	c.Assert(info.GetDataBufferSize(), Equals, uint64(65536))
	c.Assert(info.GetEntrySize(), Equals, uint64(2))
}

func (g *Given_nothing) Test_Should_not_error_if_closed_twice(c *C) {
	buffer, _ := NewRingBuffer(1, 1, 1)
	buffer.Close()
	buffer.Close()
}

func (g *Given_a_size_4_buffer) Test_Should_return_error_if_claim_batch_called_with_zero(c *C) {
	_, err := g.buffer.Claim(0, 1*time.Millisecond)

	if err == nil {
		c.Fatalf("Should not allow claim size zero: %s", err)
	}
}

func (g *Given_a_size_4_buffer) Test_Should_return_error_if_claim_batch_called_with_size_bigger_than_available(c *C) {
	_, err := g.buffer.Claim(5, 1*time.Millisecond)

	if err == nil {
		c.Fatalf("Should not allow claim size zero: %s", err)
	}
}

func (g *Given_a_size_4_buffer) Test_Should_claim_a_single_batch(c *C) {
	batch, err := g.buffer.Claim(1, 1*time.Millisecond)
	if err != nil {
		c.Fatalf("Error during claim: %s", err)
	}
	c.Assert(batch.GetBatchNum(), Equals, uint64(0))
	c.Assert(batch.GetBatchSize(), Equals, uint64(1))
	c.Assert(batch.GetSeqNum(), Equals, uint64(0xFFFFFFFF)) // Unfulfilled flag
	c.Assert(batch.GetGroupFlags(), Equals, uint64(0x00000000))
}

func (g *Given_a_size_4_buffer) Test_Should_claim_two_batches(c *C) {
	batch1, _ := g.buffer.Claim(1, 1*time.Millisecond)
	//log.Printf("Batch 1: %s", batch1)
	batch2, _ := g.buffer.Claim(1, 1*time.Millisecond)
	//log.Printf("Batch 2: %s", batch2)

	c.Assert(batch1.GetBatchNum(), Equals, uint64(0))
	c.Assert(batch1.GetBatchSize(), Equals, uint64(1))
	c.Assert(batch1.GetSeqNum(), Equals, uint64(0xFFFFFFFF)) // Unfulfilled flag
	c.Assert(batch1.GetGroupFlags(), Equals, uint64(0x00000000))
	//log.Printf("Asserted Batch 1")
	c.Assert(batch2.GetBatchNum(), Equals, uint64(1))
	c.Assert(batch2.GetBatchSize(), Equals, uint64(1))
	c.Assert(batch2.GetSeqNum(), Equals, uint64(0xFFFFFFFF)) // Unfulfilled flag
	c.Assert(batch2.GetGroupFlags(), Equals, uint64(0x00000000))
	//log.Printf("Asserted Batch 2")
}

func (g *Given_a_size_4_buffer) Test_Should_block_until_cancel_if_full(c *C) {
	g.buffer.Claim(1, 1*time.Millisecond)
	g.buffer.Claim(1, 1*time.Millisecond)
	g.buffer.Claim(1, 1*time.Millisecond)
	g.buffer.Claim(1, 1*time.Millisecond)

	claimResult := g.buffer.ClaimAsync(1, 1*time.Millisecond)

	select {
	case <-claimResult.ResultChan:
		{
			c.Fail()
		}
	case <-claimResult.CancelChan:
		{
			c.Fatalf("Should have timed out")
		}
	case <-claimResult.ErrorChan:
		{
			c.Fail()
		}
	case <-time.After(1 * time.Microsecond):
		{
			// Sufficient delay to assume it's hung
			close(claimResult.CancelChan)
		}
	}
}

func (g *Given_a_size_4_buffer) Test_Should_not_allow_release_that_overlaps_the_read_seq_num(c *C) {
	g.buffer.Claim(1, 1*time.Millisecond)
	batch, _ := g.buffer.Claim(1, 1*time.Millisecond)
	g.buffer.Publish(batch)
	err := g.buffer.Release(batch)

	if err == nil {
		c.Fatalf("Release should have failed")
	}
}

func (g *Given_a_size_4_buffer) Test_Should_return_error_if_releasing_batch_that_is_not_yet_published(c *C) {
	batch1, _ := g.buffer.Claim(1, 1*time.Millisecond)
	err := g.buffer.Release(batch1)
	if err == nil {
		c.Fail()
	}
}

func (g *Given_a_size_4_buffer) Test_Should_return_error_if_releasing_batch_that_is_already_released(c *C) {
	batch1, _ := g.buffer.Claim(1, 1*time.Millisecond)
	g.buffer.Publish(batch1)
	g.buffer.Release(batch1)
	err := g.buffer.Release(batch1)
	if err == nil {
		c.Fail()
	}
}

func (g *Given_a_size_4_buffer) Test_Should_not_block_if_previous_batches_were_released(c *C) {
	//log.Printf("State: %s", g.buffer.BatchStateString())
	batch1, _ := g.buffer.Claim(1, 1*time.Millisecond)
	//log.Printf("State: %s", g.buffer.BatchStateString())
	batch2, _ := g.buffer.Claim(1, 1*time.Millisecond)
	//log.Printf("State: %s", g.buffer.BatchStateString())
	batch3, _ := g.buffer.Claim(1, 1*time.Millisecond)
	//log.Printf("State: %s", g.buffer.BatchStateString())
	batch4, _ := g.buffer.Claim(1, 1*time.Millisecond)

	//log.Printf("State: %s", g.buffer.BatchStateString())
	//log.Printf("Buffer State: %s", g.buffer.GetStats())

	//g.buffer.Publish(batch4)

	//log.Printf("State: %s", g.buffer.BatchStateString())
	//log.Printf("Buffer State: %s", g.buffer.GetStats())

	g.buffer.Publish(batch2)

	//log.Printf("State: %s", g.buffer.BatchStateString())
	//log.Printf("Buffer State: %s", g.buffer.GetStats())

	g.buffer.Publish(batch3)

	//log.Printf("State: %s", g.buffer.BatchStateString())
	//log.Printf("Buffer State: %s", g.buffer.GetStats())

	g.buffer.Publish(batch1)

	//log.Printf("State: %s", g.buffer.BatchStateString())
	//log.Printf("Buffer State: %s", g.buffer.GetStats())
	err := g.buffer.Release(batch1)
	if err != nil {
		c.Fatalf("Failed to release batch: %s", err)
	}

	//log.Printf("State: %s", g.buffer.BatchStateString())
	//log.Printf("Buffer State: %s", g.buffer.GetStats())
	err = g.buffer.Release(batch2)
	if err != nil {
		c.Fatalf("Failed to release batch: %s", err)
	}

	//log.Printf("State: %s", g.buffer.BatchStateString())
	//log.Printf("Buffer State: %s", g.buffer.GetStats())
	err = g.buffer.Release(batch3)
	if err != nil {
		c.Fatalf("Failed to release batch: %s", err)
	}

	g.buffer.Publish(batch4)

	//log.Printf("State: %s", g.buffer.BatchStateString())
	//log.Printf("Buffer State: %s", g.buffer.GetStats())

	claimResult := g.buffer.ClaimAsync(1, 1*time.Millisecond)

	select {
	case <-claimResult.ResultChan:
		{
			//log.Printf("Result: %s", batch)
		}
	case <-claimResult.CancelChan:
		{
			c.Fail()
		}
	case <-claimResult.ErrorChan:
		{
			c.Fail()
		}
	case <-time.After(10 * time.Millisecond):
		{
			// Sufficient delay to assume it's hung
			c.Fatalf("Should not have timed out")
		}
	}
	//c.Fail()
}

func (g *Given_a_size_4_buffer) aTest_Should_not_hang(c *C) {
	b := struct {
		N int
	}{
		N: 10,
	}
	var count uint64 = 10
	log.Printf("Bench: %d\t\tCount: %d", b.N, count)
	//buffer, _ := NewRingBuffer(100000, 8, 8)

	//b.ResetTimer()

	//for count := 0; count < 10; count++ {
	//log.Printf(buffer.BatchStateString())
	chan1 := make(chan struct{})
	chan2 := make(chan struct{})
	go func() {
		for i := 0; i < b.N; i++ {
			Temp(g.buffer, count)
		}
		close(chan1)
	}()
	go func() {
		for i := 0; i < b.N; i++ {
			Temp(g.buffer, count)
		}
		close(chan2)
	}()
	//}
	<-chan1
	<-chan2
	//buffer.Close()

	//c.Fail()
}

//

//

//

//

//

//

//

//

//

/*
func (g *Given_a_size_4_buffer) Test_Should_temp(c *C) {
	//log.Printf("Stats: %s", g.buffer.GetStats())
	//g.buffer.Temp(4)

	//	log.Printf("Stats: %s", g.buffer.GetStats())
	//c.Fail()
}
*/

//

//

//

//
/*


func (g *Given_a_size_64_writer_buffer) Test_Should_not_return_error_if_claim_batch_called_with_one(c *C) {
	_, err := g.buffer.Claim(1)

	if err != nil {
		c.Fatalf("Should allow claim size one: %s", err)
	}
}

func (g *Given_a_size_64_writer_buffer) Test_Should_return_a_batch_instance(c *C) {
	batch, err := g.buffer.Claim(1)

	if err != nil {
		c.Fatalf("Error: %d", err)
	}
	if batch == nil {
		c.Fatalf("Should have returned a valid batch")
	}
	c.Assert(batch.BatchNum, Equals, uint64(0))
	c.Assert(batch.SeqNum, Equals, uint64(0))
	c.Assert(batch.BatchSize, Equals, uint64(1))
}

func (g *Given_a_size_64_writer_buffer) Test_Should_be_able_to_write_to_and_verify_single_entry_batch(c *C) {
	batch, err := g.buffer.Claim(1)
	if err != nil {
		c.Fatalf("Failed to claim batch: %d", err)
	}
	copy(g.buffer.Entry(batch.SeqNum), GetData(10))
	buffer := g.buffer.Entry(batch.SeqNum)

	c.Assert(buffer[0], Equals, uint8(10))
	c.Assert(buffer[4], Equals, uint8(6))
}


func (g *Given_a_size_64_writer_buffer) Test_Should_not_overlap_multiple_buffers(c *C) {
	batch1, err := g.buffer.Claim(1)
	if err != nil {
		c.Fatalf("Failed to claim batch 1: %d", err)
	}
	batch2, err := g.buffer.Claim(1)
	if err != nil {
		c.Fatalf("Failed to claim batch 2: %d", err)
	}

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
	batch1, err := g.buffer.Claim(1)
	if err != nil {
		c.Fatalf("Failed to claim batch 1: %d", err)
	}
	log.Printf("Got 1: %s - %d", batch1, batch1.SeqNum)
	buffer1 := g.buffer.Entry(batch1.SeqNum)
	copy(g.buffer.Entry(batch1.SeqNum), GetData(10))
	log.Printf("buffer1: %v", buffer1)

	batch2, err := g.buffer.Claim(1)
	if err != nil {
		c.Fatalf("Failed to claim batch 2: %d", err)
	}
	log.Printf("Got 2: %s - %d", batch2, batch2.SeqNum)
	buffer2 := g.buffer.Entry(batch2.SeqNum)
	copy(g.buffer.Entry(batch2.SeqNum), GetData(20))
	log.Printf("buffer2: %v", buffer2)

	log.Printf("Buffer overlap 2: %v == %v", buffer1, buffer2)
	var i int
	for i = 0; i < len(buffer1); i++ {
		if int(buffer1[i]) != 10+i {
			c.Fatalf("buffer1: %v", buffer1)
		}
		if int(buffer2[i]) != 20+i {
			c.Fatalf("buffer2: %v", buffer2)
		}
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

	c.Assert(info.GetBatchBufferSize(), Equals, uint64(64))
	c.Assert(info.GetDataBufferSize(), Equals, uint64(64))
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
	log.Printf("Batch num before: %d", g.buffer.GetStats().GetBatchNum())
	_, err := g.buffer.Claim(1)
	if err != nil {
		c.Fatalf("Failed to claim batch: %d", err)
	}
	log.Printf("Batch num after: %d", g.buffer.GetStats().GetBatchNum())
	c.Assert(g.buffer.GetStats().GetBatchNum(), Equals, uint64(1))
}

func (g *Given_a_size_64_writer_buffer) Test_Should_return_seq_num(c *C) {
	batch, _ := g.buffer.Claim(1)

	if batch == nil {
		c.Fatalf("Should have returned a valid batch")
	}
	//c.Fail()
}
*/
/*
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
		buffer, _ := NewRingBuffer(64, 64, 2)

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
	buffer, err := NewRingBuffer(64, 64, 2)
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
	buffer, err := NewRingBuffer(64, 64, 2)
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
	buffer, err := NewRingBuffer(64, 64, 2)
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
	buffer, err := NewRingBuffer(64, 64, 2)
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
	buffer, err := NewRingBuffer(64, 64, 2)
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
	buffer, err := NewRingBuffer(64, 64, 2)
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
	buffer, err := NewRingBuffer(64, 64, 2)
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

			//batch, _ := buffer.Claim(1) // 344 ns/op - down to 72 ns/op
			//buffer.Publish(batch)

			//batch, err := buffer.Claim(1)
			//if err != nil {
				//log.Printf("Err: %s", err)
			//} else {
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
			//	batch.Cancel()
			//}

	}
}
*/

func Run_Temp(b *testing.B, c uint64) {
	log.Printf("Bench: %d\t\tCount: %d", b.N, c)
	buffer, _ := NewRingBuffer(10, 8, 8)

	b.ResetTimer()

	//for count := 0; count < 10; count++ {
	//log.Printf(buffer.BatchStateString())
	chan1 := make(chan struct{})
	chan2 := make(chan struct{})
	go func() {
		for i := 0; i < b.N; i++ {
			log.Printf("C1: %d", i)
			Temp(buffer, c)
		}
		close(chan1)
	}()
	go func() {
		for i := 0; i < b.N; i++ {
			log.Printf("C2: %d", i)
			Temp(buffer, c)
		}
		close(chan2)
	}()
	//}
	<-chan1
	<-chan2
	buffer.Close()
}

func Benchmark_Claim_0001(b *testing.B) {
	Run_Temp(b, 1)
}

func Benchmark_Claim_0010(b *testing.B) {
	Run_Temp(b, 10)
}

func Benchmark_Claim_0100(b *testing.B) {
	Run_Temp(b, 100)
}

func Benchmark_Claim_1000(b *testing.B) {
	Run_Temp(b, 1000)
}

func aBenchmark_Claim(b *testing.B) {
	log.Printf("Bench: %d", b.N)
	buffer, _ := NewRingBuffer(uint64(b.N), 8, 8)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buffer.Claim(1, 1*time.Millisecond)
	}
	buffer.Close()
}

func BTemp(b *testing.B) {
	//buffer, _ := NewRingBuffer(1>>32, 4, 4)
	//log.Printf("%s", buffer.GetInfo())
	log.Printf("Bench: %d", b.N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		buffer, _ := NewRingBuffer(1000, 4, 4)
		b.StartTimer()
		//buffer.Claim(1)
		//buffer.Claim(1)
		//buffer.Claim(1)
		batch, err := buffer.Claim(1, 10*time.Millisecond)
		if err != nil {
			b.Fatalf("Failed: %s", err)
		}
		if batch == nil {
			b.Fatalf("Batch was nil")
		}

		//buffer.Temp(1000)
		buffer.Close()
		//log.Printf("Starting claim")
		//claimResult, _ := buffer.Claim(1).Wait(1 * time.Microsecond)

		//buffer.Claim(1).Wait(1 * time.Microsecond)

		//claimResult.Tomb.Kill(errors.New("Canceled"))
		//log.Printf("Staring first wait")
		/*select {
		case <-claimResult.Tomb.Dying():
			{
				b.Fatalf("Should have timed out waiting for batch slot")
			}
		case <-time.After(1 * time.Microsecond):
			{
				//log.Printf("Timed out on first wait")
				//sc.Fatalf("Should have died prior to this timer")
			}
		}
		claimResult.Tomb.Kill(errors.New("Success"))
		select {
		case <-claimResult.Tomb.Dying():
			{
				//log.Printf("Died on second wait: %s", claimResult.Tomb.Err())
				//c.Fatalf("Should have timed out")
			}
		case <-time.After(1 * time.Microsecond):
			{
				b.Fatalf("Should have died prior to this timer")
			}
		}*/
	}
}

func ring_buffer_test_ignore() {
	log.Println(fmt.Sprintf("", 10))
	log.Printf("", reflect.SliceHeader{}, errors.New("stuff"), strings.HasPrefix("s", "q"), unsafe.Pointer(nil), time.Millisecond)
}
