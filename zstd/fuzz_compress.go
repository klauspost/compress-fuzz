package fuzzzstd

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/klauspost/compress/zstd"
)

const (
	speedNotSet zstd.EncoderLevel = iota

	// SpeedFastest will choose the fastest reasonable compression.
	// This is roughly equivalent to the fastest Zstandard mode.
	SpeedFastest

	// SpeedDefault is the default "pretty fast" compression option.
	// This is roughly equivalent to the default Zstandard mode (level 3).
	SpeedDefault

	// speedLast should be kept as the last actual compression option.
	// The is not for external usage, but is used to keep track of the valid options.
	speedLast

	// SpeedBetterCompression will (in the future) yield better compression than the default,
	// but at approximately 4x the CPU usage of the default.
	// For now this is not implemented.
	SpeedBetterCompression = SpeedDefault

	// SpeedBestCompression will choose the best available compression option.
	// For now this is not implemented.
	SpeedBestCompression = SpeedDefault
)

var dec *zstd.Decoder
var encs [speedLast]*zstd.Encoder
var mu sync.Mutex

func init() {
	var err error
	dec, err = zstd.NewReader(nil, zstd.WithDecoderConcurrency(1))
	if err != nil {
		panic(err)
	}
	for level := zstd.EncoderLevel(speedNotSet + 1); level < speedLast; level++ {
		encs[level], err = zstd.NewWriter(nil, zstd.WithEncoderCRC(false), zstd.WithEncoderLevel(level), zstd.WithEncoderConcurrency(1))
	}
}

func FuzzCompress(data []byte) int {
	mu.Lock()
	defer mu.Unlock()
	// Run test against out decoder
	for level := zstd.EncoderLevel(speedNotSet + 1); level < speedLast; level++ {
		var dst bytes.Buffer
		enc := encs[level]
		//enc, err := NewWriter(nil, WithEncoderCRC(true), WithEncoderLevel(level), WithEncoderConcurrency(1))
		encoded := enc.EncodeAll(data, make([]byte, 0, len(data)))
		enc.Reset(&dst)

		n, err := enc.Write(data)
		if err != nil {
			panic(err)
		}
		if n != len(data) {
			panic(fmt.Sprintln("Level", level, "Short write, got:", n, "want:", len(data)))
		}
		err = enc.Close()
		if err != nil {
			panic(err)
		}
		got, err := dec.DecodeAll(encoded, make([]byte, 0, len(data)))
		if err != nil {
			panic(fmt.Sprintln("Level", level, "DecodeAll error:", err, "got:", got, "want:", data, "encoded", encoded))
		}
		if !bytes.Equal(got, data) {
			panic(fmt.Sprintln("Level", level, "DecodeAll output mismatch", got, "(got) != ", data, "(want)", "encoded", encoded))
		}

		encoded = dst.Bytes()
		got, err = dec.DecodeAll(encoded, make([]byte, 0, len(data)))
		if err != nil {
			panic(fmt.Sprintln("Level", level, "DecodeAll (buffer) error:", err, "got:", got, "want:", data, "encoded", encoded))
		}
		if !bytes.Equal(got, data) {
			panic(fmt.Sprintln("Level", level, "DecodeAll (buffer) output mismatch", got, "(got) != ", data, "(want)", "encoded", encoded))
		}
	}
	return 1
}
