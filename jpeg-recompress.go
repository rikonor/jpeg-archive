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
// #include "src/recompress.h"
import "C"

import (
	"fmt"
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

	ok := C.recompress(input, output, &options, nil)
	fmt.Println("Hello, World", ok)
}
