// Copyright 2015 go-fuzz project authors. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.

package fuzzflate

import (
	"bytes"
	"github.com/klauspost/compress/flate"
	"io/ioutil"
	"strconv"
)

func Fuzz(data []byte) int {
	// Enable to focus on decoding.
	if false {
		r := bytes.NewReader(data)
		fr := flate.NewReader(r)
		_, err := ioutil.ReadAll(fr)
		if _, ok := err.(flate.InternalError); ok {
			panic(err)
		}
		if err != nil {
			return 0
		}
	}
	data1 := data
	for level := -2; level <= 9; level++ {
		msg := "level " + strconv.Itoa(level) + ":"
		buf := new(bytes.Buffer)
		fw, err := flate.NewWriter(buf, level)
		if err != nil {
			panic(msg + err.Error())
		}
		n, err := fw.Write(data1)
		if n != len(data1) {
			panic(msg + "short write")
		}
		if err != nil {
			panic(msg + err.Error())
		}
		err = fw.Close()
		if err != nil {
			panic(msg + err.Error())
		}
		fr1 := flate.NewReader(buf)
		data2, err := ioutil.ReadAll(fr1)
		if err != nil {
			panic(msg + err.Error())
		}
		if !bytes.Equal(data1, data2) {
			panic(msg + "not equal")
		}
		// Do it again...
		msg = "level " + strconv.Itoa(level) + " (reset):"
		buf.Reset()
		fw.Reset(buf)
		n, err = fw.Write(data1)
		if n != len(data1) {
			panic(msg + "short write")
		}
		if err != nil {
			panic(msg + err.Error())
		}
		err = fw.Close()
		if err != nil {
			panic(msg + err.Error())
		}
		fr1 = flate.NewReader(buf)
		data2, err = ioutil.ReadAll(fr1)
		if err != nil {
			panic(msg + err.Error())
		}
		if !bytes.Equal(data1, data2) {
			panic(msg + "not equal")
		}
	}
	return 1
}
