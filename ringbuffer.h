#ifndef _RING_BUFFER_H_
#define _RING_BUFFER_H_

#ifdef __cplusplus
extern "C"
#endif

#include <stdint.h>

#ifndef __x86_64__
#warning "The program is developed for x86-64 architecture only."
#endif
#if !defined(DCACHE1_LINESIZE) || !DCACHE1_LINESIZE
#ifdef DCACHE1_LINESIZE
#undef DCACHE1_LINESIZE
#endif
#define ____cacheline_aligned	__attribute__((aligned(64)))
#else
#define ____cacheline_aligned	__attribute__((aligned(DCACHE1_LINESIZE)))
#endif

typedef enum {
	AVAILABLE = 1, 
	WRITING = 2,
	CANCELED = 3,
	PUBLISHED = 4
} rb_batch_states;
//const uint64_t AVAILABLE_WRITE;//		= 1 << 0;

typedef struct rb_buffer_info {
	uint64_t		batch_buffer_size;	/** < Number of slots allocated for use in tracking batches */
	uint64_t		batch_size_mask;	/** < Batch buffer size - 1; Used to keep batches inside the ring */
	uint64_t		data_buffer_size;	/** < Number of slots allocated per the rules above */
	uint64_t		data_size_mask;		/** < Data buffer size - 1; Used to keep data inside the ring */
	uint64_t		entry_size;			/** < Number of bytes of data for each slot */
	uint64_t		total_data_size;	/** < Total number of bytes allocated to the buffer */
} rb_buffer_info;

typedef struct rb_buffer_stats {
	volatile uint64_t	barrier_batch_num 	____cacheline_aligned;	/** < Index of the oldest batch that is still in use by at least one reader */
	volatile uint64_t	read_batch_num 		____cacheline_aligned;	/** < Index of the newest batch which has been released to readers */
	volatile uint64_t	write_batch_num 	____cacheline_aligned;	/** < Index of the next available batch to be assigned to a writer */
	volatile uint64_t	barrier_seq_num 	____cacheline_aligned;	/** < Index of the oldest seq num that is still held by at least one reader (don't overflow) */
	volatile uint64_t	write_seq_num 		____cacheline_aligned;	/** < Index of the next available entry for allocation to a batch */
} rb_buffer_stats;

typedef struct rb_process_result {
	uint64_t	releases;
	uint64_t	claims;
	uint64_t	publishes;
} rb_process_result;

typedef struct rb_batch {
	/** <  The counter used to & to find the index of this batch in the buffer */
	volatile uint64_t			batch_num;
	/** <  How many entries to claim for this batch */
	volatile uint64_t 			batch_size;
	/** <  Seq number assigend to this batch */
	volatile uint64_t 			seq_num;
	/** <  Padding to keep other data elements out of this cache line */
	volatile uint64_t			data[5];					
	/** <  Current state of the batch */
	volatile rb_batch_states 	state 				____cacheline_aligned;
	/** < Readers are assigned a group bit mask to & once all of their group are complete */
	volatile uint64_t 			group_flags			____cacheline_aligned;
	/** < Readers are assigned a bit mask to & once complete */
	volatile uint64_t 			reader_flags[8]		____cacheline_aligned;
} rb_batch;


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

#define RB_RELEASE_OVERFLOW				(RB_ERROR - 400)		/** Requested batch release was out of order */

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

// Just for testing latency of Go interop
extern void temp(rb_buffer * buffer, uint64_t count);

/* This will block until the batch is claimed */
extern rb_batch * rb_claim(rb_buffer * buffer, uint16_t count, void* cancel);
/*
This is basically a trigger to clear the batch for processing by readers
	- Readers will (should) process batches in claimed sequence
	- Can watch state at this batch slot to tell when batch is finished writing
	- Can put state info into batch data for readers
*/
extern void rb_publish(rb_buffer * buffer, rb_batch * batch);
/*
This is a trigger to update the batch's state flag
*/
extern void rb_release(rb_buffer * buffer, rb_batch * batch);
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