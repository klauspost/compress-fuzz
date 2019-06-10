package fuzzhuff0

import "github.com/klauspost/compress/huff0"

func FuzzDecompress(data []byte) int {
	s, rem, err := huff0.ReadTable(data, nil)
	if err != nil {
		return 0
	}
	_, err1 := s.Decompress1X(rem)
	_, err4 := s.Decompress4X(rem, huff0.BlockSizeMax)
	if err1 != nil && err4 != nil {
		return 0
	}
	return 1
}
