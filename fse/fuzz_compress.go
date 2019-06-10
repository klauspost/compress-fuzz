package fuzzfse

import (
	"bytes"
	"github.com/klauspost/compress/fse"
)

func FuzzCompress(data []byte) int {
	comp, err := fse.Compress(data, nil)
	if err == fse.ErrIncompressible || err == fse.ErrUseRLE {
		return 0
	}
	if err != nil {
		panic(err)
	}
	dec, err := fse.Decompress(comp, nil)
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(data, dec) {
		panic("decoder mismatch")
	}
	return 1
}
