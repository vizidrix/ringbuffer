#ifndef _RING_BUFFER_H_
#define _RING_BUFFER_H_

#ifdef __cplusplus
extern "C"
#endif

#include <stdint.h>

typedef struct rb_buffer_info {
	uint64_t		batch_buffer_size;	/** < Number of slots allocated for use in tracking batches */
	uint64_t		batch_size_mask;	/** < Batch buffer size - 1; Used to keep batches inside the ring */
	uint64_t		data_buffer_size;	/** < Number of slots allocated per the rules above */
	uint64_t		data_size_mask;		/** < Data buffer size - 1; Used to keep data inside the ring */
	uint64_t		entry_size;			/** < Number of bytes of data for each slot */
	uint64_t		total_data_size;	/** < Total number of bytes allocated to the buffer */
} rb_buffer_info;

typedef struct rb_buffer_stats {
	volatile uint64_t	barrier_batch_num;	/** < Index of the oldest batch that is still in use by at least one reader */
	volatile uint64_t	read_batch_num;		/** < Index of the newest batch which has been released to readers */
	volatile uint64_t	write_batch_num;	/** < Index of the next available batch to be assigned to a writer */
	volatile uint64_t	barrier_seq_num;	/** < Index of the oldest seq num that is still held by at least one reader (don't overflow) */
	volatile uint64_t	write_seq_num;		/** < Index of the next available entry for allocation to a batch */
	volatile uint64_t	__padding[3];		/** < 64 bytes - 40 byte struct = 3 * 8 == 24 bytes */
} rb_buffer_stats;

typedef struct rb_process_result {
	uint64_t	releases;
	uint64_t	claims;
	uint64_t	publishes;
} rb_process_result;

typedef struct rb_batch {
	volatile uint64_t	batch_num;			/** <  8 bytes - The counter used to & to find the index of this batch in the buffer */
	volatile uint64_t 	batch_size;			/** <  8 bytes - How many entries to claim for this batch */
	volatile uint64_t 	seq_num;			/** <  8 bytes - Seq number assigend to this batch */
	volatile uint16_t 	group_flags;		/** <  2 bytes - Accomodates 16 potential groups of readers */
	volatile uint8_t 	reader_flags[16];  	/** < 16 bytes - Up to 8 reader slots per group */
	volatile void *		release_callback;	/** <  8 bytes - If set to a valid pointer the buffer will push a non-zero value to the pointer location */
	/* Group size can be expanded by setting the group mask higher by << 1 for each additional 8... **/
	volatile uint8_t 	data[22];			/** < 22 bytes - Main purpose is to pad for cache lines but can be used for state */
} rb_batch; // Struct should fill a cache line: 64 bytes

typedef struct rb_reader {

} rb_reader;

typedef struct rb_slice {
	void *		data;
	uint64_t	len;
	uint64_t	cap;
} rb_slice;

typedef struct rb_buffer rb_buffer;

#define RB_SUCCESS 						0 						/** Successful result */
#define RB_ERROR						(-40600)				/** Generic error */

#define RB_ALLOC_BUFFER					(RB_ERROR - 100) 		/** Claim request violated buffer constraints */
#define RB_ALLOC_INFO 					(RB_ERROR - 101) 		/** Claim request violated buffer constraints */
#define RB_ALLOC_STATS					(RB_ERROR - 102)
#define RB_ALLOC_BATCHES				(RB_ERROR - 103) 		/** Claim request violated buffer constraints */
#define RB_ALLOC_DATA 					(RB_ERROR - 104) 		/** Claim request violated buffer constraints */

#define RB_CLAIM_PANIC 					(RB_ERROR - 200) 		/** Claim request violated buffer constraints */
#define RB_CLAIM_FULL					(RB_ERROR - 201)		
#define RB_CLAIM_CANCELED				(RB_ERROR - 202)		/** Claim request was canceled before it could be completed */
#define RB_WRITE_BUFFER_FULL			(RB_ERROR - 302)		/** Insufficient room in buffer to claim batch */


extern void rb_reset_batch(rb_batch * batch);

extern void rb_init_buffer(rb_buffer ** buffer_ptr, uint64_t batch_buffer_size, uint64_t data_buffer_size, uint64_t data_size);
extern void rb_free_buffer(rb_buffer ** buffer);

extern uint64_t rb_proces_releases(rb_buffer * buffer);
extern uint64_t rb_process_claims(rb_buffer * buffer);
extern uint64_t rb_process_publishes(rb_buffer * buffer);
extern rb_process_result rb_process_all(rb_buffer * buffer);

extern rb_batch * rb_get_batch(rb_buffer * buffer, uint64_t batch_num);
extern void * rb_get_entry(rb_buffer * buffer, uint64_t seq_num);
extern void * rb_get_entry_slice(rb_buffer * buffer, uint64_t seq_num);


/* Once a batch is claimed the writer needs to wait for seq_num to be 
	populated by the buffer */
extern rb_batch * rb_claim(rb_buffer * buffer, uint16_t count, void* cancel);
/*
This is basically a trigger to clear the batch for processing by readers
	- Readers will (should) process batches in claimed sequence
	- Can watch seq_num at this batch slot to tell when batch is finished
	- Can put state info into batch header for readers
	- Returns pointer to the callback slot, if set
*/
extern void * rb_publish(rb_batch * batch);
/*
The is a trigger to update the batch's group / reader flags 
*/
extern void rb_release(rb_reader * reader, rb_batch * batch);
/*
Registers a reader with the buffer
	- Group num translates into group mask bit position
	- Reader mask is assigned by the buffer
*/
extern rb_reader * rb_subscribe(rb_buffer * buffer, uint8_t group_num);
/*
Frees the memory for the reader
*/
extern void rb_unsubscribe(rb_reader * reader);

//extern void rb_publish(rb_buffer * buffer, rb_batch * batch);
//extern void rb_release(rb_buffer * buffer, rb_batch * batch, uint16_t group_mask, uint8_t reader_mask);

// Just for testing latency of Go interop
//extern void rb_claim_and_publish(rb_buffer * buffer, int count);


extern rb_buffer_info * rb_get_info(rb_buffer * buffer);
extern rb_buffer_stats * rb_get_stats(rb_buffer * buffer);


#ifdef __extern_golang
/* Hack to get referencing package to build */
#include "ringbuffer.c"
#endif

#ifdef __cplusplus
}
#endif


#endif /* _RINGBUFFER_H_ */