package main

import (
	"bytes"
	"testing"
)

func TestStructEncode(t *testing.T) {
	blabers := struct {
		a int
		b string
	}{1234, "let's go"}
	result := EncodeStruct(blabers)

	if !bytes.Equal(result, []byte("i1234e8:let's go")) {
		t.Error()
	}
}
