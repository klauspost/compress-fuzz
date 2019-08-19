package fuzzs2

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/klauspost/compress/s2"
)

func FuzzDecompress(data []byte) int {
	_, err1 := s2.Decode(nil, data)
	dec := s2.NewReader(bytes.NewBuffer(data))
	_, err2 := io.Copy(ioutil.Discard, dec)

	// If one is good and not CRC error, continue with that.
	if err1 == nil || err2 == nil || err2 == s2.ErrCRC {
		return 1
	}
	return 0
}
