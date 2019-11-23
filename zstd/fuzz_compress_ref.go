//+build datadog

package fuzzzstd

import (
	"bytes"
	"fmt"

	ddz "github.com/DataDog/zstd"
)

func FuzzCompressRef(data []byte) int {
	// Run test against our decoder from DataDog zstd package
	for level := 1; level < 6; level++ {
		// Create a buffer that will usually be too small.
		var bufSize = len(data)
		if bufSize > 2 {
			// Make deterministic size
			bufSize = int(data[0]) | int(data[1])<<8
			if bufSize >= len(data) {
				bufSize = len(data) / 2
			}
		}
		dst := make([]byte, bufSize)
		encoded, err := ddz.CompressLevel(dst, data, level)
		if err != nil {
			panic(err)
		}

		got, err := dec.DecodeAll(encoded, make([]byte, 0, bufSize))
		if err != nil {
			panic(fmt.Sprintln("Level", level, "DecodeAll error:", err, "got:", got, "want:", data, "encoded", encoded))
		}
		if !bytes.Equal(got, data) {
			panic(fmt.Sprintln("Level", level, "DecodeAll output mismatch", got, "(got) != ", data, "(want)", "encoded", encoded))
		}
	}
	if len(data) > 64<<10 {
		// Prefer small for now.
		return 0
	}
	return 1
}
