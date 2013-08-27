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


inline uint64_t 
rb_get_buffer_size_from_type(uint8_t buffer_type) {
	static const uint64_t buffer_sizes[17] =
	{
		0x00000040, // L0 					64	=	         64
		0x00000200, // L1				8 * 64	=           512
		0x00000400, // L2			   16 * 64	=         1,024
		0x00000800, // L3			   32 * 64	=         2,048
		0x00001000, // L4			   64 * 64 	=         4,096
		0x00008000, // L5			  8 * 4096 	=        32,768
		0x00010000, // L6			 16 * 4096 	=        65,536
		0x00020000, // L7			 32 * 4096 	=       131,072
		0x00040000, // L8			 64 * 4096 	=       262,144
		0x00200000, // L9		 8 * 64 * 4096 	=     2,097,152
		0x00400000, // L10		16 * 64 * 4096 	=     4,194,304
		0x00800000, // L11		32 * 64 * 4096 	=     8,388,608
		0x01000000, // L12		64 * 64 * 4096 	=    16,777,216
		0x02000000, // L13	   8 * 4096 * 4086  =   134,217,728
		0x04000000, // L14	  16 * 4096 * 4096  =   268,435,456
		0x08000000, // L15    32 * 4096 * 4096  =   536,870,912
		0x10000000, // L16	  64 * 4096 * 4096  = 1,073,741,824
	};
	return buffer_sizes[buffer_type];
}


typedef enum { 				//   Multiplier    | Entries
	L0		= 0, 			//	   	  1 *   64 =	     64 -- 64 * 1 [ 111111 ] 8 bit address - 1 Level Trie
	L1		= 1, 			//	   	  8 *   64 =	    512 -- 64 * 8 [ 111111 | 111 ] 16 bit address
	L2		= 2,			//		 16 *   64 =	  1,024
	L3		= 3,			//		 32 * 	64 =	  2,048
	L4		= 4,			//		  1 * 4096 =	  4,096 -- 4096 = 64 * 64 - 2 Level Trie ^
	L5 		= 5, 			//        8 * 4096 =     32,768
	L6 		= 6, 			//       16 * 4096 =     65,536 -- 64 * 64 * 16 [ 111111 | 111111 | 1111 ] 16 bit address
	L7 		= 7, 			//       32 * 4096 =    131,072 -- Here and down require 32 but addreess (technically between 16 and 32)
	L8		= 8,			//       64 * 4096 =    262,144 -- 3 Level Trie ^
	L9		= 9,			//   8 * 64 * 4096 =  2,097,152
	L10		= 10,			//  16 * 64 * 4096 =  4,194,304
	L11		= 11,			//  32 * 64 * 4096 =  8,388,608
	L12		= 12,			//  64 * 64 * 4096 = 16,777,216 -- 4 Level Trie ^
} rb_buffer_size_t;


type BUFFER_TYPES uint16

const (
	L0  BUFFER_TYPES = iota //           1 * 64 	=            64 (iota has been reset)
	L1                      //           8 * 64 	=           512
	L2                      //          16 * 64 	=         1,024
	L3                      //          32 * 64 	=         2,048
	L4                      //          64 * 64 	=         4,096
	L5                      //         8 * 4096 	=        32,768
	L6                      //        16 * 4096 	=        65,536
	L7                      //        32 * 4096 	=       131,072
	L8                      // 	      64 * 4096 	=       262,144
	L9                      //    8 * 64 * 4096 	=     2,097,152
	L10                     //   16 * 64 * 4096 	=     4,194,304
	L11                     //   32 * 64 * 4096 	=     8,388,608
	L12                     //   64 * 64 * 4096 	=    16,777,216
	L13                     //  8 * 4096 * 4086  	=   134,217,728
	L14                     // 16 * 4096 * 4096  	=   268,435,456
	L15                     // 32 * 4096 * 4096  	=   536,870,912
	L16                     // 64 * 4096 * 4096  	= 1,073,741,824
)
