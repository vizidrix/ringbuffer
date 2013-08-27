#ifndef _RING_BUFFER_H_
#define _RING_BUFFER_H_

#ifdef __cplusplus
extern "C"
#endif

#include <stdint.h>

typedef struct rb_buffer_info {
	uint64_t		buffer_size;	/** < Number of slots allocated per the rules above */
	uint64_t		size_mask;		/** < Buffer size - 1; Used to maintain scope of buffer */
	uint64_t		entry_size;		/** < Number of bytes of data for each slot */
	uint64_t		total_size;		/** < Total number of bytes allocated to the buffer */
} rb_buffer_info;

typedef struct rb_buffer_stats {
	volatile uint64_t	write_seq_num;	/** < Index of the latest entry released for production - finished read, ready for write */
	volatile uint64_t	read_seq_num;	/** < Index of the latest entry released for consumption - finished write, ready for read */
	volatile uint64_t	read_barrier;	/** < Index of the oldest entry released for consumption but still in use by at least one reader */
	
	volatile uint64_t	write_barrier;	/** < Index of the oldest entry released for production but still in use by a writer */
	volatile uint64_t	batch_num;		/** < Index of the last batch allocated to a producer */
	volatile uint64_t	__padding[3];	// 64 bytes - 40 byte struct = 3 * 8 bytes
} rb_buffer_stats;

//typedef struct rb_buffer_stats {
//	volatile uint64_t	read_seq_num;	/** < Index of the latest entry released for consumption */
//	volatile uint64_t	read_barrier;	/** < Index of the oldest entry released for consumption but still in use by at least one reader */
//	volatile uint64_t	write_seq_num;	/** < Index of the latest entry released for production */
//	volatile uint64_t	write_barrier;	/** < Index of the oldest entry released for production but still in use by a writer */
//	volatile uint64_t	batch_num;		/** < Index of the last batch allocated to a producer */
//	volatile uint64_t	__padding[3];	// 64 bytes - 40 byte struct = 3 * 8 bytes
//} rb_buffer_stats;

typedef struct rb_barrier {
	void *		seq_num_ptr;
} rb_barrier;

typedef struct rb_batch {
	uint64_t seq_num;
	uint64_t batch_num;
	uint64_t batch_size;
} rb_batch;

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
#define RB_ALLOC_STATS					(RB_ERROR - 102) 		/** Claim request violated buffer constraints */
#define RB_ALLOC_DATA 					(RB_ERROR - 103) 		/** Claim request violated buffer constraints */

#define RB_CLAIM_PANIC 					(RB_ERROR - 200) 		/** Claim request violated buffer constraints */
#define RB_WRITE_BUFFER_FULL			(RB_ERROR - 201)		/** Insufficient room in buffer to claim batch */



/*
This ring buffer conforms to certain size constraints due to tracking mechanisms
buffer_size:	Ensures that ring size fits in 3-4 depth Trie
data_size: 		Number of bytes to allocate per entry
- Keeping cache sizes in mind when picking data_size is critical to high perf
- Remember to adjust for DATA_HEADER_SIZE in order to maintain cache lines
- For example a data_size of 4096 would be bad but 4090 (+6 byte header) would fill cache line exactly
- If smaller data sizes are desired try to ensure 4096 % (data_size + 6) == 0
*/
extern void rb_init_buffer(rb_buffer** buffer_ptr, uint64_t buffer_size, uint64_t data_size);
extern void rb_release_buffer(rb_buffer * buffer);

extern rb_batch * rb_claim(rb_buffer * buffer, uint16_t count);

extern void * rb_get_entry(rb_buffer * buffer, uint64_t seq_num);
extern void * rb_get_entry_slice(rb_buffer * buffer, uint64_t seq_num);

extern void rb_publish(rb_buffer * buffer, rb_batch * batch);

extern rb_buffer_info * rb_get_info(rb_buffer * buffer);
extern rb_buffer_stats * rb_get_stats(rb_buffer * buffer);

extern void rb_claim_and_publish(rb_buffer * buffer, int count);

#ifdef __extern_golang
/* Hack to get referencing package to build */
#include "ringbuffer.c"
#endif

#ifdef __cplusplus
}
#endif


#endif /* _RINGBUFFER_H_ */