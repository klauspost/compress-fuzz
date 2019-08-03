package fuzzhuff0

import (
	"bytes"
	"github.com/klauspost/compress/huff0"
)

func FuzzCompress(data []byte) int {
	comp, _, err := huff0.Compress1X(data, nil)
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

	comp, _, err = huff0.Compress4X(data, nil)
	if err == huff0.ErrIncompressible || err == huff0.ErrUseRLE || err == huff0.ErrTooBig {
		return 0
	}
	if err != nil {
		panic(err)
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

	return 1
}
