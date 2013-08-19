#ifndef _UTIL_H_
#define _UTIL_H_

#include <stdarg.h> /* Needed for the definition of va_list */

/****************************************************************************
 *
 *		Helper Methods
 *
 ****************************************************************************/

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