#include <stdlib.h>
#include <stdint.h>
#include <assert.h>
#include <time.h>

#include "ringbuffer.h"
//#include "util.h"
//#ifndef _POOL_H_
//#include "pool.h"
//#endif
//http://contiki.sourceforge.net/docs/2.6/a00224.html

//http://stackoverflow.com/questions/9718116/improving-c-circular-buffer-efficiency

//https://github.com/pthrasher/c-generic-ring-buffer/blob/master/ringbuffer.h

//#include <xmmintrin.h>

inline uint64_t rb_get_buffer_size_from_type(uint8_t buffer_type) {
	static const uint64_t buffer_sizes[13] =
	{
		0x00000040, // L0 					64	=	      64
		0x00000200, // L1				8 * 64	=        512
		0x00000400, // L2			   16 * 64	=      1,024
		0x00000800, // L3			   32 * 64	=      2,048
		0x00001000, // L4			   64 * 64 	=      4,096
		0x00008000, // L5			  8 * 4096 	=     32,768
		0x00010000, // L6			 16 * 4096 	=     65,536
		0x00020000, // L7			 32 * 4096 	=    131,072
		0x00040000, // L8			 64 * 4096 	=    262,144
		0x00200000, // L9		 8 * 64 * 4096 	=  2,097,152
		0x00400000, // L10		16 * 64 * 4096 	=  4,194,304
		0x00800000, // L11		32 * 64 * 4096 	=  8,388,608
		0x01000000, // L12		64 * 64 * 4096 	= 16,777,216
	};
	return buffer_sizes[buffer_type];
}

//struct rb_barrier {
//	uint64_t		seq_num;		/** < Index of the last entry released for consumption */
//};

struct rb_batch {
	//rb_buffer_info *	info;			/** < Holds buffer settings */
	//uint64_t			batch_num;		/** < Sequential counter of batches to date */
	// Moving into mem pooling so need to make this as trim as possible
	// seq_num is already captured as:
	// batch.data_buffer - buffer.data_buffer
	//uint64_t			seq_num;		/** < Sequence number of the first entry in the batch */
	//uint8_t				size;			/** < The number of slots owned by this batch */
	// It's expected that any entries to skip will have their batch num set to zero and be tail aligned
	//uint8_t				pub_mask;		/** < By default this is all on (max value).  Flipping a bit off will cause the ring buffer to skip that slot */

	uint8_t **				data_buffer;	/** < Array of pointers to the slot(s) owned by this batch in the buffer */
};



struct rb_buffer {
	rb_buffer_info *	info;			/** < Holds buffer settings */

	uint64_t			size_mask;		/** < Buffer size - 1; Used to maintain scope of buffer */
	volatile uint64_t	read_seq_num;	/** < Index of the latest entry released for consumption */
	volatile uint64_t	read_barrier;	/** < Index of the oldest entry released for consumption but still in use by at least one reader */
	volatile uint64_t	write_seq_num;	/** < Index of the latest entry released for production */
	volatile uint64_t	write_barrier;	/** < Index of the oldest entry released for production but still in use by a writer */
	volatile uint64_t	batch_num;		/** < Index of the last batch allocated to a producer */

	pool *				entry_pointers;
	uint8_t *			data_buffer;
};

uint8_t const SMALL_BATCH_HEADER_SIZE = 5;
uint8_t const LARGE_BATCH_HEADER_SIZE = 6;
//#define DATA_HEADER_SIZE 6; // 4 to hold batch_num and 2 to hold batch_index (max batch size of 65,535)

int rb_init_buffer(rb_buffer** buffer_ptr, uint8_t buffer_type, rb_batching_mode_t batching_mode, uint64_t data_size) {
	// Allocate space to hold the buffer and info structs
	*buffer_ptr = malloc(sizeof(rb_buffer));
	if(!*buffer_ptr) {
		return 1;
	}
	(*buffer_ptr)->info = malloc(sizeof(rb_buffer_info));
	if(!(*buffer_ptr)->info) {
		return 1;
	}

	// Populate the info struct
	(*buffer_ptr)->info->buffer_type = buffer_type;
	(*buffer_ptr)->info->buffer_size = rb_get_buffer_size_from_type(buffer_type);
	(*buffer_ptr)->info->batching_mode = batching_mode;
	(*buffer_ptr)->info->data_size = data_size;
	switch(batching_mode) {
		case NONE: { // If batching is disabled then there is no batch header
			(*buffer_ptr)->info->entry_size = data_size; break;
		}
		case SMALL_BATCH: { // Small batches are 256 entry max but take 1 less byte in data
			(*buffer_ptr)->info->entry_size = SMALL_BATCH_HEADER_SIZE + data_size; break;
		}
		case LARGE_BATCH: { // Large batches are 65536 entry max and takes 2 bytes in data
			(*buffer_ptr)->info->entry_size = LARGE_BATCH_HEADER_SIZE + data_size; break;
		}
	}
	(*buffer_ptr)->info->total_size = (*buffer_ptr)->info->buffer_size * (*buffer_ptr)->info->entry_size;
	//rb_print_info((*buffer_ptr)->info);
	// Populate the buffer struct
	(*buffer_ptr)->size_mask = (*buffer_ptr)->info->buffer_size - 1;
	(*buffer_ptr)->read_seq_num = 0;
	(*buffer_ptr)->write_seq_num = 0;
	(*buffer_ptr)->batch_num = 1;
	
	// Init the pointer pool
	//if(ptr_pool_init(&((*buffer_ptr)->entry_pointers), (*buffer_ptr)->info->buffer_size) != 0) {
	if(pool_init(&(*buffer_ptr)->entry_pointers, (*buffer_ptr)->info->buffer_size) != 0) {
		return 1;
	}
	// Allocate giant contiguous byte array to hold the entries
	//DebugPrint("Allocating data_buffer: %d * (%d * 32) = %d", buffer_size, chunk_count, buffer_size * (chunk_count << 5));
	(*buffer_ptr)->data_buffer = malloc((*buffer_ptr)->info->total_size);
	if(!(*buffer_ptr)->data_buffer) {
		return 1;
	}

	//DebugPrint("Buffer size: %d", (*buffer_ptr)->info->buffer_size);
	//DebugPrint("Entry size: %d", (*buffer_ptr)->info->entry_size);
	//DebugPrint("Total size: %d", (*buffer_ptr)->info->total_size);
	int i = 0;
	for (i = 0; i < (*buffer_ptr)->info->total_size; i++) {
		(*buffer_ptr)->data_buffer[i] = 0xFF - (i % 0xFF);
		//char** temp = (*buffer_ptr)->data_buffer;
		//)[i] = *(char)(i & 0xFF);
		//DebugPrint("Pos[%d]: %d", i, (*buffer_ptr)->data_buffer[i]);
	}


	return 0;
}

int rb_release_buffer(rb_buffer * buffer) {
	//DebugPrint("Released Buffer: %d", buffer->seq_num);
	free(buffer->data_buffer);
	free(buffer->entry_pointers);
	free(buffer->info);
	free(buffer);
}

/*  ! ! ! ! !

rb_claim and rb_publish are NOT thread safe and must somehow be
fanned in for scenarios other than single producer

654 ns/op Buffer(L12, 2)

*/
unsigned long const MILLISECOND = 1000000L;
unsigned long const MICROSECOND = 1000L;

int rb_claim(rb_buffer * buffer, void ** entry) {

}

int rb_claim_batch(rb_buffer * buffer, void ** batch_ptr, uint16_t count) {
	// TODO: Pool batch alloc's?
	if(buffer->info->batching_mode == NONE) {
		perror("Cannot claim batches when baching mode is NONE");
		return 1;
	}
	if(count > buffer->info->buffer_size) {
		perror("Requested batch size exceeds buffer size");
		return 1;
	}
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

	char ** batch = malloc(sizeof(char *) * count);
	uint64_t pool_ptr = pool_alloc(buffer->entry_pointers, count);
	uint64_t pool_ptr2 = pool_alloc(buffer->entry_pointers, count);
	uint64_t pool_ptr3 = pool_alloc(buffer->entry_pointers, count);
	DebugPrint("pool ptr: %d", pool_ptr);
	DebugPrint("pool ptr 2: %d", pool_ptr2);
	DebugPrint("pool ptr 3: %d", pool_ptr3);

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
	(*batch_ptr) = &batch[0];

	buffer->batch_num++;
	buffer->write_seq_num += count;

	return 0;
	
}

// Internal method to share cleanup between cancel and publish
int rb_free_batch(rb_batch * batch, uint16_t count) {

}

int rb_cancel(rb_buffer * buffer, void * entry);
int rb_publish(rb_buffer * buffer, void * entry);

int rb_cancel_batch(rb_buffer * buffer, void * batch, uint16_t count) {
	//DebugPrint("Canceling batch");
	//free(batch->data_buffer);
	//free(batch);

	return 0;
}

int rb_publish_batch(rb_buffer * buffer, void * batch, uint16_t count) {
	//DebugPrint("Publishing batch");
	//free(batch);

	return 0;
}

/*
int rb_update_seq_num(rb_buffer * buffer, uint64_t value) {
	// Set seq num
	// Notify readers if applicable
}

int rb_update_write_barrier(rb_buffer * buffer, uint64_t value) {
	// Set write barrier
	// Notify writers if applicable
}

int rb_try_update_write_slot(rb_buffer * buffer, uint64_t value) {
	// Needs to be done atomically
	// Checks for room between slot and barrier
	// If no room is found do we (a) block or (b) return fail? Flag?
}
*/


void rb_print_info(rb_buffer_info* info) {
	DebugPrint("Buffer Type: %d", info->buffer_type);
	DebugPrint("Buffer Size: %d", info->buffer_size);
	//DebugPrint("Chunk Count: %d", info->chunk_count);
	DebugPrint("Data Size: %d", info->data_size);
	DebugPrint("Total Size: %d", info->total_size);
}

rb_buffer_info * rb_get_info(rb_buffer * buffer) {
	return buffer->info;
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

