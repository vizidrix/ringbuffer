#include <stdlib.h>
#include <stdint.h>
#include <assert.h>
#include <time.h>
#include <errno.h>

#include "ringbuffer.h"
#include "util.h"

//#include <xmmintrin.h>

#define __errno(err) if(errno == RB_SUCCESS) { errno = err; }
#define rb_write_space(buffer) ((buffer->stats->read_barrier + buffer->info->buffer_size) - buffer->stats->write_seq_num)

struct rb_buffer {
	rb_buffer_info *	info;			/** < Holds buffer settings */
	rb_buffer_stats *	stats;			/** < Holds buffer allocation details */
	rb_batch *			batches;
	rb_slice *			slices;
	uint8_t *			data_buffer;
};

void
rb_init_buffer(rb_buffer** buffer_ptr, uint64_t buffer_size, uint64_t entry_size) {
	// Allocate space to hold the buffer and info structs
	*buffer_ptr = malloc(sizeof(rb_buffer));
	if(!*buffer_ptr) {
		__errno(RB_ALLOC_BUFFER); goto error;
	}
	(*buffer_ptr)->info = malloc(sizeof(rb_buffer_info));
	if(!(*buffer_ptr)->info) {
		__errno(RB_ALLOC_INFO); goto error;
	}
	(*buffer_ptr)->stats = malloc(sizeof(rb_buffer_stats));
	if(!(*buffer_ptr)->stats) {
		__errno(RB_ALLOC_STATS); goto error;
	}

	buffer_size = round_up_pow_2_uint64_t(buffer_size);

	// Populate the info struct
	(*buffer_ptr)->info->buffer_size = buffer_size;
	(*buffer_ptr)->info->entry_size = entry_size;
	(*buffer_ptr)->info->total_size = (*buffer_ptr)->info->buffer_size * (*buffer_ptr)->info->entry_size;
	(*buffer_ptr)->info->size_mask = (*buffer_ptr)->info->buffer_size - 1;

	// Populate the buffer struct
	(*buffer_ptr)->stats->read_seq_num = 0;
	(*buffer_ptr)->stats->read_barrier = 0;
	(*buffer_ptr)->stats->write_seq_num = 0;
	(*buffer_ptr)->stats->write_barrier = 0;
	(*buffer_ptr)->stats->batch_num = 0;
	
	// Allocate a pool of batches to hold claimed set data
	// TODO: find a more memory efficient/compact way of allocating these pools
	// TODO: As it stands there is a 32 byte overhead to every entry due to fixed sized pools
	(*buffer_ptr)->batches = malloc(sizeof(rb_batch) * (*buffer_ptr)->info->buffer_size);
	(*buffer_ptr)->slices = malloc(sizeof(rb_slice) * (*buffer_ptr)->info->buffer_size);

	// Allocate giant contiguous byte array to hold the entries
	(*buffer_ptr)->data_buffer = malloc((*buffer_ptr)->info->total_size);
	if(!(*buffer_ptr)->data_buffer) {
		__errno(RB_ALLOC_DATA); goto error;
	}
	
	// Setting default values - should zero in real setup
	//int i = 0;
	//for (i = 0; i < (*buffer_ptr)->info->total_size; i++) {
	//	(*buffer_ptr)->data_buffer[i] = 0xFF - (i % 0xFF);
	//}

	__errno(RB_SUCCESS);
	return;
error:
	__errno(RB_ERROR);
	rb_release_buffer(*buffer_ptr);
}

void
rb_release_buffer(rb_buffer * buffer) {
	if(buffer) {
		if(buffer->data_buffer) { free(buffer->data_buffer); }
		if(buffer->slices) { free(buffer->slices); }
		if(buffer->batches) { free(buffer->batches); }
		if(buffer->stats) { free(buffer->stats); }
		if(buffer->info) { free(buffer->info); }
		free(buffer);
	}
}

/*  ! ! ! ! !

rb_claim and rb_publish are NOT thread safe and must somehow be
fanned in for scenarios other than single producer

654 ns/op Buffer(L12, 2)
314 ns/op Buffer(L12, 2)
 72 ns/op Buffer(L16, 2)
*/
rb_batch *
rb_claim(rb_buffer * buffer, uint16_t count) {
	// TODO: fix pool impl
	
	if(count == 0 || count > buffer->info->buffer_size) {
		__errno(RB_CLAIM_PANIC);
		goto error;
	} // Must be > 0 and < buffer size
	if(count > rb_write_space(buffer)) {
		__errno(RB_WRITE_BUFFER_FULL);
		goto error;
	}

	// With fixed pool size this should always be a safe operation
	rb_batch * batch = &buffer->batches[buffer->stats->write_seq_num&buffer->info->size_mask];
	batch->seq_num = buffer->stats->write_seq_num;
	batch->batch_num = buffer->stats->batch_num++;
	batch->batch_size = count;
	
	buffer->stats->write_seq_num += count;

	__errno(RB_SUCCESS);
	return batch;
error:
	__errno(RB_ERROR);
	return NULL;
}

void *
rb_get_entry(rb_buffer * buffer, uint64_t seq_num) {
	return &buffer->data_buffer[seq_num&buffer->info->size_mask];
}

void *
rb_get_entry_slice(rb_buffer * buffer, uint64_t seq_num) {
	// With fixed pool size this should always be a safe operation
	buffer->slices[seq_num].data = rb_get_entry(buffer, seq_num);
	return &buffer->slices[seq_num&buffer->info->size_mask];
}

void
rb_publish(rb_buffer * buffer, rb_batch * batch) {
	buffer->stats->write_barrier+=batch->batch_size;
	//free(batch);

	// REMOVE THIS - Simulates readers immediately consuming
	buffer->stats->read_barrier+=batch->batch_size;
	buffer->stats->read_seq_num+=batch->batch_size;


	__errno(RB_SUCCESS);
	return;
error:
	__errno(RB_ERROR);
}

void rb_claim_and_publish(rb_buffer * buffer, int count) {
	int i = 0;
	for(i = 0; i < count; i++) {
		rb_batch * batch = rb_claim(buffer, 1);
		char * entry = (char *)rb_get_entry(buffer, batch->seq_num);
		//char[] data = { 1, 2 };
		rb_publish(buffer, batch);
	}
}




rb_buffer_info * 
rb_get_info(rb_buffer * buffer) {
	return buffer->info;
}

rb_buffer_stats * 
rb_get_stats(rb_buffer * buffer) {
	return buffer->stats;
}

void
rb_print_buffer(rb_buffer * buffer) {
	DebugPrint("buffer_size: %d", buffer->info->buffer_size);
	DebugPrint("size_mask: %d", buffer->info->size_mask);
	DebugPrint("entry_size: %d", buffer->info->entry_size);
	DebugPrint("total_size: %d", buffer->info->total_size);
}

void
rb_print_state(rb_buffer * buffer) {
	DebugPrint("read_seq_num: %d", buffer->stats->read_seq_num);
	DebugPrint("read_barrier: %d", buffer->stats->read_barrier);
	DebugPrint("write_seq_num: %d", buffer->stats->write_seq_num);
	DebugPrint("write_barrier: %d", buffer->stats->write_barrier);
	DebugPrint("batch_num: %d", buffer->stats->batch_num);
}
