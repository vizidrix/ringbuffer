#include <stdlib.h>
#include <stdint.h>
#include <assert.h>
#include <time.h>
#include <errno.h>
#include <sched.h>

#include "ringbuffer.h"
#include "util.h"

#include <xmmintrin.h>



//http://gcc.gnu.org/onlinedocs/gcc-4.1.2/gcc/Atomic-Builtins.html

#define __errno(err) if(errno == RB_SUCCESS) { errno = err; }
// Makes sure that the write cursor cannot lap the slowest reader
//#define rb_write_space(buffer) ((buffer->stats->read_barrier + buffer->info->data_buffer_size) - buffer->stats->write_seq_num)

struct rb_buffer {
	struct rb_buffer_info		info;			/** < Holds buffer settings */
	struct rb_buffer_stats		stats;			/** < Holds buffer allocation details */
	rb_batch *					batches;
	rb_slice *					slices;
	uint8_t *					data_buffer;
};

void
rb_reset_batch(struct rb_batch * batch) {
	int i = 0;
	//DebugPrint("Set reader flags");
	for(; i < 8; i++) {
		batch->reader_flags[i] = 0;
	}
	batch->state = AVAILABLE;
	// TODO: make these atomic?
	batch->group_flags = 0x00000000; // Same as: all groups complete
	batch->seq_num = 0xFFFFFFFF; // Buffer process should ignore 0xFFFFFFFF
	batch->batch_size = 0; // On claim this will set to > 0 and seq_num == 0
}

void
rb_init_buffer(struct rb_buffer** buffer_ptr, uint64_t batch_buffer_size, uint64_t data_buffer_size, uint64_t entry_size) {
	// Allocate space to hold the buffer and info structs
	*buffer_ptr = malloc(sizeof(rb_buffer));
	if(!*buffer_ptr) {
		__errno(RB_ALLOC_BUFFER); goto error;
	}

	batch_buffer_size = round_up_pow_2_uint64_t(batch_buffer_size);
	data_buffer_size = round_up_pow_2_uint64_t(data_buffer_size);
	
	// Populate the info struct
	(*buffer_ptr)->info.batch_buffer_size = batch_buffer_size;
	(*buffer_ptr)->info.batch_size_mask = batch_buffer_size - 1;
	(*buffer_ptr)->info.data_buffer_size = data_buffer_size;
	(*buffer_ptr)->info.data_size_mask = data_buffer_size - 1;
	(*buffer_ptr)->info.entry_size = entry_size;
	(*buffer_ptr)->info.total_data_size = 
		(*buffer_ptr)->info.data_buffer_size * 
		(*buffer_ptr)->info.entry_size;

	// Populate the stats struct
	(*buffer_ptr)->stats.barrier_batch_num = 0;
	(*buffer_ptr)->stats.read_batch_num = 0;
	(*buffer_ptr)->stats.write_batch_num = 0;
	(*buffer_ptr)->stats.barrier_seq_num = 0;
	(*buffer_ptr)->stats.write_seq_num = 0;

	int i = 0;
	// Allocate a pool of batches to hold claimed set data
	(*buffer_ptr)->batches = malloc(sizeof(rb_batch) * (*buffer_ptr)->info.batch_buffer_size);
	if(!(*buffer_ptr)->batches) {
		__errno(RB_ALLOC_BATCHES); goto error;
	}
	for(i = 0; i < (*buffer_ptr)->info.batch_buffer_size; i++) {
		(*buffer_ptr)->batches[i].batch_num = 0;
		//DebugPrint("Resetting Batch: %d", i);
		rb_reset_batch(&(*buffer_ptr)->batches[i]);
	}
	
	// Flag to set if we alloc the slices later
	//(*buffer_ptr)->slices = NULL;

	// Allocate giant contiguous byte array to hold the entries
	(*buffer_ptr)->data_buffer = malloc((*buffer_ptr)->info.total_data_size);
	if(!(*buffer_ptr)->data_buffer) {
		__errno(RB_ALLOC_DATA); goto error;
	}
	// Flush the data buffer to all zeros
	for (i = 0; i < (*buffer_ptr)->info.total_data_size; i++) {
		(*buffer_ptr)->data_buffer[i] = 0;//0xFF - (i % 0xFF);
	}
	__errno(RB_SUCCESS);
	return;
error:
	DebugPrint("ERROR IN CREATE BUFFER");
	__errno(RB_ERROR);
	rb_free_buffer(buffer_ptr);
}

void
rb_free_buffer(struct rb_buffer ** buffer) {
	if(*buffer != NULL) {
		//free((*buffer)->slices);
		free((*buffer)->data_buffer);
		free((*buffer)->batches);
		(*buffer)=(free(*buffer),NULL);
	}
}

inline uint64_t
rb_process_releases(rb_buffer * buffer) {
	DebugPrint("Processing releases...");
}

inline uint64_t
rb_process_claims(rb_buffer * buffer) {
	DebugPrint("Processing claims...");
}

inline uint64_t
rb_process_publishes(rb_buffer * buffer) {
	DebugPrint("Processing publishes...");
}

inline rb_process_result
rb_process_all(rb_buffer * buffer) {
	DebugPrint("Processing all...");
	rb_process_result result;
	// First we want to free up any available batches/seqs
	result.releases = rb_process_releases(buffer);
	// Then we want to let writers start doing their work
	result.claims = rb_process_claims(buffer);
	// Finally we release pubished batches to the reader pool
	result.publishes = rb_process_publishes(buffer);

	// Release finished read batches -> make batch available for writers
	// - Starting at 
	// - Find batch(es) where group_flags == 0xFFFF and reset them
	// Release finished write batches -> make batch available for readers
	// - Find batch(es) where 
	// Claim batches for pending writers -> allocate seq_num(s) to batches
	// - Starting at write_batch_num - barrier between writer claimed batches and open/reader batches
	// - Find batch(es) where batch_size > 0 and seq_num is 0xFFFFFFFF
	// - # have been requested but seq_num has not been assigned yet

	/*
	// Find the position of the oldest released write
	uint64_t batch_index = rb_buffer->stats->read_batch_num & rb_buffer->info->batch_size_mask;
	while(!buffer->batches)

	int i = 0;
	//DebugPrint("Scanning");
	// Update looping logic... go forever?
	for(i = 0; i < buffer->info->batch_buffer_size; i++) {
		//DebugPrint("Batch size[%d]: %d", i, buffer->batches[i].batch_size);
		//DebugPrint("Seq Num[%d]: %d", i, buffer->batches[i].seq_num);
		// Claimed but unfulfilled

		if(buffer->batches[i].batch_size > 0 && buffer->batches[i].seq_num == 0xFFFFFFFF) {
			//DebugPrint("if(%d > 0 && %d == 0xFFFFFFFF", buffer->batches[i].batch_size, buffer->batches[i].seq_num);
			// Claim the number of requested slots by inc the seq_num
			uint64_t seq_num = (uint64_t)__sync_fetch_and_add(&buffer->stats->write_seq_num, buffer->batches[i].batch_size);
			DebugPrint("Process[%d]: %d size %d -> write_seq_num: %d", i, seq_num, buffer->batches[i].batch_size, buffer->stats->write_seq_num);
			if(!__sync_bool_compare_and_swap(&buffer->batches[i].seq_num, 0xFFFFFFFF, seq_num)){
				DebugPrint("Error in process"); // Could be caused by multiple processors
				__errno(RB_ERROR);
				goto error;
			}
			//DebugPrint("Set seq num[%d]: %d", i, buffer->batches[i].seq_num);
		}
	}
	*/
	return result;
error:
	__errno(RB_ERROR);
	return result;
}

void temp(rb_buffer * buffer, uint64_t count) {
	//rb_buffer * buffer;
	//rb_init_buffer(&buffer, 1000, 1, 1024);
	int i = 0;
	for(i = 0; i < count; i++) {
		char cancel = 0;
		rb_batch * batch = rb_claim(buffer, 1, &cancel);

		//if(batch->batch_num != i) {
		//	__errno(100);
		//	return;
		//}
	}
	//rb_free_buffer(&buffer);
	return;
}

void // Publish is not guaranteed to be sequential
rb_publish(rb_buffer * buffer, rb_batch * batch) {
	batch->state = PUBLISHED;
	BARRIER();
	DebugPrint("%d != %d", batch->batch_num, buffer->stats.read_batch_num);
	if(batch->batch_num != buffer->stats.read_batch_num) {
		return; // All done here
	}
	uint64_t index = batch->batch_num;
	do {
		DebugPrint("1 Writing read_batch_num %d to new value %d", buffer->stats.read_batch_num, index);
		// Scan across all batches starting at the next slot looking for other published batches
		while(buffer->batches[(++index) & buffer->info.batch_size_mask].state == PUBLISHED) {}
		// While loop overshoots by one... but this pointer should point to the oldest publish?
		DebugPrint("2 Writing read_batch_num %d to new value %d", buffer->stats.read_batch_num, index);
		__sync_bool_compare_and_swap(&buffer->stats.read_batch_num, batch->batch_num, index);
	// Handle a potential edge case where the next batch was published and returned in between 
	// the prev while and the CAS.  Should be very rare, if ever
	} while(buffer->batches[(index) & buffer->info.batch_size_mask].state == PUBLISHED);
}

void
rb_release(rb_buffer * buffer, rb_batch * batch) {
	if(batch->state != PUBLISHED || batch->batch_num >= buffer->stats.read_batch_num) {
		__errno(RB_RELEASE_OVERFLOW); goto error;
	}
	rb_reset_batch(batch);

	__sync_add_and_fetch(&buffer->stats.barrier_batch_num, 1);
	__errno(RB_SUCCESS);
	return;
error:
	__errno(RB_ERROR);
	return;
}

rb_batch *
rb_claim(rb_buffer * buffer, uint16_t count, void* cancel) {
	if(count == 0 || count > buffer->info.data_buffer_size) {
		DebugPrint("count [%d] == 0 || > Data buffer size: %d", count, buffer->info.data_buffer_size);
		__errno(RB_CLAIM_PANIC);
		goto error;
	} // Must be > 0 and < buffer size
	uint64_t index = buffer->stats.write_batch_num;
	// Scan forward trying to put your count in the slot first
	DebugPrint("%d >= (%d + %d)", index, buffer->stats.barrier_batch_num, buffer->info.batch_buffer_size);
	while(__builtin_expect(index >= (buffer->stats.barrier_batch_num + buffer->info.batch_buffer_size), 0) ||
		!__sync_bool_compare_and_swap(&buffer->batches[index++ & buffer->info.batch_size_mask].batch_size, 0, count)) {
		sched_yield();
		usleep(1);
		if(__builtin_expect(cancel == NULL, 0)) { __errno(RB_CLAIM_CANCELED); goto error; }
	}
	// Increment the starting spot for the next claim
	__sync_add_and_fetch(&buffer->stats.write_batch_num, 1);

	index--; // Loop causes the index to overshoot by 1
	buffer->batches[index & buffer->info.batch_size_mask].batch_num = index;
	buffer->batches[index & buffer->info.batch_size_mask].state = WRITING;

	uint64_t seq_num = 0;
	uint64_t prev_count = 0;
	// Look backwards looking for a previously assigned seq num
	// Count batch size along the way for a running offset
	// If you reach the Barrier_Buf_Num then use
	//while()
	// Fake reader releases
	//__sync_add_and_fetch(&buffer->stats.barrier_batch_num, 1);
	//buffer->stats.barrier_batch_num++;
	
	__errno(RB_SUCCESS);
	return &buffer->batches[index & buffer->info.batch_size_mask];

	/*
	rb_process(buffer);

	rb_batch * batch = &buffer->batches[(index-1) & buffer->info->batch_size_mask];

	// TODO: Return the batch and let the caller decide how to wait
	retries = 0;
	// What happens if a claimed slot is abandoned by the writer?
	// Spin wait until the buffer process has allocated a seq_num for the batch
	while((*batch).seq_num == 0xFFFFFFFF) {
		retries++;
		if(retries > 10) {
			__errno(RB_CLAIM_FULL);
			DebugPrint("Retries exceeded");
			goto error;
		}
	}
	DebugPrint("Claim seq_num: %d - size: %d", (*batch).seq_num, (*batch).batch_size);
	//DebugPrint("after Index: %d", index-1);
	//DebugPrint("after Value: %d", buffer->batches[index-1].batch_size);
	*/
	// batch num update should be done by buffer process as it allocates

	

	//batch->seq_num = buffer->stats->write_seq_num;
	
	//buffer->stats->write_seq_num += count;

error:
	__errno(RB_ERROR);
	return NULL;
}

rb_batch *
rb_get_batch(rb_buffer * buffer, uint64_t batch_num) {
	return &buffer->batches[batch_num & buffer->info.batch_size_mask];
}

void *
rb_get_entry(rb_buffer * buffer, uint64_t seq_num) {
	return &buffer->data_buffer[seq_num & buffer->info.data_size_mask];
}

void *
rb_get_entry_slice(rb_buffer * buffer, uint64_t seq_num) {
	// Initialize slice mem pool
	// TODO: find a more memory efficient/compact way of allocating this pool
	// TODO: As it stands there is a 24 byte overhead to every entry due to fixed sized pool
	if(buffer->slices == NULL) {
		// Lazy load the pool on first call since this is for Go interop
		// and we don't want to use the mem unless it's needed
		buffer->slices = malloc(sizeof(rb_slice) * buffer->info.data_buffer_size);
		int i = 0;
		for(i = 0; i < buffer->info.data_buffer_size; i++) {
			buffer->slices[i].data = 0;
			buffer->slices[i].len = buffer->info.entry_size;
			buffer->slices[i].cap = buffer->info.entry_size;
		}
	}
	// With fixed pool size this should always be a safe operation
	buffer->slices[seq_num].data = rb_get_entry(buffer, seq_num);
	return &buffer->slices[seq_num & buffer->info.data_size_mask];
}

/*
void
rb_publish(rb_buffer * buffer, rb_batch * batch) {
	buffer->stats->write_barrier+=batch->batch_size;

	// REMOVE THIS - Simulates readers immediately consuming
	buffer->stats->read_barrier+=batch->batch_size;
	buffer->stats->read_seq_num+=batch->batch_size;


	__errno(RB_SUCCESS);
	return;
error:
	__errno(RB_ERROR);
}
*/

/*
void rb_claim_and_publish(rb_buffer * buffer, int count) {
	int i = 0;
	for(i = 0; i < count; i++) {
		rb_batch * batch = rb_claim(buffer, 1);
		char * entry = (char *)rb_get_entry(buffer, batch->seq_num);
		//char[] data = { 1, 2 };
		//rb_publish(buffer, batch);
		rb_publish(batch);
	}
}
*/

// Reader needs to notify when it's barrier is updated
// and be notified when it's next seq_num is avail
// i.e. holds a range 'lock' from barrier to seq_num

rb_buffer_info * 
rb_get_info(rb_buffer * buffer) {
	return &buffer->info;
}

rb_buffer_stats * 
rb_get_stats(rb_buffer * buffer) {
	return &buffer->stats;
}

void
rb_print_info(rb_buffer * buffer) {
	DebugPrint("C Info - Batch# [ %d ] Data# [ %d ] Entry# [ %d ] - Entry Buffer# [ %d ]",
			buffer->info.batch_buffer_size,
			buffer->info.data_buffer_size,
			buffer->info.entry_size,
			buffer->info.total_data_size);
}

void
rb_print_stats(rb_buffer * buffer) {
	DebugPrint("C Stats - Batch [ B %d | R %d | W %d ] Seq [ B %d | W %d ]",
			buffer->stats.barrier_batch_num,
			buffer->stats.read_batch_num,
			buffer->stats.write_batch_num,
			buffer->stats.barrier_seq_num,
			buffer->stats.write_seq_num);
}

void
rb_print_buffer(rb_buffer * buffer) {
	DebugPrint("Buffer - Info | Stats | Batches | Data");
	rb_print_info(buffer);
	rb_print_stats(buffer);
}
