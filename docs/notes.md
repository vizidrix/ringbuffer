// go test -c
// ./eventstore.test -test.bench=.benchname -test.cpuprofile=cpu.out
// ./eventstore.test -test.v -test.fun xxx -test.cpuprofile=cpu.out -test.memprofile=mem.out -test.bench=.Bench_name
// go tool pprof eventstore.test cpu.out

// go test github.com/vizidrix/eventstore -bench .Trim -benchmem

// go test github.com/vizidrx/ringbuffer -gocheck.v


// C.GoBytes(unsafe.Pointer, C.int) []byte



read seq can't pass write barrier
	write seq can't pass read barrier + size

	uint64_t target_seq_num = buffer->write_seq_num + count;
	uint64_t target_position = target_seq_num & buffer->size_mask;
	uint64_t current_read_barrier = buffer->read_barrier + buffer->size_mask;
	uint64_t current_read_position = buffer->read_barrier & buffer->size_mask;
	//DebugPrint("T Seq: %d > %d && T Pos: %d > %d ?", target_seq_num, current_barrier, target_position, current_position);
	
	uint16_t backoff = 0;
	while(target_seq_num > current_read_barrier && target_position > current_read_position) {
		
		if((backoff & 0xFF) == 0x00) {
			if((backoff / 0xFF) > 0xFF) {
				//DebugPrint("Backoff: %d", ((backoff / 0xFF)+1));
				backoff = 0;
			}
			nanosleep((struct timespec[]){{0, MICROSECOND * ((backoff / 0xFF)+1)}}, NULL);
		}
		backoff++;
	}

int i = 0;
	int j = 0;
	for(i = 0; i < count; i++) {
		int index = buffer->info->entry_size * (buffer->write_seq_num + i);
		batch[i] = &buffer->data_buffer[index];
		
		batch[i][0] = (buffer->batch_num >> 24) & 0xFF;
		batch[i][1] = (buffer->batch_num >> 16) & 0xFF;
		batch[i][2] = (buffer->batch_num >> 8) & 0xFF;
		batch[i][3] = buffer->batch_num & 0xFF;
		// TODO: Change this switch to a macro?
		switch(buffer->info->batching_mode) {
			case SMALL_BATCH: {
				batch[i][4] = i & 0xFF;
				break;
			}
			case LARGE_BATCH: {
				batch[i][4] = (i >> 8) & 0xFF;
				batch[i][5] = i & 0xFF;
				break;
			}
		}
	}

func (entry *BatchEntry) GetBatchNum() uint64 {
	if len(entry.Data) < 6 {
		panic(fmt.Sprintf("Invalid data length: %d", len(entry.Data)))
	}
	//log.Printf("entry.Data[0]: %s", uint64(entry.Data[0])<<24)
	//batch_num := (uint64(entry.Data[0]) << 24) |
	//	(uint64(entry.Data[1]) << 16)
	batch_num := (uint64(entry.Data[0])<<24 |
		uint64(entry.Data[1])<<16 |
		uint64(entry.Data[2])<<8 |
		uint64(entry.Data[3])<<0)
	//log.Printf("batch_num: %d", batch_num)
	return batch_num
}

func (entry *BatchEntry) GetIndex() uint16 {
	if len(entry.Data) < 6 {
		panic(fmt.Sprintf("Invalid data length: %d", len(entry.Data)))
	}
	index := (uint16(entry.Data[0])<<8 |
		uint16(entry.Data[1])<<0)
	return index
}


//#include <stdarg.h> /* Needed for the definition of va_list */
//#include <stdlib.h>
//#include <stdint.h>
//#include <stdio.h>
//#include <string.h>

//#include <unistd.h>

//#include <sys/uio.h>

//#include <fcntl.h>

//#include <sys/mman.h>
//#include <sys/types.h>
//#include <sys/stat.h>
//#include <errno.h>

//#include <sys/socket.h>
//#include <xmmintrin.h>



//#include <liburcu.h>
//#include <trace/events/rcu.h>


// RCU on the writer lock structure
// Claim and Publish are write operations (mutate the lock structure)
// Evety publish which completes a contiguous region is a Quesiant State
// Writers have a CAS block to compete for batches
// Each batch has a counter
// On publish find the first slot with batch id < target

/*
void synchronize_kernel(void);

void call_rcu(struct rcu_head *head,
              void (*func)(void *arg),
              void *arg);

struct rcu_head {
        struct list_head list;
        void (*func)(void *obj);
        void *arg;
};

void rcu_read_lock(void);

void rcu_read_unlock(void);

https://kernel.googlesource.com/pub/scm/linux/kernel/git/rostedt/linux-trace/+/ftrace/rcu-2/kernel/rcutree.c

#424 - rcu_idle_enter


*/