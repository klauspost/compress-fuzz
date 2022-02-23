package fuzzzstd

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"sync"

	"github.com/klauspost/compress/zstd"
)

func FuzzDecompress(data []byte) int {
	const maxmem = 10 << 20
	dec, err := zstd.NewReader(ioutil.NopCloser(bytes.NewBuffer(data)),
		zstd.WithDecoderLowmem(true),
		zstd.WithDecoderConcurrency(2),
		zstd.WithDecoderMaxMemory(maxmem),
		zstd.WithDecoderDicts(dictBytes),
	)
	if err != nil {
		return 0
	}
	defer dec.Close()
	var gotBuffer []byte
	var berr error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		gotBuffer, berr = dec.DecodeAll(data, nil)
		wg.Done()
	}()

	var gotReader bytes.Buffer
	_, err = io.Copy(&gotReader, dec)
	wg.Wait()
	switch berr {
	case nil, zstd.ErrDecoderSizeExceeded, zstd.ErrCRCMismatch:
	default:
		return 0
	}

	if err != nil && berr == nil {
		panic(fmt.Errorf("blob decoder returned %v, but Reader decoder returned: %v", berr, err))
	}

	if err == nil && berr == nil && !bytes.Equal(gotReader.Bytes(), gotBuffer) {
		panic(fmt.Errorf("output mismatch, length %d, %d", gotReader.Len(), len(gotBuffer)))
	}

	// Test with concurrency = 1
	if true {
		dec.Close()
		gotReader.Reset()
		dec, err2 := zstd.NewReader(ioutil.NopCloser(bytes.NewBuffer(data)),
			zstd.WithDecoderLowmem(false),
			zstd.WithDecoderConcurrency(1),
			zstd.WithDecoderMaxMemory(maxmem),
			zstd.WithDecoderDicts(dictBytes),
		)
		_, err2 = io.Copy(&gotReader, dec)
		dec.Close()
		isBothNil := err2 == nil && err == nil
		isOneNil := (err2 == nil) == !(err == nil)
		if isBothNil {
			if berr == nil && !bytes.Equal(gotReader.Bytes(), gotBuffer) {
				panic(fmt.Errorf("output mismatch, length %d, %d", gotReader.Len(), len(gotBuffer)))
			}
		} else if isOneNil {
			panic(fmt.Errorf("buffer conc = 2 returned %v, but conc = 1 decoder returned: %v", err, err2))
		}
	}

	return 1
}
