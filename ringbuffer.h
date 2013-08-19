#ifndef _RING_BUFFER_H_
#define _RING_BUFFER_H_

#ifdef __cplusplus
extern "C"
#endif

#include <stdint.h>

typedef struct rb_barrier rb_barrier;

typedef struct rb_buffer rb_buffer;

extern int rb_init_buffer(rb_buffer** buffer_ptr, uint64_t buffer_size, uint32_t data_size);
extern int rb_release_buffer(rb_buffer * buffer);
extern int rb_claim(uint64_t position, char count);
extern int rb_publish(uint64_t position, char count);

/* Hack to get referencing package to build */
#include "ringbuffer.c"

#ifdef __cplusplus
}
#endif


#endif /* _RB_H_ */