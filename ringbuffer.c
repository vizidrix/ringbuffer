#include <stdlib.h>
#include <stdint.h>
#include <assert.h>

#include "ringbuffer.h"
#include "util.h"

//#include <xmmintrin.h>

inline uint64_t rb_get_buffer_size_from_type(uint8_t buffer_type) {
	static const uint64_t buffer_sizes[14] =
	{
		0x00000000,
		0x00000040, //             64	=	      64
		0x00000200, //         8 * 64	=        512
		0x00000400, //        16 * 64	=      1,024
		0x00000800, //        32 * 64	=      2,048
		0x00001000, //        64 * 64 	=      4,096
		0x00008000, //       8 * 4096 	=     32,768
		0x00010000, //      16 * 4096 	=     65,536
		0x00020000, //      32 * 4096 	=    131,072
		0x00040000, // 	    64 * 4096 	=    262,144
		0x00200000, //  8 * 64 * 4096 	=  2,097,152
		0x00400000, // 16 * 64 * 4096 	=  4,194,304
		0x00800000, // 32 * 64 * 4096 	=  8,388,608
		0x01000000, // 64 * 64 * 4096 	= 16,777,216
	};
	return buffer_sizes[buffer_type];
}

//struct rb_barrier {
//	uint64_t		seq_num;		/** < Index of the last entry released for consumption */
//};

struct rb_batch {
	rb_buffer_info *	info;			/** < Holds buffer settings */
	uint64_t			batch_num;		/** < Sequential counter of batches to date */
	uint64_t			seq_num;		/** < Sequence number of the first entry in the batch */
	uint8_t				size;			/** < The number of slots owned by this batch */
	uint8_t				pub_mask;		/** < By default this is all on (max value).  Flipping a bit off will cause the ring buffer to skip that slot */

	void *				data_buffer;	/** < Pointer to the slot(s) owned by this batch in the buffer */
};

struct rb_buffer {
	rb_buffer_info *	info;			/** < Holds buffer settings */

	uint64_t			size_mask;		/** < Buffer size - 1; Used to maintain scope of buffer */
	uint64_t			read_seq_num;	/** < Index of the last entry released for consumption */
	uint64_t			write_seq_num;	/** < Index of the last entry released for production */
	uint64_t			batch_num;		/** < Index of the last batch allocated to a producer */

	void *				data_buffer;

	uint64_t			write_slot;		/** < Index of the next available production slot */
	uint64_t			write_barrier;  /** < Index of the last entry released for production */
};

void rb_print_info(rb_buffer_info* info) {
	DebugPrint("Buffer Type: %d", info->buffer_type);
	DebugPrint("Buffer Size: %d", info->buffer_size);
	DebugPrint("Chunk Count: %d", info->chunk_count);
	DebugPrint("Data Size: %d", info->data_size);
}

int rb_init_buffer(rb_buffer** buffer_ptr, uint8_t buffer_type, uint8_t chunk_count) {
	// Allocate space to hold the buffer and info structs
	*buffer_ptr = malloc(sizeof(rb_buffer));
	(*buffer_ptr)->info = malloc(sizeof(rb_buffer_info));
	uint64_t buffer_size = rb_get_buffer_size_from_type(buffer_type);
	// Populate the info struct
	(*buffer_ptr)->info->buffer_type = buffer_type;
	(*buffer_ptr)->info->buffer_size = buffer_size;
	(*buffer_ptr)->info->chunk_count = chunk_count;
	(*buffer_ptr)->info->data_size = chunk_count << 5; // * 32
	//rb_print_info((*buffer_ptr)->info);
	// Populate the buffer struct
	(*buffer_ptr)->size_mask = buffer_size - 1;
	(*buffer_ptr)->read_seq_num = 0;
	(*buffer_ptr)->write_seq_num = 0;
	(*buffer_ptr)->batch_num = 1;
	
	// Allocate giant contiguous byte array to hold the entries
	//DebugPrint("Allocating data_buffer: %d * (%d * 32) = %d", buffer_size, chunk_count, buffer_size * (chunk_count << 5));
	(*buffer_ptr)->data_buffer = malloc(buffer_size * (chunk_count << 5));

	return 0;
}

int rb_release_buffer(rb_buffer * buffer) {
	//DebugPrint("Released Buffer: %d", buffer->seq_num);
	free(buffer->info);
	free(buffer);
}

rb_buffer_info * rb_get_info(rb_buffer * buffer) {
	return buffer->info;
}

//struct rb_batch {
//	rb_buffer_info *	info;			/** < Holds buffer settings */
//	uint8_t				size;			/** < The number of slots owned by this batch */
//	uint8_t				pub_mask;		/** < By default this is all on (max value).  Flipping a bit off will cause the ring buffer to skip that slot */
//
//	void *				data_buffer;	/** < Pointer to the slot(s) owned by this batch in the buffer */
//};

/*  ! ! ! ! !

rb_claim and rb_publish are NOT thread safe and must somehow be
fanned in for scenarios other than single producer

*/

//rb_batch * rb_claim(rb_buffer * buffer, rb_batch ** batch, uint8_t count) {
int rb_claim(rb_buffer * buffer, rb_batch ** batch, uint8_t count) {
	// TODO: Pool batch alloc's?
	if(count > buffer->info->buffer_size) {
		perror("Requested batch size exceeds buffer size");
		//errno(1);
		return 1;
	}
	//rb_batch * batch = malloc(sizeof(rb_batch));
	*batch = malloc(sizeof(rb_batch));
	(*batch)->info = buffer->info;
	(*batch)->batch_num = buffer->batch_num;
	(*batch)->seq_num = buffer->write_seq_num;
	(*batch)->size = count;
	(*batch)->pub_mask = 0xFF;
	(*batch)->data_buffer = (&buffer->data_buffer)[buffer->write_seq_num];

	buffer->batch_num++;
	buffer->write_seq_num += count;

	return 0;
}

int rb_cancel(rb_batch * batch) {
	free(batch);

	return 0;
}

int rb_publish(rb_batch * batch) {
	free(batch);

	return 0;
}

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

