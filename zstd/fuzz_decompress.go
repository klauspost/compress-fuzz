package fuzzzstd

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/klauspost/compress/zstd"
)

func FuzzDecompress(data []byte) int {
	const maxDecSize = 10 << 20
	dec, derr := zstd.NewReader(bytes.NewBuffer(data), zstd.WithDecoderLowmem(true), zstd.WithDecoderConcurrency(1), zstd.WithDecoderMaxMemory(maxDecSize))
	if derr != nil {
		return 0
	}
	readB := make([]byte, maxDecSize-1)
	defer dec.Close()
	n, err := ReadMax(dec, readB)
	switch err {
	case zstd.ErrCRCMismatch, nil, io.EOF:
	default:
		return 0
	}
	readB = readB[:n]

	// Test if DecodeAll can also decode.
	readA, err := dec.DecodeAll(data, nil)
	switch err {
	case nil, zstd.ErrCRCMismatch:
	default:
		if derr != zstd.ErrCRCMismatch {
			panic(fmt.Errorf("buffer decoder could decode to len(%d), but blob decoder returned: %v", len(readB), err))
		}
	}
	if !bytes.Equal(readA, readB) {
		if derr != zstd.ErrCRCMismatch {
			panic(fmt.Errorf("DecodeAll (1) bytes mismatched"))
		}
	}

	// Try again.
	readA, err = dec.DecodeAll(data, readA[:0])
	switch err {
	case nil, zstd.ErrCRCMismatch:
	default:
		if derr != zstd.ErrCRCMismatch {
			panic(fmt.Errorf("buffer decoder could decode, but blob decoder (2) returned: %v", err))
		}
	}
	if !bytes.Equal(readA, readB) {
		if derr != zstd.ErrCRCMismatch {
			panic(fmt.Errorf("DecodeAll (2) bytes mismatched"))
		}
	}

	// Reset and use stream.
	err = dec.Reset(ioutil.NopCloser(bytes.NewBuffer(data)))
	if err != nil {
		panic(err)
	}
	readA = make([]byte, maxDecSize-1)
	n, err = ReadMax(dec, readA)
	switch err {
	case zstd.ErrCRCMismatch, nil, io.EOF:
	default:
		panic(fmt.Errorf("after Reset, decoder returned error : %v", err))
	}
	readA = readA[:n]

	if !bytes.Equal(readA, readB) {
		panic(fmt.Errorf("dec.Reset bytes mismatched"))
	}

	return 1
}

func ReadMax(r io.Reader, buf []byte) (n int, err error) {
	max := len(buf)
	for n < max && err == nil {
		var nn int
		nn, err = r.Read(buf[n:])
		n += nn
	}
	if err == io.EOF {
		err = nil
	}
	return
}
