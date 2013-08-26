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

struct rb_buffer {
	rb_buffer_info *	info;			/** < Holds buffer settings */
	rb_buffer_stats *	stats;			/** < Holds buffer allocation details */
	rb_claim_result *	batches;
	uint8_t *			data_buffer;
};

void
//rb_init_buffer(rb_buffer** buffer_ptr, uint8_t buffer_type, rb_batching_mode_t batching_mode, uint64_t data_size) {
rb_init_buffer(rb_buffer** buffer_ptr, uint8_t buffer_type, uint64_t entry_size) {
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

	(*buffer_ptr)->batches = malloc(sizeof(rb_claim_result) * 100 * 1000 * 1000);

	// Populate the info struct
	(*buffer_ptr)->info->buffer_type = buffer_type;
	(*buffer_ptr)->info->buffer_size = rb_get_buffer_size_from_type(buffer_type);
	(*buffer_ptr)->info->entry_size = entry_size;
	//(*buffer_ptr)->info->entry_size = data_size; break;
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
*/

void
rb_print_buffer(rb_buffer * buffer) {
	DebugPrint("buffer_type: %d", buffer->info->buffer_type);
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

rb_claim_result *
rb_claim(rb_buffer * buffer, uint16_t count) {
	// TODO: fix pool impl
	rb_claim_result * result = &buffer->batches[buffer->stats->write_seq_num];
	
	if(count == 0 || count > buffer->info->buffer_size) {
		__errno(RB_CLAIM_PANIC);
		goto error;
	} // Must be > 0 and < buffer size
	if(count > rb_write_space(buffer)) {
		__errno(RB_WRITE_BUFFER_FULL);
		goto error;
	}

	result->seq_num = buffer->stats->write_seq_num;
	result->batch_num = buffer->stats->batch_num++;
	result->batch_size = count;
	
	buffer->stats->write_seq_num += count;

	__errno(RB_SUCCESS);
	return result;
error:
	__errno(RB_ERROR);
	return result;
}

void *
rb_get_entry(rb_buffer * buffer, uint64_t seq_num) {
	return &buffer->data_buffer[seq_num & buffer->info->size_mask];
}

void
//rb_publish(rb_buffer * buffer, uint64_t seq_num, uint16_t count) {
rb_publish(rb_buffer * buffer, rb_claim_result * batch) {
	//DebugPrint("Publishing batch");
	free(batch);


	__errno(RB_SUCCESS);
	return;
error:
	__errno(RB_ERROR);
}

rb_buffer_info * 
rb_get_info(rb_buffer * buffer) {
	return buffer->info;
}

rb_buffer_stats * 
rb_get_stats(rb_buffer * buffer) {
	return buffer->stats;
}
