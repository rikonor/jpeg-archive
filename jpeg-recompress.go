package main

// #cgo linux CFLAGS: -I/opt/mozjpeg/include
// #cgo windows CFLAGS: -I../mozjpeg
// #cgo LDFLAGS: src/util.o
// #cgo LDFLAGS: src/edit.o
// #cgo LDFLAGS: src/smallfry.o
// #cgo LDFLAGS: src/commander.o
// #cgo LDFLAGS: src/recompress.o
// #cgo LDFLAGS: src/iqa/build/release/libiqa.a
// #cgo linux,amd64 LDFLAGS: /opt/mozjpeg/lib64/libjpeg.a
// #cgo linux,386 LDFLAGS: /opt/mozjpeg/lib64/libjpeg.a
// #cgo windows LDFLAGS: ../mozjpeg/libjpeg.a
// #cgo LDFLAGS: -lm
//
// #include <stdlib.h>
// #include <stdbool.h>
// #include "src/recompress.h"
//
// typedef struct {
//     bool	ok;
//     recompress_error_t *err;
// } recompress_result_t;
//
// recompress_result_t go_recompress(const char *input, const char *output,
//                                   const recompress_options_t *options) {
//     recompress_error_t **error = malloc(sizeof(recompress_error_t *));
//     bool ok = recompress(input, output, options, error);
//     recompress_result_t res = {};
//     res.ok = ok;
//     res.err = ok ? NULL : *error;
//     free(error);
//     return res;
// }
import "C"

import (
	"fmt"
	"os"
	"unsafe"
)

func main() {
	options := C.recompress_options_t{
		method:         C.METHOD_SSIM,
		attempts:       6,
		target:         0,
		preset:         C.QUALITY_MEDIUM,
		jpegMin:        40,
		jpegMax:        95,
		strip:          false,
		noProgressive:  false,
		defishStrength: 0.0,
		defishZoom:     1.0,
		inputFiletype:  C.FILETYPE_AUTO,
		copyFiles:      true,
		accurate:       false,
		subsample:      C.SUBSAMPLE_DEFAULT,
		quiet:          false,
	}

	input := C.CString("/usr/local/google/home/pope/Pictures/background.jpg")
	defer C.free(unsafe.Pointer(input))

	output := C.CString("foo-go.jpg")
	defer C.free(unsafe.Pointer(output))

	res := C.go_recompress(input, output, &options)
	if !res.ok {
		defer C.free(unsafe.Pointer(res.err))
		fmt.Fprintln(os.Stderr, C.GoString(&res.err.message[0]))
	}

	fmt.Println("Hello, World", res.ok)
}
