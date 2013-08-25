#ifndef _UTIL_H_
#define _UTIL_H_

#include <stdarg.h> /* Needed for the definition of va_list */
#include <stdint.h>

/****************************************************************************
 *
 *		Helper Methods
 *
 ****************************************************************************/

uint64_t round_up_pow_2_uint64_t(uint64_t v) {
	v--;
	v |= v >> 1;
	v |= v >> 2;
	v |= v >> 4;
	v |= v >> 8;
	v |= v >> 16;
	v |= v >> 32;
	v++;
	return v;
}

// cache line
#define ALIGN 64

void *aligned_malloc(int size) {
    void *mem = malloc(size+ALIGN+sizeof(void*));
    void **ptr = (void**)((long)(mem+ALIGN+sizeof(void*)) & ~(ALIGN-1));
    ptr[-1] = mem;
    return ptr;
}

void aligned_free(void *ptr) {
    free(((void**)ptr)[-1]);
}

#define BARRIER() { asm volatile("" ::: "memory"); } // Compiler barrier
 
char debug_out[1000];
void vDebugPrint(const char* format, va_list args) {
	vsprintf(debug_out, format, args);
	DebugPrintf(debug_out);
}
void DebugPrint(const char* format, ...) {
	va_list args;
	va_start(args, format);
	vDebugPrint(format, args);
	va_end(args);
}

#endif