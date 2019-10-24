package fuzzhuff0

import (
	"bytes"

	"github.com/klauspost/compress/huff0"
)

func FuzzCompress(data []byte) int {
	var sc huff0.Scratch
	comp, _, err := huff0.Compress1X(data, &sc)
	if err == huff0.ErrIncompressible || err == huff0.ErrUseRLE || err == huff0.ErrTooBig {
		return 0
	}
	if err != nil {
		panic(err)
	}
	s, remain, err := huff0.ReadTable(comp, nil)
	if err != nil {
		panic(err)
	}
	out, err := s.Decompress1X(remain)
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(out, data) {
		panic("decompression 1x mismatch")
	}
	// Reuse as 4X
	sc.Reuse = huff0.ReusePolicyAllow
	comp, reUsed, err := huff0.Compress4X(data, &sc)
	if err == huff0.ErrIncompressible || err == huff0.ErrUseRLE || err == huff0.ErrTooBig {
		return 0
	}
	if err != nil {
		panic(err)
	}
	remain = comp
	if !reUsed {
		s, remain, err = huff0.ReadTable(comp, s)
		if err != nil {
			panic(err)
		}
	}
	out, err = s.Decompress4X(remain, len(data))
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(out, data) {
		panic("decompression 4x with reuse mismatch")
	}

	s.Reuse = huff0.ReusePolicyNone
	comp, reUsed, err = huff0.Compress4X(data, s)
	if err == huff0.ErrIncompressible || err == huff0.ErrUseRLE || err == huff0.ErrTooBig {
		return 0
	}
	if err != nil {
		panic(err)
	}
	if reUsed {
		panic("reused when asked not to")
	}
	s, remain, err = huff0.ReadTable(comp, nil)
	if err != nil {
		panic(err)
	}
	out, err = s.Decompress4X(remain, len(data))
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(out, data) {
		panic("decompression 4x mismatch")
	}

	// Reuse as 1X
	s.Reuse = huff0.ReusePolicyAllow
	comp, reUsed, err = huff0.Compress1X(data, &sc)
	if err == huff0.ErrIncompressible || err == huff0.ErrUseRLE || err == huff0.ErrTooBig {
		return 0
	}
	if err != nil {
		panic(err)
	}
	remain = comp
	if !reUsed {
		s, remain, err = huff0.ReadTable(comp, s)
		if err != nil {
			panic(err)
		}
	}
	out, err = s.Decompress1X(remain)
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(out, data) {
		panic("decompression 1x with reuse mismatch")
	}
	return 1
}
