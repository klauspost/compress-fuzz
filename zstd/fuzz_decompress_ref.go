//+build datadog

package fuzzzstd

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	ddzstd "github.com/DataDog/zstd"
	"github.com/klauspost/compress/zstd"
)

func FuzzDecompressRef(data []byte) int {
	// Run against reference decoder
	dec, err := zstd.NewReader(nil, zstd.WithDecoderLowmem(true), zstd.WithDecoderConcurrency(1), zstd.WithDecoderMaxMemory(10<<20))
	if err != nil {
		panic(err)
	}
	defer dec.Close()
	got, err := dec.DecodeAll(data, nil)
	if err == zstd.ErrDecoderSizeExceeded {
		// Don't run me out of memory.
		return 0
	}

	ref, refErr := ddzstd.Decompress(nil, data)

	switch {
	case err == nil:
		if refErr != nil {
			panic(fmt.Errorf("decoder returned no error, but reference returned %v", refErr))
		}
		if !bytes.Equal(ref, got) {
			panic("output mismatch")
		}
		return 1
	case refErr == nil:
		if err != nil {
			panic(fmt.Errorf("reference returned no error, but got %v", err))
		}
	}
	return 0
}
