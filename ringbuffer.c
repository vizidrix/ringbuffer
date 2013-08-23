#include <stdlib.h>
#include <stdint.h>
#include <assert.h>
#include <time.h>

#include "ringbuffer.h"
#include "util.h"

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

/* This should morph into a pool of pointers
typedef struct rb_batch_pool {
	rb_batch *			batch_pool;
	char **				entry_pool;

	uint64_t			batch_pool_count;
	uint64_t			batch_pool_index;
	uint64_t			
} rb_batch_pool;
*/

struct rb_buffer {
	rb_buffer_info *	info;			/** < Holds buffer settings */

	uint64_t			size_mask;		/** < Buffer size - 1; Used to maintain scope of buffer */
	volatile uint64_t	read_seq_num;	/** < Index of the latest entry released for consumption */
	volatile uint64_t	read_barrier;	/** < Index of the oldest entry released for consumption but still in use by at least one reader */
	volatile uint64_t	write_seq_num;	/** < Index of the latest entry released for production */
	volatile uint64_t	write_barrier;	/** < Index of the oldest entry released for production but still in use by a writer */
	volatile uint64_t	batch_num;		/** < Index of the last batch allocated to a producer */

	uint8_t *				data_buffer;
};

uint64_t const DATA_HEADER_SIZE = 6;
//#define DATA_HEADER_SIZE 6; // 4 to hold batch_num and 2 to hold batch_index (max batch size of 65,535)

int rb_init_buffer(rb_buffer** buffer_ptr, uint8_t buffer_type, uint64_t data_size) {
	// Allocate space to hold the buffer and info structs
	*buffer_ptr = malloc(sizeof(rb_buffer));
	if(!buffer_ptr) {
		return 1;
	}
	(*buffer_ptr)->info = malloc(sizeof(rb_buffer_info));
	if(!(*buffer_ptr)->info) {
		return 1;
	}
	//uint64_t buffer_size = ;
	//uint64_t entry_size = DATA_HEADER_SIZE + data_size;
	// Populate the info struct
	(*buffer_ptr)->info->buffer_type = buffer_type;
	(*buffer_ptr)->info->buffer_size = rb_get_buffer_size_from_type(buffer_type);
	//(*buffer_ptr)->info->chunk_count = chunk_count;
	(*buffer_ptr)->info->data_size = data_size;//chunk_count << 5; // * 32
	(*buffer_ptr)->info->entry_size = DATA_HEADER_SIZE + data_size;//DATA_HEADER_SIZE + data_size;
	//(*buffer_ptr)->info->total_size = buffer_size * (chunk_count << 5);
	//uint64_t total_size = buffer_size * (DATA_HEADER_SIZE + data_size);
	//(*buffer_ptr)->info->total_size = total_size;
	(*buffer_ptr)->info->total_size = (*buffer_ptr)->info->buffer_size * (*buffer_ptr)->info->entry_size;
	//rb_print_info((*buffer_ptr)->info);
	// Populate the buffer struct
	(*buffer_ptr)->size_mask = (*buffer_ptr)->info->buffer_size - 1;
	(*buffer_ptr)->read_seq_num = 0;
	(*buffer_ptr)->write_seq_num = 0;
	(*buffer_ptr)->batch_num = 1;
	
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
	free(buffer->info);
	free(buffer);
}

/*  ! ! ! ! !

rb_claim and rb_publish are NOT thread safe and must somehow be
fanned in for scenarios other than single producer

*/
unsigned long const MILLISECOND = 1000000L;
unsigned long const MICROSECOND = 1000L;


//int rb_claim(rb_buffer * buffer, void ** batch_ptr, uint16_t count) {
int rb_claim(rb_buffer * buffer, void ** batch_ptr, uint16_t count) {
	// TODO: Pool batch alloc's?
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

/*
	//	char * stuff[count];
	//stuff = ((char (*)[count])(*batch_ptr));
	//stuff = ([]char *)malloc(sizeof(char *));// = malloc(sizeof(char *) * count);
	//&(*batch_ptr) = stuff;
	int i = 0;
	int j = 0;
	for(i = 0; i < count; i++) {
		//DebugPrint("Stuff: %d", ((char (*)[count])(*batch_ptr)));
		//((char (*)[count])(*batch_ptr))[i] = malloc(sizeof(char *));// * buffer->info->entry_size);

		int index = buffer->info->entry_size * (buffer->write_seq_num + i);
		DebugPrint("Index: %d", index);
		((char (*)[count])(*batch_ptr))[i] = (char (*))&buffer->data_buffer[index];
		
		((char (*)[count])(*batch_ptr))[i][0] = (buffer->batch_num >> 24) & 0xFF;
		((char (*)[count])(*batch_ptr))[i][1] = (buffer->batch_num >> 16) & 0xFF;
		((char (*)[count])(*batch_ptr))[i][2] = (buffer->batch_num >> 8) & 0xFF;
		((char (*)[count])(*batch_ptr))[i][3] = buffer->batch_num & 0xFF;
		((char (*)[count])(*batch_ptr))[i][4] = (i >> 8) & 0xFF;
		((char (*)[count])(*batch_ptr))[i][5] = i & 0xFF;
		DebugPrint("Stuff: %d", ((char (*)[count])(*batch_ptr))[i][5]);
	}
	//(*batch_ptr) = &stuff;
*/



	//char * stuff[count] = (*batch_ptr);
	//char * stuff[count];
	char ** batch = malloc(sizeof(char *) * count);

	//char (* stuff)[count];	
	//stuff) = malloc(sizeof(char *) * count);
	//stuff = (char (*)[count])(*batch_ptr);
	//stuff = ([]char *)malloc(sizeof(char *));// = malloc(sizeof(char *) * count);
	//&(*batch_ptr) = stuff;
	int i = 0;
	int j = 0;
	for(i = 0; i < count; i++) {
		//DebugPrint("Stuff: %d", &batch[i]);
		batch[i] = malloc(sizeof(char *));// * buffer->info->entry_size);
		//DebugPrint("Stuff: %d", &batch[i]);
		int index = buffer->info->entry_size * (buffer->write_seq_num + i);
		//DebugPrint("Index: %d", index);
		batch[i] = &buffer->data_buffer[index];
		
		batch[i][0] = (buffer->batch_num >> 24) & 0xFF;
		batch[i][1] = (buffer->batch_num >> 16) & 0xFF;
		batch[i][2] = (buffer->batch_num >> 8) & 0xFF;
		batch[i][3] = buffer->batch_num & 0xFF;
		batch[i][4] = (i >> 8) & 0xFF;
		batch[i][5] = i & 0xFF;
		//DebugPrint("Stuff: %d", batch[i][5]);
	}
	//(*batch_ptr) = &stuff;
	(*batch_ptr) = &batch[0];



	//char value = ((char (*)[count])(*batch_ptr))[0][5];
	//DebugPrint("Value: %d", value);






	//char ** batch = (char**)(malloc(sizeof(char *) * count));
	//batch_ptr = batch;
	/*
	if(batch) {
		int i = 0;
		for(i = 0; i < count; i++) {
			batch[i] = malloc(sizeof(char *) * 10);
			if(batch[i]) {
				int j = 0;
				for(j = 0; j < 10; j++) {
					batch[i][j] = j;
				}
			}
		}
	}
	*/
	//(*batch_ptr) = malloc(sizeof(char*) * count);

	//batch = malloc(sizeof(char*) * count);
	//int i, index = 0;
	//for(i = 0; i < count; i++) {
		//((char**)(*batch_ptr)) = buffer->data_buffer;
		//*((char**)(batch_ptr))[i] = buffer->data_buffer;
	//}

	//(*batch_ptr) = malloc(1000);
	
	/*
	size_t size = (sizeof(char**) * count);
	DebugPrint("Size: %d", size);
	char** batch = malloc(sizeof(char**) * count);
	(*batch_ptr) = &batch;//(void **)batch;
*/
/*
	int i = 0;
	for(i = 0; i < count; i++) {
		buffer->data_buffer[i] = i;
	
		DebugPrint("BEFORE \t batch/buffer - [%x]: %x / [%x]: %x", batch, batch[i][0], buffer->data_buffer, buffer->data_buffer[i]);
		
		batch[i] = &buffer->data_buffer[i];

		DebugPrint("AFTER \t batch/buffer - [%x]: %x / [%x]: %x", batch, batch[i][0], buffer->data_buffer, buffer->data_buffer[i]);

		

		DebugPrint("CAST \t batch_ptr/buffer - [%x]: %x / [%x]: %x", ((char **)(batch_ptr)), ((char **)(batch_ptr))[i][0], buffer->data_buffer, buffer->data_buffer[i]);

		DebugPrint("_");
	}
	*.
	//free(batch);
	//*batch_ptr = malloc(sizeof(char**) * count);
	//DebugPrint("batch_ptr: %x", batch_ptr);

	/*
	//char** batch = malloc(sizeof(char**) * count);
	char** batch = *batch_ptr;
	if(!batch) {
		return 1;
	}
	uint64_t index = buffer->write_seq_num * buffer->info->entry_size;

	DebugPrint("Buffer: % x", &buffer->data_buffer);
	// Need to assign chunks from buffer to array of pointers
	uint8_t i = 0;
	for(i = 0; i < count; i++) {
		//DebugPrint("B")
		// First will always fit or we'd have already wrapped
		//batch[i] = &buffer->data_buffer + index;
		batch[i] = malloc(buffer->info->entry_size);

		// Need to set the batch num and index into the first 6 bytes
		batch[i][0] = 0x01;
		batch[i][1] = 0x01;
		batch[i][2] = 0x01;
		batch[i][3] = 0x01;
		batch[i][4] = 0x01;
		batch[i][5] = 0x01;

		int j = 0;
		for(j = 0; j < buffer->info->entry_size; j++) {
			batch[i][j] = 0x01;
		}

		// Move index to the next slot position
		index += buffer->info->entry_size;// DATA_HEADER_SIZE + buffer->info->data_size;
		// Check for overflow on next itteration
		if(buffer->write_seq_num + i == buffer->info->buffer_size) {
			// Reached the last slot, time to wrap
			index = 0;
		}
	}
	*/

	//batch = &((void *)b);

	//batch = (void**)b;
	//DebugPrint("batch: %x", *batch[0]);
	//DebugPrint("Index 0: %d", ((char**)batch)[0][1]);
	buffer->batch_num++;
	buffer->write_seq_num += count;

	return 0;
	//return 0;

	/*
	*batch = malloc(sizeof(rb_batch));
	if(!*batch || *batch == NULL) {
		DebugPrint("dead batch");
		return 1;
	}
	//(*batch)->info = buffer->info;
	//(*batch)->batch_num = buffer->batch_num;
	//(*batch)->seq_num = buffer->write_seq_num;
	//(*batch)->size = count;
	//(*batch)->pub_mask = 0xFF;
	(*batch)->data_buffer = malloc(sizeof(char**) * count);//char*[count];
	// Enfoce buffer length here, causing segfaults

	// 391317

	//(*batch)->data_buffer = (&buffer->data_buffer)[buffer->write_seq_num];
	//char * temp = buffer->data_buffer;
	uint64_t index = buffer->write_seq_num * buffer->info->entry_size;// (DATA_HEADER_SIZE + buffer->info->data_size);
	//if(buffer->write_seq_num > 300000) {
	//	DebugPrint("Begun:\t%d > %d", buffer->write_seq_num, buffer->info->buffer_size);
	//}
	//if(buffer->write_seq_num > buffer->info->buffer_size) {
	//	DebugPrint("%d > %d", buffer->write_seq_num, buffer->info->buffer_size);
	//	errno(1);
	//} else {
	//(*batch)->data_buffer = (&buffer->data_buffer)[buffer->write_seq_num];

	// Need to assign chunks from buffer to array of pointers
	uint8_t i = 0;
	for(i = 0; i < count; i++) {
		// First will always fit or we'd have already wrapped
		(*batch)->data_buffer[i] = buffer->data_buffer + index;

		// Need to set the batch num and index into the first 6 bytes

		// Move index to the next slot position
		index += DATA_HEADER_SIZE + buffer->info->data_size;
		// Check for overflow on next itteration
		if(buffer->write_seq_num + i == buffer->info->buffer_size) {
			// Reached the last slot, time to wrap
			index = 0;
		}
	}
	//(*batch)->data_buffer = temp + index;//[buffer->write_seq_num];
	//}
	/*if(buffer->write_seq_num > 300000) {
		DebugPrint("Finished:\t%d > %d", buffer->write_seq_num, buffer->info->buffer_size);
	}*/

	
}

int rb_free_batch(rb_batch * batch, uint16_t count) {

}
int rb_cancel(rb_buffer * buffer, rb_batch * batch, uint16_t count) {
	//DebugPrint("Canceling batch");
	//free(batch->data_buffer);
	//free(batch);

	return 0;
}

int rb_publish(rb_buffer * buffer, rb_batch * batch, uint16_t count) {
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

