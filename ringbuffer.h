#ifndef _RING_BUFFER_H_
#define _RING_BUFFER_H_

#ifdef __cplusplus
extern "C"
#endif

#include <stdint.h>

typedef enum { 				//   Multiplier    | Entries
	L0		= 0, 			//	   	  1 *   64 =	     64 -- 64 * 1 [ 111111 ] 8 bit address - 1 Level Trie
	L1		= 1, 			//	   	  8 *   64 =	    512 -- 64 * 8 [ 111111 | 111 ] 16 bit address
	L2		= 2,			//		 16 *   64 =	  1,024
	L3		= 3,			//		 32 * 	64 =	  2,048
	L4		= 4,			//		  1 * 4096 =	  4,096 -- 4096 = 64 * 64 - 2 Level Trie ^
	L5 		= 5, 			//        8 * 4096 =     32,768
	L6 		= 6, 			//       16 * 4096 =     65,536 -- 64 * 64 * 16 [ 111111 | 111111 | 1111 ] 16 bit address
	L7 		= 7, 			//       32 * 4096 =    131,072 -- Here and down require 32 but addreess (technically between 16 and 32)
	L8		= 8,			//       64 * 4096 =    262,144 -- 3 Level Trie ^
	L9		= 9,			//   8 * 64 * 4096 =  2,097,152
	L10		= 10,			//  16 * 64 * 4096 =  4,194,304
	L11		= 11,			//  32 * 64 * 4096 =  8,388,608
	L12		= 12,			//  64 * 64 * 4096 = 16,777,216 -- 4 Level Trie ^
} rb_buffer_size_t;

typedef struct rb_buffer_info {
	uint8_t			buffer_type;	/** < Determines number of slots as specified in rb_buffer_size_t enum */
	uint64_t		buffer_size;	/** < Actual number of slots allocated per the rules above */
	uint8_t			chunk_count;	/** < Number of 32 byte chunks allocated for each buffer slot (must be an even number on systems with 64 byte L1 cache) */
	uint64_t		data_size;		/** < Actual number of bytes in length for each slot */
} rb_buffer_info;

//typedef struct rb_barrier rb_barrier;

typedef struct rb_batch rb_batch;

typedef struct rb_buffer rb_buffer;

/*
This ring buffer conforms to certain size constraints due to tracking mechanisms
buffer_size:	Ensures that ring size fits in 3-4 depth Trie
data_size: 		Number of 64 byte blocks to allocate per entry
*/
extern int rb_init_buffer(rb_buffer** buffer_ptr, uint8_t buffer_size, uint8_t data_size);
extern int rb_release_buffer(rb_buffer * buffer);
extern rb_buffer_info * rb_get_info(rb_buffer * buffer);
//extern rb_batch * rb_claim(rb_buffer * buffer, uint8_t count);
extern int rb_claim(rb_buffer * buffer, rb_batch ** batch, uint8_t count);
extern int rb_cancel(rb_buffer * buffer, rb_batch * batch);
extern int rb_publish(rb_buffer * buffer, rb_batch * batch);

#ifdef __extern_golang
/* Hack to get referencing package to build */
#include "ringbuffer.c"
#endif

#ifdef __cplusplus
}
#endif


#endif /* _RB_H_ */