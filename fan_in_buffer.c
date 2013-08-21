/*
#include <stdlib.h>
#include <stdint.h>

#include "util.h"

** The goal of this future impl is to fan multiple producers into
	a single thread which can handle actually claiming and publishing
	on the ring buffer instance which alleviates the need to think
	about multi-producer and optimize for single producer/multi-consumer
	pattern which is the primary purpose and strength of the ring buffer
	pattern

	I think that putting something like this non-blocking fan-in
	buffer in front of the producer side will help us acheive low contention
	write buffering on the front end to handle cases where entries might
	be generated from highly concurrant locations such as web servers

	On the publish side the fan-in-buffer would have to also guarantee
	ordering by batch id and never publish out of order batches otherwise
	there is no way to make sure that readers are handed in-order messages
	and that all the efforts to fit data into cache lines and contiguous blocks
	are actually utilized during reader forward scanning...  both are important

	

// Likely shape of request struct
typedef struct rb_batch_request {
	rb_batch * 		batch;
	uint8_t 		count;
	uint64_t		error;
} rb_batch_request;
*/

// Article that describes the structure used to do non-blocking queue using CAS
// http://www.drdobbs.com/parallel/practical-lock-free-buffers/219500200

/*
BARRIER();
	
__int128_t x = 0;
DebugPrint("Before: %d", x);
__sync_fetch_and_add(&x, 10);
DebugPrint("After: %d", x);
*/

/*

//Each writer would init a local rb_batch_request:

rb_batch_request request = malloc(sizeof(rb_batch_request));
request->batch = 0; // Key to know when batch has been allocated
request->count = 8; // How many slots to allocate for batch

// Allocator will cas set the batch with the ptr once allocated

//Then call to the fan_in_buffer instance for claims:

// Call signature would be something like:
// int rb_request_claim(rb_fan_in_buffer * buffer, rb_batch_request * request)
rb_request_claim(fi_claim_buffer, request);

// Then writer would spinlock waiting for batch ptr to != 0
for (;;;) {
	if(request->batch == 0 || request->error != 0) {
		// handle error
		spinlock!!!
		continue;
	}
	// Request fullfilled
	break;
}

// Writer now has an allocated batch to fill
fill_the_batch(batch);

// Now put the same batch request ptr into the fan_in_buffer instance for publish:
rb_request_publish(fi_publish_buffer, request);

// If we care about confirming the result
// Allocator will cas set the batch to 0 (zero) once published
for (;;;) {
	if(request->batch != 0 || request->error != 0) {
		spinlock!!!
		continue;
	}
	// Publish completed
	break;
}

*/