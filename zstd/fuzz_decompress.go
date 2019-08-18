package fuzzzstd

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/klauspost/compress/zstd"
)

func FuzzDecompress(data []byte) int {
	dec, err := zstd.NewReader(bytes.NewBuffer(data), zstd.WithDecoderLowmem(true), zstd.WithDecoderConcurrency(1), zstd.WithDecoderMaxMemory(10<<20))
	if err != nil {
		return 0
	}
	defer dec.Close()
	_, err = io.Copy(ioutil.Discard, dec)
	switch err {
	case nil, zstd.ErrCRCMismatch:
	default:
		return 0
	}

	_, err = dec.DecodeAll(data, nil)
	switch err {
	case nil, zstd.ErrCRCMismatch:
	default:
		panic(fmt.Errorf("buffer decoder could decode, but blob decoder returned: %v", err))
	}
	return 1
}
