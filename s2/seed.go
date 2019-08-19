//+build ignore

package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/klauspost/compress/s2"
)

func main() {
	filepath.Walk("compress/corpus", func(path string, info os.FileInfo, err error) error {
		b, err := ioutil.ReadFile(path)
		if err == nil {
			comp := s2.Encode(nil, b)
			ioutil.WriteFile("decompress/corpus/"+filepath.Base(path)+"-block", comp, os.ModePerm)

			var buf bytes.Buffer
			enc := s2.NewWriter(&buf)
			_, err := enc.Write(b)
			if err != nil {
				return nil
			}
			enc.Flush()
			ioutil.WriteFile("decompress/corpus/"+filepath.Base(path)+"-frame", buf.Bytes(), os.ModePerm)
		}
		return nil
	})
}
