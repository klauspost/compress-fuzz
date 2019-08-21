package fuzzfse

import (
	"strings"

	"github.com/klauspost/compress/fse"
)

func FuzzDecompress(data []byte) int {
	s := fse.Scratch{}
	// Max output 1 MB.
	s.DecompressLimit = 1 << 20
	dec, err := fse.Decompress(data, &s)
	if err != nil && !strings.Contains(err.Error(), "DecompressLimit") {
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
