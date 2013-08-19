//#include <stdarg.h> /* Needed for the definition of va_list */
//#include <stdlib.h>
//#include <stdint.h>
//#include <stdio.h>
//#include <string.h>

//#include <unistd.h>

//#include <sys/uio.h>

//#include <fcntl.h>

//#include <sys/mman.h>
//#include <sys/types.h>
//#include <sys/stat.h>
//#include <errno.h>

//#include <sys/socket.h>
//#include <xmmintrin.h>

#include <stdlib.h>
#include <stdint.h>
#include <assert.h>

#include "ringbuffer.h"
#include "util.h"

struct rb_barrier {
	uint64_t		seq_num;		/** < Index of the last entry released */
};

struct rb_buffer {
	//rb_barrier *	insert_barrier;
	uint64_t		seq_num;
	uint64_t		size_mask;		/** < Buffer size - 1; Used to maintain scope of buffer */
	void *			data_buffer;
};

int rb_init_buffer(rb_buffer** buffer_ptr, uint64_t buffer_size, uint32_t data_size) {
	*buffer_ptr = malloc(sizeof(rb_buffer));
	(*buffer_ptr)->seq_num = 10;
	return 0;
}

int rb_release_buffer(rb_buffer * buffer) {
	DebugPrint("Released Buffer: %d", buffer->seq_num);
	free(buffer);
}

int rb_claim(uint64_t position, char count) {
	return 0;
}

int rb_publish(uint64_t position, char count) {
	return 0;
}
