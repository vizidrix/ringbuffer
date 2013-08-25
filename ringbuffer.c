#include <stdlib.h>
#include <stdint.h>
#include <assert.h>
#include <time.h>
#include <errno.h>

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

struct rb_buffer {
	rb_buffer_info *	info;			/** < Holds buffer settings */
	rb_buffer_stats *	stats;			/** < Holds buffer allocation details */
	
	uint8_t *			data_buffer;
};

uint8_t const SMALL_BATCH_HEADER_SIZE = 5;
uint8_t const LARGE_BATCH_HEADER_SIZE = 6;

int rb_init_buffer(rb_buffer** buffer_ptr, uint8_t buffer_type, rb_batching_mode_t batching_mode, uint64_t data_size) {
	// Allocate space to hold the buffer and info structs
	*buffer_ptr = malloc(sizeof(rb_buffer));
	if(!*buffer_ptr) {
		errno = 1;
		return 1;
	}
	(*buffer_ptr)->info = malloc(sizeof(rb_buffer_info));
	if(!(*buffer_ptr)->info) {
		errno = 1;
		return 1;
	}
	(*buffer_ptr)->stats = malloc(sizeof(rb_buffer_stats));
	if(!(*buffer_ptr)->stats) {
		errno = 1;
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
	(*buffer_ptr)->info->size_mask = (*buffer_ptr)->info->buffer_size - 1;
	//rb_print_info((*buffer_ptr)->info);

	// Populate the buffer struct
	(*buffer_ptr)->stats->read_seq_num = 0;
	(*buffer_ptr)->stats->read_barrier = 0;
	(*buffer_ptr)->stats->write_seq_num = 0;
	(*buffer_ptr)->stats->write_barrier = 0;
	(*buffer_ptr)->stats->batch_num = 0;
	
	// Allocate giant contiguous byte array to hold the entries
	//DebugPrint("Allocating data_buffer: %d * (%d * 32) = %d", buffer_size, chunk_count, buffer_size * (chunk_count << 5));
	(*buffer_ptr)->data_buffer = malloc((*buffer_ptr)->info->total_size);
	if(!(*buffer_ptr)->data_buffer) {
		errno = 1;
		return 1;
	}
	
	// Setting default values - should zero in real setup
	int i = 0;
	for (i = 0; i < (*buffer_ptr)->info->total_size; i++) {
		(*buffer_ptr)->data_buffer[i] = 0xFF - (i % 0xFF);
	}

	return 0;
}

int rb_release_buffer(rb_buffer * buffer) {
	free(buffer->data_buffer);
	free(buffer->stats);
	free(buffer->info);
	free(buffer);
}

/*  ! ! ! ! !

rb_claim and rb_publish are NOT thread safe and must somehow be
fanned in for scenarios other than single producer

654 ns/op Buffer(L12, 2)
314 ns/op Buffer(L12, 2)
*/

rb_print_buffer(rb_buffer * buffer) {
	DebugPrint("buffer_type: %d", buffer->info->buffer_type);
	DebugPrint("buffer_size: %d", buffer->info->buffer_size);
	DebugPrint("size_mask: %d", buffer->info->size_mask);
	DebugPrint("batching_mode: %d", buffer->info->batching_mode);
	DebugPrint("data_size: %d", buffer->info->data_size);
	DebugPrint("entry_size: %d", buffer->info->entry_size);
	DebugPrint("total_size: %d", buffer->info->total_size);
}

rb_print_state(rb_buffer * buffer) {
	DebugPrint("read_seq_num: %d", buffer->stats->read_seq_num);
	DebugPrint("read_barrier: %d", buffer->stats->read_barrier);
	DebugPrint("write_seq_num: %d", buffer->stats->write_seq_num);
	DebugPrint("write_barrier: %d", buffer->stats->write_barrier);
	DebugPrint("batch_num: %d", buffer->stats->batch_num);
}

#define space(buffer) ((buffer->stats->read_barrier + buffer->info->buffer_size) - buffer->stats->write_seq_num)
uint64_t rb_claim(rb_buffer * buffer, uint16_t count) {
	// Cannot claim batch count zero
	// Requested batch size exceeds buffer size
	if(count == 0 || count > buffer->info->buffer_size) {
		goto error;
	} // Must be > 0
	if(count > 1) { // Cannot claim batch count greater than one when baching mode is NONE
		if(buffer->info->batching_mode == NONE) {
			goto error;
		}
	}
	if(count > space(buffer)) {
		goto error;
	}
	// Copy the seq_num before it gets modified
	uint64_t seq_num = buffer->stats->write_seq_num;
	
	buffer->stats->batch_num++;
	buffer->stats->write_seq_num += count;

	errno = 0;
	return seq_num;
	error:
		errno = -1;
		return 0;
}

void * rb_get_entry(rb_buffer * buffer, uint64_t seq_num) {
	return &buffer->data_buffer[seq_num & buffer->info->size_mask];
}

int rb_cancel(rb_buffer * buffer, uint64_t seq_num, uint16_t count) {
	//DebugPrint("Canceling batch");
	//free(batch->data_buffer);
	//free(batch);
	// TODO: Need to move write buffer

	return 0;
}

int rb_publish(rb_buffer * buffer, uint64_t seq_num, uint16_t count) {
	//DebugPrint("Publishing batch");
	//free(batch);

	return 0;
}

rb_buffer_info * rb_get_info(rb_buffer * buffer) {
	return buffer->info;
}

rb_buffer_stats * rb_get_stats(rb_buffer * buffer) {
	return buffer->stats;
}
