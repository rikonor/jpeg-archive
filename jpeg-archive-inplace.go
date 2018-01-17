package main

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
	"context"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	humanize "github.com/dustin/go-humanize"
)

func recompress(src, dest string) error {
	options := C.recompress_options_t{
		method:         C.METHOD_SSIM,
		attempts:       6,
		target:         0,
		preset:         C.QUALITY_VERYHIGH,
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
		quiet:          true,
	}

	input := C.CString(src)
	defer C.free(unsafe.Pointer(input))
	output := C.CString(dest)
	defer C.free(unsafe.Pointer(output))

	res := C.go_recompress(input, output, &options)
	if !res.ok {
		defer C.free(unsafe.Pointer(res.err))
		return errors.New(C.GoString(&res.err.message[0]))
	}

	return nil
}

func recompressInplace(ctx context.Context, src string) (err error) {
	srcFi, err := os.Stat(src)
	if err != nil {
		return
	}

	tmp, err := ioutil.TempFile(filepath.Dir(src), "recompress")
	if err != nil {
		return
	}
	defer func() {
		// If err is not nil, then it means there was some problem processing
		// things. So we should remove the tmp file.
		if err != nil {
			os.Remove(tmp.Name())
		}
	}()
	if err = tmp.Close(); err != nil {
		return
	}

	recompressErr := make(chan error)
	go func() {
		recompressErr <- recompress(src, tmp.Name())
	}()

	select {
	case err = <-recompressErr:
		if err != nil {
			return
		}
	case <-ctx.Done():
		err = ctx.Err()
		return
	}

	if err = os.Rename(tmp.Name(), src); err != nil {
		return err
	}

	dstFi, err := os.Stat(src)
	if err != nil {
		return err
	}

	diff := srcFi.Size() - dstFi.Size()
	pct := (float64(diff) / float64(srcFi.Size())) * 100.0
	log.Printf("%s: savings of %s (decreased by %0.2f%%)\n", src, humanize.Bytes(uint64(diff)), pct)
	return nil
}

// filepath.Glob wasn't working for OneDrive stuff.
// See https://github.com/golang/go/issues/22579.
func findJPEGs(dir string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	res := make([]string, 0)
	for _, f := range files {
		ext := filepath.Ext(f.Name())
		if strings.EqualFold(ext, ".JPG") {
			res = append(res, filepath.Join(dir, f.Name()))
		}
	}
	return res, nil
}

func processArgs(args []string) ([]string, error) {
	if len(args) == 0 {
		dir, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		return findJPEGs(dir)
	}
	return args, nil
}

func main() {
	var err error

	jpegs, err := processArgs(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
	if len(jpegs) == 0 {
		log.Fatal("No JPEGS to recompress.")
		return
	}

	numWorkers := runtime.NumCPU()
	sem := make(chan int, numWorkers)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		cancel()
	}()

	var wg sync.WaitGroup
	wg.Add(len(jpegs))

	for _, jpeg := range jpegs {
		go func(src string) {
			defer wg.Done()

			select {
			case sem <- 1:
				if ctx.Err() != nil {
					return
				}
			case <-ctx.Done():
				return
			}

			err := recompressInplace(ctx, src)
			if err != nil {
				log.Printf("E: [%s] %q", src, err)
				cancel()
			}

			<-sem
		}(jpeg)
	}
	wg.Wait()
}
