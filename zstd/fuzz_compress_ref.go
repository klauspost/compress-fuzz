//+build datadog

package fuzzzstd

import (
	"bytes"
	"fmt"

	ddzstd "github.com/DataDog/zstd"
	"github.com/klauspost/compress/zstd"
)

func FuzzCompressRef(data []byte) int {
	// Runs Go encoder but decompresses with datadog zstd.
	mu.Lock()
	defer mu.Unlock()
	// Run test against out decoder
	for level := zstd.EncoderLevel(speedNotSet + 1); level < speedLast; level++ {
		var dst bytes.Buffer
		enc := encs[level]
		// Create a buffer that will usually be too small.
		var bufSize = len(data)
		if bufSize > 2 {
			// Make deterministic size
			bufSize = int(data[0]) | int(data[1])<<8
			if bufSize >= len(data) {
				bufSize = len(data) / 2
			}
		}
		enc.Reset(&dst)
		n, err := enc.Write(data)
		if err != nil {
			panic(err)
		}
		if n != len(data) {
			panic(fmt.Sprintln("Level", level, "Short write, got:", n, "want:", len(data)))
		}

		encoded := enc.EncodeAll(data, make([]byte, 0, bufSize))

		err = enc.Close()
		if err != nil {
			panic(err)
		}
		got, err := ddzstd.Decompress(make([]byte, 0, bufSize), encoded)
		if err != nil {
			panic(fmt.Sprintln("Level", level, "DecodeAll error:", err, "\nwant:", data, "\nencoded", encoded))
		}
		if !bytes.Equal(got, data) {
			panic(fmt.Sprintln("Level", level, "DecodeAll output mismatch\n", got, "(got) != \n", data, "(want)", "\nencoded:", encoded))
		}

		encoded = dst.Bytes()
		got, err = ddzstd.Decompress(make([]byte, 0, bufSize), encoded)
		if err != nil {
			panic(fmt.Sprintln("Level", level, "DecodeAll (buffer) error:", err, "\ngot:", got, "\nwant:", data, "\nencoded", encoded))
		}
		if !bytes.Equal(got, data) {
			panic(fmt.Sprintln("Level", level, "DecodeAll output mismatch\n", got, "(got) != \n", data, "(want)", "\nencoded:", encoded))
		}
	}
	return 1
}
