package fuzzfse

import (
	"github.com/klauspost/compress/fse"
	"strings"
)

func FuzzDecompress(data []byte) int {
	dec, err := fse.Decompress(data, nil)
	if err != nil && strings.Contains(err.Error(), "DecompressLimit") {
		panic(err)
	}
	if err != nil {
		return 0
	}
	if len(dec) == 0 {
		panic("no output")
	}
	return 1
}
