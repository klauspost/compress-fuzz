package fuzzhuff0

import (
	"bytes"
	"fmt"

	"github.com/klauspost/compress/huff0"
)

func FuzzCompress(data []byte) int {
	var enc huff0.Scratch
	enc.WantLogLess = 5
	comp, _, err := huff0.Compress1X(data, &enc)
	if err == huff0.ErrIncompressible || err == huff0.ErrUseRLE || err == huff0.ErrTooBig {
		return 0
	}
	if err != nil {
		panic(err)
	}
	dec, remain, err := huff0.ReadTable(comp, nil)
	if err != nil {
		panic(err)
	}
	if len(comp) >= len(data)-len(data)>>enc.WantLogLess {
		panic(fmt.Errorf("too large output provided. got %d, but should be < %d", len(comp), len(data)-len(data)>>enc.WantLogLess))
	}
	out, err := dec.Decompress1X(remain)
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(out, data) {
		panic("decompression 1x mismatch")
	}
	// Reuse as 4X
	enc.Reuse = huff0.ReusePolicyAllow
	comp, reUsed, err := huff0.Compress4X(data, &enc)
	if err == huff0.ErrIncompressible || err == huff0.ErrUseRLE || err == huff0.ErrTooBig {
		return 0
	}
	if err != nil {
		panic(err)
	}
	if len(comp) >= len(data)-len(data)>>enc.WantLogLess {
		panic(fmt.Errorf("too large output provided. got %d, but should be < %d", len(comp), len(data)-len(data)>>enc.WantLogLess))
	}

	remain = comp
	if !reUsed {
		dec, remain, err = huff0.ReadTable(comp, dec)
		if err != nil {
			panic(err)
		}
	}
	out, err = dec.Decompress4X(remain, len(data))
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(out, data) {
		panic("decompression 4x with reuse mismatch")
	}

	enc.Reuse = huff0.ReusePolicyNone
	comp, reUsed, err = huff0.Compress4X(data, &enc)
	if err == huff0.ErrIncompressible || err == huff0.ErrUseRLE || err == huff0.ErrTooBig {
		return 0
	}
	if err != nil {
		panic(err)
	}
	if reUsed {
		panic("reused when asked not to")
	}
	if len(comp) >= len(data)-len(data)>>enc.WantLogLess {
		panic(fmt.Errorf("too large output provided. got %d, but should be < %d", len(comp), len(data)-len(data)>>enc.WantLogLess))
	}

	dec, remain, err = huff0.ReadTable(comp, nil)
	if err != nil {
		panic(err)
	}
	out, err = dec.Decompress4X(remain, len(data))
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(out, data) {
		panic("decompression 4x mismatch")
	}

	// Reuse as 1X
	dec.Reuse = huff0.ReusePolicyAllow
	comp, reUsed, err = huff0.Compress1X(data, &enc)
	if err == huff0.ErrIncompressible || err == huff0.ErrUseRLE || err == huff0.ErrTooBig {
		return 0
	}
	if err != nil {
		panic(err)
	}
	if len(comp) >= len(data)-len(data)>>enc.WantLogLess {
		panic(fmt.Errorf("too large output provided. got %d, but should be < %d", len(comp), len(data)-len(data)>>enc.WantLogLess))
	}

	remain = comp
	if !reUsed {
		dec, remain, err = huff0.ReadTable(comp, dec)
		if err != nil {
			panic(err)
		}
	}
	out, err = dec.Decompress1X(remain)
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(out, data) {
		panic("decompression 1x with reuse mismatch")
	}
	return 1
}
